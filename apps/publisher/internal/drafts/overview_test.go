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

func TestPostGroupsKeepAllDestinationsTogether(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	for _, lang := range []string{"es", "en"} {
		for _, platform := range []string{"youtube", "tiktok", "instagram", "x"} {
			d := sample()
			d.Key.Episode = "catsoup-grouped-video"
			d.Key.Language = lang
			d.Key.Platform = platform
			d.Key.Account = "account:" + platform + lang
			if _, err := s.Save(ctx, d); err != nil {
				t.Fatal(err)
			}
		}
	}
	_, err := s.db.Exec("INSERT INTO upload_attempts VALUES('youtube','youtubees',?,'published','remote',?)", sample().VideoHash, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	page, err := s.PostGroups(ctx, PostFilter{Platform: "youtube", State: "published"})
	if err != nil || len(page.Items) != 1 {
		t.Fatal(page, err)
	}
	byTitle, e := s.PostGroups(ctx, PostFilter{Search: "Display title", EpisodeMatches: []string{"catsoup-grouped-video"}})
	if e != nil || len(byTitle.Items) != 1 {
		t.Fatal("catalog title not found", byTitle, e)
	}
	g := page.Items[0]
	if g.Total != 8 || g.Platforms != 4 || g.Counts["published"] != 1 || g.Counts["draft"] != 7 {
		t.Fatal("partial summary", g)
	}
	page, err = s.PostGroups(ctx, PostFilter{Platform: "tiktok", State: "published"})
	if err != nil || len(page.Items) != 0 {
		t.Fatal("filters must match the same destination", page, err)
	}
	details, err := s.PostStatuses(ctx, PostFilter{Episode: "catsoup-grouped-video", Language: "en"})
	if err != nil || len(details.Items) != 4 {
		t.Fatal(details, err)
	}
	for i := 0; i < 23; i++ {
		d := sample()
		d.Key.Episode = fmt.Sprintf("catsoup-other-%02d", i)
		if _, err = s.Save(ctx, d); err != nil {
			t.Fatal(err)
		}
	}
	page, err = s.PostGroups(ctx, PostFilter{})
	if err != nil || len(page.Items) != 20 || !page.More {
		t.Fatal(page, err)
	}
	next, err := s.PostGroups(ctx, PostFilter{Offset: 20})
	if err != nil || len(next.Items) != 4 || next.More {
		t.Fatal(next, err)
	}
	seen := map[string]bool{}
	for _, p := range append(page.Items, next.Items...) {
		if seen[p.Episode] {
			t.Fatal("group repeated across pages")
		}
		seen[p.Episode] = true
	}
	details, err = s.PostStatuses(ctx, PostFilter{Episode: "catsoup-grouped-video"})
	if err != nil || len(details.Items) != 8 {
		t.Fatal("unrelated episode leaked", details, err)
	}
}

func TestPostFilterCategories(t *testing.T) {
	s := openTest(t, filepath.Join(t.TempDir(), "drafts.sqlite"))
	ctx := context.Background()
	d := sample()
	d.Key.Account = "account:category-test"
	if _, err := s.Save(ctx, d); err != nil {
		t.Fatal(err)
	}
	page, err := s.PostGroups(ctx, PostFilter{State: "unsent"})
	if err != nil || len(page.Items) != 1 || page.Items[0].VideoID != d.VideoID {
		t.Fatal(page, err)
	}
	_, err = s.db.Exec("INSERT INTO upload_attempts VALUES(?,?,?,'published','remote',?)", d.Key.Platform, "category-test", d.VideoHash, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ state, category string }{
		{"published", "delivered"}, {"private", "delivered"}, {"restricted", "delivered"}, {"inbox", "delivered"},
		{"sending", "active"}, {"failed", "attention"}, {"check", "attention"},
	} {
		if _, err = s.db.Exec("UPDATE upload_attempts SET state=?", tc.state); err != nil {
			t.Fatal(err)
		}
		for _, filter := range []string{"unsent", "active", "delivered", "attention"} {
			groups, e := s.PostGroups(ctx, PostFilter{State: filter})
			details, de := s.PostStatuses(ctx, PostFilter{State: filter})
			want := 0
			if filter == tc.category {
				want = 1
			}
			if e != nil || de != nil || len(groups.Items) != want || len(details.Items) != want {
				t.Fatal(tc, filter, groups, details, e, de)
			}
			if want == 1 && details.Items[0].State != tc.state {
				t.Fatal("category replaced precise receipt state", details)
			}
		}
	}
}
