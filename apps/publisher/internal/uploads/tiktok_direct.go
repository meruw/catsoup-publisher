package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strings"
	"unicode/utf16"
)

var ErrProcessingFailed = errors.New("The platform confirmed that processing failed; inspect the video and account before another attempt")

type TikTokCreator struct {
	Username        string   `json:"creator_username"`
	Nickname        string   `json:"creator_nickname"`
	Privacy         []string `json:"privacy_level_options"`
	CommentDisabled bool     `json:"comment_disabled"`
	DuetDisabled    bool     `json:"duet_disabled"`
	StitchDisabled  bool     `json:"stitch_disabled"`
	MaxDuration     int      `json:"max_video_post_duration_sec"`
}
type TikTokOptions struct {
	Privacy    string `json:"privacy"`
	Comments   bool   `json:"comments"`
	Duet       bool   `json:"duet"`
	Stitch     bool   `json:"stitch"`
	Commercial bool   `json:"commercial"`
	OwnBrand   bool   `json:"ownBrand"`
	Branded    bool   `json:"branded"`
	Synthetic  bool   `json:"synthetic"`
	Consent    bool   `json:"consent"`
}

func (c *Client) TikTokCreator(ctx context.Context, token string) (TikTokCreator, error) {
	var out struct {
		Data  TikTokCreator
		Error struct{ Code string }
	}
	_, err := c.request(ctx, "POST", tik+"post/publish/creator_info/query/", "Bearer "+token, jsonHeaders, jsonBody(map[string]string{}), -1, &out)
	if err != nil {
		return out.Data, err
	}
	if out.Error.Code != "ok" || out.Data.Nickname == "" || out.Data.MaxDuration <= 0 || len(out.Data.Privacy) == 0 {
		return TikTokCreator{}, errors.New("TikTok cannot offer Direct Post for this account; check video.publish authorization and posting limits")
	}
	allowed := []string{"PUBLIC_TO_EVERYONE", "MUTUAL_FOLLOW_FRIENDS", "FOLLOWER_OF_CREATOR", "SELF_ONLY"}
	for _, p := range out.Data.Privacy {
		if !slices.Contains(allowed, p) {
			return TikTokCreator{}, errors.New("TikTok returned an unsupported privacy option")
		}
	}
	return out.Data, nil
}
func ValidateTikTokOptions(o TikTokOptions, creator TikTokCreator, duration float64, caption string) error {
	if !o.Consent {
		return errors.New("accept TikTok's music and applicable branded-content policies")
	}
	if !slices.Contains(creator.Privacy, o.Privacy) || o.Privacy == "" {
		return errors.New("choose an available TikTok privacy setting")
	}
	if math.IsNaN(duration) || math.IsInf(duration, 0) || duration <= 0 || duration > float64(creator.MaxDuration) {
		return errors.New("video exceeds this TikTok account's duration limit")
	}
	if len(utf16.Encode([]rune(caption))) > 2200 {
		return errors.New("shorten the TikTok caption to 2200 UTF-16 units")
	}
	if (o.Comments && creator.CommentDisabled) || (o.Duet && creator.DuetDisabled) || (o.Stitch && creator.StitchDisabled) {
		return errors.New("TikTok interaction permissions changed; reopen the confirmation")
	}
	if o.Commercial != (o.OwnBrand || o.Branded) {
		return errors.New("choose the applicable commercial-content disclosure")
	}
	if o.Branded && o.Privacy == "SELF_ONLY" {
		return errors.New("branded content cannot use Only me visibility")
	}
	return nil
}
func TikTokDirectEnabled() bool {
	cfg, err := desktopConfig("tiktok")
	return err == nil && strings.Contains(","+cfg.scope+",", ",video.publish,")
}
func SetTikTokDirectEnabled(enabled bool) error {
	raw, err := readDeveloperConfiguration("tiktok-desktop.json")
	if err != nil {
		return errMissingClient
	}
	var cfg map[string]any
	if json.Unmarshal(raw, &cfg) != nil {
		return errors.New("invalid TikTok configuration")
	}
	cfg["direct_post"] = enabled
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return Configure("tiktok", string(data))
}
func (t Token) HasScope(scope string) bool {
	return slices.Contains(strings.FieldsFunc(t.Scope, func(r rune) bool { return r == ',' || r == ' ' }), scope)
}
