package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

func (c *Client) tiktok(ctx context.Context, token string, m Media, record func(string) error) (string, error) {
	chunk := min(m.Size, int64(16<<20))
	count := m.Size / chunk
	// A whole-file transfer must declare its actual size, including the tail
	// when the file is between one and two preferred chunks.
	if count == 1 {
		chunk = m.Size
	}
	src := map[string]any{"source": "FILE_UPLOAD", "video_size": m.Size, "chunk_size": chunk, "total_chunk_count": count}
	var out struct {
		Data struct {
			ID  string `json:"publish_id"`
			URL string `json:"upload_url"`
		}
		Error struct{ Code string }
	}
	auth := "Bearer " + token
	endpoint := tik + "post/publish/inbox/video/init/"
	body := map[string]any{"source_info": src}
	if m.Mode == "publish" {
		if m.TikTok == nil {
			return "", errors.New("TikTok posting options are required")
		}
		creator, err := c.TikTokCreator(ctx, token)
		if err != nil {
			return "", err
		}
		if err = ValidateTikTokOptions(*m.TikTok, creator, m.Duration, m.Caption); err != nil {
			return "", err
		}
		o := m.TikTok
		endpoint = tik + "post/publish/video/init/"
		body["post_info"] = map[string]any{"title": m.Caption, "privacy_level": o.Privacy, "disable_comment": !o.Comments, "disable_duet": !o.Duet, "disable_stitch": !o.Stitch, "brand_content_toggle": o.Branded, "brand_organic_toggle": o.OwnBrand, "is_aigc": o.Synthetic}
		if m.CoverTimeMS != nil {
			body["post_info"].(map[string]any)["video_cover_timestamp_ms"] = *m.CoverTimeMS
		}
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_started", diagnostics.Details{})
	_, err := c.request(ctx, "POST", endpoint, auth, jsonHeaders, jsonBody(body), -1, &out)
	if err != nil {
		return "", err
	}
	if err = tikTokSession(ctx, out.Error.Code, out.Data.ID, out.Data.URL); err != nil {
		return "", err
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_finished", diagnostics.Details{})
	if err = record(out.Data.ID); err != nil {
		return "", err
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_started", diagnostics.Details{})
	for i := int64(0); i < count; i++ {
		start := i * chunk
		length := chunk
		if i == count-1 {
			length = m.Size - start
		}
		_, err = c.request(ctx, "PUT", out.Data.URL, "", map[string]string{"Content-Type": "video/mp4", "Content-Range": fmt.Sprintf("bytes %d-%d/%d", start, start+length-1, m.Size)}, io.NewSectionReader(m.File, start, length), length, nil)
		if err != nil {
			return "", err
		}
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_finished", diagnostics.Details{})
	diagnostics.Record(ctx, slog.LevelInfo, "upload_processing", diagnostics.Details{})
	err = waitFor(ctx, func() (bool, error) {
		var outStatus struct {
			Data  struct{ Status string }
			Error struct{ Code string }
		}
		_, err := c.request(ctx, "POST", tik+"post/publish/status/fetch/", auth, jsonHeaders, jsonBody(map[string]string{"publish_id": out.Data.ID}), -1, &outStatus)
		if err != nil {
			return false, err
		}
		if outStatus.Data.Status == "FAILED" && outStatus.Error.Code == "ok" {
			return false, ErrProcessingFailed
		}
		if outStatus.Error.Code != "ok" {
			return false, errors.New("TikTok could not confirm the transfer status")
		}
		return (m.Mode != "publish" && outStatus.Data.Status == "SEND_TO_USER_INBOX") || outStatus.Data.Status == "PUBLISH_COMPLETE", nil
	}, 5*time.Second)
	return out.Data.ID, err
}

func (c *Client) instagram(ctx context.Context, token string, a Account, m Media, record func(string) error) (string, error) {
	return c.instagramWithRetry(ctx, token, a, m, record, 15*time.Second)
}

func (c *Client) instagramAttempt(ctx context.Context, token string, a Account, m Media, record func(string) error) (string, error) {
	auth := "Bearer " + token
	var container struct{ ID, URI string }
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_started", diagnostics.Details{})
	body := map[string]any{"media_type": "REELS", "upload_type": "resumable", "caption": m.Caption}
	if m.CoverTimeMS != nil {
		body["thumb_offset"] = *m.CoverTimeMS
	}
	_, err := c.request(ctx, "POST", graph+a.ID+"/media", auth, jsonHeaders, jsonBody(body), -1, &container)
	if err != nil {
		return "", err
	}
	if !safeID.MatchString(container.ID) || !safeSession(container.URI, "rupload.facebook.com") {
		return "", errors.New("Meta did not return a supported local Reel upload session")
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_finished", diagnostics.Details{})
	if err = record(container.ID); err != nil {
		return "", err
	}
	var accepted struct{ Success bool }
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_started", diagnostics.Details{})
	_, err = c.request(ctx, "POST", container.URI, "OAuth "+token, map[string]string{"Content-Type": "application/octet-stream", "offset": "0", "file_size": fmt.Sprint(m.Size)}, io.NewSectionReader(m.File, 0, m.Size), m.Size, &accepted)
	if err != nil || !accepted.Success {
		transferErr := err
		if transferErr == nil {
			transferErr = errors.New("Meta declined the video transfer")
		}
		// The byte endpoint can report an error even though Meta has received
		// and processed this exact container. Reconcile with a read, never resend.
		diagnostics.Record(ctx, slog.LevelInfo, "upload_status_check", diagnostics.Details{})
		state, statusErr := c.instagramContainerStatus(ctx, auth, container.ID)
		if CanResumeInstagramBytes(state, statusErr) {
			return "", statusErr
		}
		if pendingInstagramFailure(state, statusErr) {
			// Meta may retain a phase error while the container later succeeds.
			statusErr = nil
		}
		if errors.Is(statusErr, ErrProcessingFailed) {
			return "", statusErr
		}
		if statusErr != nil || (state != "FINISHED" && state != "IN_PROGRESS") {
			return "", transferErr
		}
		if state == "FINISHED" {
			diagnostics.Record(ctx, slog.LevelWarn, "upload_reconciled", diagnostics.Details{})
		}
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_finished", diagnostics.Details{})
	diagnostics.Record(ctx, slog.LevelInfo, "upload_processing", diagnostics.Details{})
	err = c.waitInstagramContainer(ctx, auth, container.ID, time.Minute)
	if err != nil {
		return "", err
	}
	return c.publishInstagramContainer(ctx, auth, a.ID, container.ID)
}

func (c *Client) publishInstagramContainer(ctx context.Context, auth, account, container string) (string, error) {
	var result struct{ ID string }
	diagnostics.Record(ctx, slog.LevelInfo, "publish_started", diagnostics.Details{})
	_, err := c.request(ctx, "POST", graph+account+"/media_publish", auth, jsonHeaders, jsonBody(map[string]string{"creation_id": container}), -1, &result)
	if err != nil {
		return "", err
	}
	if !safeID.MatchString(result.ID) {
		return "", errors.New("Instagram did not return a media ID")
	}
	return result.ID, nil
}

func (c *Client) waitInstagramContainer(ctx context.Context, auth, id string, interval time.Duration) error {
	// Keep polling bounded even if a caller has no deadline. An uncertain
	// container is retained for checking; it is never replaced or re-uploaded.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	err := waitFor(ctx, func() (bool, error) {
		state, err := c.instagramContainerStatus(ctx, auth, id)
		if pendingInstagramFailure(state, err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		switch state {
		case "FINISHED":
			return true, nil
		case "IN_PROGRESS":
			return false, nil
		default:
			return false, errors.New("Instagram could not process the Reel")
		}
	}, interval)
	if ctx.Err() != nil {
		return errors.New("Instagram processing is not confirmed yet; the saved container was kept. Check its status before retrying")
	}
	return err
}

func (c *Client) instagramContainerStatus(ctx context.Context, auth, id string) (string, error) {
	type phase struct {
		Status string `json:"status"`
		Bytes  *int64 `json:"bytes_transferred"`
		Size   *int64 `json:"source_file_size"`
		Errors []struct {
			Code int `json:"code"`
		} `json:"errors"`
	}
	var state struct {
		Code  string `json:"status_code"`
		Video struct {
			Uploading  phase `json:"uploading_phase"`
			Processing phase `json:"processing_phase"`
		} `json:"video_status"`
	}
	_, err := c.request(ctx, "GET", graph+id+"?fields=status_code,video_status", auth, nil, nil, 0, &state)
	if err == nil && state.Code != "FINISHED" && state.Code != "PUBLISHED" {
		for i, p := range []phase{state.Video.Uploading, state.Video.Processing} {
			if p.Status != "error" {
				continue
			}
			code := 0
			if len(p.Errors) > 0 {
				code = p.Errors[0].Code
			}
			if state.Code == "IN_PROGRESS" && code == 1363008 {
				diagnostics.Record(ctx, slog.LevelWarn, "upload_processing_pending", diagnostics.Details{ProviderCode: code, ProviderErrorCode: "meta_processing_failed"})
			} else {
				diagnostics.Record(ctx, slog.LevelError, "upload_processing_failed", diagnostics.Details{ProviderCode: code, ProviderErrorCode: "meta_processing_failed"})
			}
			return state.Code, &instagramProcessingError{container: id, code: code, emptyUpload: i == 0 && p.Bytes != nil && p.Size != nil && *p.Bytes == 0 && *p.Size == 0}
		}
	}
	return state.Code, err
}
func waitFor(ctx context.Context, ready func() (bool, error), interval time.Duration) error {
	for {
		done, err := ready()
		if err != nil || done {
			return err
		}
		select {
		case <-ctx.Done():
			return errors.New("processing timed out; check the destination")
		case <-time.After(interval):
		}
	}
}
