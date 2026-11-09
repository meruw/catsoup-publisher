package drafts

import (
	"context"
	"path/filepath"
	"testing"
)

func TestFacebookDestinationIsolation(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	for _, p := range []Profile{{Platform: "facebook", Language: "es", ID: "22", Name: "Page ES"}, {Platform: "facebook", Language: "en", ID: "33", Name: "Page EN"}, {Platform: "instagram", Language: "es", ID: "22", Name: "IG ES"}} {
		if err := s.PutProfile(p, []byte("fixture"), ""); err != nil {
			t.Fatal(err)
		}
		d := sample()
		d.Key.Platform = p.Platform
		d.Key.Language = p.Language
		d.Key.Account = "account:" + p.ID
		d.DeliveryMode = "publish"
		saved, err := s.Save(ctx, d)
		if err != nil {
			t.Fatal(err)
		}
		if p.Platform == "facebook" && p.Language == "es" {
			if _, err = s.SetPostApproval(ctx, saved.Key, saved.Revision, true); err != nil {
				t.Fatal(err)
			}
			if err = s.Reserve(saved, p.ID); err != nil {
				t.Fatal(err)
			}
			if err := s.SaveReceipt(saved, p.ID, "published", "100"); err != nil {
				t.Fatal(err)
			}
		}
	}
	page, err := s.PostStatuses(ctx, PostFilter{Platform: "facebook", State: "published"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Key.Language != "es" || page.Items[0].Reference != "100" {
		t.Fatal(page, err)
	}
	for _, p := range []struct{ platform, id string }{{"facebook", "33"}, {"instagram", "22"}} {
		r, err := s.Receipt(p.platform, p.id, sample().VideoHash)
		if err != nil || r.State != "" {
			t.Fatal("receipt leaked across destinations", r, err)
		}
	}
	if err := s.SetPublishingDefault(ctx, PublishingDefault{Platform: "facebook", Language: "es", AccountID: "22", Mode: "publish"}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutProfile(Profile{Platform: "facebook", Language: "en", ID: "22", Name: "Wrong language"}, []byte("fixture"), "33"); err == nil {
		t.Fatal("assigned same Page to both languages")
	}
}
