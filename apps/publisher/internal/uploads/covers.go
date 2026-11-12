package uploads

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"time"
)

func (c *Client) SetYouTubeCover(ctx context.Context, token, id string, b []byte, mime string) error {
	if !safeID.MatchString(id) || len(b) == 0 || len(b) > 2<<20 || (mime != "image/jpeg" && mime != "image/png") {
		return errors.New("invalid YouTube thumbnail")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	err := waitFor(ctx, func() (bool, error) {
		var out struct {
			Items []struct {
				Processing struct {
					Status string `json:"processingStatus"`
				} `json:"processingDetails"`
			}
		}
		_, err := c.request(ctx, "GET", "https://www.googleapis.com/youtube/v3/videos?part=processingDetails&id="+url.QueryEscape(id), "Bearer "+token, nil, nil, 0, &out)
		if err != nil {
			return false, err
		}
		if len(out.Items) != 1 {
			return false, errors.New("YouTube video unavailable for thumbnail update")
		}
		status := out.Items[0].Processing.Status
		if status == "failed" || status == "terminated" {
			return false, errors.New("YouTube video processing failed")
		}
		return status == "succeeded", nil
	}, 5*time.Second)
	if err != nil {
		return errors.New("video kept on YouTube; its thumbnail is pending. Use Apply thumbnail after processing finishes")
	}
	var out struct{ Items []map[string]any }
	_, err = c.request(ctx, "POST", "https://www.googleapis.com/upload/youtube/v3/thumbnails/set?videoId="+url.QueryEscape(id), "Bearer "+token, map[string]string{"Content-Type": mime}, bytes.NewReader(b), int64(len(b)), &out)
	if err != nil {
		return errors.New("video kept on YouTube; thumbnail was not confirmed. Check channel custom-thumbnail eligibility, then retry Apply thumbnail")
	}
	if len(out.Items) == 0 {
		return errors.New("YouTube did not confirm the thumbnail; the video was kept")
	}
	return nil
}
