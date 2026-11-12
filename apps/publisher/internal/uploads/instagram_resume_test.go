package uploads

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestResumeInstagramBytesUsesOnlyOriginalContainer(t *testing.T) {
	for _, scenario := range []string{"empty", "ready", "published", "healthy", "nonempty", "missing_sizes", "network", "rejected"} {
		t.Run(scenario, func(t *testing.T) {
			f, e := os.CreateTemp(t.TempDir(), "video")
			if e != nil {
				t.Fatal(e)
			}
			defer f.Close()
			f.WriteString("approved")
			c := New()
			reads, sends := 0, 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.Method == "POST" {
					if r.URL.String() != "https://rupload.facebook.com/ig-api-upload/v25.0/original" {
						t.Fatalf("unexpected mutation: %s", r.URL)
					}
					sends++
					b, _ := io.ReadAll(r.Body)
					if string(b) != "approved" || r.Header.Get("offset") != "0" || r.Header.Get("file_size") != "8" || r.Header.Get("Authorization") != "OAuth fixture" {
						t.Fatal("wrong transport")
					}
					if scenario == "rejected" {
						x := response(`{}`)
						x.StatusCode = 400
						return x, nil
					}
					return response(`{"success":true}`), nil
				}
				if r.URL.Path != "/v25.0/original" {
					t.Fatal("wrong container")
				}
				reads++
				body := strings.Replace(emptyInstagramUpload, `"ERROR"`, `"IN_PROGRESS"`, 1)
				if sends > 0 && scenario != "rejected" || scenario == "ready" {
					body = `{"status_code":"FINISHED"}`
				}
				switch scenario {
				case "published":
					body = `{"status_code":"PUBLISHED"}`
				case "healthy":
					body = `{"status_code":"IN_PROGRESS"}`
				case "nonempty":
					body = strings.Replace(body, `"bytes_transferred":0`, `"bytes_transferred":10`, 1)
				case "missing_sizes":
					body = strings.Replace(body, `"bytes_transferred":0,"source_file_size":0,`, "", 1)
				case "network":
					x := response(`{}`)
					x.StatusCode = 503
					return x, nil
				}
				return response(body), nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			e = c.resumeInstagramBytes(ctx, "fixture", "original", Media{File: f, Size: 8}, time.Millisecond)
			if (e == nil) != (scenario == "empty" || scenario == "ready") {
				t.Fatal(e)
			}
			want := 0
			if scenario == "empty" || scenario == "rejected" {
				want = 1
			}
			if sends != want {
				t.Fatalf("sent %d times", sends)
			}
		})
	}
}
