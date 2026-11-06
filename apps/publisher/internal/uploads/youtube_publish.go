package uploads

import (
	"context"
	"errors"
	"net/url"
)

// PublishYouTube updates the existing upload; it never transfers the video again.
// Fetch mutable metadata first so changing visibility does not reset other settings.
func (c *Client) PublishYouTube(ctx context.Context, token, channel, id string, m Media) (string, error) {
	if !safeID.MatchString(id) {
		return "", errors.New("invalid YouTube receipt")
	}
	var out struct {
		Items []struct {
			ID      string
			Snippet map[string]any
			Status  map[string]any
		}
	}
	_, err := c.request(ctx, "GET", "https://www.googleapis.com/youtube/v3/videos?part=snippet,status&id="+url.QueryEscape(id), "Bearer "+token, nil, nil, 0, &out)
	if err != nil {
		return "", err
	}
	if len(out.Items) != 1 || out.Items[0].Snippet["channelId"] != channel {
		return "", errors.New("the recorded video does not belong to this channel")
	}
	item := out.Items[0]
	if item.Status["privacyStatus"] == "public" {
		return id, nil
	}
	if item.Status["privacyStatus"] != "private" {
		return "", errors.New("check this video's visibility in YouTube Studio before publishing")
	}
	snippet := map[string]any{}
	for _, k := range []string{"title", "description", "tags", "categoryId", "defaultLanguage", "defaultAudioLanguage"} {
		if v, ok := item.Snippet[k]; ok {
			snippet[k] = v
		}
	}
	snippet["title"] = m.Title
	snippet["description"] = m.Caption
	snippet["defaultLanguage"] = m.Language
	status := map[string]any{}
	for _, k := range []string{"embeddable", "license", "publicStatsViewable", "selfDeclaredMadeForKids", "containsSyntheticMedia"} {
		if v, ok := item.Status[k]; ok {
			status[k] = v
		}
	}
	status["privacyStatus"] = "public"
	status["selfDeclaredMadeForKids"] = m.Audience == "kids"
	status["containsSyntheticMedia"] = m.Synthetic
	var result struct {
		ID     string
		Status struct {
			Privacy string `json:"privacyStatus"`
		}
	}
	_, err = c.request(ctx, "PUT", "https://www.googleapis.com/youtube/v3/videos?part=snippet,status", "Bearer "+token, jsonHeaders, jsonBody(map[string]any{"id": id, "snippet": snippet, "status": status}), -1, &result)
	if err != nil {
		return "", err
	}
	if result.ID != id || result.Status.Privacy != "public" {
		return "", errors.New("YouTube did not confirm public visibility; check YouTube Studio")
	}
	return id, nil
}
