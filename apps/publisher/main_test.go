package main

import (
	"net/http/httptest"
	"os"
	"testing"
	"testing/fstest"

	"catsoup/publisher/internal/delivery"
)

// Explicit opt-in: reads the generated synthetic sample without modifying it.
func TestRealDeliveryMedia(t *testing.T) {
	root := os.Getenv("CATSOUP_TEST_ROOT")
	if root == "" {
		t.Skip("set CATSOUP_TEST_ROOT for read-only integration coverage")
	}
	reader, err := delivery.New(root, delivery.PythonValidator{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	handler := assetHandler(reader, fstest.MapFS{"index.html": {Data: []byte("test UI")}})
	for _, video := range []string{"es_final", "en_final"} {
		request := httptest.NewRequest("GET", "/media?episode=catsoup-demo&video="+video, nil)
		request.Header.Set("Range", "bytes=0-31")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != 206 || response.Body.Len() != 32 || response.Header().Get("Content-Type") != "video/mp4" {
			t.Fatalf("%s: status=%d, bytes=%d", video, response.Code, response.Body.Len())
		}
		if response.Header().Get("Content-Security-Policy") == "" {
			t.Fatal("missing CSP")
		}
	}
	for _, test := range []struct {
		method, url string
		code        int
	}{
		{"POST", "/media", 405},
		{"GET", "/media?episode=../../secret&video=anything", 400},
		{"GET", "/media?episode=catsoup-demo&video=../../secret", 400},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(test.method, test.url, nil))
		if response.Code != test.code {
			t.Errorf("%s: got %d, want %d", test.url, response.Code, test.code)
		}
	}
}
