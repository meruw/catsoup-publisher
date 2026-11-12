package uploads

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGoogleAuthorizationFailuresAreActionableAndRedacted(t *testing.T) {
	for _, tc := range []struct{ body, code, hint string }{
		{`{"error":"invalid_grant","error_description":"SECRET token revoked"}`, "invalid_grant", "Accounts"},
		{`{"error":"invalid_client","error_description":"SECRET client"}`, "invalid_client", "Settings"},
		{`{"error":"invalid_scope"}`, "invalid_scope", "permissions"},
		{`{"error":"temporarily_unavailable"}`, "temporarily_unavailable", "later"},
		{`{"error":"SECRET","error_description":"SECRET"}`, "google_oauth_unknown", "configuration"},
		{`{"error":{"message":"SECRET"}}`, "google_oauth_unknown", "configuration"},
		{`SECRET not json`, "google_oauth_unknown", "configuration"},
	} {
		t.Run(tc.code+tc.hint, func(t *testing.T) {
			c := New()
			requests := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				requests++
				return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			_, err := c.request(context.Background(), "POST", "https://oauth2.googleapis.com/token", "", nil, strings.NewReader("refresh_token=SECRET"), -1, nil)
			if err == nil || !strings.Contains(err.Error(), tc.code) || !strings.Contains(err.Error(), tc.hint) || strings.Contains(err.Error(), "SECRET") || requests != 1 {
				t.Fatalf("unexpected authorization failure: %v, requests=%d", err, requests)
			}
		})
	}
}
