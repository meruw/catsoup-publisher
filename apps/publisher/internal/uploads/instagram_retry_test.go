package uploads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const emptyInstagramUpload = `{"status_code":"ERROR","video_status":{"uploading_phase":{"status":"error","bytes_transferred":0,"source_file_size":0,"errors":[{"code":1363008}]},"processing_phase":{"status":"error"}}}`

func TestInstagramRecoveryNeverDuplicatesPublish(t *testing.T) {
	for _, scenario := range []string{"recover", "exhausted", "late_success", "unknown", "published", "read_error", "cancel", "publish_error", "checkpoint_error", "other_code", "nonempty", "missing_sizes"} {
		t.Run(scenario, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "video")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			payload := strings.Repeat("approved video", 10000)
			if _, err = f.WriteString(payload); err != nil {
				t.Fatal(err)
			}
			c := New()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			creates, sends, publishes, reads := 0, 0, 0, 0
			saved := ""
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/media"):
					creates++
					return response(fmt.Sprintf(`{"id":"c%d","uri":"https://rupload.facebook.com/c%d"}`, creates, creates)), nil
				case r.URL.Host == "rupload.facebook.com":
					if saved != fmt.Sprintf("c%d", creates) {
						t.Fatal("no durable checkpoint")
					}
					b, e := io.ReadAll(r.Body)
					if e != nil || string(b) != payload || r.ContentLength != int64(len(payload)) || r.Header.Get("file_size") != fmt.Sprint(len(payload)) || r.Header.Get("offset") != "0" || r.Header.Get("Authorization") != "OAuth fixture" {
						t.Fatal("incorrect bytes or upload headers")
					}
					sends++
					if creates == 1 || scenario == "exhausted" {
						res := response(`{"debug_info":{"type":"ProcessingFailedError"}}`)
						res.StatusCode = 400
						return res, nil
					}
					return response(`{"success":true}`), nil
				case strings.HasSuffix(r.URL.Path, "/media_publish"):
					publishes++
					var b map[string]string
					json.NewDecoder(r.Body).Decode(&b)
					if b["creation_id"] != saved {
						t.Fatal("wrong container")
					}
					if scenario == "publish_error" {
						return nil, errors.New("connection lost")
					}
					return response(`{"id":"reel"}`), nil
				case r.Method == "GET":
					reads++
					if creates > 1 && scenario != "exhausted" {
						return response(`{"status_code":"FINISHED"}`), nil
					}
					if reads == 1 {
						if scenario == "cancel" {
							cancel()
						}
						body := emptyInstagramUpload
						switch scenario {
						case "other_code":
							body = strings.Replace(body, "1363008", "999", 1)
						case "nonempty":
							body = strings.Replace(body, `"bytes_transferred":0`, `"bytes_transferred":10`, 1)
						case "missing_sizes":
							body = strings.Replace(body, `"bytes_transferred":0,"source_file_size":0,`, "", 1)
						}
						return response(body), nil
					}
					switch scenario {
					case "late_success":
						return response(`{"status_code":"FINISHED"}`), nil
					case "unknown":
						return response(`{}`), nil
					case "published":
						return response(`{"status_code":"PUBLISHED"}`), nil
					case "read_error":
						return nil, errors.New("offline")
					}
					return response(emptyInstagramUpload), nil
				default:
					t.Fatal("unexpected request")
					return nil, nil
				}
			})
			ref, err := c.instagramWithRetry(ctx, "fixture", Account{ID: "account"}, Media{File: f, Size: int64(len(payload))}, func(ref string) error {
				if scenario == "checkpoint_error" && ref == "c2" {
					return errors.New("disk failure")
				}
				saved = ref
				return nil
			}, time.Millisecond)
			wantCreates, wantSends, wantPublishes := 1, 1, 0
			switch scenario {
			case "recover", "publish_error":
				wantCreates, wantSends, wantPublishes = 2, 2, 1
			case "exhausted":
				wantCreates, wantSends = 2, 2
			case "late_success":
				wantPublishes = 1
			case "checkpoint_error":
				wantCreates = 2
			}
			if creates != wantCreates || sends != wantSends || publishes != wantPublishes {
				t.Fatal(creates, sends, publishes)
			}
			success := scenario == "recover" || scenario == "late_success"
			if (err == nil) != success || (success && ref != "reel") {
				t.Fatal(ref, err)
			}
		})
	}
}

func TestInstagramFailureConfirmation(t *testing.T) {
	for _, state := range []string{emptyInstagramUpload, `{"status_code":"FINISHED"}`, `{"status_code":"PUBLISHED"}`, `{"status_code":"IN_PROGRESS"}`, strings.Replace(emptyInstagramUpload, `"ERROR"`, `"IN_PROGRESS"`, 1), `{}`} {
		c := New()
		c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
			if r.Method != "GET" {
				t.Fatal("mutation")
			}
			return response(state), nil
		})
		if err := c.ConfirmInstagramFailure(context.Background(), "fixture", "container"); (err == nil) != (state == emptyInstagramUpload) {
			t.Fatal(state, err)
		}
	}
}

func TestInstagramPendingPhaseErrorCanFinish(t *testing.T) {
	c := New()
	reads, publishes := 0, 0
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		if r.Method == "GET" {
			reads++
			state := "IN_PROGRESS"
			if reads > 1 {
				state = "FINISHED"
			}
			// Meta retains the phase error even after top-level success.
			return response(strings.Replace(emptyInstagramUpload, `"ERROR"`, `"`+state+`"`, 1)), nil
		}
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/account/media_publish") {
			t.Fatal("unexpected mutation", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["creation_id"] != "original" {
			t.Fatal(body, err)
		}
		publishes++
		return response(`{"id":"reel"}`), nil
	})
	if err := c.waitInstagramContainer(context.Background(), "Bearer fixture", "original", time.Millisecond); err != nil {
		t.Fatal(err)
	}
	ref, err := c.PublishReadyInstagram(context.Background(), "fixture", "account", "original")
	if err != nil || ref != "reel" || reads != 3 || publishes != 1 {
		t.Fatal(ref, err, reads, publishes)
	}
}

func TestInstagramResumeRequiresReadyContainer(t *testing.T) {
	for _, state := range []string{"IN_PROGRESS", "PUBLISHED", "ERROR", "EXPIRED", ""} {
		t.Run(state, func(t *testing.T) {
			c := New()
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.Method != "GET" {
					t.Fatal("must not publish or re-upload")
				}
				return response(fmt.Sprintf(`{"status_code":%q}`, state)), nil
			})
			if _, err := c.PublishReadyInstagram(context.Background(), "fixture", "account", "original"); err == nil {
				t.Fatal("accepted non-ready container")
			}
		})
	}
}
