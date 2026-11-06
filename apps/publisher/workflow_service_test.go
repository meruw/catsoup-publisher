package main

import (
	"catsoup/publisher/internal/drafts"
	"testing"
)

func TestUnsupportedModesNeverReachProvider(t *testing.T) {
	svc := UploadService{}
	for _, c := range []struct{ platform, mode string }{{"tiktok", "publish"}, {"instagram", "draft"}, {"x", "draft"}, {"youtube", "other"}} {
		if _, err := svc.SendMode(drafts.Key{Platform: c.platform}, 1, "id", c.mode, true, true, "general", false); err == nil {
			t.Fatal("accepted unsupported mode", c)
		}
	}
	if _, err := svc.PublishYouTube(drafts.Key{Platform: "youtube"}, 1, "id", false, "general", false); err == nil {
		t.Fatal("unconfirmed promotion accepted")
	}
}
