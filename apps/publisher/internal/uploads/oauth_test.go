package uploads

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestBrowserSignInChecksStateAndPKCE(t *testing.T) {
	cfg := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("AppData", cfg)
	} else {
		t.Setenv("XDG_CONFIG_HOME", cfg)
	}
	dir := filepath.Join(cfg, "CATsoup Publisher")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ platform, file, config string }{{"youtube", "google-desktop.json", `{"installed":{"client_id":"test-client","client_secret":"test-secret"}}`}, {"tiktok", "tiktok-desktop.json", `{"client_key":"test-client","client_secret":"test-secret"}`}} {
		t.Run(tc.platform, func(t *testing.T) {
			if err := os.WriteFile(filepath.Join(dir, tc.file), []byte(tc.config), 0600); err != nil {
				t.Fatal(err)
			}
			var challenge, redirect string
			c := New()
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				p, err := url.ParseQuery(string(raw))
				if err != nil {
					t.Fatal(err)
				}
				sum := sha256.Sum256([]byte(p.Get("code_verifier")))
				expected := base64.RawURLEncoding.EncodeToString(sum[:])
				if tc.platform == "tiktok" {
					expected = hex.EncodeToString(sum[:])
				}
				if challenge != expected || p.Get("redirect_uri") != redirect || p.Get("code") != "test-code" {
					t.Fatal("invalid token exchange")
				}
				return response(`{"access_token":"fixture-token"}`), nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			token, err := c.SignIn(ctx, tc.platform, func(raw string) error {
				u, err := url.Parse(raw)
				if err != nil {
					return err
				}
				p := u.Query()
				redirect = p.Get("redirect_uri")
				challenge = p.Get("code_challenge")
				for _, trial := range []struct {
					state  string
					status int
				}{{"wrong", 400}, {p.Get("state"), 200}} {
					resp, err := http.Get(redirect + "?" + url.Values{"state": {trial.state}, "code": {"test-code"}}.Encode())
					if err != nil {
						return err
					}
					resp.Body.Close()
					if resp.StatusCode != trial.status {
						t.Fatal("unexpected loopback status", resp.StatusCode)
					}
				}
				return nil
			})
			if err != nil || token != "fixture-token" {
				t.Fatal(token, err)
			}
		})
	}
}
