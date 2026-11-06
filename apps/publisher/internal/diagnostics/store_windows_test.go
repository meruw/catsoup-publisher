package diagnostics

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHistoryFiltersPaginationAndRestart(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	es := Start(context.Background(), s, "instagram", "es")
	en := Start(context.Background(), s, "youtube", "en")
	for i := 0; i < 230; i++ {
		Record(es, slog.LevelDebug, "pages_received", Details{Pages: i})
	}
	Record(es, slog.LevelError, "provider_failed", Details{Reason: "provider_rejected", HTTPStatus: 403, ProviderCode: 200})
	Record(en, slog.LevelInfo, "connection_finished", Details{})
	page, err := s.Query(Query{Service: "instagram"})
	if err != nil || len(page.Events) != 200 || !page.More {
		t.Fatal("first page", len(page.Events), err)
	}
	older, err := s.Query(Query{Service: "instagram", Before: page.Events[199].ID})
	if err != nil || len(older.Events) != 32 || older.More {
		t.Fatal("older page", len(older.Events), err)
	}
	result, _ := s.Query(Query{Search: "403", Language: "es", Level: "ERROR"})
	if len(result.Events) != 1 || result.Events[0].ProviderCode != 200 {
		t.Fatal("numeric search and filters", result)
	}
	s.Close()
	// Keep backup filenames distinct when repeatedly reopening in a test.
	time.Sleep(2 * time.Millisecond)
	s = Open(dir)
	defer s.Close()
	result, _ = s.Query(Query{Search: "403"})
	if len(result.Events) != 1 {
		t.Fatal("history lost after restart", result)
	}
	current, _ := s.Query(Query{CurrentSession: true})
	if len(current.Events) != 0 {
		t.Fatal("old records leaked into current session")
	}
	Record(Start(context.Background(), s, "instagram", "en"), slog.LevelInfo, "connection_finished", Details{})
	current, _ = s.Query(Query{CurrentSession: true})
	if len(current.Events) != 2 || current.Events[0].ID <= result.Events[0].ID {
		t.Fatal("sequence did not survive restart")
	}
}

func TestBoundedMemoryAndExclusiveWriter(t *testing.T) {
	dir := t.TempDir()
	first := Open(dir)
	defer first.Close()
	second := Open(dir)
	defer second.Close()
	ctx := Start(context.Background(), second, "tiktok", "en")
	for i := 0; i < 2100; i++ {
		Record(ctx, slog.LevelInfo, "browser_opening", Details{})
	}
	if len(second.recent) != memoryLimit || second.writer != nil {
		t.Fatal("second process writer or memory limit")
	}
	result, _ := second.Query(Query{})
	if result.Warning == "" || len(result.Events) != 200 {
		t.Fatal("missing memory-only feedback")
	}
	if len(first.recent) != 0 {
		t.Fatal("instances shared events")
	}
}

func TestPrivacyAndRetention(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	ctx := Start(context.Background(), s, "instagram", "es")
	Record(ctx, slog.LevelError, "provider_failed", Details{Reason: "token-SUPER-SECRET", HTTPStatus: 401})
	Record(ctx, slog.LevelError, "secret_stage_SUPER_SECRET", Details{})
	s.Close()
	raw, err := os.ReadFile(filepath.Join(dir, "publisher.jsonl"))
	if err != nil || strings.Contains(string(raw), "SUPER") {
		t.Fatal("unexpected secret or IO error", err)
	}
	old := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(filepath.Join(dir, "publisher.jsonl"), old, old); err != nil {
		t.Fatal(err)
	}
	s = Open(dir)
	defer s.Close()
	result, _ := s.Query(Query{})
	if len(result.Events) != 0 {
		t.Fatal("expired events retained")
	}
	raw, _ = os.ReadFile(filepath.Join(dir, "publisher.jsonl"))
	if strings.Contains(string(raw), "provider_failed") {
		t.Fatal("expired active file not cleared")
	}
}

func TestRotationAndInterruptedTail(t *testing.T) {
	dir := t.TempDir()
	s := Open(dir)
	ctx := Start(context.Background(), s, "instagram", "es")
	Record(ctx, slog.LevelInfo, "connection_finished", Details{})
	if err := s.writer.Rotate(); err != nil {
		t.Fatal(err)
	}
	Record(ctx, slog.LevelDebug, "pages_received", Details{Pages: 1})
	s.Close()
	f, err := os.OpenFile(filepath.Join(dir, "publisher.jsonl"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("{incomplete")
	_ = f.Close()
	time.Sleep(2 * time.Millisecond)
	s = Open(dir)
	defer s.Close()
	result, _ := s.Query(Query{})
	if len(result.Events) != 3 {
		t.Fatal("rotation or truncated tail lost valid history", len(result.Events))
	}
	body, _ := json.Marshal(result)
	if strings.Contains(string(body), "incomplete") {
		t.Fatal("corrupt tail exposed")
	}
}

func TestRejectInvalidFiltersAndUntrustedEventFields(t *testing.T) {
	s := Open(t.TempDir())
	defer s.Close()
	for _, q := range []Query{{Search: strings.Repeat("x", 201)}, {Before: -1}, {Before: 9007199254740992}, {Level: "SECRET"}, {Service: "other"}, {Language: "other"}} {
		if _, err := s.Query(q); err == nil {
			t.Fatalf("accepted invalid filter: %+v", q)
		}
	}
	ctx := Start(context.Background(), s, "SECRET", "es")
	Record(ctx, slog.LevelError, "connection_failed", Details{})
	if len(s.recent) != 0 {
		t.Fatal("invalid service was logged")
	}
	ctx = Start(context.Background(), s, "instagram", "es")
	Record(ctx, slog.LevelError, "provider_failed", Details{})
	e := s.recent[len(s.recent)-1]
	e.Reason = "token_SECRET"
	if validEvent(e) {
		t.Fatal("unrecognized disk reason accepted")
	}
}
