package uploads

import (
	"context"
	"errors"
	"regexp"
)

// TikTokPostStatus only observes an existing upload; it never initializes media.
type TikTokPostStatus struct {
	Status     string `json:"status"`
	FailReason string `json:"fail_reason"`
}

var tikTokInboxID = regexp.MustCompile(`^v_inbox_file~v2\.[0-9]+$`)
var tikTokReason = regexp.MustCompile(`^[a-zA-Z0-9_]{1,120}$`)

func (c *Client) TikTokInboxStatus(ctx context.Context, token, reference string) (TikTokPostStatus, error) {
	if !tikTokInboxID.MatchString(reference) {
		return TikTokPostStatus{}, errors.New("invalid inbox receipt")
	}
	var out struct {
		Data  TikTokPostStatus
		Error struct{ Code string }
	}
	_, err := c.request(ctx, "POST", tik+"post/publish/status/fetch/", "Bearer "+token, jsonHeaders, jsonBody(map[string]string{"publish_id": reference}), -1, &out)
	if err != nil {
		return TikTokPostStatus{}, err
	}
	if out.Error.Code != "ok" {
		return TikTokPostStatus{}, errors.New("TikTok status unavailable")
	}
	switch out.Data.Status {
	case "PROCESSING_UPLOAD", "PROCESSING_DOWNLOAD", "SEND_TO_USER_INBOX", "PUBLISH_COMPLETE", "FAILED":
	default:
		return TikTokPostStatus{}, errors.New("unrecognized TikTok status")
	}
	if out.Data.FailReason != "" && !tikTokReason.MatchString(out.Data.FailReason) {
		out.Data.FailReason = "provider_processing_failed"
	}
	return out.Data, nil
}
