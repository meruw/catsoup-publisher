package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestInstagramDiagnosticFailureKeepsOnlyNumericProviderError(t *testing.T) {
	logs := diagnostics.Open(t.TempDir())
	defer logs.Close()
	ctx := diagnostics.Start(context.Background(), logs, "instagram", "es")
	client := New()
	client.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 403, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"SECRET_ACCESS_TOKEN","code":200,"error_subcode":123}}`))}, nil
	})
	_, err := client.InstagramPage(ctx, Token{Access: "SECRET_ACCESS_TOKEN"}, "PRIVATE_USERNAME")
	if err == nil {
		t.Fatal("expected provider failure")
	}
	result, err := logs.Query(diagnostics.Query{})
	raw, _ := json.Marshal(result)
	if err != nil || strings.Contains(string(raw), "SECRET") || strings.Contains(string(raw), "PRIVATE") {
		t.Fatal("sensitive data in diagnostics", err)
	}
	for _, event := range result.Events {
		if event.Stage == "provider_failed" && event.HTTPStatus == 403 && event.ProviderCode == 200 && event.ProviderSubcode == 123 {
			return
		}
	}
	t.Fatal("provider diagnosis missing")
}

func TestInstagramResolutionLogsEmptyPagesAndMismatch(t *testing.T) {
	for _, test := range []struct{ body, reason string }{
		{`{"data":[]}`, "no_pages"},
		{`{"data":[{"access_token":"SECRET","instagram_business_account":{"username":"different"}}]}`, "username_mismatch"},
	} {
		logs := diagnostics.Open(t.TempDir())
		ctx := diagnostics.Start(context.Background(), logs, "instagram", "en")
		client := New()
		client.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(test.body))}, nil
		})
		_, err := client.InstagramPage(ctx, Token{Access: "SECRET"}, "expected")
		result, _ := logs.Query(diagnostics.Query{Level: "ERROR"})
		logs.Close()
		if err == nil || len(result.Events) != 1 || result.Events[0].Reason != test.reason {
			t.Fatal("incorrect resolution diagnosis", result)
		}
	}
}

func TestGoogleAuthorizationDiagnosticKeepsOnlyKnownCode(t *testing.T) {
	logs := diagnostics.Open(t.TempDir())
	defer logs.Close()
	ctx := diagnostics.StartTransfer(context.Background(), logs, "youtube", "es")
	c := New()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"error":"invalid_grant","error_description":"SECRET"}`))}, nil
	})
	_, err := c.request(ctx, "POST", "https://oauth2.googleapis.com/token", "", nil, nil, 0, nil)
	if err == nil {
		t.Fatal("expected failure")
	}
	result, e := logs.Query(diagnostics.Query{Level: "ERROR"})
	raw, _ := json.Marshal(result)
	if e != nil || len(result.Events) != 1 || result.Events[0].ProviderErrorCode != "invalid_grant" || result.Events[0].HTTPStatus != 400 || strings.Contains(string(raw), "SECRET") {
		t.Fatalf("unsafe or missing diagnosis: %s %v", raw, e)
	}
}
