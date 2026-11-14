package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"errors"
	"log/slog"
	"net/url"
	"regexp"
	"time"
)

var hostedMediaName = regexp.MustCompile(`^[a-z0-9-]+\.trycloudflare\.com$`)
var hostedMediaPath = regexp.MustCompile(`^/[a-f0-9]{48}/video\.mp4$`)

// InstagramHosted creates exactly one container. Unknown results remain locked.
func (c *Client) InstagramHosted(ctx context.Context, token string, a Account, m Media, mediaURL string, record func(string) error, beforePublish func() error) (string, error) {
	u, e := url.Parse(mediaURL)
	if e != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !hostedMediaName.MatchString(u.Host) || !hostedMediaPath.MatchString(u.Path) {
		return "", errors.New("invalid temporary Instagram media URL")
	}
	body := map[string]any{"media_type": "REELS", "video_url": mediaURL, "caption": m.Caption}
	if m.CoverTimeMS != nil {
		body["thumb_offset"] = *m.CoverTimeMS
	}
	var out struct{ ID string }
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_started", diagnostics.Details{})
	if _, e = c.request(ctx, "POST", graph+a.ID+"/media", "Bearer "+token, jsonHeaders, jsonBody(body), -1, &out); e != nil {
		return "", e
	}
	if !safeID.MatchString(out.ID) {
		return "", errors.New("Instagram did not confirm the hosted container")
	}
	if e = record(out.ID); e != nil {
		return "", e
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_processing", diagnostics.Details{})
	if e = c.waitInstagramContainer(ctx, "Bearer "+token, out.ID, 10*time.Second); e != nil {
		return "", e
	}
	if e = beforePublish(); e != nil {
		return "", e
	}
	return c.PublishReadyInstagram(ctx, token, a.ID, out.ID)
}
