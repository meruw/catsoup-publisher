package main

import (
	"catsoup/publisher/internal/drafts"
	"testing"
)

func TestInboxResendRequiresExplicitConsentAndTikTok(t *testing.T) {
	s := &UploadService{}
	if s.AllowTikTokInboxResend(drafts.Key{Platform: "tiktok"}, "", "", "", false) == nil {
		t.Fatal("missing consent accepted")
	}
	if s.AllowTikTokInboxResend(drafts.Key{Platform: "instagram"}, "", "", "", true) == nil {
		t.Fatal("other platform accepted")
	}
}
