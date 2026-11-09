package main

import (
	"context"
	"os"
	"testing"

	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
)

// Explicit opt-in only. Reads counters for saved published posts. Token refresh
// may update encrypted credentials; no publication/approval methods are called.
func TestLivePostStatistics(t *testing.T) {
	path := os.Getenv("CATSOUP_LIVE_STATS_DB")
	if path == "" {
		t.Skip("live statistics check is opt-in")
	}
	s, err := drafts.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := &UploadService{store: s, client: uploads.New()}
	posts, err := s.PostStatuses(context.Background(), drafts.PostFilter{State: "published"})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range posts.Items {
		if p.Key.Platform != "youtube" && p.Key.Platform != "instagram" && p.Key.Platform != "facebook" {
			continue
		}
		t.Run(p.Key.Platform+"/"+p.Key.Language, func(t *testing.T) {
			result, err := service.PostStatistics(p.Key)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("Counters: %+v", result.Items)
		})
	}
}
