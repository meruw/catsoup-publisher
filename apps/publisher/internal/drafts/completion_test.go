package drafts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompletionRequiresEveryConnectedDestination(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	for _, lang := range []string{"es", "en"} {
		for _, platform := range []string{"youtube", "instagram", "facebook", "tiktok", "x"} {
			id := platform + "-" + lang
			if err := s.PutProfile(Profile{Platform: platform, Language: lang, ID: id, Name: id}, []byte("fixture"), ""); err != nil {
				t.Fatal(err)
			}
			// A connected destination without a draft must prevent completion.
			if platform == "facebook" && lang == "en" {
				continue
			}
			d := sample()
			d.Key.Platform = platform
			d.Key.Language = lang
			d.Key.Account = "account:" + id
			if _, err := s.Save(ctx, d); err != nil {
				t.Fatal(err)
			}
			state := "published"
			if platform == "tiktok" {
				state = "inbox"
			}
			if platform == "x" {
				state = "check"
			}
			if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES(?,?,?,?,?,1)", platform, id, d.VideoHash, state, "fixture"); err != nil {
				t.Fatal(err)
			}
		}
	}
	check := func(want int) {
		t.Helper()
		got, err := s.CompletedLanguages(ctx)
		if err != nil || len(got) != want {
			t.Fatal(got, err)
		}
		for _, c := range got {
			if c.Fingerprint != sample().DeliveryFingerprint || c.Episode != sample().Key.Episode {
				t.Fatal(c)
			}
		}
	}
	check(1) // ES is complete; missing EN Facebook is not ignored.
	d := sample()
	d.Key.Platform = "facebook"
	d.Key.Language = "en"
	d.Key.Account = "account:facebook-en"
	if _, err := s.Save(ctx, d); err != nil {
		t.Fatal(err)
	}
	check(1)
	if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES('facebook','facebook-en',?,'published','fixture',1)", d.VideoHash); err != nil {
		t.Fatal(err)
	}
	check(2) // X remains unresolved in both languages and does not block.
	for _, state := range []string{"private", "restricted", "sending", "check", "failed", "web_ready", "sent"} {
		if err := s.SaveReceipt(d, "facebook-en", state, "fixture"); err != nil {
			t.Fatal(err)
		}
		check(1)
	}
	if err := s.SaveReceipt(d, "facebook-en", "published", "fixture"); err != nil {
		t.Fatal(err)
	}
	check(2)
	if _, err := s.db.Exec("UPDATE drafts SET video_hash=? WHERE platform='facebook' AND language='en'", strings.Repeat("c", 64)); err != nil {
		t.Fatal(err)
	}
	check(1)
	if _, err := s.db.Exec("UPDATE drafts SET video_hash=? WHERE platform='facebook' AND language='en'", d.VideoHash); err != nil {
		t.Fatal(err)
	}
	check(2)
	if err := s.PutProfile(Profile{Platform: "facebook", Language: "en", ID: "replacement", Name: "new"}, []byte("fixture"), "facebook-en"); err != nil {
		t.Fatal(err)
	}
	check(1)
	if _, err := s.db.Exec("DELETE FROM accounts WHERE platform<>'x'"); err != nil {
		t.Fatal(err)
	}
	check(0)
}

func TestCompletionCannotMixDeliveryVersions(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	for i, platform := range []string{"youtube", "tiktok"} {
		id := platform + "-es"
		d := sample()
		d.Key.Platform = platform
		d.Key.Account = "account:" + id
		if i == 1 {
			d.DeliveryFingerprint = strings.Repeat("c", 64)
		}
		if err := s.PutProfile(Profile{Platform: platform, Language: "es", ID: id, Name: id}, []byte("fixture"), ""); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Save(ctx, d); err != nil {
			t.Fatal(err)
		}
		if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES(?,?,?,'published','fixture',1)", platform, id, d.VideoHash); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.CompletedLanguages(ctx)
	if err != nil || len(got) != 0 {
		t.Fatal("mixed versions hidden", got, err)
	}
}
