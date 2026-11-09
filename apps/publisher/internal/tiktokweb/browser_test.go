package tiktokweb

import (
	"context"
	"github.com/chromedp/chromedp"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProfileIsolationAndHandles(t *testing.T) {
	a, _ := ProfileKey("es", "same")
	b, _ := ProfileKey("en", "same")
	c, _ := ProfileKey("es", "different")
	if a == b || a == c || strings.Contains(a, "same") {
		t.Fatal("profile identity leaked or collided")
	}
	for _, v := range []string{"../escape", "CAT soup!", "https://tiktok.com/@catsoup", "bad.", ""} {
		if _, err := Handle(v); err == nil {
			t.Fatal("invalid handle", v)
		}
	}
	if h, err := Handle(" @CATsoup_es "); err != nil || h != "catsoup_es" {
		t.Fatal(h, err)
	}
}

// This fixture never contacts TikTok, never uses real cookies or videos, and
// counts Publish clicks to enforce the manual final action.
func TestBrowserPreparationFixture(t *testing.T) {
	if os.Getenv("CATSOUP_BROWSER_TEST") != "1" {
		t.Skip("set CATSOUP_BROWSER_TEST=1 to run isolated Chrome fixture")
	}
	const html = `<!doctype html><html><body>
 <button data-tt="Header_NewHeader_Clickable" aria-haspopup="dialog" aria-expanded="false" onclick="const x=this.getAttribute('aria-expanded')==='true';this.setAttribute('aria-expanded',String(!x));document.querySelector('a').hidden=x">Account</button>
 <a data-tt="Header_NewHeader_TUXMenuItem" href="https://www.tiktok.com/@fixture_es" hidden>Profile</a>
 <input type="file" accept="video/*" onchange="window.attached=(window.attached||0)+1;document.querySelector('#editor').hidden=false;document.querySelector('#post').disabled=false">
 <div id="editor" role="textbox" contenteditable="true" hidden></div>
 <button id="post" disabled onclick="window.published=(window.published||0)+1">Publish</button>
 </body></html>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(html)) }))
	defer server.Close()
	m := New(t.TempDir())
	m.studioURL = server.URL + "/tiktokstudio/upload"
	key, _ := ProfileKey("es", "fixture-id")
	dir := filepath.Join(m.root, key)
	os.MkdirAll(dir, 0700)
	opts := append(browserOptions(filepath.Join(dir, "chrome")), chromedp.Flag("headless", true))
	alloc, stop := chromedp.NewExecAllocator(context.Background(), opts...)
	defer stop()
	ctx, cancel := chromedp.NewContext(alloc)
	defer cancel()
	if err := chromedp.Do(ctx); err != nil {
		t.Fatal(err)
	}
	m.sessions[key] = &Session{ctx: ctx, cancel: cancel, dir: dir}
	if err := m.Open("es", "fixture-id"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Verify("es", "fixture-id", "fixture_en"); err == nil {
		t.Fatal("wrong account verified")
	}
	if _, err := m.Verify("es", "fixture-id", "fixture_es"); err != nil {
		t.Fatal(err)
	}
	if expected, err := m.Expected("es", "fixture-id"); err != nil || expected != "fixture_es" {
		t.Fatal(expected, err)
	}
	file := filepath.Join(t.TempDir(), "fixture.mp4")
	os.WriteFile(file, []byte("fixture, not a real video"), 0600)
	reserved := 0
	caption := "Title\n\nA description with 'quotes' and á\n\n#roblox #cats"
	os.WriteFile(filepath.Join(dir, "account.json"), []byte(`{"username":"fixture_en"}`), 0600)
	if err := m.Prepare("es", "fixture-id", file, caption, func() error { reserved++; return nil }, func(string) {}); err == nil || reserved != 0 {
		t.Fatal("wrong identity reached attachment", err, reserved)
	}
	os.WriteFile(filepath.Join(dir, "account.json"), []byte(`{"username":"fixture_es"}`), 0600)
	if err := m.Prepare("es", "fixture-id", file, caption, func() error { reserved++; return nil }, func(string) {}); err != nil {
		actual, _ := chromedp.Run(ctx, chromedp.Evaluate[string](`document.querySelector("#editor").outerHTML+"|"+JSON.stringify(document.querySelector("#editor").innerText)`))
		t.Log(actual)
		t.Fatal(err)
	}
	if reserved != 1 {
		t.Fatal("reservation count", reserved)
	}
	if err := m.Prepare("es", "fixture-id", file, caption, func() error { reserved++; return nil }, func(string) {}); err == nil {
		t.Fatal("attached twice")
	}
	check, stopCheck := context.WithTimeout(ctx, 5*time.Second)
	defer stopCheck()
	counts, err := chromedp.Run(check, chromedp.Evaluate[[]int](`[window.attached||0,window.published||0]`))
	if err != nil || len(counts) != 2 || counts[0] != 1 || counts[1] != 0 {
		t.Fatal(counts, err)
	}
	if !m.HasPreparedBrowser() {
		t.Fatal("missing close protection")
	}
}
