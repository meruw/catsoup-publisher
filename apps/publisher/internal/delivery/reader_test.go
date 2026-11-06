package delivery

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type validatorFunc func(context.Context, string) (Verdict, error)

func (f validatorFunc) Check(ctx context.Context, path string) (Verdict, error) { return f(ctx, path) }
func fixture(t *testing.T, id string, validate Validator) (*Reader, string, Manifest) {
	t.Helper()
	root := t.TempDir()
	base := filepath.Join(root, "deliveries", id)
	if err := os.MkdirAll(base, 0700); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "test.mp4"), []byte("test-only media"), 0600); err != nil {
		t.Fatalf("write fixture video: %v", err)
	}
	lang := "es"
	man := Manifest{Contract: "catsoup-delivery", ContractVersion: 2, ID: id, Title: "Fixture", Version: "v1", Status: "exported", Languages: []string{"es"}, Videos: []Video{{ID: "es-final", Lang: &lang, Kind: "final", Path: "test.mp4"}}}
	saveFixture(t, base, man)
	if validate == nil {
		validate = validatorFunc(func(context.Context, string) (Verdict, error) { return Verdict{}, nil })
	}
	reader, err := New(root, validate)
	if err != nil {
		t.Fatal(err)
	}
	return reader, base, man
}
func saveFixture(t *testing.T, base string, man Manifest) {
	t.Helper()
	data, err := json.Marshal(man)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(base, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestImportDoesNotWriteOrApprove(t *testing.T) {
	reader, base, _ := fixture(t, "catsoup-test", nil)
	before, _ := os.ReadFile(filepath.Join(base, "manifest.json"))
	result, err := reader.Inspect("catsoup-test")
	if err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(base, "manifest.json"))
	if string(before) != string(after) || result.Ready || result.Status != "exported" {
		t.Fatal("read-only import changed state")
	}
}
func TestRejectsIdentifierTraversal(t *testing.T) {
	reader, _, _ := fixture(t, "catsoup-test", nil)
	for _, id := range []string{"../catsoup-test", "catsoup-test/../../x", `C:\secret`, "catsoup-test?x", ""} {
		if _, err := reader.Inspect(id); err == nil {
			t.Errorf("accepted %q", id)
		}
	}
}

func TestManualUploadRequiresCurrentApprovedFinal(t *testing.T) {
	for _, tc := range []struct {
		name, id, status, lang, kind   string
		valid, stale, wrongHash, allow bool
	}{
		{"approved", "catsoup-test", "approved_for_publishing", "es", "final", true, false, false, true},
		{"validator refused", "catsoup-test", "approved_for_publishing", "es", "final", false, false, false, false},
		{"exported", "catsoup-test", "exported", "es", "final", true, false, false, false},
		{"clean", "catsoup-test", "approved_for_publishing", "es", "clean", true, false, false, false},
		{"wrong language", "catsoup-test", "approved_for_publishing", "en", "final", true, false, false, false},
		{"changed manifest", "catsoup-test", "approved_for_publishing", "es", "final", true, true, false, false},
		{"changed video", "catsoup-test", "approved_for_publishing", "es", "final", true, false, true, false},
		{"demo", "catsoup-demo", "approved_for_publishing", "es", "final", true, false, false, false},
		{"pilot", "catsoup-poc-streetlight", "approved_for_publishing", "es", "final", true, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reader, base, man := fixture(t, tc.id, validatorFunc(func(context.Context, string) (Verdict, error) {
				return Verdict{Publishable: tc.valid}, nil
			}))
			man.Status = tc.status
			man.Videos[0].Kind = tc.kind
			man.Videos[0].SHA256 = "saved-hash"
			saveFixture(t, base, man)
			episode, err := reader.Inspect(tc.id)
			if err != nil {
				t.Fatal(err)
			}
			if tc.stale {
				man.Version = "v2"
				saveFixture(t, base, man)
			}
			hash := man.Videos[0].SHA256
			if tc.wrongHash {
				hash = "another-hash"
			}
			path, err := reader.UploadPath(tc.id, tc.lang, "es-final", episode.Fingerprint, hash)
			if tc.allow {
				want, resolveErr := filepath.EvalSymlinks(filepath.Join(base, "test.mp4"))
				if resolveErr != nil {
					t.Fatal(resolveErr)
				}
				if err != nil || path != want {
					t.Fatalf("path=%q err=%v", path, err)
				}
			} else if err == nil || path != "" {
				t.Fatal("unexpected upload path", path)
			}
		})
	}
}
func TestRejectsPathEscapeBeforeValidator(t *testing.T) {
	called := false
	v := validatorFunc(func(context.Context, string) (Verdict, error) { called = true; return Verdict{}, nil })
	reader, base, man := fixture(t, "catsoup-test", v)
	outside := filepath.Join(reader.root, "outside.mp4")
	os.WriteFile(outside, []byte("outside"), 0600)
	for _, path := range []string{"../../outside.mp4", outside, `\\server\share\video.mp4`, "test.mp4:stream"} {
		man.Videos[0].Path = path
		saveFixture(t, base, man)
		if _, err := reader.Inspect("catsoup-test"); err == nil {
			t.Errorf("accepted escape %q", path)
		}
	}
	if called {
		t.Fatal("validator was called with an unsafe path")
	}
}
func TestLegacyExceptionIsOnlyForPoC(t *testing.T) {
	reader, base, man := fixture(t, "catsoup-poc-streetlight", validatorFunc(func(context.Context, string) (Verdict, error) { return Verdict{Publishable: true}, nil }))
	os.Mkdir(filepath.Join(reader.root, "render"), 0700)
	os.WriteFile(filepath.Join(reader.root, "render", "poc.mp4"), []byte("poc"), 0600)
	man.Videos[0].Path = "../../render/poc.mp4"
	saveFixture(t, base, man)
	episode, err := reader.Inspect(man.ID)
	if err != nil {
		t.Fatal(err)
	}
	if episode.Ready {
		t.Fatal("PoC must never be publishable")
	}
	other := filepath.Join(reader.root, "deliveries", "catsoup-other")
	os.Mkdir(other, 0700)
	man.ID = "catsoup-other"
	saveFixture(t, other, man)
	if _, err := reader.Inspect(man.ID); err == nil {
		t.Fatal("legacy exception leaked to another episode")
	}
}
func TestRejectsChangedManifestDuringValidation(t *testing.T) {
	v := validatorFunc(func(_ context.Context, path string) (Verdict, error) {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return Verdict{}, err
		}
		defer f.Close()
		f.WriteString("\n")
		return Verdict{}, nil
	})
	reader, _, _ := fixture(t, "catsoup-test", v)
	if _, err := reader.Inspect("catsoup-test"); err == nil {
		t.Fatal("accepted manifest changed during validation")
	}
}
func TestContractErrorsBlockImportAndMedia(t *testing.T) {
	reader, _, _ := fixture(t, "catsoup-test", validatorFunc(func(context.Context, string) (Verdict, error) {
		return Verdict{Errors: []string{"sha256 differs"}}, nil
	}))
	if _, err := reader.Inspect("catsoup-test"); err == nil {
		t.Fatal("accepted changed media")
	}
	if _, err := reader.Media("catsoup-test", "es-final"); err == nil {
		t.Fatal("served changed media")
	}
}
func TestNoGenericFileAccess(t *testing.T) {
	reader, _, _ := fixture(t, "catsoup-test", nil)
	if _, err := reader.Media("catsoup-test", "../../outside.mp4"); err == nil {
		t.Fatal("accepted arbitrary media path")
	}
}
func TestRejectsSymlinkOutsideDelivery(t *testing.T) {
	reader, base, man := fixture(t, "catsoup-test", nil)
	target := filepath.Join(reader.root, "outside.mp4")
	os.WriteFile(target, []byte("outside"), 0600)
	if err := os.Symlink(target, filepath.Join(base, "link.mp4")); err != nil {
		t.Skipf("symlink creation not available: %v", err)
	}
	man.Videos[0].Path = "link.mp4"
	saveFixture(t, base, man)
	if _, err := reader.Inspect(man.ID); err == nil {
		t.Fatal("accepted escaping symlink")
	}
}
