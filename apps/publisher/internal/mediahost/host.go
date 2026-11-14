// Package mediahost exposes only one approved snapshot for a bounded transfer.
package mediahost

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var tunnelURL = regexp.MustCompile(`^https://[a-z0-9-]+\.trycloudflare\.com$`)
var loggedURL = regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com\b`)

// Handler has no directory route, credentials, filenames, or application API.
func Handler(file *os.File, size int64, route string) http.Handler {
	slots := make(chan struct{}, 8)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != route || r.URL.RawQuery != "" || (r.Method != "GET" && r.Method != "HEAD") {
			http.NotFound(w, r)
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			http.Error(w, "Busy", 429)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.ServeContent(w, r, "video.mp4", time.Time{}, io.NewSectionReader(file, 0, size))
	})
}

// Start returns only after the public HTTPS route serves the expected bytes.
// The caller must close it on success, error, cancellation, and shutdown.
func Start(parent context.Context, executable string, file *os.File, size int64) (string, func(), error) {
	noop := func() {}
	if file == nil || size < 32 || !filepath.IsAbs(executable) {
		return "", noop, errors.New("invalid temporary media host")
	}
	info, err := os.Stat(executable)
	if err != nil || !info.Mode().IsRegular() {
		return "", noop, errors.New("install cloudflared before enabling Instagram temporary HTTPS")
	}
	ctx, cancel := context.WithTimeout(parent, 15*time.Minute)
	dir, err := os.MkdirTemp("", "catsoup-media-host-")
	if err != nil {
		cancel()
		return "", noop, err
	}
	config := filepath.Join(dir, "config.yml")
	if err = os.WriteFile(config, []byte("{}\n"), 0600); err != nil {
		cancel()
		os.RemoveAll(dir)
		return "", noop, err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		cancel()
		os.RemoveAll(dir)
		return "", noop, err
	}
	nonce := make([]byte, 24)
	if _, err = rand.Read(nonce); err != nil {
		ln.Close()
		cancel()
		os.RemoveAll(dir)
		return "", noop, err
	}
	route := "/" + hex.EncodeToString(nonce) + "/video.mp4"
	server := &http.Server{Handler: Handler(file, size, route), ReadHeaderTimeout: 10 * time.Second, WriteTimeout: 2 * time.Minute, IdleTimeout: 15 * time.Second}
	go server.Serve(ln)
	cmd := exec.CommandContext(ctx, executable, "tunnel", "--config", config, "--no-autoupdate", "--url", "http://"+ln.Addr().String(), "--protocol", "http2")
	cmd.Dir = dir
	hideProcess(cmd)
	pipe, err := cmd.StderrPipe()
	if err != nil {
		server.Close()
		cancel()
		os.RemoveAll(dir)
		return "", noop, err
	}
	if err = cmd.Start(); err != nil {
		server.Close()
		cancel()
		os.RemoveAll(dir)
		return "", noop, errors.New("could not start temporary HTTPS")
	}
	hosts := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			if host := loggedURL.FindString(scanner.Text()); tunnelURL.MatchString(host) {
				select {
				case hosts <- host:
				default:
				}
			}
		}
	}()
	done := make(chan struct{})
	go func() { cmd.Wait(); close(done) }()
	var once sync.Once
	closeHost := func() { once.Do(func() { server.Close(); cancel(); <-done; os.RemoveAll(dir) }) }
	go func() { <-ctx.Done(); closeHost() }()
	fail := func() (string, func(), error) {
		closeHost()
		return "", noop, errors.New("temporary HTTPS could not be verified; no new Instagram upload was created")
	}
	var host string
	select {
	case host = <-hosts:
	case <-done:
		return fail()
	case <-ctx.Done():
		return fail()
	case <-time.After(45 * time.Second):
		return fail()
	}
	expected := make([]byte, 32)
	if _, err = file.ReadAt(expected, 0); err != nil {
		return fail()
	}
	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	lastStatus := 0
	for i := 0; i < 20; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", host+route, nil)
		req.Header.Set("Range", "bytes=0-31")
		resp, e := client.Do(req)
		if e == nil {
			lastStatus = resp.StatusCode
			raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64))
			resp.Body.Close()
			if resp.StatusCode == 206 && bytes.Equal(raw, expected) {
				return host + route, closeHost, nil
			}
		}
		select {
		case <-ctx.Done():
			return fail()
		case <-done:
			return fail()
		case <-time.After(2 * time.Second):
		}
	}
	closeHost()
	return "", noop, fmt.Errorf("temporary HTTPS verification failed (HTTP %d; 0 means no response); no new Instagram upload was created", lastStatus)
}
