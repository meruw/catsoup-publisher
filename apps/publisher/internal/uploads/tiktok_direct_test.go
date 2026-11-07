package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestTikTokPostingChoices(t *testing.T) {
	creator := TikTokCreator{Nickname: "Creator", Privacy: []string{"SELF_ONLY", "PUBLIC_TO_EVERYONE"}, MaxDuration: 60, DuetDisabled: true}
	base := TikTokOptions{Privacy: "SELF_ONLY", Consent: true}
	cases := []struct {
		name     string
		edit     func(*TikTokOptions)
		duration float64
		caption  string
	}{
		{"no consent", func(o *TikTokOptions) { o.Consent = false }, 12, "caption"},
		{"no privacy", func(o *TikTokOptions) { o.Privacy = "" }, 12, "caption"},
		{"unavailable privacy", func(o *TikTokOptions) { o.Privacy = "FOLLOWER_OF_CREATOR" }, 12, "caption"},
		{"duration", func(o *TikTokOptions) {}, 61, "caption"},
		{"NaN", func(o *TikTokOptions) {}, math.NaN(), "caption"},
		{"UTF16 limit", func(o *TikTokOptions) {}, 12, strings.Repeat("🐱", 1101)},
		{"disabled Duet", func(o *TikTokOptions) { o.Duet = true }, 12, "caption"},
		{"missing commercial choice", func(o *TikTokOptions) { o.Commercial = true }, 12, "caption"},
		{"hidden disclosure", func(o *TikTokOptions) { o.OwnBrand = true }, 12, "caption"},
		{"private branded post", func(o *TikTokOptions) { o.Commercial = true; o.Branded = true }, 12, "caption"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := base
			tc.edit(&o)
			if ValidateTikTokOptions(o, creator, tc.duration, tc.caption) == nil {
				t.Fatal("accepted invalid post")
			}
		})
	}
	if err := ValidateTikTokOptions(base, creator, 12, strings.Repeat("🐱", 1100)); err != nil {
		t.Fatal(err)
	}
	base.Privacy = "PUBLIC_TO_EVERYONE"
	base.Commercial = true
	base.Branded = true
	if err := ValidateTikTokOptions(base, creator, 12, "caption"); err != nil {
		t.Fatal(err)
	}
}
func TestTikTokDirectOptInAndScope(t *testing.T) {
	cfg := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", cfg)
	} else {
		t.Setenv("XDG_CONFIG_HOME", cfg)
	}
	if err := Configure("tiktok", `{"client_key":"fixture-key","client_secret":"fixture-secret"}`); err != nil {
		t.Fatal(err)
	}
	if TikTokDirectEnabled() {
		t.Fatal("direct post must be opt-in")
	}
	if err := SetTikTokDirectEnabled(true); err != nil {
		t.Fatal(err)
	}
	c, err := desktopConfig("tiktok")
	if err != nil || !strings.Contains(c.scope, "video.publish") {
		t.Fatal("scope not added", err)
	}
	if (Token{}).HasScope("video.publish") || (Token{Scope: "video.publish.extra"}).HasScope("video.publish") {
		t.Fatal("missing or partial grant accepted")
	}
	if !(Token{Scope: "user.info.basic,video.upload video.publish"}).HasScope("video.publish") {
		t.Fatal("exact grant missing")
	}
	if err := SetTikTokDirectEnabled(false); err != nil || TikTokDirectEnabled() {
		t.Fatal("could not opt out", err)
	}
	raw, err := readDeveloperConfiguration("tiktok-desktop.json")
	if err != nil || !strings.Contains(string(raw), "fixture-key") || !strings.Contains(string(raw), "fixture-secret") {
		t.Fatal("configuration lost")
	}
}
func TestTikTokDirectTransferAndDefinitiveFailure(t *testing.T) {
	for _, status := range []string{"PUBLISH_COMPLETE", "FAILED", "network", "host"} {
		t.Run(status, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "video-*.mp4")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			f.WriteString("fixture")
			c := New()
			recorded := ""
			sent := false
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host == "open-upload.tiktokapis.com" {
					if recorded != "direct1" || r.Header.Get("Authorization") != "" {
						t.Fatal("checkpoint or upload credential invalid")
					}
					raw, _ := io.ReadAll(r.Body)
					if string(raw) != "fixture" {
						t.Fatal("wrong media")
					}
					sent = true
					return response(""), nil
				}
				if r.URL.Host != "open.tiktokapis.com" {
					t.Fatal("unexpected destination", r.URL.Host)
				}
				if r.Header.Get("Authorization") != "Bearer fixture-token" {
					t.Fatal("missing provider auth")
				}
				switch {
				case strings.Contains(r.URL.Path, "creator_info"):
					return response(`{"data":{"creator_nickname":"CATsoup","privacy_level_options":["SELF_ONLY"],"max_video_post_duration_sec":60},"error":{"code":"ok"}}`), nil
				case strings.Contains(r.URL.Path, "/video/init/"):
					if strings.Contains(r.URL.Path, "inbox") {
						t.Fatal("used inbox endpoint")
					}
					var body struct {
						Post map[string]any `json:"post_info"`
					}
					if json.NewDecoder(r.Body).Decode(&body) != nil {
						t.Fatal("invalid body")
					}
					for k, want := range map[string]any{"title": "caption", "privacy_level": "SELF_ONLY", "disable_comment": false, "disable_duet": true, "disable_stitch": true, "brand_content_toggle": false, "brand_organic_toggle": false, "is_aigc": true} {
						if body.Post[k] != want {
							t.Fatalf("%s: got %v", k, body.Post[k])
						}
					}
					host := "open-upload.tiktokapis.com"
					if status == "host" {
						host = "example.com"
					}
					return response(`{"data":{"publish_id":"direct1","upload_url":"https://` + host + `/transfer"},"error":{"code":"ok"}}`), nil
				case strings.Contains(r.URL.Path, "status/fetch"):
					if status == "network" {
						return nil, errors.New("lost response")
					}
					return response(`{"data":{"status":"` + status + `"},"error":{"code":"ok"}}`), nil
				default:
					t.Fatal("unexpected request", r.URL.Path)
					return nil, nil
				}
			})
			m := Media{File: f, Size: 7, Mode: "publish", Duration: 12, Caption: "caption", TikTok: &TikTokOptions{Privacy: "SELF_ONLY", Consent: true, Comments: true, Synthetic: true}}
			ref, err := c.Send(context.Background(), "tiktok", "fixture-token", Account{ID: "creator"}, m, func(s string) error { recorded = s; return nil })
			if status == "PUBLISH_COMPLETE" && (err != nil || ref != "direct1" || !sent) {
				t.Fatal(ref, err, sent)
			}
			if status == "FAILED" && !errors.Is(err, ErrProcessingFailed) {
				t.Fatal("failure not definitive", err)
			}
			if status == "network" && (err == nil || errors.Is(err, ErrProcessingFailed)) {
				t.Fatal("uncertain network result mislabeled")
			}
			if status == "host" && (err == nil || sent || recorded != "") {
				t.Fatal("unsafe host accepted")
			}
		})
	}
}
