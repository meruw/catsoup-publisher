package main

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
)

type statisticsTransport func(*http.Request) (*http.Response, error)

func (f statisticsTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPostStatisticsUsesSavedReceiptAndLanguage(t *testing.T) {
	store, err := drafts.Open(filepath.Join(t.TempDir(), "drafts.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	client := uploads.New()
	calls := 0
	client.HTTP.Transport = statisticsTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "GET" || r.URL.Query().Get("id") != "abcdefghijk" || r.Header.Get("Authorization") != "Bearer fixture" {
			t.Fatal("wrong read-only request")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"items":[{"id":"abcdefghijk","snippet":{"channelId":"channel"},"statistics":{"viewCount":"2"}}]}`))}, nil
	})
	s := &UploadService{store: store, client: client}
	secret, err := (uploads.Token{Access: "fixture"}).Seal("youtube/es/channel")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.PutProfile(drafts.Profile{Platform: "youtube", Language: "es", ID: "channel", Name: "Fixture"}, secret, ""); err != nil {
		t.Fatal(err)
	}
	d := drafts.Draft{Key: drafts.Key{Episode: "catsoup-demo", Language: "es", Platform: "youtube", Account: "account:channel"}, VideoID: "es_final", VideoHash: strings.Repeat("a", 64), DeliveryFingerprint: strings.Repeat("b", 64)}
	d, err = store.Save(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.PostStatistics(d.Key); err != nil || got.Note == "" || calls != 0 {
		t.Fatal("draft should not contact provider", got, err)
	}
	if _, err = store.SetPostApproval(context.Background(), d.Key, d.Revision, true); err != nil {
		t.Fatal(err)
	}
	if err = store.Reserve(d, "channel"); err != nil {
		t.Fatal(err)
	}
	if err = store.SaveReceipt(d, "channel", "published", "abcdefghijk"); err != nil {
		t.Fatal(err)
	}
	before, _ := store.Get(context.Background(), d.Key)
	if got, err := s.PostStatistics(d.Key); err != nil || len(got.Items) != 1 || calls != 1 || len(got.History) != 1 || got.History[0].Items[0].Value != "2" {
		t.Fatal(got, err)
	}
	after, _ := store.Get(context.Background(), d.Key)
	if before != after {
		t.Fatal("statistics changed draft or approval")
	}
	r, _ := store.Receipt("youtube", "channel", d.VideoHash)
	if r.State != "published" || r.Reference != "abcdefghijk" {
		t.Fatal("statistics changed receipt")
	}
	wrong := d.Key
	wrong.Language = "en"
	if _, err = s.PostStatistics(wrong); err == nil || calls != 1 {
		t.Fatal("language isolation failed")
	}
	d.VideoHash = strings.Repeat("c", 64)
	d.Revision = before.Revision
	d, err = store.Save(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := s.PostStatistics(d.Key); err != nil || len(got.Items) != 0 || calls != 1 {
		t.Fatal("new file reused old receipt", got, err)
	}
}

func TestUnsupportedStatisticsNeverUsesPaidAPI(t *testing.T) {
	s := &UploadService{}
	for _, p := range []string{"x", "tiktok"} {
		got, err := s.PostStatistics(drafts.Key{Episode: "catsoup-demo", Language: "es", Platform: p, Account: "account:123"})
		if err != nil || got.Note == "" || len(got.Items) != 0 {
			t.Fatal(got, err)
		}
	}
}
