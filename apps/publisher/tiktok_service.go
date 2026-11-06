package main

import (
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"context"
	"errors"
	"time"
)

func (s *UploadService) TikTokDirectEnabled() bool { return uploads.TikTokDirectEnabled() }
func (s *UploadService) SetTikTokDirectEnabled(enabled bool) error {
	return uploads.SetTikTokDirectEnabled(enabled)
}
func (s *UploadService) tikTokAccess(ctx context.Context, key drafts.Key, id string) (string, error) {
	if key.Platform != "tiktok" {
		return "", errors.New("select a TikTok post")
	}
	if err := drafts.CheckDestination(key, id); err != nil {
		return "", err
	}
	if !uploads.TikTokDirectEnabled() {
		return "", errors.New("enable Direct Post in TikTok developer settings first")
	}
	token, err := s.access(ctx, "tiktok", key.Language, id)
	if err != nil {
		return "", err
	}
	raw, err := s.store.Credential("tiktok", key.Language, id)
	if err != nil {
		return "", err
	}
	saved, err := uploads.Unseal(raw, "tiktok/"+key.Language+"/"+id)
	if err != nil || !saved.HasScope("video.publish") {
		return "", errors.New("reconnect this TikTok account to authorize video.publish")
	}
	return token, nil
}
func (s *UploadService) TikTokCreator(key drafts.Key, id string) (uploads.TikTokCreator, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	token, err := s.tikTokAccess(ctx, key, id)
	if err != nil {
		return uploads.TikTokCreator{}, err
	}
	return s.client.TikTokCreator(ctx, token)
}
func (s *UploadService) SendTikTok(key drafts.Key, revision int64, id string, confirmed bool, options uploads.TikTokOptions) (drafts.Receipt, error) {
	if key.Platform != "tiktok" {
		return drafts.Receipt{}, errors.New("select a TikTok post")
	}
	return s.sendMode(key, revision, id, "publish", confirmed, false, "", false, &options)
}

// OpenTikTokPolicy opens only the declarations required by TikTok's posting UI.
func (s *UploadService) OpenTikTokPolicy(policy string) error {
	switch policy {
	case "music":
		return s.openBrowser("https://www.tiktok.com/legal/page/global/music-usage-confirmation/en")
	case "branded":
		return s.openBrowser("https://www.tiktok.com/legal/page/global/bc-policy/en")
	default:
		return errors.New("unknown TikTok policy")
	}
}
