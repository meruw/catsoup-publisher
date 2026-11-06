package main

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestCopyResultsIncludesWholeAttemptAndPreservesClipboardOnEmpty(t *testing.T) {
	store := diagnostics.Open(t.TempDir())
	defer store.Close()
	ctx := diagnostics.Start(context.Background(), store, "instagram", "es")
	for i := 0; i < 250; i++ {
		diagnostics.Record(ctx, slog.LevelInfo, "pages_received", diagnostics.Details{Pages: i})
	}
	other := diagnostics.Start(context.Background(), store, "youtube", "en")
	diagnostics.Record(other, slog.LevelInfo, "connection_finished", diagnostics.Details{})
	page, _ := store.Query(diagnostics.Query{Service: "instagram"})
	clipboard := "unchanged"
	svc := DiagnosticsService{store: store, copyText: func(text string) bool { clipboard = text; return true }}
	count, err := svc.CopyResults(diagnostics.Query{Attempt: page.Events[0].Attempt, Before: page.Events[199].ID}, 0)
	if err != nil || count != 251 {
		t.Fatalf("whole attempt: %d %v", count, err)
	}
	lines := strings.Split(strings.TrimSpace(clipboard), "\n")
	var last int64
	for _, line := range lines {
		var event diagnostics.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event.ID <= last || event.Service != "instagram" {
			t.Fatal("incorrect order or filter")
		}
		last = event.ID
	}
	previous := clipboard
	count, err = svc.CopyResults(diagnostics.Query{Search: "no-match"}, 1)
	if err != nil || count != 0 || clipboard != previous {
		t.Fatal("empty copy replaced clipboard")
	}
	svc.copyText = func(string) bool { return false }
	if _, err := svc.CopyResults(diagnostics.Query{}, 0); err == nil {
		t.Fatal("clipboard failure ignored")
	}
}
