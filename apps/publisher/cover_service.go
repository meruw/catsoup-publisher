package main

import (
	"catsoup/publisher/internal/drafts"
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

func (s *DraftService) ImportCover(data string) (string, error) {
	if len(data) > 2_800_000 {
		return "", errors.New("choose a PNG or JPEG up to 2 MiB")
	}
	_, encoded, ok := strings.Cut(data, ",")
	if !ok || (!strings.HasPrefix(data, "data:image/png;base64,") && !strings.HasPrefix(data, "data:image/jpeg;base64,")) {
		return "", errors.New("choose a PNG or JPEG")
	}
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("invalid image encoding")
	}
	return s.store.ImportCover(context.Background(), b)
}
func (s *DraftService) CoverPreview(id string) (string, error) {
	b, mime, err := s.store.CoverImage(context.Background(), id)
	if err != nil {
		return "", err
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}
func (s *UploadService) CoverStatus(key drafts.Key) (string, error) {
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return "", err
	}
	return s.store.CoverResult(context.Background(), d), nil
}

// This action retries only the cover of a confirmed YouTube video, never its bytes.
func (s *UploadService) ApplyYouTubeCover(key drafts.Key, revision int64) (string, error) {
	if key.Platform != "youtube" || !strings.HasPrefix(key.Account, "account:") {
		return "", errors.New("select a connected YouTube account")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	d, err := s.store.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if d.Revision != revision || d.ApprovedRevision != revision || revision == 0 {
		return "", errors.New("approve the current post before applying its cover")
	}
	id := strings.TrimPrefix(key.Account, "account:")
	receipt, err := s.store.Receipt(key.Platform, id, d.VideoHash)
	if err != nil {
		return "", err
	}
	if receipt.State != "published" && receipt.State != "private" {
		return "", errors.New("a confirmed YouTube video is required")
	}
	token, err := s.access(ctx, key.Platform, key.Language, id)
	if err != nil {
		return "", err
	}
	account, err := s.client.Identify(ctx, key.Platform, token)
	if err != nil {
		return "", err
	}
	if account.ID != id {
		return "", errors.New("the connected account changed")
	}
	err = s.applyYouTubeCover(ctx, d, token, receipt.Reference)
	if err != nil {
		return "", err
	}
	return "Thumbnail applied", nil
}
func (s *UploadService) applyYouTubeCover(ctx context.Context, d drafts.Draft, token, reference string) error {
	if d.Cover == nil {
		return nil
	}
	b, mime, err := s.store.CoverImage(ctx, d.Cover.Hash)
	if err != nil {
		return err
	}
	if err = s.store.SaveCoverResult(ctx, d, "pending"); err != nil {
		return err
	}
	err = s.client.SetYouTubeCover(ctx, token, reference, b, mime)
	state := "applied"
	if err != nil {
		state = "needs_attention"
	}
	// Preserve a recoverable result even when the provider wait times out.
	if saveErr := s.store.SaveCoverResult(context.Background(), d, state); saveErr != nil {
		return saveErr
	}
	return err
}
