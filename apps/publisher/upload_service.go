package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"catsoup/publisher/internal/delivery"
	"catsoup/publisher/internal/diagnostics"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
)

type UploadService struct {
	diagnostics *diagnostics.Store
	store       *drafts.Store
	reader      *delivery.Reader
	client      *uploads.Client
	openBrowser func(string) error
}

func (s *UploadService) OpenXComposer(key drafts.Key, revision int64) error {
	if key.Platform != "x" {
		return errors.New("select an X draft")
	}
	if strings.HasPrefix(key.Account, "account:") {
		if _, err := s.store.Credential(key.Platform, key.Language, strings.TrimPrefix(key.Account, "account:")); err != nil {
			return err
		}
	}
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if d.Revision == 0 || revision != d.Revision {
		return drafts.ErrConflict
	}
	if _, err = s.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash); err != nil {
		return err
	}
	if err = s.store.RequireReview(context.Background(), d); err != nil {
		return err
	}
	text := strings.TrimSpace(d.Title + "\n\n" + d.Description + "\n\n" + d.Hashtags)
	return s.openBrowser("https://x.com/intent/tweet?" + url.Values{"text": {text}}.Encode())
}
func (s *UploadService) Receipt(key drafts.Key, id string) (drafts.Receipt, error) {
	if err := drafts.CheckDestination(key, id); err != nil {
		return drafts.Receipt{}, err
	}
	if _, err := s.store.Credential(key.Platform, key.Language, id); err != nil {
		return drafts.Receipt{}, err
	}
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return drafts.Receipt{}, err
	}

	if key.Platform == "instagram" {
		if err = s.store.RecoverStaleInstagram(id, d.VideoHash); err != nil {
			return drafts.Receipt{}, err
		}
	}
	return s.store.Receipt(key.Platform, id, d.VideoHash)
}
func (s *UploadService) ClearUncertain(key drafts.Key, id, expectedHash string, checked bool) error {
	if err := drafts.CheckDestination(key, id); err != nil {
		return err
	}
	if !checked {
		return errors.New("check the platform before allowing another attempt")
	}
	if _, err := s.store.Credential(key.Platform, key.Language, id); err != nil {
		return err
	}
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if d.VideoHash != expectedHash {
		return errors.New("linked video changed; reopen the draft")
	}
	return s.store.ResetAttempt(d, id)
}

// Stage only bytes matching the approved hash so an external edit cannot change
// the media while it is streaming. The temporary snapshot is removed on return.
func stage(path, hash string) (*os.File, int64, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, 0, errors.New("could not open approved video")
	}
	defer src.Close()
	file, err := os.CreateTemp("", "catsoup-upload-*.mp4")
	if err != nil {
		return nil, 0, errors.New("could not stage video")
	}
	digest := sha256.New()
	size, err := io.Copy(io.MultiWriter(file, digest), src)
	if err != nil || hex.EncodeToString(digest.Sum(nil)) != hash {
		file.Close()
		os.Remove(file.Name())
		return nil, 0, errors.New("could not stage the exact approved video; refresh the delivery")
	}
	return file, size, nil
}
func (s *UploadService) Send(key drafts.Key, revision int64, id string, confirmed bool) (drafts.Receipt, error) {
	mode := "publish"
	if key.Platform == "youtube" || key.Platform == "tiktok" {
		mode = "draft"
	}
	return s.SendMode(key, revision, id, mode, confirmed, false, "", false)
}
func (s *UploadService) SendX(key drafts.Key, revision int64, id string, confirmed, paidAccepted bool) (drafts.Receipt, error) {
	if key.Platform != "x" {
		return drafts.Receipt{}, errors.New("select an X draft")
	}
	return s.SendMode(key, revision, id, "publish", confirmed, paidAccepted, "", false)
}

// SendMode accepts only capabilities actually implemented by each connector.
func (s *UploadService) SendMode(key drafts.Key, revision int64, id, mode string, confirmed, paidAccepted bool, audience string, synthetic bool) (drafts.Receipt, error) {
	return s.sendMode(key, revision, id, mode, confirmed, paidAccepted, audience, synthetic, nil)
}
func (s *UploadService) sendMode(key drafts.Key, revision int64, id, mode string, confirmed, paidAccepted bool, audience string, synthetic bool, tikOptions *uploads.TikTokOptions) (receipt drafts.Receipt, failure error) {
	if err := drafts.CheckDestination(key, id); err != nil {
		return drafts.Receipt{}, err
	}
	if mode != "draft" && mode != "publish" {
		return drafts.Receipt{}, errors.New("choose a delivery mode")
	}
	if mode == "draft" && key.Platform != "youtube" && key.Platform != "tiktok" {
		return drafts.Receipt{}, errors.New("this platform supports local drafts only")
	}
	if mode == "publish" && key.Platform == "tiktok" && tikOptions == nil {
		return drafts.Receipt{}, errors.New("TikTok Direct Post is not enabled; send to the inbox instead")
	}
	if key.Platform == "x" && !paidAccepted {
		return drafts.Receipt{}, errors.New("acknowledge X API usage charges before publishing")
	}
	if !confirmed {
		return drafts.Receipt{}, errors.New("confirm the destination and upload action")
	}
	credential, err := s.store.Credential(key.Platform, key.Language, id)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if key.Platform == "x" && len(credential) == 0 {
		return drafts.Receipt{}, errors.New("connect X in API mode in Accounts before publishing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	ctx = diagnostics.StartTransfer(ctx, s.diagnostics, key.Platform, key.Language)
	defer func() {
		if failure != nil {
			diagnostics.Record(ctx, slog.LevelError, "transfer_failed", diagnostics.Details{})
		} else {
			diagnostics.Record(ctx, slog.LevelInfo, "transfer_finished", diagnostics.Details{})
		}
	}()
	d, err := s.store.Get(ctx, key)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if d.Revision == 0 || revision != d.Revision {
		return drafts.Receipt{}, drafts.ErrConflict
	}
	if strings.HasPrefix(key.Account, "account:") && d.DeliveryMode != mode {
		return drafts.Receipt{}, errors.New("delivery mode changed; reopen the publishing confirmation")
	}
	path, err := s.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if err = s.store.RequireReview(ctx, d); err != nil {
		return drafts.Receipt{}, err
	}
	file, size, err := stage(path, d.VideoHash)
	if err != nil {
		return drafts.Receipt{}, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	uploadHash := d.VideoHash
	caption := strings.TrimSpace(d.Description + "\n\n" + d.Hashtags)
	if (key.Platform == "instagram" || key.Platform == "x" || (key.Platform == "tiktok" && mode == "publish")) && strings.TrimSpace(d.Title) != "" {
		caption = strings.TrimSpace(d.Title + "\n\n" + caption)
	}
	m := uploads.Media{Title: d.Title, Caption: caption, Language: key.Language, File: file, Size: size, Mode: mode, Audience: audience, Synthetic: synthetic}
	if d.Cover != nil {
		if e := drafts.ValidateCover(key.Platform, d.Cover); e != nil {
			return drafts.Receipt{}, e
		}
		if _, _, e := s.store.CoverImage(ctx, d.Cover.Hash); e != nil {
			return drafts.Receipt{}, e
		}
		if d.Cover.Kind == "frame" {
			value := d.Cover.TimeMS
			m.CoverTimeMS = &value
		}
	}
	if err = uploads.Validate(key.Platform, m); err != nil {
		return drafts.Receipt{}, err
	}
	var token string
	if key.Platform == "tiktok" && mode == "publish" {
		token, err = s.tikTokAccess(ctx, key, id)
		if err != nil {
			return drafts.Receipt{}, err
		}
		ep, e := s.reader.Inspect(key.Episode)
		if e != nil {
			return drafts.Receipt{}, e
		}
		for _, v := range ep.Videos {
			if v.ID == d.VideoID {
				m.Duration = v.Measured.Duration
			}
		}
		m.TikTok = tikOptions
		creator, e := s.client.TikTokCreator(ctx, token)
		if e != nil {
			return drafts.Receipt{}, e
		}
		if e = uploads.ValidateTikTokOptions(*tikOptions, creator, m.Duration, m.Caption); e != nil {
			return drafts.Receipt{}, e
		}
	} else {
		token, err = s.access(ctx, key.Platform, key.Language, id)
	}
	if err != nil {
		return drafts.Receipt{}, err
	}
	current, err := s.client.Identify(ctx, key.Platform, token)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if current.ID != id {
		return drafts.Receipt{}, errors.New("the authorized account changed; reconnect before sending")
	}
	if key.Platform == "instagram" {
		if err = s.store.RecoverStaleInstagram(id, d.VideoHash); err != nil {
			return drafts.Receipt{}, err
		}
	}
	previous, err := s.store.Receipt(key.Platform, id, d.VideoHash)
	if err != nil {
		return drafts.Receipt{}, err
	}
	resumeBytes := false
	if key.Platform == "instagram" && (previous.State == "failed" || previous.State == "check") {
		state, stateErr := s.client.InstagramContainerState(ctx, token, previous.Reference)
		if stateErr == nil && state == "FINISHED" {
			if err = s.store.ReserveInstagramResume(d, id, previous.Reference); err != nil {
				return drafts.Receipt{}, err
			}
			ref, publishErr := s.client.PublishReadyInstagram(ctx, token, id, previous.Reference)
			state = "published"
			if publishErr != nil {
				state = "check_publish"
			}
			if err = s.store.SaveReceipt(d, id, state, ref); err != nil {
				return drafts.Receipt{}, errors.New("receipt write failed; check Instagram before retrying")
			}
			if publishErr != nil {
				return drafts.Receipt{}, publishErr
			}
			return s.store.Receipt(key.Platform, id, d.VideoHash)
		}
		resumeBytes = uploads.CanResumeInstagramBytes(state, stateErr)
		if !resumeBytes {
			if err = s.client.ConfirmInstagramFailure(ctx, token, previous.Reference); err != nil {
				return drafts.Receipt{}, err
			}
		}
	}
	if key.Platform == "instagram" {
		if previous.State != "" && previous.State != "failed" && previous.State != "check" {
			return drafts.Receipt{}, errors.New("an Instagram upload already exists; check its receipt")
		}
		diagnostics.Record(ctx, slog.LevelInfo, "upload_preparing", diagnostics.Details{})
		prepared, preparedSize, hash, cleanup, prepareErr := s.reader.PrepareInstagramAudio(ctx, file.Name())
		if prepareErr != nil {
			return drafts.Receipt{}, prepareErr
		}
		defer cleanup()
		m.File, m.Size, uploadHash = prepared, preparedSize, hash
		diagnostics.Record(ctx, slog.LevelInfo, "upload_prepared", diagnostics.Details{})
	}
	if resumeBytes {
		if err = s.store.ReserveInstagramRetransfer(d, id, previous.Reference, uploadHash); err != nil {
			return drafts.Receipt{}, err
		}
		if err = s.client.ResumeInstagramBytes(ctx, token, previous.Reference, m); err != nil {
			if saveErr := s.store.SaveReceipt(d, id, "check", ""); saveErr != nil {
				return drafts.Receipt{}, errors.New("recovery receipt write failed; check Instagram before retrying")
			}
			return drafts.Receipt{}, err
		}
		// Persist the ambiguous-publish boundary before sending media_publish.
		if err = s.store.SaveReceipt(d, id, "publishing", ""); err != nil {
			return drafts.Receipt{}, err
		}
		ref, publishErr := s.client.PublishReadyInstagram(ctx, token, id, previous.Reference)
		state := "published"
		if publishErr != nil {
			state = "check_publish"
		}
		if err = s.store.SaveReceipt(d, id, state, ref); err != nil {
			return drafts.Receipt{}, errors.New("receipt write failed; check Instagram before retrying")
		}
		if publishErr != nil {
			return drafts.Receipt{}, publishErr
		}
		return s.store.Receipt(key.Platform, id, d.VideoHash)
	}
	if key.Platform == "instagram" && (previous.State == "failed" || previous.State == "check") {
		if err = s.client.ConfirmInstagramFailure(ctx, token, previous.Reference); err != nil {
			return drafts.Receipt{}, err
		}
		err = s.store.ReserveInstagramRetry(d, id, previous.Reference)
	} else {
		err = s.store.Reserve(d, id)
	}
	if err != nil {
		return drafts.Receipt{}, err
	}
	ref, sendErr := s.client.Send(ctx, key.Platform, token, current, m, func(ref string) error {
		if key.Platform == "instagram" {
			return s.store.SaveInstagramCheckpoint(d, id, ref, uploadHash)
		}
		return s.store.SaveReceipt(d, id, "sending", ref)
	})
	state := "published"
	if key.Platform == "youtube" && mode == "draft" {
		state = "private"
	}
	if key.Platform == "tiktok" {
		state = "inbox"
		if mode == "publish" {
			state = "published"
			if tikOptions.Privacy == "SELF_ONLY" {
				state = "private"
			} else if tikOptions.Privacy != "PUBLIC_TO_EVERYONE" {
				state = "restricted"
			}
		}
	}
	if sendErr != nil {
		state = "check"
		if errors.Is(sendErr, uploads.ErrProcessingFailed) {
			state = "failed"
		}
	}
	if err = s.store.SaveReceipt(d, id, state, ref); err != nil {
		return drafts.Receipt{}, errors.New("receipt write failed; check the platform before retrying")
	}
	if sendErr != nil {
		return drafts.Receipt{}, sendErr
	}
	if key.Platform == "youtube" && d.Cover != nil {
		if e := s.applyYouTubeCover(ctx, d, token, ref); e != nil {
			diagnostics.Record(ctx, slog.LevelWarn, "cover_pending", diagnostics.Details{Reason: "thumbnail_not_confirmed"})
		}
	}
	return s.store.Receipt(key.Platform, id, d.VideoHash)
}
