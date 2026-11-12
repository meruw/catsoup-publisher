package drafts

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestTikTokTrackingRestartIsolationAndBackoff(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d := sample()
	d.Key.Platform = "tiktok"
	d.Key.Account = "account:es"
	d, err := s.Save(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PutProfile(Profile{Platform: "tiktok", Language: "es", ID: "es", Name: "Spanish"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Unix()
	old := "v_inbox_file~v2.111"
	latest := "v_inbox_file~v2.222"
	_, err = s.db.Exec(`INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('tiktok','es',?,'check',?,?)`, d.VideoHash, latest, now-2000)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`INSERT INTO inbox_receipt_history VALUES('es',?,?,?,0,?)`, d.VideoHash, old, now-3000, now-2000)
	if err != nil {
		t.Fatal(err)
	}
	due, err := s.DueTikTok(ctx, now)
	if err != nil || len(due) != 2 {
		t.Fatal(due, err)
	}
	w := TikTokWatch{Account: "es", Language: "es", Reference: latest, Started: now - 2000}
	if err = s.SaveTikTokObservation(ctx, w, "SEND_TO_USER_INBOX", "", "", now); err != nil {
		t.Fatal(err)
	}
	r, err := s.Receipt("tiktok", "es", d.VideoHash)
	if err != nil || r.State != "inbox" || r.Tracking.Status != "SEND_TO_USER_INBOX" {
		t.Fatal(r, err)
	}
	// An old resend completing must not mark the current copy as published.
	archived := w
	archived.Reference = old
	if err = s.SaveTikTokObservation(ctx, archived, "PUBLISH_COMPLETE", "", "", now); err != nil {
		t.Fatal(err)
	}
	r, _ = s.Receipt("tiktok", "es", d.VideoHash)
	if r.State != "inbox" || r.InboxHistory[0].Tracking.Status != "PUBLISH_COMPLETE" {
		t.Fatal(r)
	}
	if err = s.SaveTikTokObservation(ctx, w, "", "", "network error with secret", now+1); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	r, err = s.Receipt("tiktok", "es", d.VideoHash)
	if err != nil || r.Tracking.Status != "SEND_TO_USER_INBOX" || r.Tracking.Problem == "" || r.RetryAllowed {
		t.Fatal(r, err)
	}
	if due, err = s.DueTikTok(ctx, now+2); err != nil || len(due) != 0 {
		t.Fatal("schedule lost", due, err)
	}
	if due, err = s.DueTikTok(ctx, now+130); err != nil || len(due) != 1 || due[0].Failures != 1 {
		t.Fatal(due, err)
	}
	if err = s.SaveTikTokObservation(ctx, w, "PUBLISH_COMPLETE", "", "", now+131); err != nil {
		t.Fatal(err)
	}
	r, _ = s.Receipt("tiktok", "es", d.VideoHash)
	if r.State != "published" || r.RetryAllowed {
		t.Fatal(r)
	}
	if due, err = s.DueTikTok(ctx, now+100000); err != nil || len(due) != 0 {
		t.Fatal("terminal receipts polled", due, err)
	}
}

func TestTikTokTrackingSkipsLiveSendAndDisconnectedAccounts(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "db"))
	now := time.Now().Unix()
	ctx := context.Background()
	if err := s.PutProfile(Profile{Platform: "tiktok", Language: "es", ID: "es", Name: "Spanish"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	_, err := s.db.Exec(`INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES
 ('tiktok','es','live','sending','v_inbox_file~v2.111',?),
 ('tiktok','removed','old','inbox','v_inbox_file~v2.222',1),
 ('tiktok','es','web','web_ready','web:123',1)`, now)
	if err != nil {
		t.Fatal(err)
	}
	due, err := s.DueTikTok(ctx, now)
	if err != nil || len(due) != 0 {
		t.Fatal(due, err)
	}
	due, err = s.DueTikTok(ctx, now+1000)
	if err != nil || len(due) != 1 || due[0].Reference != "v_inbox_file~v2.111" {
		t.Fatal(due, err)
	}
	// Cancellation must stop the worker without starting more provider work.
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	done := StartTikTokMonitor(cancelled, s, func(context.Context, TikTokWatch) (string, string, error) {
		t.Error("request after shutdown")
		return "", "", nil
	}, nil)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("monitor did not stop")
	}
}

func TestTikTokTrackingMigrationPreservesReceipts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "db")
	s := openTest(t, path)
	_, err := s.db.Exec(`DROP TABLE tiktok_tracking; UPDATE schema_version SET version=12; INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('tiktok','es','hash','inbox','v_inbox_file~v2.123',123)`)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	r, err := s.Receipt("tiktok", "es", "hash")
	if err != nil || r.State != "inbox" || r.Reference != "v_inbox_file~v2.123" {
		t.Fatal(r, err)
	}
}

func TestTikTokMonitorPersistsObservationBeforeNotification(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "db"))
	if err := s.PutProfile(Profile{Platform: "tiktok", Language: "en", ID: "en", Name: "English"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('tiktok','en','hash','check','v_inbox_file~v2.123',1)`); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := StartTikTokMonitor(ctx, s, func(_ context.Context, w TikTokWatch) (string, string, error) {
		if w.Language != "en" || w.Account != "en" || w.Reference != "v_inbox_file~v2.123" {
			t.Error("wrong destination", w)
		}
		return "SEND_TO_USER_INBOX", "", nil
	}, func() {
		r, err := s.Receipt("tiktok", "en", "hash")
		if err != nil || r.Tracking == nil || r.Tracking.Status != "SEND_TO_USER_INBOX" {
			t.Error("notification before commit", r, err)
		}
		cancel()
	})
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not complete")
	}
}
