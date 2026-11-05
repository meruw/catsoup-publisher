package uploads

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	for _, tc := range []struct{ platform, file, config string }{{"youtube", "google-desktop.json", `{"installed":{"client_id":"test-client","client_secret":"test-secret"}}`}, {"tiktok", "tiktok-desktop.json", `{"client_key":"test-client","client_secret":"test-secret"}`}, {"instagram", "meta-desktop.json", `{"app_id":"test-client","app_secret":"test-secret","config_id":"test-config"}`}} {
		t.Run(tc.platform, func(t *testing.T) {
			const customURL = "https://studio.example.org/team/meta/"
			if tc.platform == "instagram" {
				if _, err := SetMetaCallbackURL(customURL); err != nil {
					t.Fatal(err)
				}
			}
			if ClientStatus(tc.platform) != "missing" {
				t.Fatal("expected missing client")
			}
			if err := os.WriteFile(filepath.Join(dir, tc.file), []byte("{}"), 0600); err != nil {
				t.Fatal(err)
			}
			if ClientStatus(tc.platform) != "invalid" {
				t.Fatal("expected invalid client")
			}
			if err := os.WriteFile(filepath.Join(dir, tc.file), []byte(tc.config), 0600); err != nil {
				t.Fatal(err)
			}
			if ClientStatus(tc.platform) != "saved" {
				t.Fatal("expected locally valid client")
			}
			if tc.platform == "instagram" {
				if _, err := SetMetaCallbackURL(""); err != nil {
					t.Fatal(err)
				}
				_, err := New().SignIn(context.Background(), tc.platform, func(string) error { t.Fatal("opened browser without a callback choice"); return nil })
				if err == nil || !strings.Contains(err.Error(), "callback hosting") {
					t.Fatal("expected setup error", err)
				}
				if _, err := SetMetaCallbackURL(customURL); err != nil {
					t.Fatal(err)
				}
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
				if tc.platform == "instagram" {
					expected = ""
					if p.Get("code_verifier") != "" {
						t.Fatal("unexpected Meta verifier")
					}
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
				var reply string
				callback := p.Get("redirect_uri")
				if tc.platform == "instagram" {
					boot, err := url.ParseQuery(u.Fragment)
					if err != nil {
						return err
					}
					if u.Scheme+"://"+u.Host+u.Path != customURL || boot.Get("return_uri") != customURL || u.RawQuery != "" || boot.Get("config") != "test-config" || boot.Get("app") != "test-client" {
						t.Fatal("wrong Meta launch")
					}
					p = url.Values{"state": {boot.Get("state")}, "redirect_uri": {customURL}}
					callback = "http://127.0.0.1:" + boot.Get("port") + "/callback/"
					key, err := base64.RawURLEncoding.DecodeString(boot.Get("key"))
					if err != nil {
						return err
					}
					block, err := aes.NewCipher(key)
					if err != nil {
						return err
					}
					box, err := cipher.NewGCM(block)
					if err != nil {
						return err
					}
					iv := make([]byte, 12)
					reply = base64.RawURLEncoding.EncodeToString(box.Seal(iv, iv, []byte("test-code"), []byte(boot.Get("state"))))
					res, err := http.Get(callback + "?" + url.Values{"state": {boot.Get("state")}, "code": {"test-code"}, "reply": {"invalid"}}.Encode())
					if err != nil {
						return err
					}
					res.Body.Close()
					if res.StatusCode != 400 {
						t.Fatal("unauthenticated Meta code accepted")
					}
				}
				redirect = p.Get("redirect_uri")
				challenge = p.Get("code_challenge")
				for _, trial := range []struct {
					state  string
					status int
				}{{"wrong", 400}, {p.Get("state"), 200}} {
					resp, err := http.Get(callback + "?" + url.Values{"state": {trial.state}, "code": {"test-code"}, "reply": {reply}}.Encode())
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
			if err != nil || token.Access != "fixture-token" {
				t.Fatal(token, err)
			}
			if ClientStatus(tc.platform) != "confirmed" {
				t.Fatal("successful exchange should confirm this client")
			}
			if err := Configure(tc.platform, strings.Replace(tc.config, "test-secret", "new-secret", 1)); err != nil {
				t.Fatal(err)
			}
			if ClientStatus(tc.platform) != "saved" {
				t.Fatal("replacement credentials must be confirmed separately")
			}
		})
	}
}
