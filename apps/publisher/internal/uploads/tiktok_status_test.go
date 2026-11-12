package uploads

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTikTokStatusIsReadOnlyAndValidatesResponses(t *testing.T) {
	for _, status := range []string{"PROCESSING_UPLOAD", "SEND_TO_USER_INBOX", "PUBLISH_COMPLETE", "FAILED", "unexpected"} {
		c := New()
		calls := 0
		c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
			calls++
			b, _ := io.ReadAll(r.Body)
			if r.URL.String() != tik+"post/publish/status/fetch/" || r.Method != "POST" || !strings.Contains(string(b), "v_inbox_file~v2.123") {
				t.Fatal("not a status query")
			}
			return response(`{"data":{"status":"` + status + `","fail_reason":"https://secret"},"error":{"code":"ok"}}`), nil
		})
		got, err := c.TikTokInboxStatus(context.Background(), "token", "v_inbox_file~v2.123")
		if (err != nil) != (status == "unexpected") || calls != 1 {
			t.Fatal(got, err, calls)
		}
		if err == nil && got.FailReason != "provider_processing_failed" {
			t.Fatal("unsafe reason")
		}
		if _, err = c.TikTokInboxStatus(context.Background(), "token", "https://other"); err == nil || calls != 1 {
			t.Fatal("invalid receipt was requested")
		}
	}
}
