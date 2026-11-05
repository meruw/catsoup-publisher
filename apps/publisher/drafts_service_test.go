package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"catsoup/publisher/internal/delivery"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
)

type draftTestValidator struct{ calls int }

func (v *draftTestValidator) Check(context.Context, string) (delivery.Verdict, error) {
	v.calls++
	return delivery.Verdict{}, nil
}
func TestDraftServicePreservesApprovalAndSource(t *testing.T) {
	root := t.TempDir()
	folder := filepath.Join(root, "deliveries", "catsoup-demo")
	os.MkdirAll(folder, 0700)
	media := []byte("fixture bytes")
	os.WriteFile(filepath.Join(folder, "video.mp4"), media, 0600)
	hash := sha256.Sum256(media)
	lang := "es"
	man := delivery.Manifest{Contract: "catsoup-delivery", ContractVersion: 2, ID: "catsoup-demo", Title: "Demo", Version: "v1", Status: "exported", Languages: []string{"es"}, Videos: []delivery.Video{{ID: "es_final", Lang: &lang, Kind: "final", Path: "video.mp4", SHA256: hex.EncodeToString(hash[:])}}}
	path := filepath.Join(folder, "manifest.json")
	raw, _ := json.Marshal(man)
	os.WriteFile(path, raw, 0600)
	validator := &draftTestValidator{}
	reader, err := delivery.New(root, validator)
	if err != nil {
		t.Fatal(err)
	}
	store, err := drafts.Open(filepath.Join(root, "data", "drafts.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	copied, revealed := "", ""
	svc := DraftService{store: store, reader: reader, copyText: func(s string) bool { copied = s; return true }, reveal: func(s string) error { revealed = s; return nil }}
	input := SaveDraftInput{Key: drafts.Key{Episode: "catsoup-demo", Language: "es", Platform: "youtube", Account: "@catsoup"}, Title: "Title", Description: "Description", Hashtags: "#cats", VideoID: "es_final"}
	d, err := svc.SaveDraft(input)
	if err != nil {
		t.Fatal(err)
	}
	if validator.calls != 1 {
		t.Fatal("new source was not validated")
	}
	originalFingerprint := d.DeliveryFingerprint
	// A later delivery does not silently rebind this draft on an ordinary text save.
	man.Version = "v2"
	changed, _ := json.Marshal(man)
	os.WriteFile(path, changed, 0600)
	input.Revision = d.Revision
	input.Title = "Updated"
	d, err = svc.SaveDraft(input)
	if err != nil || validator.calls != 1 || d.DeliveryFingerprint != originalFingerprint {
		t.Fatalf("autosave revalidated/rebound source: %+v %v", d, err)
	}
	if err = svc.CopyDraft(input.Key, "caption"); err != nil || copied != "Description\n\n#cats" {
		t.Fatalf("copy: %q %v", copied, err)
	}
	if err = svc.CopyDraft(input.Key, "unknown"); err == nil {
		t.Fatal("invalid clipboard operation")
	}
	input.Revision = d.Revision
	input.AttachVideo = true
	d, err = svc.SaveDraft(input)
	if err != nil || d.DeliveryFingerprint == originalFingerprint || validator.calls != 2 {
		t.Fatalf("explicit rebind failed: %v", err)
	}
	if _, err = svc.SaveDraft(input); err != drafts.ErrConflict {
		t.Fatal("accepted stale revision", err)
	}
	invalid := input
	invalid.Revision = d.Revision
	invalid.VideoID = "../../private"
	if _, err = svc.SaveDraft(invalid); err == nil {
		t.Fatal("accepted invalid video")
	}
	invalid = input
	invalid.Key.Language = "en"
	invalid.Revision = 0
	if _, err = svc.SaveDraft(invalid); err == nil {
		t.Fatal("accepted other language's video")
	}
	if err = svc.OpenVideoFolder("catsoup-demo", "es_final"); err != nil || revealed != filepath.Join(folder, "video.mp4") {
		t.Fatalf("reveal: %q %v", revealed, err)
	}
	revealed = ""
	if err = svc.OpenVideoFolder("../../private", "es_final"); err == nil || revealed != "" {
		t.Fatal("arbitrary path reached desktop")
	}
	after, _ := os.ReadFile(path)
	copied = ""
	for _, attempt := range []struct {
		revision  int64
		confirmed bool
	}{{d.Revision, false}, {d.Revision - 1, true}, {d.Revision, true}} {
		if err := svc.CopyUploadPath(input.Key, attempt.revision, attempt.confirmed); err == nil || copied != "" {
			t.Fatal("unauthorized handoff reached the clipboard")
		}
	}
	if string(after) != string(changed) {
		t.Fatal("draft actions changed the manifest")
	}
	// No provider client is attached: even a confirmed, connected demo draft
	// must fail locally before a provider is contacted.
	linked := input.Key
	linked.Episode = ""
	upload := UploadService{store: store, reader: reader, sessions: map[drafts.Key]connection{linked: {account: uploads.Account{ID: "test-channel"}}}}
	if _, err := upload.Send(input.Key, d.Revision, "test-channel", true); err == nil || !strings.Contains(err.Error(), "approved delivery") {
		t.Fatalf("demo was not blocked by approval: %v", err)
	}
}
