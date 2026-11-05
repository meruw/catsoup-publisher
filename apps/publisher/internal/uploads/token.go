package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

type Token struct {
	Access, Refresh string
	Expires         int64
}

func (t Token) Due() bool { return t.Expires > 0 && time.Now().Add(time.Minute).Unix() >= t.Expires }
func (t Token) Seal(scope string) ([]byte, error) {
	raw, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return protect(raw, scope, true)
}
func Unseal(raw []byte, scope string) (Token, error) {
	b, err := protect(raw, scope, false)
	if err != nil {
		return Token{}, err
	}
	var t Token
	err = json.Unmarshal(b, &t)
	if err != nil || t.Access == "" {
		return Token{}, errors.New("invalid credential; reconnect from Accounts")
	}
	return t, nil
}
func (c *Client) exchange(ctx context.Context, endpoint string, form url.Values) (Token, error) {
	var out struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
		Seconds int64  `json:"expires_in"`
	}
	_, err := c.request(ctx, "POST", endpoint, "", map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, strings.NewReader(form.Encode()), -1, &out)
	if err != nil {
		return Token{}, err
	}
	if out.Access == "" {
		return Token{}, errors.New("provider declined authorization; reconnect from Accounts")
	}
	t := Token{Access: out.Access, Refresh: out.Refresh}
	if out.Seconds > 0 {
		t.Expires = time.Now().Unix() + out.Seconds
	}
	return t, nil
}
func (c *Client) RefreshToken(ctx context.Context, platform string, t Token) (Token, error) {
	if !t.Due() {
		return t, nil
	}
	if t.Refresh == "" {
		return Token{}, errors.New("authorization expired; reconnect from Accounts")
	}
	cfg, err := desktopConfig(platform)
	if err != nil {
		return Token{}, err
	}
	next, err := c.exchange(ctx, cfg.tokenURL, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {t.Refresh}, cfg.keyField: {cfg.id}, "client_secret": {cfg.secret}})
	if err != nil {
		return Token{}, err
	}
	if next.Refresh == "" {
		next.Refresh = t.Refresh
	}
	return next, nil
}

// Resolve the Page credential for the explicitly requested Instagram username.
func (c *Client) InstagramPage(ctx context.Context, user Token, handle string) (Token, error) {
	endpoint := graph + "me/accounts?fields=access_token%2Cinstagram_business_account%7Bid%2Cusername%7D&limit=100"
	after := ""
	for page := 0; page < 25; page++ {
		var out struct {
			Data []struct {
				Token string                    `json:"access_token"`
				IG    struct{ Username string } `json:"instagram_business_account"`
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
		_, err := c.request(ctx, "GET", target, "Bearer "+user.Access, nil, nil, 0, &out)
		if err != nil {
			return Token{}, err
		}
		for _, p := range out.Data {
			if p.Token != "" && strings.EqualFold(p.IG.Username, handle) {
				return Token{Access: p.Token}, nil
			}
		}
		if out.Paging.Next == "" || out.Paging.Cursors.After == "" || out.Paging.Cursors.After == after {
			break
		}
		after = out.Paging.Cursors.After
	}
	return Token{}, errors.New("no authorized Page matches this Instagram username")
}
