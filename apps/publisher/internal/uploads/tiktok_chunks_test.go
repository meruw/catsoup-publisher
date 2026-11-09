package uploads

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTikTokDeclaredChunksMatchTransfer(t *testing.T) {
	for _, tc := range []struct{ size, chunk, count int64 }{
		{1024, 1024, 1},
		{5 << 20, 5 << 20, 1},
		{16 << 20, 16 << 20, 1},
		{(16 << 20) + 1, (16 << 20) + 1, 1},
		{31007917, 31007917, 1},
		{(32 << 20) - 1, (32 << 20) - 1, 1},
		{32 << 20, 16 << 20, 2},
		{(32 << 20) + 1, 16 << 20, 2},
		{65 << 20, 16 << 20, 4},
	} {
		t.Run(fmt.Sprint(tc.size), func(t *testing.T) {
			file, err := os.Create(filepath.Join(t.TempDir(), "video.mp4"))
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			if err := file.Truncate(tc.size); err != nil {
				t.Fatal(err)
			}
			var bytesSent, chunksSent int64
			checkpoint := false
			c := New()
			c.HTTP.Transport = transport(func(r *http.Request) (*http.Response, error) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/inbox/video/init/"):
					var payload struct {
						Source struct {
							Method string `json:"source"`
							Size   int64  `json:"video_size"`
							Chunk  int64  `json:"chunk_size"`
							Count  int64  `json:"total_chunk_count"`
						} `json:"source_info"`
					}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Fatal(err)
					}
					s := payload.Source
					if s.Method != "FILE_UPLOAD" || s.Size != tc.size || s.Chunk != tc.chunk || s.Count != tc.count {
						t.Fatalf("incorrect upload declaration: %+v, expected %+v", s, tc)
					}
					return response(`{"data":{"publish_id":"fixture","upload_url":"https://open-upload-sg.tiktokapis.com/video/"},"error":{"code":"ok"}}`), nil
				case r.Method == "PUT":
					if !checkpoint {
						t.Fatal("video sent before saving the receipt")
					}
					n, err := io.Copy(io.Discard, r.Body)
					if err != nil {
						t.Fatal(err)
					}
					wantSize := tc.chunk
					if chunksSent == tc.count-1 {
						wantSize = tc.size - bytesSent
					}
					wantRange := fmt.Sprintf("bytes %d-%d/%d", bytesSent, bytesSent+n-1, tc.size)
					if n != wantSize || r.ContentLength != n || r.Header.Get("Content-Range") != wantRange {
						t.Fatalf("incorrect chunk: size=%d length=%d range=%q", n, r.ContentLength, r.Header.Get("Content-Range"))
					}
					bytesSent += n
					chunksSent++
					return response(""), nil
				case strings.HasSuffix(r.URL.Path, "/status/fetch/"):
					return response(`{"data":{"status":"SEND_TO_USER_INBOX"},"error":{"code":"ok"}}`), nil
				default:
					t.Fatalf("unexpected request: %s", r.URL.Path)
					return nil, nil
				}
			})
			_, err = c.tiktok(context.Background(), "fixture", Media{File: file, Size: tc.size}, func(string) error { checkpoint = true; return nil })
			if err != nil || bytesSent != tc.size || chunksSent != tc.count {
				t.Fatalf("transfer: bytes=%d chunks=%d err=%v", bytesSent, chunksSent, err)
			}
		})
	}
}
