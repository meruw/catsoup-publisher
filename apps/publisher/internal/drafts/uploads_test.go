package drafts

import (
	"context"
	"path/filepath"
	"testing"
)

func TestV1MigrationAndUploadReconciliation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d, err := s.Save(context.Background(), sample())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec("DROP TABLE accounts; DROP TABLE upload_attempts; DROP TABLE reviews; UPDATE schema_version SET version=1"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	if got, err := s.Get(context.Background(), d.Key); err != nil || got != d {
		t.Fatal("lost v1 draft")
	}
	if err = s.PutProfile(Profile{Platform: d.Key.Platform, Language: d.Key.Language, ID: "channel", Name: "Channel"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	reviewForTest(t, s, d)
	if err = s.Reserve(d, "channel"); err != nil {
		t.Fatal(err)
	}
	second := openTest(t, path)
	if second.Reserve(d, "channel") == nil || second.ResetAttempt(d, "channel") == nil {
		t.Fatal("second instance duplicated a live upload")
	}
	if err = s.SaveReceipt(d, "channel", "sending", "container"); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveReceipt(d, "channel", "check", ""); err != nil {
		t.Fatal(err)
	}
	r, err := second.Receipt(d.Key.Platform, "channel", d.VideoHash)
	if err != nil || !r.RetryAllowed || r.Reference != "container" {
		t.Fatal(r, err)
	}
	if err = second.ResetAttempt(d, "channel"); err != nil {
		t.Fatal(err)
	}
	if err = second.Reserve(d, "channel"); err != nil {
		t.Fatal(err)
	}
	if err = second.SaveReceipt(d, "channel", "sent", "post"); err != nil {
		t.Fatal(err)
	}
	if s.ResetAttempt(d, "channel") == nil || s.Reserve(d, "channel") == nil {
		t.Fatal("allowed repeat after completion")
	}
}
