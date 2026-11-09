package uploads

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestStatisticsCountersAndOwnership(t *testing.T) {
	for _, tc := range []struct {
		name, platform, id, account, body string
		want                              int
		fail                              bool
	}{
		{"youtube", "youtube", "abcdefghijk", "channel", `{"items":[{"id":"abcdefghijk","snippet":{"channelId":"channel"},"statistics":{"viewCount":"9007199254740993","likeCount":"0"}}]}`, 2, false},
		{"instagram", "instagram", "123", "456", `{"id":"123","owner":{"id":"456"},"like_count":0,"comments_count":12}`, 2, false},
		{"facebook", "facebook", "123", "456", `{"id":"123","from":{"id":"456"},"likes":{"summary":{"total_count":4}},"comments":{"summary":{"total_count":0}}}`, 2, false},
		{"missing is not zero", "instagram", "123", "456", `{"id":"123","owner":{"id":"456"},"like_count":null,"comments_count":2}`, 1, false},
		{"other owner", "instagram", "123", "456", `{"id":"123","owner":{"id":"789"},"like_count":1}`, 0, true},
		{"other channel", "youtube", "abcdefghijk", "channel", `{"items":[{"id":"abcdefghijk","snippet":{"channelId":"other"},"statistics":{"viewCount":"1"}}]}`, 0, true},
		{"deleted video", "youtube", "abcdefghijk", "channel", `{"items":[]}`, 0, true},
		{"missing counters", "instagram", "123", "456", `{"id":"123","owner":{"id":"456"}}`, 0, true},
		{"negative", "instagram", "123", "456", `{"id":"123","owner":{"id":"456"},"like_count":-1}`, 0, true},
		{"fractional", "instagram", "123", "456", `{"id":"123","owner":{"id":"456"},"like_count":1.2}`, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := New()
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.Method != "GET" || r.Header.Get("Authorization") != "Bearer fixture" || r.URL.Query().Has("access_token") {
					t.Fatal("unsafe metrics request")
				}
				if r.URL.Host != "www.googleapis.com" && r.URL.Host != "graph.facebook.com" {
					t.Fatal("unexpected host")
				}
				return response(tc.body), nil
			})
			got, err := c.Statistics(context.Background(), tc.platform, "fixture", tc.account, tc.id)
			if (err != nil) != tc.fail {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.fail && (len(got.Items) != tc.want || got.UpdatedAt == "") {
				t.Fatalf("missing result: %+v", got)
			}
			if tc.name == "youtube" && got.Items[0].Value != "9007199254740993" {
				t.Fatal("counter lost precision")
			}
		})
	}
}

func TestStatisticsRejectsArbitraryRequestsAndRedactsErrors(t *testing.T) {
	for _, platform := range []string{"youtube", "instagram", "facebook", "x", "tiktok"} {
		c := New()
		c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
			t.Fatal("invalid reference caused network access")
			return nil, nil
		})
		if _, err := c.Statistics(context.Background(), platform, "secret", "123", "https://attacker.invalid/?token=secret"); err == nil {
			t.Fatal("accepted URL")
		}
	}
	c := New()
	c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) {
		r := response(`{"error":{"message":"secret","code":190}}`)
		r.StatusCode = 403
		return r, nil
	})
	_, err := c.Statistics(context.Background(), "instagram", "secret", "456", "123")
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("provider failure not safely reported")
	}
}
