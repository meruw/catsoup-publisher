package uploads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(body string) *http.Response {
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
func TestTransfers(t *testing.T) {
	for _, platform := range []string{"youtube", "tiktok", "instagram"} {
		t.Run(platform, func(t *testing.T) {
			raw := bytes.Repeat([]byte("z"), 1024)
			if platform == "tiktok" {
				raw = bytes.Repeat([]byte("cat"), 12<<20)
			} // Two chunks with a larger final chunk.
			f, err := os.Create(filepath.Join(t.TempDir(), "final.mp4"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err = f.Write(raw); err != nil {
				t.Fatal(err)
			}
			c := New()
			var sent []byte
			var recorded string
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				auth := r.Header.Get("Authorization")
				switch r.URL.Host {
				case "www.googleapis.com":
					if auth != "Bearer local" {
						t.Fatal("missing Google authorization")
					}
					if r.Method == "POST" {
						var meta struct {
							Status  struct{ PrivacyStatus string }
							Snippet struct{ Title, Description, DefaultLanguage string }
						}
						if err := json.NewDecoder(r.Body).Decode(&meta); err != nil {
							t.Fatal(err)
						}
						if meta.Status.PrivacyStatus != "private" || meta.Snippet.Title != "title" || meta.Snippet.DefaultLanguage != "es" || meta.Snippet.Description != "#cats" {
							t.Fatal("bad YouTube metadata", meta)
						}
						resp := response("")
						resp.Header.Set("Location", "https://www.googleapis.com/upload/session")
						return resp, nil
					}
					sent, err = io.ReadAll(r.Body)
					return response(`{"id":"video1"}`), err
				case "open.tiktokapis.com":
					if auth != "Bearer local" {
						t.Fatal("missing TikTok authorization")
					}
					if strings.Contains(r.URL.Path, "inbox/video/init") {
						return response(`{"data":{"publish_id":"inbox1","upload_url":"https://open-upload.tiktokapis.com/transfer"},"error":{"code":"ok"}}`), nil
					}
					return response(`{"data":{"status":"SEND_TO_USER_INBOX"},"error":{"code":"ok"}}`), nil
				case "open-upload.tiktokapis.com":
					if auth != "" || recorded != "inbox1" {
						t.Fatal("unexpected upload credential or missing checkpoint")
					}
					piece, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}
					if r.Header.Get("Content-Range") != fmt.Sprintf("bytes %d-%d/%d", len(sent), len(sent)+len(piece)-1, len(raw)) {
						t.Fatal("bad TikTok chunk range")
					}
					sent = append(sent, piece...)
					return response(""), nil
				case "graph.facebook.com":
					if auth != "Bearer local" {
						t.Fatal("missing Meta authorization")
					}
					switch {
					case strings.HasSuffix(r.URL.Path, "/media"):
						return response(`{"id":"container1","uri":"https://rupload.facebook.com/ig-api-upload/container1"}`), nil
					case strings.HasSuffix(r.URL.Path, "/media_publish"):
						return response(`{"id":"reel1"}`), nil
					default:
						return response(`{"status_code":"FINISHED"}`), nil
					}
				case "rupload.facebook.com":
					if recorded != "container1" || auth != "OAuth local" || r.Header.Get("file_size") != "1024" {
						t.Fatal("invalid Meta transfer")
					}
					sent, err = io.ReadAll(r.Body)
					return response(`{"success":true}`), err
				}
				t.Fatalf("unexpected network request %s", r.URL.Host)
				return nil, nil
			})
			ref, err := c.Send(context.Background(), platform, "local", Account{"account", "CATsoup"}, Media{"title", "#cats", "es", f, int64(len(raw))}, func(ref string) error { recorded = ref; return nil })
			if err != nil || ref == "" || !bytes.Equal(raw, sent) {
				t.Fatalf("ref=%s err=%v transferred=%d", ref, err, len(sent))
			}
		})
	}
}
func TestRejectsForeignTransferHosts(t *testing.T) {
	for _, url := range []string{"http://rupload.facebook.com/x", "https://rupload.facebook.com.evil.test/x", "https://rupload.facebook.com:8888/", "https://rupload.facebook.com@evil.test/"} {
		if safeSession(url, "rupload.facebook.com") {
			t.Fatal("unsafe transfer host accepted")
		}
	}
}
