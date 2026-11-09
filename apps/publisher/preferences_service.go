package main

import (
	"catsoup/publisher/internal/drafts"
	"context"
	"time"
)

func (s *DraftService) GetPublishingDefaults() ([]drafts.PublishingDefault, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.PublishingDefaults(ctx)
}
func (s *DraftService) SetPublishingDefault(p drafts.PublishingDefault) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.SetPublishingDefault(ctx, p)
}

func (s *DraftService) PostStatuses(filter drafts.PostFilter) (drafts.PostPage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.PostStatuses(ctx, filter)
}

func (s *DraftService) PostGroups(filter drafts.PostFilter) (drafts.PostGroupPage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.PostGroups(ctx, filter)
}

// CompletedLanguages reads local receipts only; it never publishes or hides data.
func (s *DraftService) CompletedLanguages() ([]drafts.Completion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.CompletedLanguages(ctx)
}
