package main

import (
	"bytes"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"context"
	"database/sql"
	"image"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

type coverTransport func(*http.Request) (*http.Response, error)

func (f coverTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestCoverFailureKeepsConfirmedVideo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	store, err := drafts.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 16, 16)))
	hash, err := store.ImportCover(context.Background(), b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	d := drafts.Draft{Key: drafts.Key{Episode: "sample", Language: "es", Platform: "youtube", Account: "account:channel"}, VideoHash: strings.Repeat("a", 64), Cover: &drafts.Cover{Kind: "image", Hash: hash}}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('youtube','channel',?,'published','confirmed-video',1)", d.VideoHash); err != nil {
		t.Fatal(err)
	}
	client := uploads.New()
	requests := 0
	client.HTTP.Transport = coverTransport(func(r *http.Request) (*http.Response, error) {
		requests++
		code := 200
		body := `{"items":[{"processingDetails":{"processingStatus":"succeeded"}}]}`
		if r.Method == "POST" {
			if r.URL.Path != "/upload/youtube/v3/thumbnails/set" {
				t.Fatal("video upload attempted")
			}
			code = 403
			body = `{"error":{"code":403,"errors":[{"reason":"forbidden"}]}}`
		}
		return &http.Response{StatusCode: code, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	svc := UploadService{store: store, client: client}
	if err = svc.applyYouTubeCover(context.Background(), d, "fixture", "confirmed-video"); err == nil {
		t.Fatal("expected thumbnail error")
	}
	got, err := store.Receipt("youtube", "channel", d.VideoHash)
	if err != nil || got.State != "published" || got.Reference != "confirmed-video" || requests != 2 {
		t.Fatal(got, requests, err)
	}
	if store.CoverResult(context.Background(), d) != "needs_attention" {
		t.Fatal("missing separate cover result")
	}
}
