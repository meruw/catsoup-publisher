package uploads

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestYouTubeCoverWaitsForProcessing(t *testing.T) {
	for _, state := range []string{"succeeded", "processing", "failed"} {
		t.Run(state, func(t *testing.T) {
			c := New()
			posts := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.Header.Get("Authorization") != "Bearer fixture" {
					t.Fatal("missing authorization")
				}
				if r.Method == "GET" {
					return response(`{"items":[{"processingDetails":{"processingStatus":"` + state + `"}}]}`), nil
				}
				posts++
				if r.URL.Path != "/upload/youtube/v3/thumbnails/set" || r.URL.Query().Get("videoId") != "original" {
					t.Fatal("wrong target")
				}
				b, _ := io.ReadAll(r.Body)
				if string(b) != "image" || r.Header.Get("Content-Type") != "image/jpeg" {
					t.Fatal("wrong image")
				}
				return response(`{"items":[{"default":{"url":"https://example.invalid/thumb"}}]}`), nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			err := c.SetYouTubeCover(ctx, "fixture", "original", []byte("image"), "image/jpeg")
			if (err == nil) != (state == "succeeded") || posts != map[bool]int{true: 1, false: 0}[state == "succeeded"] {
				t.Fatal(posts, err)
			}
		})
	}
}
func TestInstagramEmptyTransferRecoversSameContainerOnce(t *testing.T) {
	f, e := os.CreateTemp(t.TempDir(), "video")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	f.WriteString("x")
	c := New()
	creates, sends, publishes := 0, 0, 0
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		if strings.HasSuffix(r.URL.Path, "/media") {
			creates++
			return response(`{"id":"original","uri":"https://rupload.facebook.com/original"}`), nil
		}
		if r.URL.Host == "rupload.facebook.com" {
			sends++
			if sends == 1 {
				res := response(`{"debug_info":{"type":"ProcessingFailedError"}}`)
				res.StatusCode = 400
				return res, nil
			}
			if !strings.HasSuffix(r.URL.Path, "/original") {
				t.Fatal("changed container")
			}
			return response(`{"success":true}`), nil
		}
		if r.Method == "GET" {
			if sends > 1 {
				return response(`{"status_code":"FINISHED"}`), nil
			}
			return response(strings.Replace(emptyInstagramUpload, `"ERROR"`, `"IN_PROGRESS"`, 1)), nil
		}
		if strings.HasSuffix(r.URL.Path, "/media_publish") {
			publishes++
			return response(`{"id":"published"}`), nil
		}
		t.Fatal("unexpected request")
		return nil, nil
	})
	ref, err := c.instagramWithRetry(context.Background(), "fixture", Account{ID: "account"}, Media{File: f, Size: 1}, func(string) error { return nil }, time.Millisecond)
	if err != nil || ref != "published" || creates != 1 || sends != 2 || publishes != 1 {
		t.Fatal(ref, err, creates, sends, publishes)
	}
}
