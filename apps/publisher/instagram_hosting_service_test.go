package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstagramHostingConsentAndPersistence(t *testing.T) {
	root := t.TempDir()
	s := &UploadService{root: root}
	if s.InstagramHosting().Enabled {
		t.Fatal("enabled by default")
	}
	if s.SetInstagramHosting(true, true) == nil {
		t.Fatal("missing helper accepted")
	}
	os.MkdirAll(filepath.Join(root, ".tools"), 0700)
	os.MkdirAll(filepath.Join(root, "data", "publisher"), 0700)
	os.WriteFile(filepath.Join(root, ".tools", "cloudflared.exe"), []byte("fixture"), 0600)
	if s.SetInstagramHosting(true, false) == nil {
		t.Fatal("missing consent")
	}
	if e := s.SetInstagramHosting(true, true); e != nil {
		t.Fatal(e)
	}
	restarted := &UploadService{root: root}
	if !restarted.InstagramHosting().Enabled {
		t.Fatal("not persistent")
	}
	if e := restarted.SetInstagramHosting(false, false); e != nil {
		t.Fatal(e)
	}
	if restarted.InstagramHosting().Enabled {
		t.Fatal("not disabled")
	}
}
