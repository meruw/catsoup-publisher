package main

import (
	"catsoup/publisher/internal/drafts"
	"context"
	"errors"
	"time"
)

func (s *DraftService) GetReview(episode, language, video string) (drafts.Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.store.GetReview(ctx, episode, language, video)
}

func (s *DraftService) SaveReview(input drafts.Review, confirmed bool) (drafts.Review, error) {
	if input.State == "reviewed" && !confirmed {
		return drafts.Review{}, errors.New("confirm that you watched and listened to this version")
	}
	ep, err := s.reader.Inspect(input.Episode)
	if err != nil {
		return drafts.Review{}, err
	}
	if ep.Fingerprint != input.Fingerprint {
		return drafts.Review{}, errors.New("delivery changed; refresh before recording a review")
	}
	for _, v := range ep.Videos {
		if v.ID == input.VideoID && v.Lang != nil && *v.Lang == input.Language && v.Kind == "final" && v.SHA256 == input.VideoHash {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return s.store.SaveReview(ctx, input)
		}
	}
	return drafts.Review{}, errors.New("select the final video for this language")
}
