//go:build windows

package uploads

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestDeveloperMigrationPreservesClientsReceiptsAndTokens(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	fixtures := []struct{ platform, file, raw string }{
		{"youtube", "google-desktop.json", "{\n\"installed\":{\"client_id\":\"fixture-id\",\"client_secret\":\"fixture-secret\"}}"},
		{"tiktok", "tiktok-desktop.json", `{"client_key":"fixture-key","client_secret":"fixture-secret","direct_post":true}`},
		{"instagram", "meta-desktop.json", `{"app_id":"fixture-app","app_secret":"fixture-secret","config_id":"fixture-config"}`},
		{"x", "x-desktop.json", `{"client_id":"fixture-x","redirect_uri":"http://127.0.0.1:48173/callback/"}`},
	}
	fingerprints := map[string]string{}
	for _, f := range fixtures {
		if err := Configure(f.platform, f.raw); err != nil {
			t.Fatal(err)
		}
		cfg, err := desktopConfig(f.platform)
		if err != nil {
			t.Fatal(err)
		}
		fingerprints[f.platform] = cfg.fingerprint()
		receipt, _ := receiptPath(f.platform)
		if err := os.WriteFile(receipt, []byte(cfg.fingerprint()), 0600); err != nil {
			t.Fatal(err)
		}
		p, _ := developerPath(f.file)
		// Simulate an installation with existing legacy files and authorization receipts.
		if err := os.WriteFile(p, []byte(f.raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	before, err := (Token{Access: "fixture-access", Refresh: "fixture-refresh", Client: fingerprints["youtube"]}).Seal("youtube/es/fixture-account")
	if err != nil {
		t.Fatal(err)
	}
	dir, _ := os.UserConfigDir()
	tokenPath := filepath.Join(dir, "existing-account-token")
	if err := os.WriteFile(tokenPath, before, 0600); err != nil {
		t.Fatal(err)
	}
	ready, err := MigrateDeveloperCredentials()
	if err != nil || len(ready) != 4 {
		t.Fatal("migration failed", err)
	}
	for _, f := range fixtures {
		p, _ := developerPath(f.file)
		sealed, _ := os.ReadFile(p)
		if !bytes.HasPrefix(sealed, []byte(developerHeader)) || bytes.Contains(sealed, []byte("fixture-secret")) {
			t.Fatal("credentials not encrypted")
		}
		decoded, err := readDeveloperConfiguration(f.file)
		if err != nil || string(decoded) != f.raw {
			t.Fatal("migration changed configuration bytes")
		}
		cfg, err := desktopConfig(f.platform)
		if err != nil || cfg.fingerprint() != fingerprints[f.platform] {
			t.Fatal("client identity changed")
		}
		if ClientStatus(f.platform) != "confirmed" {
			t.Fatal("authorization receipt changed", ClientStatus(f.platform))
		}
		if _, err := MigrateDeveloperCredentials(); err != nil {
			t.Fatal(err)
		}
		again, _ := os.ReadFile(p)
		if !bytes.Equal(sealed, again) {
			t.Fatal("migration is not idempotent")
		}
	}
	after, _ := os.ReadFile(tokenPath)
	if !bytes.Equal(before, after) {
		t.Fatal("account token changed")
	}
	token, err := Unseal(after, "youtube/es/fixture-account")
	if err != nil || token.Access != "fixture-access" || token.Refresh != "fixture-refresh" {
		t.Fatal("existing token no longer opens")
	}
	if !TikTokDirectEnabled() {
		t.Fatal("direct post setting lost")
	}
}

func TestDeveloperStorageRejectsSwappedCorruptAndOversizedFiles(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	raw := `{"installed":{"client_id":"fixture","client_secret":"fixture-secret"}}`
	if err := Configure("youtube", raw); err != nil {
		t.Fatal(err)
	}
	google, _ := developerPath("google-desktop.json")
	tik, _ := developerPath("tiktok-desktop.json")
	blob, _ := os.ReadFile(google)
	if err := os.WriteFile(tik, blob, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readDeveloperConfiguration("tiktok-desktop.json"); err == nil {
		t.Fatal("provider-swapped file accepted")
	}
	for _, invalid := range [][]byte{[]byte(developerHeader + "corrupt"), []byte("CATSOUP-DEVELOPER-DPAPI-v9\n{}"), bytes.Repeat([]byte(" "), 60000), []byte("broken JSON")} {
		if err := os.WriteFile(google, invalid, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readDeveloperConfiguration("google-desktop.json"); err == nil {
			t.Fatal("invalid file accepted")
		}
		unchanged, _ := os.ReadFile(google)
		if !bytes.Equal(invalid, unchanged) {
			t.Fatal("invalid original overwritten")
		}
	}
	if _, err := readDeveloperConfiguration("../other.json"); err == nil {
		t.Fatal("arbitrary file accepted")
	}
}

func TestFailedMigrationKeepsOriginalAndRemovesTemporaryFile(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	p, _ := developerPath("google-desktop.json")
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"installed":{"client_id":"fixture","client_secret":"fixture-secret"}}`)
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	name, _ := windows.UTF16PtrFromString(p)
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	if _, err := readDeveloperConfiguration("google-desktop.json"); err == nil {
		t.Fatal("replacement should fail on locked destination")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(raw, after) {
		t.Fatal("failed migration changed original")
	}
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(p), "client-*.tmp"))
	if len(files) != 0 {
		t.Fatal("temporary credential file left behind")
	}
}
