package uploads

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type desktopOAuth struct{ authURL, tokenURL, keyField, id, secret, scope, callback string }

func desktopConfig(platform string) (desktopOAuth, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return desktopOAuth{}, err
	}
	if platform == "instagram" {
		raw, err := os.ReadFile(filepath.Join(configDir, "CATsoup Publisher", "meta-desktop.json"))
		if err != nil {
			return desktopOAuth{}, errors.New("configure meta-desktop.json for Facebook Login, or use a Page token")
		}
		var cfg struct {
			ID     string `json:"app_id"`
			Secret string `json:"app_secret"`
			Config string `json:"config_id"`
		}
		if json.Unmarshal(raw, &cfg) != nil || cfg.ID == "" || cfg.Secret == "" || cfg.Config == "" {
			return desktopOAuth{}, errors.New("invalid Meta app configuration")
		}
		return desktopOAuth{"https://www.facebook.com/v25.0/dialog/oauth", graph + "oauth/access_token", "client_id", cfg.ID, cfg.Secret, cfg.Config, "/callback/"}, nil
	}
	if platform == "tiktok" {
		raw, err := os.ReadFile(filepath.Join(configDir, "CATsoup Publisher", "tiktok-desktop.json"))
		if err != nil {
			return desktopOAuth{}, errors.New("save your TikTok Login Kit credentials in %APPDATA%/CATsoup Publisher/tiktok-desktop.json")
		}
		var cfg struct {
			Key    string `json:"client_key"`
			Secret string `json:"client_secret"`
		}
		if json.Unmarshal(raw, &cfg) != nil || cfg.Key == "" || cfg.Secret == "" {
			return desktopOAuth{}, errors.New("invalid TikTok desktop configuration")
		}
		return desktopOAuth{"https://www.tiktok.com/v2/auth/authorize/", tik + "oauth/token/", "client_key", cfg.Key, cfg.Secret, "user.info.basic,video.upload", "/callback/"}, nil
	}
	path := filepath.Join(configDir, "CATsoup Publisher", "google-desktop.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return desktopOAuth{}, errors.New("set up the YouTube connection in Accounts and import your Google Desktop app client")
	}
	var cfg struct {
		Installed struct {
			ID     string `json:"client_id"`
			Secret string `json:"client_secret"`
		}
	}
	if err = json.Unmarshal(raw, &cfg); err != nil || cfg.Installed.ID == "" || cfg.Installed.Secret == "" {
		return desktopOAuth{}, errors.New("Google desktop client configuration is invalid")
	}
	return desktopOAuth{"https://accounts.google.com/o/oauth2/v2/auth", "https://oauth2.googleapis.com/token", "client_id", cfg.Installed.ID, cfg.Installed.Secret, "https://www.googleapis.com/auth/youtube.upload https://www.googleapis.com/auth/youtube.readonly", "/"}, nil
}
func GoogleClientReady() bool {
	_, err := desktopConfig("youtube")
	return err == nil
}

func (c *Client) SignIn(ctx context.Context, platform string, open func(string) error) (Token, error) {
	cfg, err := desktopConfig(platform)
	if err != nil {
		return Token{}, err
	}
	addr := "127.0.0.1:0"
	if platform == "instagram" {
		addr = "127.0.0.1:48172"
	}
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		return Token{}, errors.New("could not start local sign-in")
	}
	defer ln.Close()
	host := ln.Addr().String()
	if platform == "instagram" {
		host = "localhost:48172"
	}
	redirect := "http://" + host + cfg.callback
	state, verifier := rand.Text(), rand.Text()+rand.Text()
	sum := sha256.Sum256([]byte(verifier))
	params := url.Values{cfg.keyField: {cfg.id}, "response_type": {"code"}, "redirect_uri": {redirect}, "state": {state}, "scope": {cfg.scope}, "code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"}}
	if platform == "tiktok" {
		params.Set("code_challenge", hex.EncodeToString(sum[:]))
	} else if platform == "instagram" {
		params.Del("scope")
		params.Del("code_challenge")
		params.Del("code_challenge_method")
		params.Set("config_id", cfg.scope)
		params.Set("override_default_response_type", "true")
	} else {
		params.Set("prompt", "select_account consent")
		params.Set("access_type", "offline")
	}
	code := make(chan string, 1)
	srv := &http.Server{ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 5 * time.Second, MaxHeaderBytes: 8192, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" || r.URL.Path != cfg.callback || r.Host != host || subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("state")), []byte(state)) != 1 {
			http.Error(w, "Invalid callback", 400)
			return
		}
		select {
		case code <- r.URL.Query().Get("code"):
		default:
		}
		w.Write([]byte("You can close this tab and return to CATsoup Publisher."))
	})}
	defer srv.Close()
	go srv.Serve(ln)
	if err = open(cfg.authURL + "?" + params.Encode()); err != nil {
		return Token{}, errors.New("could not open the browser for sign-in")
	}
	var authCode string
	select {
	case authCode = <-code:
	case <-ctx.Done():
		return Token{}, errors.New("sign-in timed out")
	}
	if authCode == "" {
		return Token{}, errors.New("sign-in was cancelled")
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {authCode}, "redirect_uri": {redirect}, cfg.keyField: {cfg.id}, "client_secret": {cfg.secret}, "code_verifier": {verifier}}
	if platform == "instagram" {
		form.Del("code_verifier")
	}
	return c.exchange(ctx, cfg.tokenURL, form)
}
