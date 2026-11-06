package drafts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewPersistsAndRequiresExactSource(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d := sample()
	if s.RequireReview(ctx, d) == nil {
		t.Fatal("unreviewed video accepted")
	}
	r := Review{Episode: d.Key.Episode, Language: d.Key.Language, VideoID: d.VideoID, Fingerprint: d.DeliveryFingerprint, VideoHash: d.VideoHash, State: "reviewed", Notes: "Voice checked"}
	saved, err := s.SaveReview(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	if err = s.RequireReview(ctx, d); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SaveReview(ctx, r); err == nil {
		t.Fatal("stale review overwrite")
	}
	for _, edit := range []func(*Draft){func(d *Draft) { d.Key.Language = "en" }, func(d *Draft) { d.VideoID = "other" }, func(d *Draft) { d.VideoHash = strings.Repeat("c", 64) }, func(d *Draft) { d.DeliveryFingerprint = strings.Repeat("d", 64) }} {
		changed := d
		edit(&changed)
		if s.RequireReview(ctx, changed) == nil {
			t.Fatal("review leaked to different source")
		}
	}
	saved.State = "changes_requested"
	if _, err = s.SaveReview(ctx, saved); err != nil {
		t.Fatal(err)
	}
	if s.RequireReview(ctx, d) == nil {
		t.Fatal("requested changes allowed send")
	}
}

func reviewForTest(t *testing.T, s *Store, d Draft) {
	t.Helper()
	_, err := s.SaveReview(context.Background(), Review{Episode: d.Key.Episode, Language: d.Key.Language, VideoID: d.VideoID, Fingerprint: d.DeliveryFingerprint, VideoHash: d.VideoHash, State: "reviewed"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPromotionLocksExistingReceiptWithoutAllowingDuplicate(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	d, err := s.Save(context.Background(), sample())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PutProfile(Profile{Platform: "youtube", Language: "es", ID: "channel", Name: "Channel"}, []byte("test"), ""); err != nil {
		t.Fatal(err)
	}
	if s.Reserve(d, "channel") == nil {
		t.Fatal("unreviewed upload reserved")
	}
	reviewForTest(t, s, d)
	if err = s.Reserve(d, "channel"); err != nil {
		t.Fatal(err)
	}
	s.SaveReceipt(d, "channel", "private", "video1")
	if s.ReservePromotion(d, "other", "video1") == nil || s.ReservePromotion(d, "channel", "other") == nil {
		t.Fatal("wrong destination/reference promoted")
	}
	if err = s.ReservePromotion(d, "channel", "video1"); err != nil {
		t.Fatal(err)
	}
	if s.ReservePromotion(d, "channel", "video1") == nil {
		t.Fatal("concurrent promotion accepted")
	}
	s.SaveReceipt(d, "channel", "check_publish", "")
	if s.ResetAttempt(d, "channel") == nil || s.Reserve(d, "channel") == nil {
		t.Fatal("uncertain promotion allowed duplicate upload")
	}
	if err = s.ReservePromotion(d, "channel", "video1"); err != nil {
		t.Fatal(err)
	}
	s.SaveReceipt(d, "channel", "published", "video1")
	if s.ReservePromotion(d, "channel", "video1") == nil || s.Reserve(d, "channel") == nil {
		t.Fatal("repeated publication allowed")
	}
	r, err := s.Receipt("youtube", "channel", d.VideoHash)
	if err != nil || r.Reference != "video1" || r.State != "published" {
		t.Fatal(r, err)
	}
}
