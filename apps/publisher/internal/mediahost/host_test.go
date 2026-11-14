package mediahost

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSingleFileHandler(t *testing.T) {
	f, e := os.CreateTemp(t.TempDir(), "video")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	f.WriteString("0123456789approved-video")
	route := "/" + strings.Repeat("a", 48) + "/video.mp4"
	h := Handler(f, 23, route)
	for _, path := range []string{"/", "/credentials", route + "?token=foo", route + "/../other"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatal(path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", route, nil)
	r.Header.Set("Range", "bytes=2-5")
	h.ServeHTTP(w, r)
	if w.Code != 206 || w.Body.String() != "2345" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", route, nil))
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("HEAD", route, nil))
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatal(w)
	}
}

func TestHostRejectsMissingHelper(t *testing.T) {
	_, close, e := Start(context.Background(), filepath.Join(t.TempDir(), "absent.exe"), nil, 1)
	close()
	if e == nil {
		t.Fatal("accepted missing input")
	}
}

// Opt-in checks the production tunnel lifecycle with synthetic bytes only.
func TestLiveSyntheticTunnel(t *testing.T) {
	exe := os.Getenv("CATSOUP_TEST_CLOUDFLARED")
	if exe == "" {
		t.Skip("explicit synthetic network check only")
	}
	f, e := os.CreateTemp(t.TempDir(), "synthetic")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	payload := strings.Repeat("synthetic-video-fixture-", 10)
	f.WriteString(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	url, close, e := Start(ctx, exe, f, int64(len(payload)))
	if e != nil {
		t.Fatal(e)
	}
	defer close()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal("public fetch failed")
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(b) != payload {
		t.Fatal("incorrect bytes")
	}
	close()
	close()
	t.Log("Synthetic HTTPS verified and closed")
}
