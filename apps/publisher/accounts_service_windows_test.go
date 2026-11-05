package main

import (
	"bytes"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type accountTransport func(*http.Request) (*http.Response, error)

func (f accountTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestAccountRoutingSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	store, err := drafts.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	client := uploads.New()
	client.HTTP.Transport = accountTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "graph.facebook.com" {
			t.Fatal("unexpected host")
		}
		id := "EN"
		if r.Header.Get("Authorization") == "Bearer es-only-token" {
			id = "ES"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"instagram_business_account":{"id":"` + id + `","username":"fixture` + id + `"}}`))}, nil
	})
	service := UploadService{store: store, client: client}
	if _, err := service.LinkAccount("instagram", "es", "", "fixtureES", "es-only-token"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.LinkAccount("instagram", "en", "", "fixtureEN", "en-only-token"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.LinkAccount("instagram", "en", "EN", "fixtureES", "es-only-token"); err == nil {
		t.Fatal("shared identity allowed")
	}
	raw, err := store.Credential("instagram", "es", "ES")
	if err != nil || bytes.Contains(raw, []byte("es-only-token")) {
		t.Fatal("unprotected token")
	}
	if _, err = uploads.Unseal(raw, "instagram/en/ES"); err == nil {
		t.Fatal("credential opened for the wrong slot")
	}
	store.Close()
	store, err = drafts.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service.store = store
	rows, err := service.Accounts()
	if err != nil || len(rows) != 2 {
		t.Fatal(rows, err)
	}
	for _, tc := range []struct{ lang, id, token string }{{"es", "ES", "es-only-token"}, {"en", "EN", "en-only-token"}} {
		token, err := service.access(context.Background(), "instagram", tc.lang, tc.id)
		if err != nil || token != tc.token {
			t.Fatal("credential routing failed", err)
		}
	}
	if _, err := service.access(context.Background(), "instagram", "es", "EN"); err == nil {
		t.Fatal("cross-language access allowed")
	}
	if err := service.UnlinkAccount("instagram", "es", "ES"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.access(context.Background(), "instagram", "es", "ES"); err == nil {
		t.Fatal("credential retained after unlink")
	}
	if _, err := service.access(context.Background(), "instagram", "en", "EN"); err != nil {
		t.Fatal("unlink affected other language", err)
	}
}

func TestRefreshReplacesOnlyTheExpiredSlot(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AppData", dir)
	if err := uploads.Configure("tiktok", `{"client_key":"fixture","client_secret":"fixture"}`); err != nil {
		t.Fatal(err)
	}
	store, err := drafts.Open(filepath.Join(dir, "drafts.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	old := uploads.Token{Access: "expired", Refresh: "first-refresh", Expires: time.Now().Unix() - 1}
	raw, err := old.Seal("tiktok/es/es-id")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.PutProfile(drafts.Profile{Platform: "tiktok", Language: "es", ID: "es-id", Name: "ES"}, raw, ""); err != nil {
		t.Fatal(err)
	}
	calls := 0
	client := uploads.New()
	client.HTTP.Transport = accountTransport(func(r *http.Request) (*http.Response, error) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("refresh_token") != "first-refresh" {
			t.Fatal("unexpected refresh credential")
		}
		calls++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"new-access","refresh_token":"next-refresh","expires_in":3600}`))}, nil
	})
	svc := UploadService{store: store, client: client}
	for i := 0; i < 2; i++ {
		token, err := svc.access(context.Background(), "tiktok", "es", "es-id")
		if err != nil || token != "new-access" {
			t.Fatal("refresh failed", err)
		}
	}
	updated, err := store.Credential("tiktok", "es", "es-id")
	if err != nil {
		t.Fatal(err)
	}
	tkn, err := uploads.Unseal(updated, "tiktok/es/es-id")
	if err != nil || tkn.Refresh != "next-refresh" || calls != 1 {
		t.Fatal("rotation was not persisted", err)
	}
}
