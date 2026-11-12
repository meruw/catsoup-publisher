package drafts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstagramRetryReservation(t *testing.T) {
	for _, scenario := range []string{"failed", "sending", "published", "check", "check_publish", "publishing", "wrong_reference", "revoked", "edited", "wrong_account"} {
		t.Run(scenario, func(t *testing.T) {
			s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
			ctx := context.Background()
			d := sample()
			d.Key.Platform = "instagram"
			d.Key.Account = "account:ig1"
			d.DeliveryMode = "publish"
			if err := s.PutProfile(Profile{Platform: "instagram", Language: "es", ID: "ig1", Name: "IG"}, []byte("fixture"), ""); err != nil {
				t.Fatal(err)
			}
			var err error
			d, err = s.Save(ctx, d)
			if err != nil {
				t.Fatal(err)
			}
			d, err = s.SetPostApproval(ctx, d.Key, d.Revision, true)
			if err != nil {
				t.Fatal(err)
			}
			if err = s.Reserve(d, "ig1"); err != nil {
				t.Fatal(err)
			}
			state := "failed"
			if scenario == "sending" || scenario == "published" || scenario == "check" || scenario == "check_publish" || scenario == "publishing" {
				state = scenario
			}
			if err = s.SaveReceipt(d, "ig1", state, "old"); err != nil {
				t.Fatal(err)
			}
			ref, account := "old", "ig1"
			switch scenario {
			case "wrong_reference":
				ref = "other"
			case "wrong_account":
				account = "ig2"
			case "revoked":
				if _, err = s.SetPostApproval(ctx, d.Key, d.Revision, false); err != nil {
					t.Fatal(err)
				}
			case "edited":
				changed := d
				changed.Title = "changed"
				if _, err = s.Save(ctx, changed); err != nil {
					t.Fatal(err)
				}
			}
			err = s.ReserveInstagramRetry(d, account, ref)
			if (err == nil) != (scenario == "failed" || scenario == "check") {
				t.Fatal(scenario, err)
			}
			if scenario == "failed" {
				if err = s.SaveInstagramCheckpoint(d, "ig1", "replacement", strings.Repeat("b", 64)); err != nil {
					t.Fatal(err)
				}
				var source, uploaded, preparation string
				if err = s.db.QueryRow("SELECT video_hash,upload_hash,preparation FROM upload_attempts WHERE platform='instagram' AND account_id='ig1'").Scan(&source, &uploaded, &preparation); err != nil || source != d.VideoHash || uploaded != strings.Repeat("b", 64) || preparation != "instagram-mp4-v2" {
					t.Fatal(source, uploaded, preparation, err)
				}
				r, e := s.Receipt("instagram", "ig1", d.VideoHash)
				if e != nil || r.State != "sending" || r.Reference != "replacement" {
					t.Fatal(r, e)
				}
				if s.ReserveInstagramRetry(d, "ig1", "old") == nil {
					t.Fatal("duplicate reservation")
				}
			}
		})
	}
}

func TestUploadPreparationMigrationPreservesReceipt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('instagram','ig','hash','published','reel',123); DROP TABLE cover_assets; DROP TABLE cover_results; ALTER TABLE drafts DROP COLUMN cover; ALTER TABLE upload_attempts DROP COLUMN upload_hash; ALTER TABLE upload_attempts DROP COLUMN preparation; UPDATE schema_version SET version=8"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s = openTest(t, path)
	r, err := s.Receipt("instagram", "ig", "hash")
	if err != nil || r.State != "published" || r.Reference != "reel" {
		t.Fatal(r, err)
	}
	var hash, preparation string
	if err = s.db.QueryRow("SELECT upload_hash,preparation FROM upload_attempts").Scan(&hash, &preparation); err != nil || hash != "" || preparation != "" {
		t.Fatal(hash, preparation, err)
	}
}

func TestInstagramResumeReservation(t *testing.T) {
	for _, state := range []string{"failed", "check", "publishing", "check_publish", "published", "sending"} {
		t.Run(state, func(t *testing.T) {
			s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
			ctx := context.Background()
			d := sample()
			d.Key.Platform, d.Key.Account, d.DeliveryMode = "instagram", "account:ig1", "publish"
			if err := s.PutProfile(Profile{Platform: "instagram", Language: "es", ID: "ig1", Name: "IG"}, []byte("fixture"), ""); err != nil {
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
			if err = s.Reserve(d, "ig1"); err != nil {
				t.Fatal(err)
			}
			if err = s.SaveReceipt(d, "ig1", state, "original"); err != nil {
				t.Fatal(err)
			}
			if err = s.ReserveInstagramResume(d, "ig1", "different"); err == nil {
				t.Fatal("wrong container accepted")
			}
			err = s.ReserveInstagramResume(d, "ig1", "original")
			if (err == nil) != (state == "failed" || state == "check") {
				t.Fatal(state, err)
			}
			if err == nil {
				r, e := s.Receipt("instagram", "ig1", d.VideoHash)
				if e != nil || r.State != "publishing" || r.Reference != "original" {
					t.Fatal(r, e)
				}
				if s.ReserveInstagramResume(d, "ig1", "original") == nil || s.ReserveInstagramRetry(d, "ig1", "original") == nil {
					t.Fatal("duplicate operation allowed")
				}
			}
		})
	}
}

func TestInstagramPreparationVersion(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	for _, version := range []string{"instagram-aac128-v1", "instagram-mp4-v2", "", "unknown"} {
		_, err := s.db.Exec("INSERT OR REPLACE INTO upload_attempts(platform,account_id,video_hash,state,reference,started,preparation) VALUES('instagram','ig','hash','check','container',123,?)", version)
		if err != nil {
			t.Fatal(err)
		}
		legacy, err := s.InstagramLegacyPreparation("ig", "hash", "container")
		known := version == "instagram-aac128-v1" || version == "instagram-mp4-v2"
		if (err == nil) != known || legacy != (version == "instagram-aac128-v1") {
			t.Fatal(version, legacy, err)
		}
		if _, err = s.InstagramLegacyPreparation("ig", "hash", "other"); err == nil {
			t.Fatal("accepted another container")
		}
	}
}
