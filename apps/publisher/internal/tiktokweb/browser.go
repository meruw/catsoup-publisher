// Package tiktokweb prepares a post in a visible, isolated Chrome session.
// Publishing is always performed by the user in TikTok, never by this package.
package tiktokweb

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp"
)

const StudioURL = "https://www.tiktok.com/tiktokstudio/upload"

var handlePattern = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.]{1,23}$`)

func Handle(raw string) (string, error) {
	h := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(raw), "@"))
	if !handlePattern.MatchString(h) || strings.HasSuffix(h, ".") {
		return "", errors.New("enter the TikTok @username, not its display name")
	}
	return h, nil
}

func ProfileKey(language, account string) (string, error) {
	if (language != "es" && language != "en") || account == "" || len(account) > 200 {
		return "", errors.New("invalid TikTok destination")
	}
	return fmt.Sprintf("%s-%x", language, sha256.Sum256([]byte(account))), nil
}

type Session struct {
	mu     sync.Mutex
	ctx    context.Context
	cancel func()
	dir    string
	// Once attached, never navigate this tab or attach again. The user may be
	// reviewing or publishing in it, including after an ambiguous timeout.
	attached atomic.Bool
}
type Manager struct {
	mu        sync.Mutex
	root      string
	sessions  map[string]*Session
	studioURL string
	bindings  sync.Mutex
}

func New(root string) *Manager {
	return &Manager{root: root, sessions: map[string]*Session{}, studioURL: StudioURL}
}
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		s.cancel()
	}
}

func (m *Manager) session(language, account string) (*Session, error) {
	key, err := ProfileKey(language, account)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if s := m.sessions[key]; s != nil {
		if s.ctx.Err() == nil {
			return s, nil
		}
		s.cancel()
		delete(m.sessions, key)
	}
	dir := filepath.Join(m.root, key)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	// Use Chrome's normal security/keychain defaults. CDP travels through an
	// inherited pipe, with no listening debugging port and no personal profile.
	opts := browserOptions(filepath.Join(dir, "chrome"))
	alloc, stopAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, stopTab := chromedp.NewContext(alloc)
	s := &Session{ctx: ctx, cancel: func() { stopTab(); stopAlloc() }, dir: dir}
	if err := chromedp.Do(ctx); err != nil {
		s.cancel()
		return nil, errors.New("could not start Chrome; install Chrome and close leftover Publisher browser windows")
	}
	m.sessions[key] = s
	return s, nil
}

// Open is a setup action only. A user signs in in this browser, never in Publisher.
func (m *Manager) Open(language, account string) error {
	s, err := m.session(language, account)
	if err != nil {
		return err
	}
	if !s.mu.TryLock() {
		return errors.New("this TikTok browser is busy preparing a video")
	}
	defer s.mu.Unlock()
	if s.attached.Load() {
		return errors.New("this browser contains a prepared video; finish or discard it in TikTok before closing Publisher")
	}
	ctx, cancel := context.WithTimeout(s.ctx, 45*time.Second)
	defer cancel()
	err = chromedp.Do(ctx, chromedp.Navigate(m.studioURL))
	if err != nil {
		return errors.New("could not open TikTok Studio; install Chrome and close any leftover Publisher TikTok browser windows, then restart Publisher")
	}
	return nil
}

const menuSelector = `button[data-tt="Header_NewHeader_Clickable"][aria-haspopup="dialog"]`

// Read only the visible, first-party own-profile menu. Never infer identity from
// avatar/name, cached cookies, arbitrary page links, or the selected language.
const identityScript = `(() => {
 const visible=e=>!!(e.getClientRects().length) && getComputedStyle(e).visibility!=='hidden';
 const links=[...document.querySelectorAll('a[data-tt="Header_NewHeader_TUXMenuItem"]')].filter(visible);
 const names=links.map(e=>{try{const u=new URL(e.href);return u.origin==='https://www.tiktok.com'&&/^\/@[a-zA-Z0-9_.]+\/?$/.test(u.pathname)?u.pathname.replace(/^\/@/,'').replace(/\/$/,'').toLowerCase():''}catch{return ''}}).filter(Boolean);
 return names.length===1?names[0]:'';
})()`

func trustedPage(ctx context.Context, studioURL string) error {
	u, err := url.Parse(studioURL)
	if err != nil {
		return err
	}
	origin, _ := json.Marshal(u.Scheme + "://" + u.Host)
	ok, err := chromedp.Run(ctx, chromedp.Evaluate[bool](`location.origin===`+string(origin)+` && location.pathname.startsWith('/tiktokstudio/')`))
	if err != nil || !ok {
		return errors.New("the browser is not on TikTok Studio; sign in there before preparing the video")
	}
	return nil
}
func identity(ctx context.Context, studioURL string) (string, error) {
	if err := trustedPage(ctx, studioURL); err != nil {
		return "", err
	}
	if err := chromedp.Do(ctx, chromedp.WaitVisible(chromedp.CSS(menuSelector))); err != nil {
		return "", errors.New("sign in to TikTok in the dedicated browser, then verify the account")
	}
	expanded, err := chromedp.Run(ctx, chromedp.Evaluate[bool](`document.querySelector('`+menuSelector+`')?.getAttribute('aria-expanded')==='true'`))
	if err != nil {
		return "", errors.New("could not read TikTok's account menu")
	}
	if !expanded {
		if err = chromedp.Do(ctx, chromedp.Click(chromedp.CSS(menuSelector))); err != nil {
			return "", errors.New("could not open TikTok's account menu")
		}
	}
	current, err := chromedp.Run(ctx, chromedp.Poll[string](identityScript, chromedp.WithPollingInterval(250*time.Millisecond), chromedp.WithPollingTimeout(8*time.Second)))
	if err != nil || current == "" {
		return "", errors.New("could not verify the signed-in TikTok username; sign in in the dedicated browser and retry verification")
	}
	if err = chromedp.Do(ctx, chromedp.Click(chromedp.CSS(menuSelector))); err != nil {
		return "", errors.New("could not close TikTok's account menu")
	}
	return Handle(current)
}

func (m *Manager) Expected(language, account string) (string, error) {
	key, err := ProfileKey(language, account)
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(filepath.Join(m.root, key, "account.json"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var binding struct {
		Username string `json:"username"`
	}
	if json.Unmarshal(raw, &binding) != nil {
		return "", errors.New("browser account binding is damaged")
	}
	return Handle(binding.Username)
}

func (m *Manager) Verify(language, account, expected string) (string, error) {
	expected, err := Handle(expected)
	if err != nil {
		return "", err
	}
	s, err := m.session(language, account)
	if err != nil {
		return "", err
	}
	if !s.mu.TryLock() {
		return "", errors.New("this TikTok browser is busy preparing a video")
	}
	defer s.mu.Unlock()
	if s.attached.Load() {
		return "", errors.New("finish the prepared video before changing its browser account")
	}
	ctx, cancel := context.WithTimeout(s.ctx, 20*time.Second)
	defer cancel()
	actual, err := identity(ctx, m.studioURL)
	if err != nil {
		return "", err
	}
	if actual != expected {
		return "", fmt.Errorf("wrong TikTok account: expected @%s, found @%s; switch accounts in this browser", expected, actual)
	}
	m.bindings.Lock()
	defer m.bindings.Unlock()
	other := "es"
	if language == "es" {
		other = "en"
	}
	bindings, _ := filepath.Glob(filepath.Join(m.root, other+"-*", "account.json"))
	for _, path := range bindings {
		raw, e := os.ReadFile(path)
		if e != nil {
			return "", e
		}
		var saved struct {
			Username string `json:"username"`
		}
		if json.Unmarshal(raw, &saved) == nil && saved.Username == actual {
			return "", errors.New("this TikTok username is already assigned to the other language's browser session")
		}
	}
	raw, _ := json.Marshal(struct {
		Username string `json:"username"`
	}{actual})
	tmp := filepath.Join(s.dir, "account.json.tmp")
	if err = os.WriteFile(tmp, raw, 0600); err != nil {
		return "", err
	}
	if err = os.Rename(tmp, filepath.Join(s.dir, "account.json")); err != nil {
		return "", err
	}
	return actual, nil
}

// Prepare reserves durably only after identity verification and before the first
// possible upload byte. beforeAttach must revalidate approval and file integrity.
func (m *Manager) Prepare(language, account, path, caption string, beforeAttach func() error, progress func(string)) error {
	expected, err := m.Expected(language, account)
	if err != nil {
		return err
	}
	if expected == "" {
		return errors.New("set up this language's TikTok browser session in Accounts first")
	}
	s, err := m.session(language, account)
	if err != nil {
		return err
	}
	if !s.mu.TryLock() {
		return errors.New("this TikTok browser is already preparing a video")
	}
	defer s.mu.Unlock()
	if s.attached.Load() {
		return errors.New("a video is already prepared in this browser; finish it in TikTok and close that browser window before preparing another")
	}
	ctx, cancel := context.WithTimeout(s.ctx, 20*time.Minute)
	defer cancel()
	progress("Opening TikTok Studio")
	navCtx, navCancel := context.WithTimeout(ctx, 45*time.Second)
	err = chromedp.Do(navCtx, chromedp.Navigate(m.studioURL))
	navCancel()
	if err != nil {
		return errors.New("could not open TikTok Studio")
	}
	check, stop := context.WithTimeout(ctx, 20*time.Second)
	actual, err := identity(check, m.studioURL)
	stop()
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("wrong TikTok account: expected @%s, found @%s; no video attached", expected, actual)
	}
	progress("Account verified · @" + actual)
	inputCtx, inputCancel := context.WithTimeout(ctx, 15*time.Second)
	err = chromedp.Do(inputCtx, chromedp.WaitReady(chromedp.CSS(`input[type="file"][accept*="video"]`)))
	inputCancel()
	if err != nil {
		return errors.New("TikTok's upload control changed; no video attached")
	}
	if err = beforeAttach(); err != nil {
		return err
	}
	if err = trustedPage(ctx, m.studioURL); err != nil {
		return err
	}
	s.attached.Store(true)
	progress("Uploading in browser")
	if err = chromedp.Do(ctx, chromedp.SetUploadFiles(chromedp.CSS(`input[type="file"][accept*="video"]`), []string{path})); err != nil {
		return errors.New("check the dedicated TikTok browser; attachment was attempted and will not be repeated automatically")
	}
	progress("Preparing caption")
	if err = fillCaption(ctx, caption, m.studioURL); err != nil {
		return err
	}
	progress("Waiting for TikTok processing")
	if _, err = chromedp.Run(ctx, chromedp.Poll[bool](readyScript, chromedp.WithPollingInterval(time.Second), chromedp.WithPollingTimeout(18*time.Minute))); err != nil {
		return errors.New("TikTok has not confirmed the form is ready; check the open browser without uploading again")
	}
	if err = trustedPage(ctx, m.studioURL); err != nil {
		return err
	}
	if err = checkCaption(ctx, caption); err != nil {
		return err
	}
	return nil
}

// Support Draft.js/editor forms, fail closed if more than one
// visible editor exists. No button named Post/Publish is ever clicked.
const editorSelector = `[contenteditable="true"][role="textbox"], .public-DraftEditor-content[contenteditable="true"], .tiptap[contenteditable="true"]`

func fillCaption(ctx context.Context, caption, studioURL string) error {
	encoded, _ := json.Marshal(caption)
	script := `(() => {const es=[...new Set(document.querySelectorAll('` + editorSelector + `'))].filter(e=>e.getClientRects().length);if(es.length!==1)return false; const e=es[0];e.focus();const s=getSelection();const r=document.createRange();r.selectNodeContents(e);s.removeAllRanges();s.addRange(r);document.execCommand('insertText',false,` + string(encoded) + `);e.dispatchEvent(new InputEvent('input',{bubbles:true,inputType:'insertText'}));return e.innerText.replace(/\r/g,'').replace(/\n{2,}/g,'\n\n').trim()===` + string(encoded) + `.replace(/\r/g,'').replace(/\n{2,}/g,'\n\n').trim()})()`
	if _, err := chromedp.Run(ctx, chromedp.Poll[bool](`(()=>[...document.querySelectorAll('`+editorSelector+`')].filter(e=>e.getClientRects().length).length===1)()`, chromedp.WithPollingTimeout(2*time.Minute))); err != nil {
		return errors.New("video attached, but no unique caption editor was found; check TikTok without uploading again")
	}
	if err := trustedPage(ctx, studioURL); err != nil {
		return err
	}
	ok, err := chromedp.Run(ctx, chromedp.Evaluate[bool](script))
	if err != nil || !ok {
		return errors.New("video attached, but the caption could not be verified; check TikTok's editor without uploading again")
	}

	return nil
}

const readyScript = `(() => {const es=[...document.querySelectorAll('button')].filter(e=>e.getClientRects().length && /^(Post|Publish|Publicar)$/.test(e.innerText.trim()));return es.length===1&&!es[0].disabled&&es[0].getAttribute('aria-disabled')!=='true'})()`

func (m *Manager) HasPreparedBrowser() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.attached.Load() && s.ctx.Err() == nil {
			return true
		}
	}
	return false
}
func checkCaption(ctx context.Context, caption string) error {
	encoded, _ := json.Marshal(caption)
	ok, err := chromedp.Run(ctx, chromedp.Evaluate[bool](`(()=>{const es=[...document.querySelectorAll('`+editorSelector+`')].filter(e=>e.getClientRects().length);return es.length===1&&es[0].innerText.replace(/\r/g,'').replace(/\n{2,}/g,'\n\n').trim()===`+string(encoded)+`.replace(/\r/g,'').replace(/\n{2,}/g,'\n\n').trim()})()`))
	if err != nil || !ok {
		return errors.New("the caption changed while TikTok processed the file; review it in the browser")
	}
	return nil
}

// Keep the staged file available if a form/processing error occurs after Chrome
// has begun reading it. Cleanup waits for this isolated browser to close.
func (m *Manager) RemoveWhenClosed(language, account, path string) {
	key, err := ProfileKey(language, account)
	if err != nil {
		return
	}
	m.mu.Lock()
	s := m.sessions[key]
	m.mu.Unlock()
	if s != nil {
		go func() { <-s.ctx.Done(); os.Remove(path) }()
	}
}

func browserOptions(profile string) []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{chromedp.NoFirstRun, chromedp.NoDefaultBrowserCheck, chromedp.Flag("enable-automation", true), chromedp.UserDataDir(profile), chromedp.WindowSize(1200, 900)}
}
