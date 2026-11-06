package main

import (
	"catsoup/publisher/internal/drafts"
	"testing"
)

func TestXPaidConsentRequiredBeforeAccess(t *testing.T) {
	// Deliberately no store or client: rejected input must never reach either.
	s := UploadService{}
	key := drafts.Key{Platform: "x", Language: "es"}
	if _, err := s.Send(key, 1, "123", true); err == nil {
		t.Fatal("generic Send bypassed paid consent")
	}
	if _, err := s.SendX(key, 1, "123", true, false); err == nil {
		t.Fatal("missing paid consent accepted")
	}
	if _, err := s.SendX(key, 1, "123", false, true); err == nil {
		t.Fatal("missing destination consent accepted")
	}
	if _, err := s.LinkXAccount("es", "", "cats_es", false); err == nil {
		t.Fatal("sign-in without cost acknowledgment")
	}
	if _, err := s.LinkXAccount("es", "", "invalid name", true); err == nil {
		t.Fatal("sign-in without valid expected username")
	}
	key.Platform = "youtube"
	if _, err := s.SendX(key, 1, "123", true, true); err == nil {
		t.Fatal("X method accepts another provider")
	}
}
