package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type desktopOAuth struct{ authURL, tokenURL, keyField, id, secret, scope, callback string }

var errMissingClient = errors.New("set up the provider credentials in Settings")

func fieldsPresent(values ...string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			return false
		}
	}
	return true
}

func desktopConfig(platform string) (desktopOAuth, error) {
	file := map[string]string{"instagram": "meta-desktop.json", "youtube": "google-desktop.json", "tiktok": "tiktok-desktop.json", "x": "x-desktop.json"}[platform]
	if file == "" {
		return desktopOAuth{}, errors.New("unsupported provider")
	}
	raw, err := readDeveloperConfiguration(file)
	if errors.Is(err, os.ErrNotExist) {
		return desktopOAuth{}, errMissingClient
	}
	if err != nil {
		return desktopOAuth{}, errors.New("could not read the credentials saved in Settings")
	}
	if platform == "x" {
		var cfg struct {
			ID       string `json:"client_id"`
			Redirect string `json:"redirect_uri"`
		}
		if json.Unmarshal(raw, &cfg) != nil || !fieldsPresent(cfg.ID) || !validXCallback(cfg.Redirect) {
			return desktopOAuth{}, errors.New("invalid X Native App configuration")
		}
		return desktopOAuth{"https://x.com/i/oauth2/authorize", "https://api.x.com/2/oauth2/token", "client_id", cfg.ID, "", "tweet.read users.read tweet.write media.write offline.access", cfg.Redirect}, nil
	}
	if platform == "instagram" {
		var cfg struct {
			ID     string `json:"app_id"`
			Secret string `json:"app_secret"`
			Config string `json:"config_id"`
		}
		if json.Unmarshal(raw, &cfg) != nil || !fieldsPresent(cfg.ID, cfg.Secret, cfg.Config) {
			return desktopOAuth{}, errors.New("invalid Meta app configuration")
		}
		return desktopOAuth{"https://www.facebook.com/v25.0/dialog/oauth", graph + "oauth/access_token", "client_id", cfg.ID, cfg.Secret, cfg.Config, "/callback/"}, nil
	}
	if platform == "tiktok" {
		var cfg struct {
			Direct bool   `json:"direct_post"`
			Key    string `json:"client_key"`
			Secret string `json:"client_secret"`
		}
		if json.Unmarshal(raw, &cfg) != nil || !fieldsPresent(cfg.Key, cfg.Secret) {
			return desktopOAuth{}, errors.New("invalid TikTok desktop configuration")
		}
		scope := "user.info.basic,video.upload"
		if cfg.Direct {
			scope += ",video.publish"
		}
		return desktopOAuth{"https://www.tiktok.com/v2/auth/authorize/", tik + "oauth/token/", "client_key", cfg.Key, cfg.Secret, scope, "/callback/"}, nil
	}
	var cfg struct {
		Installed struct {
			ID     string `json:"client_id"`
			Secret string `json:"client_secret"`
		}
	}
	if err = json.Unmarshal(raw, &cfg); err != nil || !fieldsPresent(cfg.Installed.ID, cfg.Installed.Secret) {
		return desktopOAuth{}, errors.New("Google desktop client configuration is invalid")
	}
	return desktopOAuth{"https://accounts.google.com/o/oauth2/v2/auth", "https://oauth2.googleapis.com/token", "client_id", cfg.Installed.ID, cfg.Installed.Secret, "https://www.googleapis.com/auth/youtube.upload https://www.googleapis.com/auth/youtube.readonly https://www.googleapis.com/auth/youtube.force-ssl", "/"}, nil
}

// The receipt identifies the client used in a successful exchange, not an active account.
func (cfg desktopOAuth) fingerprint() string {
	data, _ := json.Marshal([]string{cfg.authURL, cfg.id, cfg.secret, cfg.scope})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
func receiptPath(platform string) (string, error) {
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "CATsoup Publisher", platform+"-client.receipt"), err
}
func ClientStatus(platform string) string {
	cfg, err := desktopConfig(platform)
	if errors.Is(err, errMissingClient) {
		return "missing"
	}
	if err != nil {
		return "invalid"
	}
	if path, err := receiptPath(platform); err == nil {
		receipt, _ := os.ReadFile(path)
		if string(receipt) == cfg.fingerprint() {
			return "confirmed"
		}
	}
	return "saved"
}

func ClientReady(platform string) bool {
	if platform != "youtube" && platform != "tiktok" && platform != "instagram" && platform != "x" {
		return false
	}
	_, err := desktopConfig(platform)
	return err == nil
}

func (c *Client) SignIn(ctx context.Context, platform string, open func(string) error) (Token, error) {
	cfg, err := desktopConfig(platform)
	if err != nil {
		return Token{}, err
	}
	listenAddress, callbackPath := "127.0.0.1:0", cfg.callback
	if platform == "x" {
		u, _ := url.Parse(cfg.callback)
		listenAddress, callbackPath = u.Host, u.Path
	}
	ln, err := net.Listen("tcp4", listenAddress)
	if err != nil {
		if platform == "x" {
			return Token{}, errors.New("X callback port is occupied; close the other sign-in or change the port in both Settings and the X Developer Console")
		}
		return Token{}, errors.New("could not start local sign-in")
	}
	defer ln.Close()
	host := ln.Addr().String()
	redirect := "http://" + host + callbackPath
	var bridge *metaReturn
	if platform == "instagram" {
		redirect, err = MetaCallbackURL()
		if err != nil {
			return Token{}, err
		}
		if redirect == "" {
			return Token{}, errors.New("choose callback hosting in Settings before signing in with Facebook")
		}
		bridge, err = newMetaReturn()
		if err != nil {
			return Token{}, err
		}
	}
	state, verifier := rand.Text(), rand.Text()+rand.Text()
	sum := sha256.Sum256([]byte(verifier))
	params := url.Values{cfg.keyField: {cfg.id}, "response_type": {"code"}, "redirect_uri": {redirect}, "state": {state}, "scope": {cfg.scope}, "code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"}}
	if platform == "tiktok" {
		params.Set("code_challenge", hex.EncodeToString(sum[:]))
	} else if platform == "youtube" {
		params.Set("prompt", "select_account consent")
		params.Set("access_type", "offline")
	}
	code := make(chan string, 1)
	srv := &http.Server{ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 5 * time.Second, MaxHeaderBytes: 8192, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" || r.URL.Path != callbackPath || r.Host != host || subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(state)) != 1 {
			if r.URL.Path != "/favicon.ico" {
				diagnostics.Record(ctx, slog.LevelWarn, "callback_rejected", diagnostics.Details{Reason: "invalid_callback"})
			}
			c.writeBrowserReturn(w, platform, 400, "Sign-in interrupted", "This browser response does not match the sign-in in Publisher. Open Accounts and try again.")
			return
		}
		result := r.URL.Query().Get("code")
		if bridge != nil {
			var ok bool
			result, ok = bridge.code(r.URL.Query().Get("reply"), state)
			if !ok {
				diagnostics.Record(ctx, slog.LevelWarn, "callback_rejected", diagnostics.Details{Reason: "invalid_callback"})
				c.writeBrowserReturn(w, platform, 400, "Sign-in interrupted", "Publisher could not verify the return from Meta. Start again from Accounts and use the same browser tab.")
				return
			}
		}
		if result == "" {
			c.writeBrowserReturn(w, platform, 200, "Sign-in cancelled", "You can close this tab. To try again, open Accounts in Publisher.")
		} else {
			c.writeBrowserReturn(w, platform, 200, "Return to Publisher", "Close this tab and open Accounts. Check the account name under the language you selected to confirm the connection.")
		}
		diagnostics.Record(ctx, slog.LevelInfo, "callback_received", diagnostics.Details{})
		// Finish the browser response before the sign-in goroutine closes the listener.
		_ = http.NewResponseController(w).Flush()
		select {
		case code <- result:
		default:
		}
	})}
	defer srv.Close()
	go srv.Serve(ln)
	launch := cfg.authURL + "?" + params.Encode()
	if bridge != nil {
		launch = bridge.launch(redirect, strconv.Itoa(ln.Addr().(*net.TCPAddr).Port), state, cfg)
	}
	diagnostics.Record(ctx, slog.LevelInfo, "browser_opening", diagnostics.Details{})
	if err = open(launch); err != nil {
		diagnostics.Record(ctx, slog.LevelError, "sign_in_failed", diagnostics.Details{Reason: "unknown"})
		return Token{}, errors.New("could not open the browser for sign-in")
	}
	var authCode string
	select {
	case authCode = <-code:
	case <-ctx.Done():
		diagnostics.Record(ctx, slog.LevelWarn, "sign_in_failed", diagnostics.Details{Reason: "timeout"})
		return Token{}, errors.New("sign-in timed out")
	}
	if authCode == "" {
		diagnostics.Record(ctx, slog.LevelWarn, "sign_in_failed", diagnostics.Details{Reason: "cancelled"})
		return Token{}, errors.New("sign-in was cancelled")
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {authCode}, "redirect_uri": {redirect}, cfg.keyField: {cfg.id}, "client_secret": {cfg.secret}, "code_verifier": {verifier}}
	if platform == "instagram" {
		form.Del("code_verifier")
	}
	if platform == "x" {
		form.Del("client_secret")
	}
	diagnostics.Record(ctx, slog.LevelInfo, "exchange_started", diagnostics.Details{})
	token, err := c.exchange(ctx, cfg.tokenURL, form)
	if err == nil {
		if platform == "x" {
			token.Client = cfg.fingerprint()
		}
		diagnostics.Record(ctx, slog.LevelInfo, "exchange_finished", diagnostics.Details{})
		// A status indicator must not prevent a successful sign-in.
		if path, e := receiptPath(platform); e == nil {
			_ = os.WriteFile(path, []byte(cfg.fingerprint()), 0600)
		}
	}
	return token, err
}
