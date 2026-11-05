// Package uploads sends approved media through official platform APIs.
package uploads

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Media struct {
	Title, Caption, Language string
	File                     *os.File
	Size                     int64
}
type Client struct{ HTTP *http.Client }

func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 10 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (c *Client) request(ctx context.Context, method, endpoint, auth string, headers map[string]string, body io.Reader, size int64, result any) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, errors.New("invalid provider request")
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if size >= 0 {
		req.ContentLength = size
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, errors.New("provider connection interrupted; check the destination before retrying")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("provider returned HTTP %d; check authorization, account permissions and upload limits", resp.StatusCode)
	}
	if result != nil && json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(result) != nil {
		return nil, errors.New("provider returned an unreadable response")
	}
	return resp.Header, nil
}
func jsonBody(value any) io.Reader { b, _ := json.Marshal(value); return bytes.NewReader(b) }

var jsonHeaders = map[string]string{"Content-Type": "application/json"}

func safeSession(raw string, hosts ...string) bool {
	u, e := url.Parse(raw)
	if e != nil || u.Scheme != "https" || u.User != nil || u.Fragment != "" {
		return false
	}
	for _, host := range hosts {
		if u.Host == host {
			return true
		}
	}
	return false
}

var safeID = regexp.MustCompile("^[A-Za-z0-9_.~-]{1,200}$")

const graph = "https://graph.facebook.com/v25.0/"
const tik = "https://open.tiktokapis.com/v2/"

func (c *Client) Identify(ctx context.Context, platform, token string) (Account, error) {
	var a Account
	auth := "Bearer " + token
	switch platform {
	case "youtube":
		var out struct {
			Items []struct {
				ID      string
				Snippet struct{ Title string }
			}
		}
		_, err := c.request(ctx, "GET", "https://www.googleapis.com/youtube/v3/channels?part=snippet&mine=true", auth, nil, nil, 0, &out)
		if err != nil {
			return a, err
		}
		if len(out.Items) == 1 {
			a = Account{out.Items[0].ID, out.Items[0].Snippet.Title}
		}
	case "tiktok":
		var out struct {
			Data struct {
				User struct {
					ID   string `json:"open_id"`
					Name string `json:"display_name"`
				}
			}
			Error struct{ Code string }
		}
		_, err := c.request(ctx, "GET", tik+"user/info/?fields=open_id,display_name", auth, nil, nil, 0, &out)
		if err != nil {
			return a, err
		}
		if out.Error.Code == "ok" {
			a = Account{out.Data.User.ID, out.Data.User.Name}
		}
	case "instagram":
		var out struct {
			IG struct{ ID, Username string } `json:"instagram_business_account"`
		}
		_, err := c.request(ctx, "GET", graph+"me?fields=instagram_business_account%7Bid%2Cusername%7D", auth, nil, nil, 0, &out)
		if err != nil {
			return a, err
		}
		a = Account{out.IG.ID, out.IG.Username}
	default:
		return a, errors.New("this platform uses manual upload")
	}
	if !safeID.MatchString(a.ID) || a.Name == "" {
		return Account{}, errors.New("could not identify a single eligible account")
	}
	return a, nil
}

func Validate(platform string, m Media) error {
	if m.Size < 1 {
		return errors.New("the video is empty")
	}
	switch platform {
	case "youtube":
		if strings.TrimSpace(m.Title) == "" || utf8.RuneCountInString(m.Title) > 100 || len(m.Caption) > 5000 || strings.ContainsAny(m.Title+m.Caption, "<>") {
			return errors.New("YouTube requires a title up to 100 characters and a description up to 5000 bytes, without angle brackets")
		}
	case "instagram":
		if utf8.RuneCountInString(m.Caption) > 2200 {
			return errors.New("shorten the Instagram caption to 2200 characters")
		}
	case "tiktok":
		if m.Size > 4*1024*1024*1024 {
			return errors.New("video exceeds the TikTok transfer limit")
		}
	default:
		return errors.New("this platform uses manual upload")
	}
	return nil
}

// Record each non-secret remote reference as soon as the platform creates it.
func (c *Client) Send(ctx context.Context, platform, token string, a Account, m Media, record func(string) error) (string, error) {
	switch platform {
	case "youtube":
		return c.youtube(ctx, token, m)
	case "tiktok":
		return c.tiktok(ctx, token, m, record)
	case "instagram":
		return c.instagram(ctx, token, a, m, record)
	}
	return "", errors.New("unsupported upload destination")
}

func (c *Client) youtube(ctx context.Context, token string, m Media) (string, error) {
	auth := "Bearer " + token
	meta := map[string]any{"snippet": map[string]string{"title": m.Title, "description": m.Caption, "defaultLanguage": m.Language, "categoryId": "24"}, "status": map[string]string{"privacyStatus": "private"}}
	h, err := c.request(ctx, "POST", "https://www.googleapis.com/upload/youtube/v3/videos?uploadType=resumable&part=snippet,status&notifySubscribers=false", auth, map[string]string{"Content-Type": "application/json", "X-Upload-Content-Type": "video/mp4", "X-Upload-Content-Length": fmt.Sprint(m.Size)}, jsonBody(meta), -1, nil)
	if err != nil {
		return "", err
	}
	session := h.Get("Location")
	if !safeSession(session, "www.googleapis.com", "youtube.googleapis.com") {
		return "", errors.New("unexpected Google upload host")
	}
	var out struct{ ID string }
	_, err = c.request(ctx, "PUT", session, auth, map[string]string{"Content-Type": "video/mp4"}, io.NewSectionReader(m.File, 0, m.Size), m.Size, &out)
	if err != nil {
		return "", err
	}
	if !safeID.MatchString(out.ID) {
		return "", errors.New("YouTube did not return a video ID")
	}
	return out.ID, nil
}
