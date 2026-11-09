package drafts

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDeliveryModeMigrationPreservesDraftAndRevision(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	d, err := s.Save(context.Background(), sample())
	if err != nil {
		t.Fatal(err)
	}
	// Recreate the last supported schema, with its original draft still present.
	if _, err = s.db.Exec("DROP TABLE statistics_samples; ALTER TABLE upload_attempts DROP COLUMN upload_hash; ALTER TABLE upload_attempts DROP COLUMN preparation; ALTER TABLE drafts DROP COLUMN approved_revision; ALTER TABLE drafts DROP COLUMN delivery_mode; UPDATE schema_version SET version=4"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	got, err := s.Get(context.Background(), d.Key)
	if err != nil || got != d {
		t.Fatalf("migration changed draft: %#v %v", got, err)
	}
	got.DeliveryMode = "publish"
	got, err = s.Save(context.Background(), got)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	again, err := s.Get(context.Background(), d.Key)
	if err != nil || again != got {
		t.Fatalf("mode not durable: %#v %v", again, err)
	}
	again.DeliveryMode = "invalid"
	if _, err = s.Save(context.Background(), again); err == nil {
		t.Fatal("accepted invalid mode")
	}
	again.Key.Platform = "tiktok"
	again.Revision = 0
	again.DeliveryMode = "publish"
	if _, err = s.Save(context.Background(), again); err != nil {
		t.Fatal("could not prepare TikTok direct post", err)
	}
}

func TestAccountIdentitySeparatesDrafts(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	d := sample()
	d.Key.Account = "account:first"
	d, err := s.Save(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if err = CheckDestination(d.Key, "second"); err == nil {
		t.Fatal("accepted another account")
	}
	if err = CheckDestination(d.Key, "first"); err != nil {
		t.Fatal(err)
	}
	d.Key.Account = "account:second"
	got, err := s.Get(context.Background(), d.Key)
	if err != nil || got.Revision != 0 {
		t.Fatalf("leaked text: %#v %v", got, err)
	}
}
