package drafts

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestPublicationDatesMigrationAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	d := sample()
	d.Key.Account = "account:channel"
	if _, err := s.Save(ctx, d); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES(?,?,?,?,?,?)", d.Key.Platform, "channel", d.VideoHash, "published", "remote", 1700000000); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("DROP TABLE cover_assets; DROP TABLE cover_results; ALTER TABLE drafts DROP COLUMN cover; DROP TABLE receipt_times;UPDATE schema_version SET version=9"); err != nil {
		t.Fatal(err)
	}
	before, err := s.Receipt(d.Key.Platform, "channel", d.VideoHash)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	page, err := s.PostStatuses(ctx, PostFilter{})
	if err != nil || len(page.Items) != 1 {
		t.Fatal(page, err)
	}
	p := page.Items[0]
	if p.PublishedAt != "" || p.ConfirmedAt != "" || p.StartedAt != "2023-11-14T22:13:20Z" {
		t.Fatal("invented publication date", p)
	}
	at := "2024-01-02T04:05:06-06:00"
	for _, args := range [][3]string{{"youtube", "other", "remote"}, {"youtube", "channel", "other"}} {
		if err = s.RecordPublicationTime(ctx, args[0], args[1], args[2], at); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	s.db.QueryRow("SELECT COUNT(*) FROM receipt_times").Scan(&n)
	if n != 0 {
		t.Fatal("cross account/reference write")
	}
	if err = s.RecordPublicationTime(ctx, d.Key.Platform, "channel", "remote", at); err != nil {
		t.Fatal(err)
	}
	page, err = s.PostStatuses(ctx, PostFilter{})
	if err != nil || page.Items[0].PublishedAt != "2024-01-02T10:05:06Z" || page.Items[0].ConfirmedAt != "" {
		t.Fatal(page, err)
	}
	after, err := s.Receipt(d.Key.Platform, "channel", d.VideoHash)
	if err != nil || before != after {
		t.Fatal("metadata changed upload receipt", err)
	}
	for _, bad := range []string{"bad", time.Now().Add(time.Hour).Format(time.RFC3339)} {
		if s.RecordPublicationTime(ctx, "youtube", "channel", "remote", bad) == nil {
			t.Fatal("invalid date accepted")
		}
	}
}

func TestReceiptConfirmationDoesNotBecomePublication(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	d := sample()
	d.Key.Account = "account:channel"
	if _, err := s.Save(ctx, d); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES(?,?,?,?,?,?)", d.Key.Platform, "channel", d.VideoHash, "sending", "remote", 1700000000); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"private", "inbox", "web_ready", "published"} {
		if err := s.SaveReceipt(d, "channel", state, "remote"); err != nil {
			t.Fatal(err)
		}
		page, err := s.PostStatuses(ctx, PostFilter{})
		if err != nil {
			t.Fatal(err)
		}
		p := page.Items[0]
		if p.ConfirmedAt == "" || p.PublishedAt != "" || p.ReceiptState != state {
			t.Fatal(p)
		}
		if _, err = s.db.Exec("UPDATE receipt_times SET confirmed_at=1700000123"); err != nil {
			t.Fatal(err)
		}
		if err = s.SaveReceipt(d, "channel", state, "remote"); err != nil {
			t.Fatal(err)
		}
		page, err = s.PostStatuses(ctx, PostFilter{})
		if err != nil || page.Items[0].ConfirmedAt != "2023-11-14T22:15:23Z" {
			t.Fatal("confirmation overwritten", page, err)
		}
		if err = s.RecordPublicationTime(ctx, d.Key.Platform, "channel", "remote", "2024-01-02T10:05:06Z"); err != nil {
			t.Fatal(err)
		}
		page, err = s.PostStatuses(ctx, PostFilter{})
		if err != nil {
			t.Fatal(err)
		}
		if state != "published" && page.Items[0].PublishedAt != "" {
			t.Fatal("non-public receipt given publication date")
		}
	}
	if err := s.SaveReceipt(d, "channel", "published", "remote"); err != nil {
		t.Fatal(err)
	}
	page, err := s.PostStatuses(ctx, PostFilter{})
	if err != nil || page.Items[0].PublishedAt != "2024-01-02T10:05:06Z" {
		t.Fatal("provider date lost", page, err)
	}
	if _, err = s.db.Exec("UPDATE upload_attempts SET state='sending'"); err != nil {
		t.Fatal(err)
	}
	page, err = s.PostStatuses(ctx, PostFilter{})
	if err != nil || page.Items[0].ConfirmedAt != "" || page.Items[0].PublishedAt != "" {
		t.Fatal("stale receipt date leaked", page, err)
	}
}
