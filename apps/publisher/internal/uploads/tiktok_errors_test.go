package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestTikTokInitDiagnosticsAndNoTransfer(t *testing.T) {
	for _, tc := range []struct{ name, code, id, address, want string }{
		{"rejected", "scope_not_authorized", "", "", "scope_not_authorized"},
		{"unknown", "token-DO-NOT-LOG", "", "", "unrecognized error"},
		{"id", "ok", "bad/id", "https://open-upload.tiktokapis.com/video/?upload_token=secret", "invalid publishing ID"},
		{"host", "ok", "v_inbox_file~v2.123", "https://other.example/video/?upload_token=secret", "unsupported upload host (other.example)"},
		{"scheme", "ok", "valid", "http://open-upload.tiktokapis.com/video/?upload_token=secret", "invalid upload address"},
		{"userinfo", "ok", "valid", "https://secret@open-upload.tiktokapis.com/video/", "invalid upload address"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := diagnostics.Open(t.TempDir())
			defer store.Close()
			ctx := diagnostics.StartTransfer(context.Background(), store, "tiktok", "en")
			c := New()
			calls := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if !strings.HasSuffix(r.URL.Path, "/inbox/video/init/") {
					t.Fatal("unexpected request", r.URL.Path)
				}
				raw, _ := json.Marshal(map[string]any{"data": map[string]string{"publish_id": tc.id, "upload_url": tc.address}, "error": map[string]string{"code": tc.code, "message": "secret"}})
				return response(string(raw)), nil
			})
			f, err := os.CreateTemp(t.TempDir(), "video")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			f.WriteString("fixture")
			recorded := false
			_, err = c.tiktok(ctx, "secret", Media{File: f, Size: 7}, func(string) error { recorded = true; return nil })
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "DO-NOT-LOG") || calls != 1 || recorded {
				t.Fatal(err, calls, recorded)
			}
			result, err := store.Query(diagnostics.Query{})
			if err != nil || len(result.Events) == 0 {
				t.Fatal(result, err)
			}
			raw, _ := json.Marshal(result)
			for _, s := range []string{"secret", "DO-NOT-LOG", "https://", "other.example"} {
				if strings.Contains(string(raw), s) {
					t.Fatal("sensitive data in diagnostics")
				}
			}
		})
	}
}
func TestTikTokDocumentedUploadSession(t *testing.T) {
	for _, id := range []string{"v_inbox_file~v2.123456789", "v_pub_file~v2-1.123456789"} {
		if err := tikTokSession(context.Background(), "ok", id, "https://open-upload.tiktokapis.com/video/?upload_token=secret"); err != nil {
			t.Fatal(err)
		}
	}
}
func TestTikTokHTTPRejection(t *testing.T) {
	c := New()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		r := response(`{"error":{"code":"unaudited_client_can_only_post_to_private_accounts","message":"secret"}}`)
		r.StatusCode = 403
		return r, nil
	})
	_, err := c.request(context.Background(), "POST", tik+"post/publish/video/init/", "", nil, nil, -1, nil)
	if err == nil || !strings.Contains(err.Error(), "HTTP 403") || !strings.Contains(err.Error(), "unaudited_client_can_only_post_to_private_accounts") || strings.Contains(err.Error(), "secret") {
		t.Fatal(err)
	}
}

func TestTikTokHTTPErrorEnvelopesAndSingleEvent(t *testing.T) {
	for _, tc := range []struct{ body, code string }{
		{`{"error":{"code":"new_validation_error","message":"secret"}}`, "new_validation_error"},
		{`{"error":"invalid_request","error_description":"secret"}`, "invalid_request"},
		{`{"code":"new_validation_error","message":"secret"}`, "new_validation_error"},
		{`{"error":{"code":40001,"message":"secret"}}`, "40001"},
		{`{"error":{"code":"https://secret.test/token","message":"secret"}}`, ""},
		{`{"error":{"code":"token-DO-NOT-LOG"}}`, ""},
		{`<html>secret</html>`, ""},
	} {
		t.Run(tc.body, func(t *testing.T) {
			store := diagnostics.Open(t.TempDir())
			defer store.Close()
			ctx := diagnostics.StartTransfer(context.Background(), store, "tiktok", "es")
			c := New()
			calls := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				calls++
				resp := response(tc.body)
				resp.StatusCode = 400
				return resp, nil
			})
			_, err := c.request(ctx, "POST", tik+"post/publish/inbox/video/init/", "Bearer secret", nil, nil, -1, nil)
			if err == nil || calls != 1 || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "DO-NOT-LOG") || (tc.code != "" && !strings.Contains(err.Error(), tc.code)) {
				t.Fatal(err, calls)
			}
			result, queryErr := store.Query(diagnostics.Query{Level: "ERROR"})
			if queryErr != nil || len(result.Events) != 1 {
				t.Fatal(result, queryErr)
			}
			e := result.Events[0]
			if e.HTTPStatus != 400 || e.ProviderErrorCode != tc.code {
				t.Fatal(e)
			}
			encoded, _ := json.Marshal(result)
			if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "DO-NOT-LOG") {
				t.Fatal("unsafe diagnostic")
			}
			if tc.code != "" {
				found, _ := store.Query(diagnostics.Query{Search: tc.code})
				if len(found.Events) != 1 {
					t.Fatal("code is not searchable")
				}
			}
		})
	}
}

func TestTikTokRegionalUploadHosts(t *testing.T) {
	for _, host := range []string{"open-upload.tiktokapis.com", "open-upload-us.tiktokapis.com", "open-upload-eu.tiktokapis.com", "open-upload-sg.tiktokapis.com"} {
		if err := tikTokSession(context.Background(), "ok", "v_inbox_file~v2.123", "https://"+host+"/video/?upload_token=fixture"); err != nil {
			t.Fatal(err)
		}
	}
	for _, address := range []string{
		"https://open-upload-sg.tiktokapis.com.evil.test/video/",
		"https://evil-open-upload-sg.tiktokapis.com/video/",
		"https://open-upload-sg.tiktokapis.com:8443/video/",
		"https://open-upload-sg.tiktokapis.com@evil.test/video/",
		"https://open-upload-sg.tiktokapis.com/video/#fragment",
		"http://open-upload-sg.tiktokapis.com/video/",
	} {
		if tikTokSession(context.Background(), "ok", "valid", address) == nil {
			t.Fatal("unsafe upload URL accepted")
		}
	}
}
