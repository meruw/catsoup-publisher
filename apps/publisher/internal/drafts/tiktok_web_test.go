package drafts

import (
	"context"
	"path/filepath"
	"testing"
)

func TestWebAndAPIShareDuplicateProtection(t *testing.T) {
	for _, first := range []string{"inbox", "web_preparing", "web_ready", "web_check"} {
		t.Run(first, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "drafts.sqlite")
			s := openTest(t, path)
			ctx := context.Background()
			d := sample()
			d.Key.Platform = "tiktok"
			d.Key.Account = "account:es-id"
			if err := s.PutProfile(Profile{Platform: "tiktok", Language: "es", ID: "es-id", Name: "Same display name"}, []byte("fixture"), ""); err != nil {
				t.Fatal(err)
			}
			d, err := s.Save(ctx, d)
			if err != nil {
				t.Fatal(err)
			}
			if s.ReserveTikTokWeb(d, "es-id") == nil {
				t.Fatal("unapproved post accepted")
			}
			d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
			if err != nil {
				t.Fatal(err)
			}
			if s.ReserveTikTokWeb(d, "en-id") == nil {
				t.Fatal("wrong account accepted")
			}
			if err = s.ReserveTikTokWeb(d, "es-id"); err != nil {
				t.Fatal(err)
			}
			if err = s.SaveReceipt(d, "es-id", first, "original-receipt"); err != nil {
				t.Fatal(err)
			}
			other := openTest(t, path)
			if other.ReserveTikTokWeb(d, "es-id") == nil || other.Reserve(d, "es-id") == nil {
				t.Fatal("duplicate allowed")
			}
			if other.ResetAttempt(d, "es-id") == nil {
				t.Fatal("web or inbox receipt cleared")
			}
			r, err := other.Receipt("tiktok", "es-id", d.VideoHash)
			if err != nil || r.State != first || r.Reference != "original-receipt" || r.RetryAllowed {
				t.Fatalf("receipt changed: %+v %v", r, err)
			}
		})
	}
}
