package drafts

import (
	"context"
	"path/filepath"
	"testing"
)

func TestPublishingDefaultsAreAccountScopedAndDoNotEditSavedPosts(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	profile := Profile{Platform: "youtube", Language: "es", ID: "first"}
	if err := s.PutProfile(profile, []byte("fixture"), ""); err != nil {
		t.Fatal(err)
	}
	key := Key{"catsoup-demo", "es", "youtube", "account:first"}
	d, err := s.Get(ctx, key)
	if err != nil || d.DeliveryMode != "draft" {
		t.Fatal(d, err)
	}
	d = sample()
	d.Key = key
	d, err = s.Save(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	d, err = s.SetPostApproval(ctx, key, d.Revision, true)
	if err != nil {
		t.Fatal(err)
	}
	pref := PublishingDefault{"youtube", "es", "first", "publish"}
	if err = s.SetPublishingDefault(ctx, pref); err != nil {
		t.Fatal(err)
	}
	existing, err := s.Get(ctx, key)
	if err != nil || existing != d {
		t.Fatal("preference changed saved post or approval", err)
	}
	key.Episode = "catsoup-new"
	fresh, err := s.Get(ctx, key)
	if err != nil || fresh.DeliveryMode != "publish" || fresh.Revision != 0 || fresh.ApprovedRevision != 0 {
		t.Fatal(fresh, err)
	}
	s.Close()
	s = openTest(t, path)
	fresh, err = s.Get(ctx, key)
	if err != nil || fresh.DeliveryMode != "publish" {
		t.Fatal("lost preference", err)
	}
	if err = s.PutProfile(Profile{Platform: "youtube", Language: "es", ID: "second"}, []byte("fixture"), "first"); err != nil {
		t.Fatal(err)
	}
	if err = s.SetPublishingDefault(ctx, pref); err == nil {
		t.Fatal("saved preference for stale account")
	}
	key.Account = "account:second"
	fresh, err = s.Get(ctx, key)
	if err != nil || fresh.DeliveryMode != "draft" {
		t.Fatal("new account inherited mode", err)
	}
	key.Language = "en"
	fresh, err = s.Get(ctx, key)
	if err != nil || fresh.DeliveryMode != "draft" {
		t.Fatal("language inherited mode", err)
	}
	pref.Platform = "tiktok"
	if err = s.SetPublishingDefault(ctx, pref); err == nil {
		t.Fatal("accepted TikTok direct post")
	}
	pref.Platform = "youtube"
	pref.Mode = "invalid"
	if err = s.SetPublishingDefault(ctx, pref); err == nil {
		t.Fatal("accepted invalid mode")
	}
}
func TestV6PublishingDefaultsMigration(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d, err := s.Save(ctx, sample())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec("DROP TABLE publishing_defaults; UPDATE schema_version SET version=6"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	got, err := s.Get(ctx, d.Key)
	if err != nil || got != d {
		t.Fatal("migration changed draft", err)
	}
	preferences, err := s.PublishingDefaults(ctx)
	if err != nil || len(preferences) != 0 {
		t.Fatal(preferences, err)
	}
}
