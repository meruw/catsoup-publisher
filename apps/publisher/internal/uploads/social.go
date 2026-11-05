package uploads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

func (c *Client) tiktok(ctx context.Context, token string, m Media, record func(string) error) (string, error) {
	chunk := min(m.Size, int64(16<<20))
	count := m.Size / chunk
	src := map[string]any{"source": "FILE_UPLOAD", "video_size": m.Size, "chunk_size": chunk, "total_chunk_count": count}
	var out struct {
		Data struct {
			ID  string `json:"publish_id"`
			URL string `json:"upload_url"`
		}
		Error struct{ Code string }
	}
	auth := "Bearer " + token
	_, err := c.request(ctx, "POST", tik+"post/publish/inbox/video/init/", auth, jsonHeaders, jsonBody(map[string]any{"source_info": src}), -1, &out)
	if err != nil {
		return "", err
	}
	if out.Error.Code != "ok" || !safeID.MatchString(out.Data.ID) || !safeSession(out.Data.URL, "open-upload.tiktokapis.com", "open-upload-us.tiktokapis.com", "open-upload-eu.tiktokapis.com") {
		return "", errors.New("TikTok declined the inbox upload or returned an unexpected upload address")
	}
	if err = record(out.Data.ID); err != nil {
		return "", err
	}
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
	err = waitFor(ctx, func() (bool, error) {
		var outStatus struct {
			Data  struct{ Status string }
			Error struct{ Code string }
		}
		_, err := c.request(ctx, "POST", tik+"post/publish/status/fetch/", auth, jsonHeaders, jsonBody(map[string]string{"publish_id": out.Data.ID}), -1, &outStatus)
		if err != nil {
			return false, err
		}
		if outStatus.Error.Code != "ok" || outStatus.Data.Status == "FAILED" {
			return false, errors.New("TikTok could not deliver the inbox upload")
		}
		return outStatus.Data.Status == "SEND_TO_USER_INBOX" || outStatus.Data.Status == "PUBLISH_COMPLETE", nil
	}, 5*time.Second)
	return out.Data.ID, err
}

func (c *Client) instagram(ctx context.Context, token string, a Account, m Media, record func(string) error) (string, error) {
	auth := "Bearer " + token
	var container struct{ ID, URI string }
	_, err := c.request(ctx, "POST", graph+a.ID+"/media", auth, jsonHeaders, jsonBody(map[string]string{"media_type": "REELS", "upload_type": "resumable", "caption": m.Caption}), -1, &container)
	if err != nil {
		return "", err
	}
	if !safeID.MatchString(container.ID) || !safeSession(container.URI, "rupload.facebook.com") {
		return "", errors.New("Meta did not return a supported local Reel upload session")
	}
	if err = record(container.ID); err != nil {
		return "", err
	}
	var accepted struct{ Success bool }
	_, err = c.request(ctx, "POST", container.URI, "OAuth "+token, map[string]string{"Content-Type": "application/octet-stream", "offset": "0", "file_size": fmt.Sprint(m.Size)}, io.NewSectionReader(m.File, 0, m.Size), m.Size, &accepted)
	if err != nil {
		return "", err
	}
	if !accepted.Success {
		return "", errors.New("Meta declined the video transfer")
	}
	err = waitFor(ctx, func() (bool, error) {
		var state struct {
			Code string `json:"status_code"`
		}
		_, err := c.request(ctx, "GET", graph+container.ID+"?fields=status_code", auth, nil, nil, 0, &state)
		if err != nil {
			return false, err
		}
		switch state.Code {
		case "FINISHED":
			return true, nil
		case "IN_PROGRESS":
			return false, nil
		default:
			return false, errors.New("Instagram could not process the Reel")
		}
	}, time.Minute)
	if err != nil {
		return "", err
	}
	var result struct{ ID string }
	_, err = c.request(ctx, "POST", graph+a.ID+"/media_publish", auth, jsonHeaders, jsonBody(map[string]string{"creation_id": container.ID}), -1, &result)
	if err != nil {
		return "", err
	}
	if !safeID.MatchString(result.ID) {
		return "", errors.New("Instagram did not return a media ID")
	}
	return result.ID, nil
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
