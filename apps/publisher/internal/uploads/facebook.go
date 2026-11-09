package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

var facebookID = regexp.MustCompile(`^[0-9]{1,32}$`)

// FacebookPage resolves only an explicitly selected, authorized Page. The
// Instagram handle is a convenience selector, never the publishing destination.
func (c *Client) FacebookPage(ctx context.Context, user Token, selector string) (Token, error) {
	var permissions struct {
		Data []struct{ Permission, Status string }
	}
	if _, err := c.request(ctx, "GET", graph+"me/permissions", "Bearer "+user.Access, nil, nil, 0, &permissions); err != nil {
		return Token{}, err
	}
	granted := false
	for _, p := range permissions.Data {
		if p.Permission == "pages_manage_posts" && p.Status == "granted" {
			granted = true
		}
	}
	if !granted {
		return Token{}, errors.New("Facebook publishing permission is missing. Add pages_manage_posts to the Meta app and login configuration, then reconnect Facebook")
	}
	diagnostics.Record(ctx, slog.LevelInfo, "pages_started", diagnostics.Details{})
	endpoint := graph + "me/accounts?fields=id,name,access_token,tasks,instagram_business_account%7Busername%7D&limit=100"
	after := ""
	for page := 0; page < 25; page++ {
		var out struct {
			Data []struct {
				ID          string
				AccessToken string `json:"access_token"`
				Tasks       []string
				IG          struct{ Username string } `json:"instagram_business_account"`
			}
			Paging struct {
				Next    string
				Cursors struct{ After string }
			}
		}
		target := endpoint
		if after != "" {
			target += "&after=" + url.QueryEscape(after)
		}
		if _, err := c.request(ctx, "GET", target, "Bearer "+user.Access, nil, nil, 0, &out); err != nil {
			return Token{}, err
		}
		diagnostics.Record(ctx, slog.LevelDebug, "pages_received", diagnostics.Details{Pages: len(out.Data)})
		for _, p := range out.Data {
			if p.ID != selector && (p.IG.Username == "" || !strings.EqualFold(p.IG.Username, selector)) {
				continue
			}
			canCreate := false
			for _, task := range p.Tasks {
				if slices.Contains([]string{"CREATE_CONTENT", "MANAGE", "PROFILE_PLUS_CREATE_CONTENT", "PROFILE_PLUS_MANAGE", "PROFILE_PLUS_FULL_CONTROL"}, task) {
					canCreate = true
				}
			}
			if !facebookID.MatchString(p.ID) || p.AccessToken == "" || !canCreate {
				return Token{}, errors.New("this Facebook Page does not grant content creation access; check Page access and reconnect")
			}
			return Token{Access: p.AccessToken}, nil
		}
		if out.Paging.Next == "" || out.Paging.Cursors.After == "" || out.Paging.Cursors.After == after {
			break
		}
		after = out.Paging.Cursors.After
	}
	return Token{}, errors.New("no authorized Facebook Page matches this ID or linked Instagram username; select the correct Page during Facebook sign-in")
}

func (c *Client) facebook(ctx context.Context, token string, a Account, m Media, record func(string) error) (string, error) {
	if !facebookID.MatchString(a.ID) {
		return "", errors.New("invalid Facebook Page ID")
	}
	if err := Validate("facebook", m); err != nil {
		return "", err
	}
	auth := "Bearer " + token
	endpoint := graph + a.ID + "/video_reels"
	var session struct {
		ID  string `json:"video_id"`
		URL string `json:"upload_url"`
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_started", diagnostics.Details{})
	if _, err := c.request(ctx, "POST", endpoint, auth, jsonHeaders, jsonBody(map[string]string{"upload_phase": "start"}), -1, &session); err != nil {
		return "", err
	}
	if !facebookID.MatchString(session.ID) {
		return "", errors.New("Facebook did not return a valid Reel ID")
	}
	if !safeSession(session.URL, "rupload.facebook.com") {
		return session.ID, errors.New("Facebook returned an unsupported upload address; no video data was sent")
	}
	if err := record(session.ID); err != nil {
		return session.ID, err
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_init_finished", diagnostics.Details{})
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_started", diagnostics.Details{})
	var uploaded struct{ Success bool }
	_, err := c.request(ctx, "POST", session.URL, "OAuth "+token, map[string]string{"offset": "0", "file_size": strconv.FormatInt(m.Size, 10), "Content-Type": "application/octet-stream"}, io.NewSectionReader(m.File, 0, m.Size), m.Size, &uploaded)
	if err != nil {
		return session.ID, err
	}
	if !uploaded.Success {
		return session.ID, errors.New("Facebook did not confirm the Reel upload; check the Page before retrying")
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_bytes_finished", diagnostics.Details{})
	var finished struct{ Success bool }
	_, err = c.request(ctx, "POST", endpoint, auth, jsonHeaders, jsonBody(map[string]string{"upload_phase": "finish", "video_id": session.ID, "video_state": "PUBLISHED", "title": m.Title, "description": m.Caption}), -1, &finished)
	if err != nil {
		return session.ID, err
	}
	if !finished.Success {
		return session.ID, errors.New("Facebook did not accept Reel publication; check the Page before retrying")
	}
	diagnostics.Record(ctx, slog.LevelInfo, "upload_processing", diagnostics.Details{})
	err = waitFor(ctx, func() (bool, error) {
		var out struct {
			Status struct {
				Video   string                  `json:"video_status"`
				Upload  struct{ Status string } `json:"uploading_phase"`
				Process struct{ Status string } `json:"processing_phase"`
				Publish struct{ Status string } `json:"publishing_phase"`
			}
		}
		if _, e := c.request(ctx, "GET", graph+session.ID+"?fields=status", auth, nil, nil, 0, &out); e != nil {
			return false, e
		}
		for _, state := range []string{out.Status.Video, out.Status.Upload.Status, out.Status.Process.Status, out.Status.Publish.Status} {
			if state == "error" || state == "failed" {
				return false, ErrProcessingFailed
			}
		}
		return out.Status.Video == "ready" && out.Status.Publish.Status == "complete", nil
	}, 5*time.Second)
	return session.ID, err
}
