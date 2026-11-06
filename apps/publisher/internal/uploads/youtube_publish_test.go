package uploads

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestYouTubeUploadVisibilityAndAudience(t *testing.T) {
	for _, mode := range []string{"draft", "publish"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "video.mp4")
			os.WriteFile(path, []byte("video"), 0600)
			f, _ := os.Open(path)
			defer f.Close()
			privacy := "private"
			if mode == "publish" {
				privacy = "public"
			}
			c := New()
			calls := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					if r.Method != "POST" {
						t.Fatal(r.Method)
					}
					var body struct {
						Status struct {
							Privacy   string `json:"privacyStatus"`
							Kids      bool   `json:"selfDeclaredMadeForKids"`
							Synthetic bool   `json:"containsSyntheticMedia"`
						}
					}
					json.NewDecoder(r.Body).Decode(&body)
					if body.Status.Privacy != privacy || !body.Status.Kids || !body.Status.Synthetic {
						t.Fatal(body)
					}
					res := response(`{}`)
					res.Header.Set("Location", "https://www.googleapis.com/upload/session")
					return res, nil
				}
				if calls != 2 || r.Method != "PUT" {
					t.Fatal("duplicate/unexpected transfer")
				}
				return response(`{"id":"video1","status":{"privacyStatus":"` + privacy + `"}}`), nil
			})
			m := Media{Title: "Title", File: f, Size: 5, Mode: mode, Audience: "kids", Synthetic: true}
			if err := Validate("youtube", m); err != nil {
				t.Fatal(err)
			}
			id, err := c.Send(context.Background(), "youtube", "test", Account{}, m, func(string) error { return nil })
			if err != nil || id != "video1" || calls != 2 {
				t.Fatal(id, err, calls)
			}
			m.Audience = ""
			if Validate("youtube", m) == nil {
				t.Fatal("missing audience accepted")
			}
		})
	}
}

func TestYouTubePromotionOwnershipAndPreservation(t *testing.T) {
	for _, owner := range []string{"channel", "other"} {
		t.Run(owner, func(t *testing.T) {
			c := New()
			writes := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.Method == "GET" {
					return response(`{"items":[{"id":"video1","snippet":{"channelId":"` + owner + `","title":"Old","categoryId":"24","tags":["cat"],"defaultAudioLanguage":"es"},"status":{"privacyStatus":"private","license":"creativeCommon","embeddable":false,"publishAt":"later"}}]}`), nil
				}
				if r.Method != "PUT" || r.URL.Path != "/youtube/v3/videos" {
					t.Fatal("promotion attempted upload")
				}
				writes++
				var body struct {
					ID      string
					Snippet map[string]any
					Status  map[string]any
				}
				json.NewDecoder(r.Body).Decode(&body)
				if body.ID != "video1" || body.Snippet["title"] != "New" || body.Snippet["categoryId"] != "24" || body.Snippet["defaultAudioLanguage"] != "es" || body.Status["license"] != "creativeCommon" || body.Status["embeddable"] != false || body.Status["privacyStatus"] != "public" || body.Status["publishAt"] != nil {
					t.Fatal(body)
				}
				return response(`{"id":"video1","status":{"privacyStatus":"public"}}`), nil
			})
			id, err := c.PublishYouTube(context.Background(), "test", "channel", "video1", Media{Title: "New", Audience: "general", Language: "es"})
			if owner == "other" {
				if err == nil || writes != 0 {
					t.Fatal("foreign video changed")
				}
			} else if err != nil || id != "video1" || writes != 1 {
				t.Fatal(id, err, writes)
			}
		})
	}
}
