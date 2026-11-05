package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"catsoup/publisher/internal/delivery"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
)

type connection struct {
	account uploads.Account
	token   string
}
type UploadService struct {
	store       *drafts.Store
	reader      *delivery.Reader
	client      *uploads.Client
	openBrowser func(string) error
	mu          sync.Mutex
	sessions    map[drafts.Key]connection
}

func (s *UploadService) Connect(key drafts.Key, token string) (uploads.Account, error) {
	if err := drafts.ValidateKey(key); err != nil {
		return uploads.Account{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var err error
	if key.Platform == "youtube" || key.Platform == "tiktok" {
		token, err = s.client.SignIn(ctx, key.Platform, s.openBrowser)
		if err != nil {
			return uploads.Account{}, err
		}
	}
	if len(token) == 0 || len(token) > 8192 || strings.ContainsAny(token, " \r\n\t") {
		return uploads.Account{}, errors.New("enter the authorized access token")
	}
	account, err := s.client.Identify(ctx, key.Platform, token)
	if err != nil {
		return account, err
	}
	s.mu.Lock()
	key.Episode = ""
	s.sessions[key] = connection{account, token}
	s.mu.Unlock()
	return account, nil
}
func (s *UploadService) Disconnect(key drafts.Key) {
	s.mu.Lock()
	key.Episode = ""
	delete(s.sessions, key)
	s.mu.Unlock()
}
func (s *UploadService) Connected(key drafts.Key) uploads.Account {
	s.mu.Lock()
	defer s.mu.Unlock()
	key.Episode = ""
	return s.sessions[key].account
}
func (s *UploadService) OpenXComposer(key drafts.Key, revision int64) error {
	if key.Platform != "x" {
		return errors.New("select an X draft")
	}
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if d.Revision == 0 || revision != d.Revision {
		return drafts.ErrConflict
	}
	if _, err = s.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash); err != nil {
		return err
	}
	text := strings.TrimSpace(d.Title + "\n\n" + d.Description + "\n\n" + d.Hashtags)
	return s.openBrowser("https://x.com/intent/tweet?" + url.Values{"text": {text}}.Encode())
}
func (s *UploadService) session(key drafts.Key, id string) (connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key.Episode = ""
	con, ok := s.sessions[key]
	if !ok || id == "" || con.account.ID != id {
		return connection{}, errors.New("connect and confirm the destination account first")
	}
	return con, nil
}
func (s *UploadService) Receipt(key drafts.Key, id string) (drafts.Receipt, error) {
	if _, err := s.session(key, id); err != nil {
		return drafts.Receipt{}, err
	}
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return drafts.Receipt{}, err
	}
	return s.store.Receipt(key.Platform, id, d.VideoHash)
}
func (s *UploadService) ClearUncertain(key drafts.Key, id, expectedHash string, checked bool) error {
	if !checked {
		return errors.New("check the platform before allowing another attempt")
	}
	if _, err := s.session(key, id); err != nil {
		return err
	}
	d, err := s.store.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if d.VideoHash != expectedHash {
		return errors.New("linked video changed; reopen the draft")
	}
	return s.store.ResetAttempt(d, id)
}

// Stage only bytes matching the approved hash so an external edit cannot change
// the media while it is streaming. The temporary snapshot is removed on return.
func stage(path, hash string) (*os.File, int64, error) {
	src, err := os.Open(path)
	if err != nil {
		return nil, 0, errors.New("could not open approved video")
	}
	defer src.Close()
	file, err := os.CreateTemp("", "catsoup-upload-*.mp4")
	if err != nil {
		return nil, 0, errors.New("could not stage video")
	}
	digest := sha256.New()
	size, err := io.Copy(io.MultiWriter(file, digest), src)
	if err != nil || hex.EncodeToString(digest.Sum(nil)) != hash {
		file.Close()
		os.Remove(file.Name())
		return nil, 0, errors.New("could not stage the exact approved video; refresh the delivery")
	}
	return file, size, nil
}
func (s *UploadService) Send(key drafts.Key, revision int64, id string, confirmed bool) (drafts.Receipt, error) {
	if !confirmed {
		return drafts.Receipt{}, errors.New("confirm the destination and upload action")
	}
	con, err := s.session(key, id)
	if err != nil {
		return drafts.Receipt{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	d, err := s.store.Get(ctx, key)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if d.Revision == 0 || revision != d.Revision {
		return drafts.Receipt{}, drafts.ErrConflict
	}
	path, err := s.reader.UploadPath(key.Episode, key.Language, d.VideoID, d.DeliveryFingerprint, d.VideoHash)
	if err != nil {
		return drafts.Receipt{}, err
	}
	file, size, err := stage(path, d.VideoHash)
	if err != nil {
		return drafts.Receipt{}, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	caption := strings.TrimSpace(d.Description + "\n\n" + d.Hashtags)
	if key.Platform == "instagram" && strings.TrimSpace(d.Title) != "" {
		caption = strings.TrimSpace(d.Title + "\n\n" + caption)
	}
	m := uploads.Media{Title: d.Title, Caption: caption, Language: key.Language, File: file, Size: size}
	if err = uploads.Validate(key.Platform, m); err != nil {
		return drafts.Receipt{}, err
	}
	current, err := s.client.Identify(ctx, key.Platform, con.token)
	if err != nil {
		return drafts.Receipt{}, err
	}
	if current.ID != id {
		return drafts.Receipt{}, errors.New("the authorized account changed; reconnect before sending")
	}
	if err = s.store.Reserve(d, id); err != nil {
		return drafts.Receipt{}, err
	}
	ref, sendErr := s.client.Send(ctx, key.Platform, con.token, current, m, func(ref string) error { return s.store.SaveReceipt(d, id, "sending", ref) })
	state := "sent"
	if sendErr != nil {
		state = "check"
	}
	if err = s.store.SaveReceipt(d, id, state, ref); err != nil {
		return drafts.Receipt{}, errors.New("receipt write failed; check the platform before retrying")
	}
	if sendErr != nil {
		return drafts.Receipt{}, sendErr
	}
	return s.store.Receipt(key.Platform, id, d.VideoHash)
}
