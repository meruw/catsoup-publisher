package main

import (
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"context"
	"errors"
	"strings"
	"time"
)

func (s *UploadService) PublishYouTube(key drafts.Key, revision int64, id string, confirmed bool, audience string, synthetic bool) (drafts.Receipt, error) {
	if key.Platform != "youtube" || !confirmed {
		return drafts.Receipt{}, errors.New("confirm publishing the private YouTube video")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if _, err := s.store.Credential(key.Platform, key.Language, id); err != nil {
		return drafts.Receipt{}, err
	}
	d, err := s.store.Get(ctx, key)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if d.Revision == 0 || d.Revision != revision {
		return drafts.Receipt{}, drafts.ErrConflict
	}
	if err = s.store.RequireReview(ctx, d); err != nil {
		return drafts.Receipt{}, err
	}
	if _, err = s.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash); err != nil {
		return drafts.Receipt{}, err
	}
	m := uploads.Media{Title: d.Title, Caption: strings.TrimSpace(d.Description + "\n\n" + d.Hashtags), Language: key.Language, Mode: "publish", Audience: audience, Synthetic: synthetic, Size: 1}
	if err = uploads.Validate("youtube", m); err != nil {
		return drafts.Receipt{}, err
	}
	receipt, err := s.store.Receipt("youtube", id, d.VideoHash)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if receipt.Reference == "" {
		return drafts.Receipt{}, errors.New("no private YouTube upload was recorded")
	}
	token, err := s.access(ctx, "youtube", key.Language, id)
	if err != nil {
		return drafts.Receipt{}, err
	}
	account, err := s.client.Identify(ctx, "youtube", token)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if account.ID != id {
		return drafts.Receipt{}, errors.New("authorized channel changed; reconnect")
	}
	if err = s.store.ReservePromotion(d, id, receipt.Reference); err != nil {
		return drafts.Receipt{}, err
	}
	_, sendErr := s.client.PublishYouTube(ctx, token, id, receipt.Reference, m)
	state := "published"
	if sendErr != nil {
		state = "check_publish"
	}
	if err = s.store.SaveReceipt(d, id, state, receipt.Reference); err != nil {
		return drafts.Receipt{}, errors.New("receipt write failed; check YouTube Studio")
	}
	if sendErr != nil {
		return drafts.Receipt{}, errors.New("could not confirm public visibility; check YouTube Studio and reconnect YouTube if its permissions are missing: " + sendErr.Error())
	}
	return s.store.Receipt("youtube", id, d.VideoHash)
}
