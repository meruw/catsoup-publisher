package delivery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func importFixture(t *testing.T) (*Reader, string, Manifest) {
	t.Helper()
	_, source, man := fixture(t, "catsoup-import", nil)
	b, _ := os.ReadFile(filepath.Join(source, "test.mp4"))
	h := sha256.Sum256(b)
	man.Videos[0].SHA256 = hex.EncodeToString(h[:])
	saveFixture(t, source, man)
	r, _, _ := fixture(t, "catsoup-existing", nil)
	return r, source, man
}
func TestImportCopyAndRepeat(t *testing.T) {
	r, source, man := importFixture(t)
	before, _ := os.ReadFile(filepath.Join(source, "manifest.json"))
	os.WriteFile(filepath.Join(source, "not-declared.secret"), []byte("must not copy"), 0600)
	got, err := r.Import(source)
	if err != nil || got.ID != man.ID || got.Existing {
		t.Fatal(got, err)
	}
	target := filepath.Join(r.root, "deliveries", man.ID)
	after, _ := os.ReadFile(filepath.Join(target, "manifest.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("manifest modified")
	}
	if _, err = os.Stat(filepath.Join(target, "not-declared.secret")); !os.IsNotExist(err) {
		t.Fatal("copied undeclared file")
	}
	ep, err := r.Inspect(man.ID)
	if err != nil || ep.Ready || ep.Status != "exported" {
		t.Fatal(ep, err)
	}
	got, err = r.Import(source)
	if err != nil || !got.Existing {
		t.Fatal(got, err)
	}
	original, _ := os.ReadFile(filepath.Join(source, "manifest.json"))
	if !bytes.Equal(before, original) {
		t.Fatal("source changed")
	}
	man.Version = "v2"
	saveFixture(t, source, man)
	if _, err = r.Import(source); err == nil {
		t.Fatal("overwrote existing episode")
	}
	after, _ = os.ReadFile(filepath.Join(target, "manifest.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("existing changed")
	}
}
func TestImportRejectsIncompleteAndUnsafePackages(t *testing.T) {
	for _, kind := range []string{"hash", "missing", "traversal", "absolute", "colon", "duplicate", "contract", "validator"} {
		t.Run(kind, func(t *testing.T) {
			r, source, man := importFixture(t)
			switch kind {
			case "hash":
				os.WriteFile(filepath.Join(source, "test.mp4"), []byte("changed"), 0600)
			case "missing":
				os.Remove(filepath.Join(source, "test.mp4"))
			case "traversal":
				man.Videos[0].Path = "../test.mp4"
			case "absolute":
				man.Videos[0].Path = filepath.Join(source, "test.mp4")
			case "colon":
				man.Videos[0].Path = "test.mp4:hidden"
			case "duplicate":
				man.Videos = append(man.Videos, man.Videos[0])
			case "contract":
				man.ContractVersion = 99
			case "validator":
				r.validator = validatorFunc(func(context.Context, string) (Verdict, error) { return Verdict{}, errors.New("rejected") })
			}
			saveFixture(t, source, man)
			if _, err := r.Import(source); err == nil {
				t.Fatal("accepted unsafe or incomplete delivery")
			}
			if _, err := os.Stat(filepath.Join(r.root, "deliveries", man.ID)); !os.IsNotExist(err) {
				t.Fatal("partial import became visible")
			}
			stages, _ := filepath.Glob(filepath.Join(r.root, "deliveries", ".import-*"))
			if len(stages) != 0 {
				t.Fatal("staging files left behind")
			}
		})
	}
}
func TestRealPackageImport(t *testing.T) {
	source := os.Getenv("CATSOUP_IMPORT_SOURCE")
	workspace := os.Getenv("CATSOUP_TEST_ROOT")
	if source == "" || workspace == "" {
		t.Skip("optional real-validator import")
	}
	r, err := New(t.TempDir(), PythonValidator{Root: workspace})
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Import(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = r.Inspect(result.ID); err != nil {
		t.Fatal(err)
	}
}
