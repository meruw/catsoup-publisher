package drafts

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestPostStatusesIsolationAndResults(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	for _, id := range []string{"first", "second"} {
		d := sample()
		d.Key.Account = "account:" + id
		d.Title = "Same title"
		if _, err := s.Save(ctx, d); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.PutProfile(Profile{Platform: "youtube", Language: "es", ID: "first", Name: "CATsoup"}, []byte("test"), ""); err != nil {
		t.Fatal(err)
	}
	_, err := s.db.Exec("INSERT INTO upload_attempts VALUES('youtube','first',?,'published','remote1',?)", sample().VideoHash, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	page, err := s.PostStatuses(ctx, PostFilter{State: "published"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Key.Account != "account:first" || !page.Items[0].Connected {
		t.Fatal(page, err)
	}
	page, err = s.PostStatuses(ctx, PostFilter{State: "draft"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Connected {
		t.Fatal(page, err)
	}
	// Failed and uncertain receipts must remain distinct; old sending is uncertain.
	for _, tc := range []struct{ raw, want string }{{"failed", "failed"}, {"check", "check"}, {"sending", "check"}, {"private", "private"}, {"inbox", "inbox"}} {
		_, err = s.db.Exec("UPDATE upload_attempts SET state=?,started=?", tc.raw, time.Now().Add(-time.Hour).Unix())
		if err != nil {
			t.Fatal(err)
		}
		page, err = s.PostStatuses(ctx, PostFilter{State: tc.want})
		if err != nil || len(page.Items) != 1 {
			t.Fatal(tc, page, err)
		}
	}
	page, err = s.PostStatuses(ctx, PostFilter{Search: "%"})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("search must be literal", page, err)
	}
	page, err = s.PostStatuses(ctx, PostFilter{Language: "en"})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("language leak", page, err)
	}
	for i := 0; i < 52; i++ {
		d := sample()
		d.Key.Episode = fmt.Sprintf("catsoup-page-%d", i)
		if _, err = s.Save(ctx, d); err != nil {
			t.Fatal(err)
		}
	}
	page, err = s.PostStatuses(ctx, PostFilter{Search: "catsoup-page"})
	if err != nil || len(page.Items) != 50 || !page.More {
		t.Fatal(page, err)
	}
	second, err := s.PostStatuses(ctx, PostFilter{Search: "catsoup-page", Offset: 50})
	if err != nil || len(second.Items) != 2 || second.More {
		t.Fatal(second, err)
	}
	for _, f := range []PostFilter{{State: "invalid"}, {Platform: "yt"}, {Language: "fr"}, {Offset: -1}} {
		if _, err = s.PostStatuses(ctx, f); err == nil {
			t.Fatal("accepted invalid filter", f)
		}
	}
}
