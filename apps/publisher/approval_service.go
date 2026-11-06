package main

import (
	"catsoup/publisher/internal/drafts"
	"context"
	"errors"
	"strings"
	"time"
)

// Approval confirms the post in Publisher; it never modifies production approval.
func (s *DraftService) SetPostApproval(key drafts.Key, revision int64, approved bool) (drafts.Draft, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	d, err := s.store.Get(ctx, key)
	if err != nil {
		return drafts.Draft{}, err
	}
	if d.Revision != revision || revision < 1 {
		return drafts.Draft{}, drafts.ErrConflict
	}
	if approved {
		if strings.TrimSpace(d.Title+d.Description+d.Hashtags) == "" {
			return drafts.Draft{}, errors.New("add post text before approving")
		}
		ep, err := s.reader.Inspect(key.Episode)
		if err != nil {
			return drafts.Draft{}, err
		}
		if ep.Fingerprint != d.DeliveryFingerprint {
			return drafts.Draft{}, errors.New("delivery changed; refresh and link the current video")
		}
		found := false
		for _, v := range ep.Videos {
			if v.ID == d.VideoID && v.Kind == "final" && v.Lang != nil && *v.Lang == key.Language && v.SHA256 == d.VideoHash {
				found = true
				break
			}
		}
		if !found {
			return drafts.Draft{}, errors.New("select the current final video for this language")
		}
	}
	return s.store.SetPostApproval(ctx, key, revision, approved)
}
