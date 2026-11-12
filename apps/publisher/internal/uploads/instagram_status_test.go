package uploads

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestInstagramUnsupportedVideoStatus(t *testing.T) {
	for _, state := range []string{"FINISHED", "IN_PROGRESS", "PUBLISHED", "ERROR", "EXPIRED"} {
		t.Run(state, func(t *testing.T) {
			c := New()
			calls := 0
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "GET" {
					t.Fatal("unexpected mutation")
				}
				if calls == 1 {
					r := response(`{"error":{"code":100,"error_subcode":2207065}}`)
					r.StatusCode = 400
					return r, nil
				}
				if calls > 2 || r.URL.Query().Get("fields") != "status_code,status" {
					t.Fatal("unexpected fallback")
				}
				return response(`{"status_code":"` + state + `","status":"Error: Media upload has failed with error code 2207076 https://private.invalid/token"}`), nil
			})
			got, err := c.InstagramContainerState(context.Background(), "fixture", "123")
			if got != state || calls != 2 {
				t.Fatal(got, calls, err)
			}
			terminal := state == "ERROR" || state == "EXPIRED"
			if errors.Is(err, ErrProcessingFailed) != terminal {
				t.Fatal(err)
			}
			if terminal && (!strings.Contains(err.Error(), "2207076") || strings.Contains(err.Error(), "private")) {
				t.Fatal(err)
			}
			if CanResumeInstagramBytes(got, err) {
				t.Fatal("URL container offered byte retransfer")
			}
		})
	}
}
func TestInstagramStatusDoesNotHideOtherProviderErrors(t *testing.T) {
	c := New()
	calls := 0
	c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
		calls++
		x := response(`{"error":{"code":190,"error_subcode":463}}`)
		x.StatusCode = 400
		return x, nil
	})
	_, err := c.InstagramContainerState(context.Background(), "fixture", "123")
	if err == nil || calls != 1 || errors.Is(err, ErrProcessingFailed) {
		t.Fatal(calls, err)
	}
}
