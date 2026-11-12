package drafts

import (
	"context"
	"path/filepath"
	"testing"
)

func TestInboxResendPreservesHistoryAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d := sample()
	d.Key.Platform = "tiktok"
	d.Key.Account = "account:es"
	var err error
	d, err = s.Save(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PutProfile(Profile{Platform: "tiktok", Language: "es", ID: "es", Name: "Spanish"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	insert := func(ref string) {
		t.Helper()
		_, e := s.db.Exec(`INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('tiktok','es',?,'inbox',?,123)`, d.VideoHash, ref)
		if e != nil {
			t.Fatal(e)
		}
		if e = s.SaveReceipt(d, "es", "inbox", ref); e != nil {
			t.Fatal(e)
		}
	}
	insert("first")
	if s.AllowInboxResend(d, "es", "stale") == nil {
		t.Fatal("stale reference accepted")
	}
	wrong := d
	wrong.Key.Language = "en"
	if s.AllowInboxResend(wrong, "es", "first") == nil {
		t.Fatal("wrong language accepted")
	}
	stale := d
	stale.Revision++
	if s.AllowInboxResend(stale, "es", "first") == nil {
		t.Fatal("stale draft accepted")
	}
	if err = s.AllowInboxResend(d, "es", "first"); err != nil {
		t.Fatal(err)
	}
	if s.AllowInboxResend(d, "es", "first") == nil {
		t.Fatal("repeated reset accepted")
	}
	r, err := s.Receipt("tiktok", "es", d.VideoHash)
	if err != nil || r.State != "" || len(r.InboxHistory) != 1 || r.InboxHistory[0].Reference != "first" || r.InboxHistory[0].ConfirmedAt == 0 {
		t.Fatal(r, err)
	}
	insert("second")
	if err = s.AllowInboxResend(d, "es", "second"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	r, err = s.Receipt("tiktok", "es", d.VideoHash)
	if err != nil || len(r.InboxHistory) != 2 {
		t.Fatal(r, err)
	}
	for _, key := range [][2]string{{"en", d.VideoHash}, {"es", "other"}} {
		h, e := s.InboxHistory(key[0], key[1])
		if e != nil || len(h) != 0 {
			t.Fatal("history leaked", e)
		}
	}
	for _, state := range []string{"sending", "publishing", "check", "failed", "published", "private", "web_ready", "restricted"} {
		_, err = s.db.Exec(`INSERT OR REPLACE INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('tiktok','es',?,?,?,123)`, d.VideoHash, state, state)
		if err != nil {
			t.Fatal(err)
		}
		if s.AllowInboxResend(d, "es", state) == nil {
			t.Fatal("unlocked", state)
		}
	}
	d.Key.Platform = "instagram"
	if s.AllowInboxResend(d, "es", "first") == nil {
		t.Fatal("other platform accepted")
	}
}
