package main

import (
	"context"
	"strings"
	"time"

	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/statistics"
	"catsoup/publisher/internal/uploads"
)

// PostStatistics accepts a local draft key, never a remote URL or access token.
// It performs no uploads and never changes publication receipts or approvals.
func (s *UploadService) PostStatistics(key drafts.Key) (uploads.Statistics, error) {
	out := uploads.Statistics{Items: []uploads.Metric{}}
	id := strings.TrimPrefix(key.Account, "account:")
	if err := drafts.CheckDestination(key, id); err != nil {
		return out, err
	}
	switch key.Platform {
	case "tiktok":
		out.Note = "TikTok statistics need video.list authorization and a published video ID. Inbox receipts do not identify a published post. This connector does not read TikTok statistics yet."
		return out, nil
	case "x":
		out.Note = "X statistics are not requested: its API can incur charges. Check the post in X."
		return out, nil
	}
	r, err := s.Receipt(key, id)
	if err != nil {
		return out, err
	}
	if r.Reference == "" || (r.State != "published" && r.State != "private" && r.State != "sent") {
		out.Note = "Statistics are available after a confirmed upload. No published video is recorded for this file."
		return out, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	token, err := s.access(ctx, key.Platform, key.Language, id)
	if err != nil {
		return out, err
	}
	out, err = s.client.Statistics(ctx, key.Platform, token, id, r.Reference)
	if err != nil {
		return out, err
	}
	// Persistence gets its own timeout even when the provider used most of its budget.
	saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer saveCancel()
	out.History, err = s.store.RecordStatistics(saveCtx, key.Platform, id, r.Reference, statistics.Sample{Items: out.Items, UpdatedAt: out.UpdatedAt})
	if err != nil {
		out.HistoryNote = "Current counts loaded, but local history could not be saved."
	}
	return out, nil
}
