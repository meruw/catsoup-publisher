package delivery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// PrepareInstagramAudio receives the already hash-checked temporary snapshot.
// It preserves the delivery and checks both the produced file and video stream.
// The caller owns cleanup, including on cancelled/failed external transfers.
func (r *Reader) PrepareInstagramAudio(ctx context.Context, source string, legacy bool) (*os.File, int64, string, func(), error) {
	noop := func() {}
	dir, err := os.MkdirTemp("", "catsoup-instagram-*")
	if err != nil {
		return nil, 0, "", noop, errors.New("could not prepare Instagram audio")
	}
	cleanup := func() { os.RemoveAll(dir) }
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	path := filepath.Join(dir, "upload.mp4")
	args := []string{"-I", "-B", filepath.Join(r.root, "apps", "publisher", "support", "prepare_instagram.py"), source, path}
	if legacy {
		args = append(args, "--legacy")
	}
	cmd := exec.CommandContext(ctx, filepath.Join(r.root, ".venv", "Scripts", "python.exe"), args...)
	hideProcess(cmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err = cmd.Run(); err != nil {
		cleanup()
		return nil, 0, "", noop, errors.New("could not prepare Instagram audio; check the workspace FFmpeg installation")
	}
	var result struct {
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}
	if json.Unmarshal(stdout.Bytes(), &result) != nil {
		cleanup()
		return nil, 0, "", noop, errors.New("invalid Instagram preparation result")
	}
	f, err := os.Open(path)
	if err != nil {
		cleanup()
		return nil, 0, "", noop, errors.New("prepared Instagram video is unavailable")
	}
	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil || size < 1 || size != result.Size || hex.EncodeToString(h.Sum(nil)) != result.SHA256 {
		f.Close()
		cleanup()
		return nil, 0, "", noop, errors.New("prepared Instagram video failed integrity checks")
	}
	return f, size, result.SHA256, func() { f.Close(); cleanup() }, nil
}
