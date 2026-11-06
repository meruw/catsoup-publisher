package uploads

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func xConfigTest(t *testing.T, callback string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := Configure("x", fmt.Sprintf(`{"client_id":"fixture-client","redirect_uri":%q}`, callback)); err != nil {
		t.Fatal(err)
	}
}

func TestXNativeSignInAndRefresh(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	callback := "http://" + listener.Addr().String() + "/callback/"
	listener.Close()
	xConfigTest(t, callback)
	client := New()
	challenge := ""
	client.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != xAPI+"oauth2/token" {
			t.Fatal("unexpected endpoint", r.URL)
		}
		r.ParseForm()
		if r.Form.Has("client_secret") || r.Form.Get("client_id") != "fixture-client" || r.Header.Get("Authorization") != "" {
			t.Fatal("not a public native-client exchange")
		}
		if r.Form.Get("grant_type") == "authorization_code" {
			sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			if challenge != base64.RawURLEncoding.EncodeToString(sum[:]) || r.Form.Get("redirect_uri") != callback || r.Form.Get("code") != "fixture-code" {
				t.Fatal("invalid PKCE exchange")
			}
		} else if r.Form.Get("grant_type") != "refresh_token" || r.Form.Get("refresh_token") != "fixture-refresh" {
			t.Fatal("invalid refresh")
		}
		return response(`{"access_token":"fixture-access","refresh_token":"fixture-refresh","expires_in":7200}`), nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	token, err := client.SignIn(ctx, "x", func(raw string) error {
		u, _ := url.Parse(raw)
		p := u.Query()
		challenge = p.Get("code_challenge")
		if u.Scheme != "https" || u.Host != "x.com" || p.Get("redirect_uri") != callback || p.Get("code_challenge_method") != "S256" || !strings.Contains(p.Get("scope"), "media.write") {
			t.Fatal("invalid X authorization request")
		}
		for _, state := range []string{"wrong-state", p.Get("state")} {
			resp, e := http.Get(callback + "?" + url.Values{"code": {"fixture-code"}, "state": {state}}.Encode())
			if e != nil {
				return e
			}
			resp.Body.Close()
			expected := 200
			if state == "wrong-state" {
				expected = 400
			}
			if resp.StatusCode != expected {
				t.Fatal("invalid callback accepted", resp.StatusCode)
			}
		}
		return nil
	})
	if err != nil || token.Access != "fixture-access" || token.Client == "" || CheckXClient(token) != nil || ClientStatus("x") != "confirmed" {
		t.Fatal("sign-in failed", err)
	}
	token.Expires = time.Now().Unix() - 1
	refreshed, err := client.RefreshToken(ctx, "x", token)
	if err != nil || refreshed.Client != token.Client || refreshed.Due() {
		t.Fatal("refresh binding failed", err)
	}
	if err = Configure("x", fmt.Sprintf(`{"client_id":"different-project","redirect_uri":%q}`, callback)); err != nil {
		t.Fatal(err)
	}
	if CheckXClient(refreshed) == nil {
		t.Fatal("accepted a different billing client")
	}
	if _, err = client.RefreshToken(ctx, "x", token); err == nil {
		t.Fatal("refreshed against a different client")
	}
}

func TestXCallbackAndOccupiedPort(t *testing.T) {
	for _, raw := range []string{"https://example.org/callback/", "http://localhost:48173/callback/", "http://127.0.0.1:0/callback/", "http://127.0.0.1:80/callback/", "http://127.0.0.1:70000/callback/", "http://127.0.0.1:48173/callback/?x=1", "http://127.0.0.1:48173/callback/#secret", "http://user@127.0.0.1:48173/callback/", "http://127.0.0.1:48173/%63allback/"} {
		if validXCallback(raw) {
			t.Fatal("unsafe callback accepted", raw)
		}
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	xConfigTest(t, "http://"+listener.Addr().String()+"/callback/")
	_, err = New().SignIn(context.Background(), "x", func(string) error { t.Fatal("opened browser on occupied port"); return nil })
	if err == nil || !strings.Contains(err.Error(), "occupied") {
		t.Fatal(err)
	}
}

func TestXTransferChunksAndReceipt(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "video")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	raw := bytes.Repeat([]byte("cat"), 1500000)
	f.Write(raw)
	client := New()
	sent := []byte{}
	stages := []string{}
	checkpoint := ""
	client.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.x.com" || r.Header.Get("Authorization") != "Bearer fixture" {
			t.Fatal("wrong host/auth")
		}
		stages = append(stages, r.Method+" "+r.URL.Path)
		switch r.URL.Path {
		case "/2/users/me":
			return response(`{"data":{"id":"123","username":"cats_es"}}`), nil
		case "/2/media/upload/initialize":
			var payload struct {
				Size int64  `json:"total_bytes"`
				Type string `json:"media_type"`
			}
			json.NewDecoder(r.Body).Decode(&payload)
			if payload.Size != int64(len(raw)) || payload.Type != "video/mp4" {
				t.Fatal("wrong initialization")
			}
			return response(`{"data":{"id":"456"}}`), nil
		case "/2/media/upload/456/append":
			if checkpoint != "media:456" {
				t.Fatal("upload began before durable checkpoint")
			}
			if err := r.ParseMultipartForm(5 << 20); err != nil {
				t.Fatal(err)
			}
			defer r.MultipartForm.RemoveAll()
			file, _, err := r.FormFile("media")
			if err != nil {
				t.Fatal(err)
			}
			b, _ := io.ReadAll(file)
			file.Close()
			if len(b) > 4<<20 || r.FormValue("segment_index") != fmt.Sprint(len(sent)/(4<<20)) {
				t.Fatal("bad chunk")
			}
			sent = append(sent, b...)
			return response(""), nil
		case "/2/media/upload/456/finalize":
			return response(`{"data":{"id":"456","processing_info":{"state":"pending","check_after_secs":1}}}`), nil
		case "/2/media/upload":
			if r.Method != "GET" || r.URL.Query().Get("command") != "STATUS" || r.URL.Query().Get("media_id") != "456" {
				t.Fatal("bad status")
			}
			return response(`{"data":{"processing_info":{"state":"succeeded"}}}`), nil
		case "/2/tweets":
			var post struct {
				Text  string
				Media struct {
					IDs []string `json:"media_ids"`
				}
			}
			json.NewDecoder(r.Body).Decode(&post)
			if post.Text != "Caption" || len(post.Media.IDs) != 1 || post.Media.IDs[0] != "456" || !bytes.Equal(sent, raw) {
				t.Fatal("post before complete video")
			}
			return response(`{"data":{"id":"789"}}`), nil
		default:
			t.Fatal("unexpected request", r.URL)
			return nil, errors.New("unexpected request")
		}
	})
	a, err := client.Identify(context.Background(), "x", "fixture")
	if err != nil || a.Name != "@cats_es" {
		t.Fatal(a, err)
	}
	ref, err := client.Send(context.Background(), "x", "fixture", a, Media{Caption: "Caption", File: f, Size: int64(len(raw))}, func(s string) error { checkpoint = s; return nil })
	if err != nil || ref != "https://x.com/i/status/789" || len(stages) != 7 {
		t.Fatal(ref, stages, err)
	}
}

func TestXFailureNeverRetriesPost(t *testing.T) {
	for _, failure := range []string{"checkpoint", "append", "processing", "post"} {
		t.Run(failure, func(t *testing.T) {
			f, _ := os.CreateTemp(t.TempDir(), "video")
			defer f.Close()
			f.WriteString("video")
			c := New()
			posts := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if strings.HasSuffix(r.URL.Path, "initialize") {
					return response(`{"data":{"id":"123"}}`), nil
				}
				if strings.HasSuffix(r.URL.Path, "append") {
					if failure == "append" {
						return nil, errors.New("fixture network error")
					}
					return response(""), nil
				}
				if strings.HasSuffix(r.URL.Path, "finalize") {
					if failure == "processing" {
						return response(`{"data":{"processing_info":{"state":"failed"}}}`), nil
					}
					return response(`{"data":{"id":"123"}}`), nil
				}
				if strings.HasSuffix(r.URL.Path, "tweets") {
					posts++
					return nil, errors.New("unknown post outcome")
				}
				t.Fatal("unexpected call")
				return nil, nil
			})
			_, err := c.Send(context.Background(), "x", "fixture", Account{}, Media{File: f, Size: 5}, func(string) error {
				if failure == "checkpoint" {
					return errors.New("disk full")
				}
				return nil
			})
			expected := 0
			if failure == "post" {
				expected = 1
			}
			if err == nil || posts != expected {
				t.Fatal("unsafe retry", posts, err)
			}
		})
	}
}

func TestXValidation(t *testing.T) {
	for _, m := range []Media{{Size: 513 << 20}, {Size: 1, Caption: strings.Repeat("a", 281)}, {Size: 1, Caption: strings.Repeat("猫", 141)}, {Size: 1, Caption: strings.Repeat("a", 260) + " https://x.co"}} {
		if Validate("x", m) == nil {
			t.Fatal("invalid media accepted")
		}
	}
	if Validate("x", Media{Size: 1, Caption: strings.Repeat("a", 280)}) != nil {
		t.Fatal("valid text rejected")
	}
}
