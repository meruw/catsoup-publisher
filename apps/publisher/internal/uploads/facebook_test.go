package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFacebookPageSelection(t *testing.T) {
	for _, tc := range []struct {
		name, selector, permission, tasks string
		want                              bool
	}{
		{"linked Instagram", "catsoup_es", "granted", `["PROFILE_PLUS_CREATE_CONTENT"]`, true},
		{"explicit Page", "22", "granted", `["CREATE_CONTENT"]`, true},
		{"wrong destination", "catsoup_en", "granted", `["CREATE_CONTENT"]`, false},
		{"missing publishing permission", "22", "declined", `["CREATE_CONTENT"]`, false},
		{"read-only Page", "22", "granted", `["ANALYZE"]`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New()
			calls := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "GET" || r.URL.Host != "graph.facebook.com" || r.Header.Get("Authorization") != "Bearer user-fixture" {
					t.Fatal("unsafe Page query")
				}
				if strings.HasSuffix(r.URL.Path, "permissions") {
					return response(fmt.Sprintf(`{"data":[{"permission":"pages_manage_posts","status":%q}]}`, tc.permission)), nil
				}
				if r.URL.Query().Get("after") == "" {
					return response(`{"data":[{"id":"11","access_token":"wrong","tasks":["MANAGE"]}],"paging":{"next":"https://attacker.invalid/","cursors":{"after":"cursor+second"}}}`), nil
				}
				if r.URL.Query().Get("after") != "cursor+second" {
					t.Fatal("lost cursor")
				}
				return response(fmt.Sprintf(`{"data":[{"id":"22","access_token":"page-fixture","tasks":%s,"instagram_business_account":{"username":"catsoup_es"}}]}`, tc.tasks)), nil
			})
			token, err := c.FacebookPage(context.Background(), Token{Access: "user-fixture"}, tc.selector)
			if (err == nil) != tc.want {
				t.Fatalf("want success %t got %v", tc.want, err)
			}
			if tc.want && token.Access != "page-fixture" {
				t.Fatal("wrong page token")
			}
			if tc.permission != "granted" && calls != 1 {
				t.Fatal("queried Pages without publish grant")
			}
		})
	}
}
func TestFacebookReelTransfer(t *testing.T) {
	for _, mode := range []string{"success", "bad host", "checkpoint failure", "upload rejected", "finish rejected", "processing failed", "unfinished"} {
		t.Run(mode, func(t *testing.T) {
			f, err := os.Create(filepath.Join(t.TempDir(), "video.mp4"))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			f.WriteString("approved bytes")
			c := New()
			recorded := false
			sent := false
			finished := false
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Query().Get("access_token") != "" {
					t.Fatal("token in URL")
				}
				if r.URL.Host == "rupload.facebook.com" {
					if !recorded || r.Method != "POST" || r.Header.Get("Authorization") != "OAuth page-fixture" || r.Header.Get("offset") != "0" || r.Header.Get("file_size") != "14" {
						t.Fatal("invalid upload or missing checkpoint")
					}
					raw, _ := io.ReadAll(r.Body)
					if string(raw) != "approved bytes" {
						t.Fatal("wrong bytes")
					}
					sent = true
					return response(fmt.Sprintf(`{"success":%t}`, mode != "upload rejected")), nil
				}
				if r.URL.Host != "graph.facebook.com" || r.Header.Get("Authorization") != "Bearer page-fixture" {
					t.Fatal("unexpected request")
				}
				if r.Method == "POST" {
					if r.URL.Path != "/v25.0/22/video_reels" {
						t.Fatal("wrong page")
					}
					var body map[string]string
					json.NewDecoder(r.Body).Decode(&body)
					if body["upload_phase"] == "start" {
						host := "rupload.facebook.com"
						if mode == "bad host" {
							host = "rupload.facebook.com.attacker.invalid"
						}
						return response(fmt.Sprintf(`{"video_id":"123","upload_url":"https://%s/video-upload/v25.0/123"}`, host)), nil
					}
					if !sent || body["video_state"] != "PUBLISHED" || body["video_id"] != "123" || body["title"] != "title" || body["description"] != "caption" {
						t.Fatal("invalid finish")
					}
					finished = true
					return response(fmt.Sprintf(`{"success":%t}`, mode != "finish rejected")), nil
				}
				if !finished || r.URL.Path != "/v25.0/123" || r.URL.Query().Get("fields") != "status" {
					t.Fatal("status before finish or wrong receipt")
				}
				if mode == "processing failed" {
					return response(`{"status":{"video_status":"error"}}`), nil
				}
				if mode == "unfinished" {
					cancel()
					return response(`{"status":{"video_status":"ready","publishing_phase":{"status":"not_started"}}}`), nil
				}
				return response(`{"status":{"video_status":"ready","publishing_phase":{"status":"complete"}}}`), nil
			})
			ref, err := c.Send(ctx, "facebook", "page-fixture", Account{ID: "22"}, Media{File: f, Size: 14, Mode: "publish", Title: "title", Caption: "caption"}, func(id string) error {
				if id != "123" {
					t.Fatal("bad checkpoint")
				}
				recorded = true
				if mode == "checkpoint failure" {
					return errors.New("disk full")
				}
				return nil
			})
			if (err == nil) != (mode == "success") {
				t.Fatalf("unexpected outcome %v", err)
			}
			if ref != "123" {
				t.Fatal("lost remote ID")
			}
			if (mode == "bad host" || mode == "checkpoint failure") && sent {
				t.Fatal("unsafe upload")
			}
			if mode == "processing failed" && !errors.Is(err, ErrProcessingFailed) {
				t.Fatal("failure not classified")
			}
		})
	}
}
