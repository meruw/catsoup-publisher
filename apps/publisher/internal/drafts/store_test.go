package drafts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func sample() Draft {
	return Draft{Key: Key{"catsoup-demo", "es", "youtube", "CATsoup"}, Title: "A título ' ; DROP TABLE drafts; --", Description: "Una descripción\ncon acentos", Hashtags: "#gatos #cats", VideoID: "es_final", VideoHash: strings.Repeat("a", 64), DeliveryFingerprint: strings.Repeat("b", 64)}
}
func openTest(t *testing.T, path string) *Store {
	t.Helper()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func TestPersistenceMigrationAndIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data/drafts.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	d, err := s.Save(ctx, sample())
	if err != nil {
		t.Fatal(err)
	}
	if d.Revision != 1 {
		t.Fatal(d)
	}
	// Reopening reruns migrations without erasing existing edits.
	s.Close()
	s = openTest(t, path)
	got, err := s.Get(ctx, d.Key)
	if err != nil || got != d {
		t.Fatalf("restart: %#v %v", got, err)
	}
	for _, key := range []Key{{"catsoup-other", "es", "youtube", "CATsoup"}, {"catsoup-demo", "en", "youtube", "CATsoup"}, {"catsoup-demo", "es", "tiktok", "CATsoup"}, {"catsoup-demo", "es", "youtube", "other"}} {
		got, err = s.Get(ctx, key)
		if err != nil || got.Revision != 0 {
			t.Fatalf("leaked draft: %#v %v", got, err)
		}
	}
	d.Description = "Updated"
	updated, err := s.Save(ctx, d)
	if err != nil || updated.Revision != 2 {
		t.Fatalf("update: %v", err)
	}
	if _, err = s.Save(ctx, d); err != ErrConflict {
		t.Fatalf("stale update: %v", err)
	}
	if _, err = s.Save(ctx, sample()); err != ErrConflict {
		t.Fatalf("duplicate insert: %v", err)
	}
	list, err := s.List(ctx, d.Key.Episode, d.Key.Language)
	if err != nil || len(list) != 1 || list[0] != updated {
		t.Fatalf("list: %v %v", list, err)
	}
}
func TestConcurrentWritersDoNotLoseEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	a := openTest(t, path)
	b := openTest(t, path)
	ctx := context.Background()
	initial, err := a.Save(ctx, sample())
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, store := range []*Store{a, b} {
		wg.Add(1)
		go func(s *Store) { defer wg.Done(); _, err := s.Save(ctx, initial); results <- err }(store)
	}
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if err == ErrConflict {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflict)
	}
}
func TestRejectInvalidAndFutureDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s := openTest(t, path)
	ctx := context.Background()
	for _, mutate := range []func(*Draft){func(d *Draft) { d.Key.Episode = "../../secret" }, func(d *Draft) { d.Key.Platform = "unknown" }, func(d *Draft) { d.Key.Language = "fr" }, func(d *Draft) { d.Key.Account = " " }, func(d *Draft) { d.Description = strings.Repeat("x", 10001) }, func(d *Draft) { d.Title = "x\x00" }, func(d *Draft) { d.VideoHash = "invalid" }} {
		d := sample()
		mutate(&d)
		if _, err := s.Save(ctx, d); err == nil {
			t.Fatalf("accepted invalid draft: %+v", d)
		}
	}
	if _, err := s.db.Exec("UPDATE schema_version SET version=99"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	before, _ := os.ReadFile(path)
	if newer, err := Open(path); err == nil {
		newer.Close()
		t.Fatal("opened future database")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("modified unsupported database")
	}
	corrupt := filepath.Join(t.TempDir(), "broken.sqlite")
	os.WriteFile(corrupt, []byte("not a database"), 0600)
	if db, err := Open(corrupt); err == nil {
		db.Close()
		t.Fatal("accepted corruption")
	}
	after, _ = os.ReadFile(corrupt)
	if string(after) != "not a database" {
		t.Fatal("reset corrupt file")
	}
}
