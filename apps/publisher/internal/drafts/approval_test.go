package drafts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostApprovalRevisionAndAccountIsolation(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d := sample()
	d.Key.Account = "account:first"
	if err := s.PutProfile(Profile{Platform: "youtube", Language: "es", ID: "first", Name: "Same name"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	d, err := s.Save(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	reviewForTest(t, s, d)
	if s.RequireReview(ctx, d) == nil || s.Reserve(d, "first") == nil {
		t.Fatal("legacy review must not approve a bound draft")
	}
	d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	if err = s.RequireReview(ctx, d); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Draft){func(d *Draft) { d.Title = "Edited" }, func(d *Draft) { d.DeliveryMode = "publish" }, func(d *Draft) { d.VideoHash = strings.Repeat("c", 64) }, func(d *Draft) { d.DeliveryFingerprint = strings.Repeat("d", 64) }} {
		old := d
		change(&d)
		d, err = s.Save(ctx, d)
		if err != nil {
			t.Fatal(err)
		}
		if d.ApprovedRevision != 0 || s.RequireReview(ctx, old) == nil || s.Reserve(old, "first") == nil || s.Reserve(d, "first") == nil {
			t.Fatal("edit retained approval")
		}
		if _, err = s.SetPostApproval(ctx, d.Key, old.Revision, true); err == nil {
			t.Fatal("approved stale revision")
		}
		d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.SetPostApproval(ctx, d.Key, d.Revision, false); err != nil {
		t.Fatal(err)
	}
	if s.Reserve(d, "first") == nil {
		t.Fatal("reservation accepted revoked approval from stale object")
	}
	d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PutProfile(Profile{Platform: "youtube", Language: "es", ID: "second", Name: "Same name"}, []byte("fixture"), "first"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetPostApproval(ctx, d.Key, d.Revision, true); err == nil {
		t.Fatal("approved disconnected account")
	}
	if s.Reserve(d, "second") == nil || s.Reserve(d, "first") == nil {
		t.Fatal("sent to replaced account")
	}
	d.Key.Account = "account:second"
	d.Revision = 0
	d, err = s.Save(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	if s.RequireReview(ctx, d) == nil {
		t.Fatal("new account inherited approval")
	}
}

func TestPostApprovalPromotionRechecksRevocation(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	d := sample()
	d.Key.Account = "account:channel"
	if err := s.PutProfile(Profile{Platform: "youtube", Language: "es", ID: "channel"}, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	d, err := s.Save(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Reserve(d, "channel"); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveReceipt(d, "channel", "private", "video-id"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetPostApproval(ctx, d.Key, d.Revision, false); err != nil {
		t.Fatal(err)
	}
	if s.ReservePromotion(d, "channel", "video-id") == nil {
		t.Fatal("promoted revoked post")
	}
	d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.ReservePromotion(d, "channel", "video-id"); err != nil {
		t.Fatal(err)
	}
}

func TestV5ApprovalMigrationDoesNotApproveExistingDraft(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	d, err := s.Save(ctx, sample())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec("ALTER TABLE drafts DROP COLUMN approved_revision; UPDATE schema_version SET version=5"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	got, err := s.Get(ctx, d.Key)
	if err != nil || got != d || got.ApprovedRevision != 0 {
		t.Fatalf("migration: %#v %v", got, err)
	}
}
