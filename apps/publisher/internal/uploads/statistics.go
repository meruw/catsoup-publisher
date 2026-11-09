package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"catsoup/publisher/internal/statistics"
)

// Values stay decimal strings across the desktop bridge (no JavaScript rounding).
type Metric = statistics.Metric
type Statistics struct {
	Items       []Metric            `json:"items"`
	UpdatedAt   string              `json:"updatedAt"`
	Note        string              `json:"note"`
	History     []statistics.Sample `json:"history"`
	HistoryNote string              `json:"historyNote"`
}

var youtubeVideoID = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)
var decimalCounter = regexp.MustCompile(`^[0-9]+$`)

// Statistics only reads provider counters. Endpoints and fields are fixed here;
// callers must obtain the remote ID from the account's saved upload receipt.
func (c *Client) Statistics(ctx context.Context, platform, token, account, reference string) (Statistics, error) {
	out := Statistics{Items: []Metric{}}
	var raw map[string]json.RawMessage
	switch platform {
	case "youtube":
		if !youtubeVideoID.MatchString(reference) {
			return out, errors.New("no valid YouTube video ID is recorded")
		}
		var result struct {
			Items []struct {
				ID      string
				Snippet struct {
					ChannelID string `json:"channelId"`
				}
				Statistics map[string]json.RawMessage
			}
		}
		_, err := c.request(ctx, "GET", "https://www.googleapis.com/youtube/v3/videos?"+url.Values{"part": {"snippet,statistics"}, "id": {reference}}.Encode(), "Bearer "+token, nil, nil, -1, &result)
		if err != nil {
			return out, err
		}
		if len(result.Items) != 1 || result.Items[0].ID != reference || result.Items[0].Snippet.ChannelID != account {
			return out, errors.New("video unavailable for this connected channel")
		}
		raw = result.Items[0].Statistics
	case "instagram", "facebook":
		if !facebookID.MatchString(reference) || !facebookID.MatchString(account) {
			return out, errors.New("no valid Meta media ID is recorded")
		}
		fields := "id,owner{id},like_count,comments_count"
		if platform == "facebook" {
			fields = "id,from{id},likes.limit(0).summary(true),comments.limit(0).summary(true)"
		}
		var result struct {
			ID    string
			Owner struct{ ID string }
			From  struct{ ID string }
			Likes struct {
				Summary struct {
					Total json.RawMessage `json:"total_count"`
				}
			}
			Comments struct {
				Summary struct {
					Total json.RawMessage `json:"total_count"`
				}
			}
			LikeCount    json.RawMessage `json:"like_count"`
			CommentCount json.RawMessage `json:"comments_count"`
		}
		_, err := c.request(ctx, "GET", graph+reference+"?"+url.Values{"fields": {fields}}.Encode(), "Bearer "+token, nil, nil, -1, &result)
		if err != nil {
			return out, err
		}
		owner := result.Owner.ID
		raw = map[string]json.RawMessage{"likeCount": result.LikeCount, "commentCount": result.CommentCount}
		if platform == "facebook" {
			owner = result.From.ID
			raw = map[string]json.RawMessage{"likeCount": result.Likes.Summary.Total, "commentCount": result.Comments.Summary.Total}
		}
		if result.ID != reference || owner != account {
			return out, errors.New("media unavailable for this connected account")
		}
		out.Note = "Views, reach and watch time are not included."
		if platform == "facebook" {
			out.Note += " Likes do not include other reactions."
		}
	default:
		return out, errors.New("statistics are not supported for this platform")
	}
	for _, field := range []struct{ key, label string }{{"viewCount", "Views"}, {"likeCount", "Likes"}, {"commentCount", "Comments"}} {
		value := strings.TrimSpace(string(raw[field.key]))
		if value == "" || value == "null" {
			continue
		}
		if strings.HasPrefix(value, `"`) {
			if json.Unmarshal(raw[field.key], &value) != nil {
				return out, errors.New("unreadable provider counter")
			}
		}
		if !decimalCounter.MatchString(value) {
			return out, errors.New("unreadable provider counter")
		}
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return out, errors.New("unreadable provider counter")
		}
		out.Items = append(out.Items, Metric{Label: field.label, Value: strconv.FormatUint(n, 10)})
	}
	if len(out.Items) == 0 {
		return out, errors.New("the provider did not return counters for this video")
	}
	out.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return out, nil
}
