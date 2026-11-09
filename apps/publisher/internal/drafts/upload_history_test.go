package drafts

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestUploadHistorySurvivesRestartAndPaginatesVideos(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	for i := 0; i < 23; i++ {
		for _, lang := range []string{"es", "en"} {
			d := sample()
			d.Key.Episode = fmt.Sprintf("catsoup-history-%02d", i)
			d.Key.Language = lang
			d.Key.Account = "account:" + fmt.Sprintf("destination%d%s", i, lang)
			if _, err := s.Save(ctx, d); err != nil {
				t.Fatal(err)
			}
			state := "published"
			if i == 0 {
				state = "check"
			}
			if _, err := s.db.Exec("INSERT INTO upload_attempts(platform,account_id,video_hash,state,reference,started) VALUES(?,?,?,?,?,?)", d.Key.Platform, d.Key.Account[8:], d.VideoHash, state, "remote", time.Now().Unix()+int64(i)); err != nil {
				t.Fatal(err)
			}
		}
	}
	unsent := sample()
	unsent.Key.Episode = "catsoup-never-sent"
	if _, err := s.Save(ctx, unsent); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s = openTest(t, path)
	page, err := s.PostGroups(ctx, PostFilter{State: "transfers"})
	if err != nil || len(page.Items) != 20 || !page.More {
		t.Fatal(page, err)
	}
	next, err := s.PostGroups(ctx, PostFilter{State: "transfers", Offset: 20})
	if err != nil || len(next.Items) != 3 || next.More {
		t.Fatal(next, err)
	}
	seen := map[string]bool{}
	for _, g := range append(page.Items, next.Items...) {
		if seen[g.Episode] || g.Total != 2 || g.Episode == "catsoup-never-sent" {
			t.Fatal("split, repeated or unsent video", g)
		}
		seen[g.Episode] = true
	}
	checks, err := s.PostStatuses(ctx, PostFilter{State: "transfers", Episode: "catsoup-history-00"})
	if err != nil || len(checks.Items) != 2 {
		t.Fatal(checks, err)
	}
	for _, p := range checks.Items {
		if p.State != "check" {
			t.Fatal("uncertain result became success", p)
		}
	}
}
