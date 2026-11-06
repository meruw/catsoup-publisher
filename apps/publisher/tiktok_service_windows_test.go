package main

import (
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"context"
	"path/filepath"
	"testing"
)

func TestTikTokDirectRequiresPerAccountGrant(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	if err := uploads.Configure("tiktok", `{"client_key":"fixture","client_secret":"fixture","direct_post":true}`); err != nil {
		t.Fatal(err)
	}
	store, err := drafts.Open(filepath.Join(t.TempDir(), "drafts.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	for _, lang := range []string{"es", "en"} {
		scope := "video.upload"
		if lang == "es" {
			scope += ",video.publish"
		}
		sealed, err := (uploads.Token{Access: "fixture", Scope: scope}).Seal("tiktok/" + lang + "/" + lang)
		if err != nil {
			t.Fatal(err)
		}
		if err = store.PutProfile(drafts.Profile{Platform: "tiktok", Language: lang, ID: lang, Name: "Same name"}, sealed, ""); err != nil {
			t.Fatal(err)
		}
	}
	svc := UploadService{store: store, client: uploads.New()}
	rows, err := svc.Accounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range rows {
		if p.DirectPost != (p.Language == "es") {
			t.Fatal("grant leaked across languages")
		}
	}
	key := drafts.Key{Episode: "fixture", Platform: "tiktok", Language: "en", Account: "account:en"}
	if _, err = svc.tikTokAccess(context.Background(), key, "en"); err == nil {
		t.Fatal("inbox-only token accepted")
	}
	key.Language = "es"
	key.Account = "account:es"
	if _, err = svc.tikTokAccess(context.Background(), key, "en"); err == nil {
		t.Fatal("wrong destination accepted")
	}
	if err = uploads.SetTikTokDirectEnabled(false); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.tikTokAccess(context.Background(), key, "es"); err == nil {
		t.Fatal("disabled connector accepted")
	}
	if _, err = svc.SendTikTok(key, 1, "es", false, uploads.TikTokOptions{}); err == nil {
		t.Fatal("unconfirmed send accepted")
	}
}
func TestTikTokPolicyLinksAreAllowlisted(t *testing.T) {
	var opened string
	svc := UploadService{openBrowser: func(url string) error { opened = url; return nil }}
	if err := svc.OpenTikTokPolicy("music"); err != nil || opened != "https://www.tiktok.com/legal/page/global/music-usage-confirmation/en" {
		t.Fatal(err, opened)
	}
	opened = ""
	if svc.OpenTikTokPolicy("https://example.com") == nil || opened != "" {
		t.Fatal("arbitrary URL accepted")
	}
}
