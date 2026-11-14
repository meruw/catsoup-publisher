package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestInstagramHostedBoundaries(t *testing.T) {
	for _, mode := range []string{"success", "checkpoint_failure", "processing_failure", "publish_uncertain", "publish_lock_failure"} {
		t.Run(mode, func(t *testing.T) {
			creates, publishes := 0, 0
			recorded, locked := false, false
			c := New()
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "graph.facebook.com" {
					t.Fatal("unexpected media upload")
				}
				switch {
				case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/media"):
					creates++
					var body map[string]any
					json.NewDecoder(r.Body).Decode(&body)
					if body["upload_type"] != nil || body["video_url"] == nil || body["caption"] != "approved" || body["thumb_offset"] != float64(100) {
						t.Fatal(body)
					}
					return response(`{"id":"container"}`), nil
				case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/media_publish"):
					if !recorded || !locked {
						t.Fatal("publishing without durable checkpoints")
					}
					publishes++
					if mode == "publish_uncertain" {
						return nil, errors.New("interrupted")
					}
					return response(`{"id":"published"}`), nil
				case r.Method == "GET":
					if !recorded {
						t.Fatal("missing checkpoint")
					}
					if mode == "processing_failure" {
						return response(`{"status_code":"ERROR"}`), nil
					}
					return response(`{"status_code":"FINISHED"}`), nil
				}
				t.Fatal("unexpected request")
				return nil, nil
			})
			cover := int64(100)
			id, e := c.InstagramHosted(context.Background(), "fixture", Account{ID: "account"}, Media{Caption: "approved", CoverTimeMS: &cover}, "https://fixture.trycloudflare.com/"+strings.Repeat("a", 48)+"/video.mp4", func(string) error {
				if mode == "checkpoint_failure" {
					return errors.New("db")
				}
				recorded = true
				return nil
			}, func() error {
				if mode == "publish_lock_failure" {
					return errors.New("db")
				}
				locked = true
				return nil
			})
			if creates != 1 {
				t.Fatal("duplicate container", creates)
			}
			if mode == "success" {
				if e != nil || id != "published" || publishes != 1 {
					t.Fatal(id, e)
				}
			} else {
				if e == nil {
					t.Fatal("expected failure")
				}
				if mode != "publish_uncertain" && publishes != 0 {
					t.Fatal("unexpected publication")
				}
				if publishes > 1 {
					t.Fatal("duplicate publication")
				}
			}
		})
	}
}

func TestInstagramHostedRejectsForeignURLs(t *testing.T) {
	for _, url := range []string{"http://fixture.trycloudflare.com/a/video.mp4", "https://evil.test/file", "https://fixture.trycloudflare.com:444/file", "https://user@fixture.trycloudflare.com/file"} {
		c := New()
		c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { t.Fatal("network for invalid URL"); return nil, nil })
		_, e := c.InstagramHosted(context.Background(), "", Account{}, Media{}, url, func(string) error { return nil }, func() error { return nil })
		if e == nil {
			t.Fatal(url)
		}
	}
}
