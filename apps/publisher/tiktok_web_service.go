package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"catsoup/publisher/internal/diagnostics"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/tiktokweb"
)

type TikTokWebService struct {
	uploads  *UploadService
	browser  *tiktokweb.Manager
	mu       sync.Mutex
	progress map[string]string
}

func (s *TikTokWebService) account(language, id string) (drafts.Profile, error) {
	profiles, err := s.uploads.Accounts()
	if err != nil {
		return drafts.Profile{}, err
	}
	for _, p := range profiles {
		if p.Platform == "tiktok" && p.Language == language && p.ID == id {
			return p, nil
		}
	}
	return drafts.Profile{}, errors.New("TikTok destination changed; reload Accounts")
}

func (s *TikTokWebService) BrowserUsername(language, id string) (string, error) {
	if _, err := s.account(language, id); err != nil {
		return "", err
	}
	return s.browser.Expected(language, id)
}
func (s *TikTokWebService) OpenBrowser(language, id string) error {
	if _, err := s.account(language, id); err != nil {
		return err
	}
	return s.browser.Open(language, id)
}
func (s *TikTokWebService) VerifyBrowser(language, id, username string) (string, error) {
	p, err := s.account(language, id)
	if err != nil {
		return "", err
	}
	username, err = tiktokweb.Handle(username)
	if err != nil {
		return "", err
	}
	// When the token grants creator-info access, independently bind the browser
	// identity to the OAuth destination, rather than just its display name.
	if p.DirectPost {
		creator, err := s.uploads.TikTokCreator(drafts.Key{Platform: "tiktok", Language: language, Account: "account:" + id}, id)
		if err != nil {
			return "", errors.New("could not confirm the connected TikTok username; check its connection before verifying the browser")
		}
		actual, err := tiktokweb.Handle(creator.Username)
		if err != nil || actual != username {
			return "", errors.New("this username does not match the TikTok account connected to this language")
		}
	}
	return s.browser.Verify(language, id, username)
}
func (s *TikTokWebService) PreparationStatus(language, id string) (string, error) {
	if _, err := s.account(language, id); err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.progress[language+":"+id], nil
}

// Prepare keeps the Wails call asynchronous in the frontend upload queue. It
// never calls a publish API or clicks TikTok's final confirmation button.
func (s *TikTokWebService) Prepare(key drafts.Key, revision int64, id string) (string, error) {
	if key.Platform != "tiktok" || key.Account != "account:"+id {
		return "", errors.New("select the connected TikTok account")
	}
	if _, err := s.account(key.Language, id); err != nil {
		return "", err
	}
	ctx := diagnostics.StartTransfer(context.Background(), s.uploads.diagnostics, "tiktok", key.Language)
	d, err := s.uploads.store.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if d.Revision != revision {
		return "", drafts.ErrConflict
	}
	if err = s.uploads.store.RequireReview(ctx, d); err != nil {
		return "", err
	}
	r, err := s.uploads.store.Receipt("tiktok", id, d.VideoHash)
	if err != nil {
		return "", err
	}
	if r.State != "" {
		return "", errors.New("this video already has a TikTok receipt; no browser upload started. Check its existing result in Posts")
	}
	path, err := s.uploads.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash)
	if err != nil {
		return "", err
	}
	caption := strings.TrimSpace(strings.Join(nonempty(d.Title, d.Description, d.Hashtags), "\n\n"))
	if caption == "" || utf8.RuneCountInString(caption) > 2200 {
		return "", errors.New("use a caption of 1–2200 characters for browser preparation")
	}
	f, _, err := stage(path, d.VideoHash)
	if err != nil {
		return "", err
	}
	staged := f.Name()
	f.Close()
	reserved := false
	defer func() {
		if reserved {
			s.browser.RemoveWhenClosed(key.Language, id, staged)
		} else {
			os.Remove(staged)
		}
	}()
	progress := func(message string) {
		s.mu.Lock()
		if s.progress == nil {
			s.progress = map[string]string{}
		}
		s.progress[key.Language+":"+id] = message
		s.mu.Unlock()
	}
	err = s.browser.Prepare(key.Language, id, staged, caption, func() error {
		// The user may have edited or disconnected while Chrome was opening.
		if _, err := s.account(key.Language, id); err != nil {
			return err
		}
		if _, err := s.uploads.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash); err != nil {
			return err
		}
		if err := s.uploads.store.ReserveTikTokWeb(d, id); err != nil {
			return err
		}
		reserved = true
		diagnostics.Record(ctx, slog.LevelInfo, "web_attaching", diagnostics.Details{})
		return nil
	}, progress)
	if err != nil {
		if reserved {
			if saveErr := s.uploads.store.SaveReceipt(d, id, "web_check", ""); saveErr != nil {
				return "", errors.New("browser result could not be saved; do not upload again")
			}
		}
		progress("Check the TikTok browser")
		diagnostics.Record(ctx, slog.LevelError, "web_attention", diagnostics.Details{})
		return "", err
	}
	if err = s.uploads.store.SaveReceipt(d, id, "web_ready", ""); err != nil {
		return "", errors.New("browser is ready but its receipt could not be saved; do not upload again")
	}
	progress("Ready in TikTok · review and click Publish")
	diagnostics.Record(ctx, slog.LevelInfo, "web_ready", diagnostics.Details{})
	return "Ready in TikTok — review visibility and click Publish in the browser", nil
}
func nonempty(values ...string) []string {
	var result []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			result = append(result, strings.TrimSpace(v))
		}
	}
	return result
}

func (s *TikTokWebService) HasPreparedBrowser() bool { return s.browser.HasPreparedBrowser() }
