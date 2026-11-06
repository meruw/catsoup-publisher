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
