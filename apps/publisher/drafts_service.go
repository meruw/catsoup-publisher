package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"catsoup/publisher/internal/delivery"
	"catsoup/publisher/internal/drafts"
)

type SaveDraftInput struct {
	Key          drafts.Key `json:"key"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Hashtags     string     `json:"hashtags"`
	DeliveryMode string     `json:"deliveryMode"`
	Revision     int64      `json:"revision"`
	VideoID      string     `json:"videoId"`
	AttachVideo  bool       `json:"attachVideo"`
}
type DraftService struct {
	store            *drafts.Store
	reader           *delivery.Reader
	copyText         func(string) bool
	reveal           func(string) error
	enableCloseGuard func()
	confirmClose     func()
}

func (s *DraftService) ListDrafts(episode, language string) ([]drafts.Draft, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.List(ctx, episode, language)
}
func (s *DraftService) GetDraft(key drafts.Key) (drafts.Draft, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.Get(ctx, key)
}
func (s *DraftService) SaveDraft(input SaveDraftInput) (drafts.Draft, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d, err := s.store.Get(ctx, input.Key)
	if err != nil {
		return d, err
	}
	if d.Revision != input.Revision {
		return drafts.Draft{}, drafts.ErrConflict
	}
	d.Title = input.Title
	d.Description = input.Description
	d.Hashtags = input.Hashtags
	if input.DeliveryMode != "" {
		d.DeliveryMode = input.DeliveryMode
	}
	if d.Revision == 0 || input.AttachVideo {
		episode, err := s.reader.Inspect(input.Key.Episode)
		if err != nil {
			return drafts.Draft{}, err
		}
		found := false
		for _, video := range episode.Videos {
			if video.ID == input.VideoID && (video.Lang == nil || *video.Lang == input.Key.Language) {
				d.VideoID = video.ID
				d.VideoHash = video.SHA256
				d.DeliveryFingerprint = episode.Fingerprint
				found = true
				break
			}
		}
		if !found {
			return drafts.Draft{}, errors.New("choose a video matching the draft language")
		}
	}
	return s.store.Save(ctx, d)
}

// Copy reads saved metadata; no clipboard-reading or arbitrary native path bridge.
func (s *DraftService) CopyDraft(key drafts.Key, part string) error {
	d, err := s.GetDraft(key)
	if err != nil {
		return err
	}
	if d.Revision == 0 {
		return errors.New("save the draft first")
	}
	var text string
	switch part {
	case "title":
		text = d.Title
	case "caption":
		text = strings.TrimSpace(d.Description + "\n\n" + d.Hashtags)
	default:
		return errors.New("unsupported text selection")
	}
	if text == "" {
		return errors.New("nothing to copy")
	}
	if s.copyText == nil || !s.copyText(text) {
		return errors.New("clipboard unavailable; select and copy the text manually")
	}
	return nil
}
func (s *DraftService) OpenVideoFolder(episodeID, videoID string) error {
	path, err := s.reader.Media(episodeID, videoID)
	if err != nil {
		return err
	}
	if s.reveal == nil {
		return errors.New("file manager unavailable")
	}
	return s.reveal(path)
}

func (s *DraftService) CopyUploadPath(key drafts.Key, revision int64, accountConfirmed bool) error {
	if !accountConfirmed {
		return errors.New("confirm the account open on the platform")
	}
	d, err := s.GetDraft(key)
	if err != nil {
		return err
	}
	if d.Revision == 0 || d.Revision != revision {
		return errors.New("open the latest saved draft before preparing an upload")
	}
	path, err := s.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash)
	if err != nil {
		return err
	}
	if err = s.store.RequireReview(context.Background(), d); err != nil {
		return err
	}
	if s.copyText == nil || !s.copyText(path) {
		return errors.New("clipboard unavailable; try again")
	}
	return nil
}

func (s *DraftService) EnableCloseGuard() {
	if s.enableCloseGuard != nil {
		s.enableCloseGuard()
	}
}
func (s *DraftService) ConfirmClose() {
	if s.confirmClose != nil {
		s.confirmClose()
	}
}
