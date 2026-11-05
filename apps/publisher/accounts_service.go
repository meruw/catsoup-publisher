package main

import (
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

func (s *UploadService) Accounts() ([]drafts.Profile, error) { return s.store.Profiles() }
func validateSlot(platform, lang string) error {
	switch platform {
	case "youtube", "instagram", "tiktok", "x":
	default:
		return errors.New("unsupported platform")
	}
	if lang != "es" && lang != "en" {
		return errors.New("choose Spanish or English")
	}
	return nil
}

var xHandle = regexp.MustCompile("^[A-Za-z0-9_]{1,15}$")
var igHandle = regexp.MustCompile("^[A-Za-z0-9._]{1,30}$")

func (s *UploadService) LinkAccount(platform, language, expectedID, handle, pageToken string) (drafts.Profile, error) {
	var p drafts.Profile
	if err := validateSlot(platform, language); err != nil {
		return p, err
	}
	handle = strings.TrimPrefix(strings.TrimSpace(handle), "@")
	if platform == "x" {
		if !xHandle.MatchString(handle) {
			return p, errors.New("enter a valid X handle")
		}
		p = drafts.Profile{Platform: platform, Language: language, ID: strings.ToLower(handle), Name: "@" + handle}
		return p, s.store.PutProfile(p, []byte{}, expectedID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	var token uploads.Token
	var err error
	if pageToken != "" {
		if platform != "instagram" || len(pageToken) > 8192 || strings.ContainsAny(pageToken, " \t\r\n") {
			return p, errors.New("invalid Page credential")
		}
		token.Access = pageToken
	} else {
		if platform == "instagram" && !igHandle.MatchString(handle) {
			return p, errors.New("enter the Instagram username attached to your Page")
		}
		token, err = s.client.SignIn(ctx, platform, s.openBrowser)
		if err != nil {
			return p, err
		}
		if platform == "instagram" {
			token, err = s.client.InstagramPage(ctx, token, handle)
			if err != nil {
				return p, err
			}
		}
	}
	account, err := s.client.Identify(ctx, platform, token.Access)
	if err != nil {
		return p, err
	}
	if platform == "instagram" && handle != "" && !strings.EqualFold(handle, account.Name) {
		return p, errors.New("the connected Instagram account does not match this username")
	}
	p = drafts.Profile{Platform: platform, Language: language, ID: account.ID, Name: account.Name}
	sealed, err := token.Seal(platform + "/" + language + "/" + account.ID)
	if err != nil {
		return p, err
	}
	return p, s.store.PutProfile(p, sealed, expectedID)
}
func (s *UploadService) UnlinkAccount(platform, language, id string) error {
	if err := validateSlot(platform, language); err != nil {
		return err
	}
	return s.store.RemoveProfile(platform, language, id)
}
func (s *UploadService) access(ctx context.Context, platform, language, id string) (string, error) {
	scope := platform + "/" + language + "/" + id
	raw, err := s.store.Credential(platform, language, id)
	if err != nil {
		return "", err
	}
	token, err := uploads.Unseal(raw, scope)
	if err != nil {
		return "", err
	}
	if !token.Due() {
		return token.Access, nil
	}
	renewCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err = s.store.RotateCredential(renewCtx, platform, language, id, func(latest []byte) ([]byte, error) {
		token, err = uploads.Unseal(latest, scope)
		if err != nil || !token.Due() {
			return nil, err
		}
		token, err = s.client.RefreshToken(renewCtx, platform, token)
		if err != nil {
			return nil, err
		}
		return token.Seal(scope)
	})
	return token.Access, err
}
func (s *UploadService) ConfigureClient(platform, raw string) error {
	return uploads.Configure(platform, raw)
}

func (s *UploadService) GoogleClientReady() bool { return uploads.ClientReady("youtube") }

func (s *UploadService) ClientReady(platform string) bool { return uploads.ClientReady(platform) }

func (s *UploadService) CredentialStatuses() map[string]string {
	result := make(map[string]string, 3)
	for _, platform := range []string{"youtube", "tiktok", "instagram"} {
		result[platform] = uploads.ClientStatus(platform)
	}
	return result
}

func (s *UploadService) MetaCallbackURL() (string, error) { return uploads.MetaCallbackURL() }
func (s *UploadService) SetMetaCallbackURL(raw string) (string, error) {
	return uploads.SetMetaCallbackURL(raw)
}

// Only these setup pages may be opened by the account manager.
func (s *UploadService) OpenGoogleSetup(step string) error {
	switch step {
	case "api":
		return s.openBrowser("https://console.cloud.google.com/apis/library/youtube.googleapis.com")
	case "audience":
		return s.openBrowser("https://console.cloud.google.com/auth/audience")
	case "client":
		return s.openBrowser("https://console.cloud.google.com/auth/clients")
	default:
		return errors.New("unknown Google setup page")
	}
}
