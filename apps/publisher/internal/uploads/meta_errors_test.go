package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMetaErrorIdentifiersAndPrivacy(t *testing.T) {
	for _, tc := range []struct {
		body          string
		code, subcode int
		kind          string
	}{
		{`{"error":{"code":100,"error_subcode":33,"message":"SECRET"}}`, 100, 33, ""},
		{`{"debug_info":{"type":"ProcessingFailedError","message":"{\"error\":{\"code\":-1,\"error_subcode\":2207053,\"message\":\"SECRET\"}}"}}`, -1, 2207053, "meta_processing_failed"},
		{`{"debug_info":{"type":"NotAuthorizedError","message":"SECRET"}}`, 0, 0, "meta_not_authorized"},
		{`{"debug_info":{"type":"SECRET","message":"https://secret.test"}}`, 0, 0, ""},
		{`<html>SECRET</html>`, 0, 0, ""},
	} {
		for _, host := range []string{"graph.facebook.com", "rupload.facebook.com"} {
			s := diagnostics.Open(t.TempDir())
			ctx := diagnostics.StartTransfer(context.Background(), s, "instagram", "en")
			c := New()
			c.HTTP.Transport = transport(func(*http.Request) (*http.Response, error) { r := response(tc.body); r.StatusCode = 400; return r, nil })
			_, err := c.request(ctx, "POST", "https://"+host+"/fixture", "Bearer SECRET", nil, nil, 0, nil)
			if err == nil || strings.Contains(err.Error(), "SECRET") {
				t.Fatal(err)
			}
			result, _ := s.Query(diagnostics.Query{Level: "ERROR"})
			if len(result.Events) != 1 {
				t.Fatal(result)
			}
			e := result.Events[0]
			if e.ProviderCode != tc.code || e.ProviderSubcode != tc.subcode || e.ProviderErrorCode != tc.kind || e.HTTPStatus != 400 {
				t.Fatal(e)
			}
			raw, _ := json.Marshal(result)
			if strings.Contains(string(raw), "SECRET") || strings.Contains(string(raw), "secret.test") {
				t.Fatal("response text leaked")
			}
			s.Close()
		}
	}
}

func TestInstagramDetailedPhaseFailure(t *testing.T) {
	for _, phase := range []string{"uploading_phase", "processing_phase"} {
		for _, state := range []string{"IN_PROGRESS", "ERROR", "FINISHED", "PUBLISHED"} {
			t.Run(phase+"/"+state, func(t *testing.T) {
				s := diagnostics.Open(t.TempDir())
				defer s.Close()
				ctx := diagnostics.StartTransfer(context.Background(), s, "instagram", "es")
				c := New()
				c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
					if r.Method != "GET" || r.URL.Query().Get("fields") != "status_code,video_status" {
						t.Fatal("missing detailed read")
					}
					b, _ := json.Marshal(map[string]any{"status_code": state, "video_status": map[string]any{phase: map[string]any{"status": "error", "errors": []any{map[string]any{"code": 1363008, "message": "SECRET"}}}}})
					return response(string(b)), nil
				})
				got, err := c.instagramContainerStatus(ctx, "Bearer SECRET", "container1")
				failed := state != "FINISHED" && state != "PUBLISHED"
				if got != state || errors.Is(err, ErrProcessingFailed) != failed {
					t.Fatal(got, err)
				}
				if err != nil && strings.Contains(err.Error(), "SECRET") {
					t.Fatal("raw error leaked")
				}
				level, stage := "ERROR", "upload_processing_failed"
				if state == "IN_PROGRESS" {
					level, stage = "WARN", "upload_processing_pending"
					errorsOnly, _ := s.Query(diagnostics.Query{Level: "ERROR"})
					if len(errorsOnly.Events) != 0 {
						t.Fatal("pending processing reported as a confirmed failure", errorsOnly)
					}
				}
				result, _ := s.Query(diagnostics.Query{Level: level})
				if failed && (len(result.Events) != 1 || result.Events[0].ProviderCode != 1363008 || result.Events[0].Stage != stage) {
					t.Fatal(result)
				}
				raw, _ := json.Marshal(result)
				if strings.Contains(string(raw), "SECRET") {
					t.Fatal("raw error logged")
				}
			})
		}
	}
}

func TestInstagramReconcilesSavedContainer(t *testing.T) {
	for _, status := range []string{"FINISHED", "IN_PROGRESS", "PROGRESS_THEN_FINISHED", "PROGRESS_THEN_ERROR", "ERROR", "PUBLISHED", "EXPIRED", "", "HTTP_ERROR"} {
		t.Run(status, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "video")
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			if _, err = f.WriteString("video"); err != nil {
				t.Fatal(err)
			}
			c := New()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			uploads, creates, publishes, reads := 0, 0, 0, 0
			saved := ""
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				switch {
				case r.URL.Host == "rupload.facebook.com":
					if saved != "container1" {
						t.Fatal("missing checkpoint")
					}
					uploads++
					io.Copy(io.Discard, r.Body)
					res := response(`{"debug_info":{"type":"ProcessingFailedError","message":"SECRET"}}`)
					res.StatusCode = 400
					return res, nil
				case strings.HasSuffix(r.URL.Path, "/media"):
					creates++
					return response(`{"id":"container1","uri":"https://rupload.facebook.com/ig-api-upload/container1"}`), nil
				case strings.HasSuffix(r.URL.Path, "/media_publish"):
					if (status != "FINISHED" && status != "PROGRESS_THEN_FINISHED") || reads < 1 {
						t.Fatal("published unconfirmed upload")
					}
					var body map[string]string
					json.NewDecoder(r.Body).Decode(&body)
					if body["creation_id"] != saved {
						t.Fatal("container changed")
					}
					publishes++
					return response(`{"id":"reel1"}`), nil
				case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/container1"):
					reads++
					if status == "HTTP_ERROR" {
						res := response(`{"error":{"code":190}}`)
						res.StatusCode = 400
						return res, nil
					}
					next := status
					if strings.HasPrefix(status, "PROGRESS_THEN_") {
						next = "IN_PROGRESS"
						if reads > 1 {
							next = strings.TrimPrefix(status, "PROGRESS_THEN_")
						}
					}
					if status == "IN_PROGRESS" && reads > 1 {
						cancel()
					}
					b, _ := json.Marshal(map[string]string{"status_code": next})
					return response(string(b)), nil
				default:
					t.Fatal("unexpected request")
					return nil, nil
				}
			})
			ref, err := c.instagram(ctx, "SECRET", Account{ID: "account1"}, Media{File: f, Size: 5}, func(ref string) error { saved = ref; return nil })
			if creates != 1 || uploads != 1 {
				t.Fatal("duplicate transfer", creates, uploads)
			}
			if status == "FINISHED" || status == "PROGRESS_THEN_FINISHED" {
				if err != nil || ref != "reel1" || publishes != 1 {
					t.Fatal(ref, err, publishes)
				}
			} else if err == nil || publishes != 0 || ref != "" {
				t.Fatal("unsafe recovery", ref, err, publishes)
			}
		})
	}
}

func TestInstagramProcessingWait(t *testing.T) {
	for _, final := range []string{"FINISHED", "ERROR", "PUBLISHED", "EXPIRED", ""} {
		t.Run(final, func(t *testing.T) {
			c := New()
			reads := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				if r.Method != "GET" || !strings.HasSuffix(r.URL.Path, "/container1") {
					t.Fatal("unexpected mutation or container", r.Method, r.URL.Path)
				}
				reads++
				state := "IN_PROGRESS"
				if reads == 3 {
					state = final
				}
				b, _ := json.Marshal(map[string]string{"status_code": state})
				return response(string(b)), nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := c.waitInstagramContainer(ctx, "Bearer SECRET", "container1", time.Millisecond)
			if reads != 3 || (err == nil) != (final == "FINISHED") {
				t.Fatal(reads, err)
			}
		})
	}
}
