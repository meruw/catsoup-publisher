// Package diagnostics records a fixed vocabulary of account events.
// Raw provider messages, URLs, usernames and credentials never enter this API.
package diagnostics

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"
)

var messages = map[string]string{
	"connection_started":  "Account connection started",
	"browser_opening":     "Opening the provider in your browser",
	"callback_received":   "Browser returned a matching authorization response",
	"callback_rejected":   "Rejected an unmatched browser response",
	"exchange_started":    "Exchanging authorization code",
	"exchange_finished":   "Authorization code exchange finished",
	"pages_started":       "Looking up authorized Facebook Pages",
	"pages_received":      "Received a page of Facebook results",
	"profile_matched":     "Found the requested Instagram profile",
	"identity_started":    "Checking the destination account",
	"identity_finished":   "Destination account identified",
	"credential_saving":   "Saving the connection on this PC",
	"connection_finished": "Account assigned to the selected language",
	"connection_failed":   "Account connection failed",
	"connection_removed":  "Saved connection removed",
	"provider_response":   "Provider request completed",
	"provider_failed":     "Provider request failed",
	"sign_in_failed":      "Browser sign-in failed",
	"pages_failed":        "Could not resolve the Instagram profile",
}

type Details struct {
	Reason                                                    string
	HTTPStatus, ProviderCode, ProviderSubcode, Pages, Matches int
}

type attemptKey struct{}
type attempt struct {
	store                 *Store
	id, service, language string
	start                 time.Time
}

func Start(ctx context.Context, store *Store, service, language string) context.Context {
	if store == nil || (language != "es" && language != "en") {
		return ctx
	}
	switch service {
	case "youtube", "tiktok", "instagram", "x":
	default:
		return ctx
	}
	a := &attempt{store, rand.Text(), service, language, time.Now()}
	ctx = context.WithValue(ctx, attemptKey{}, a)
	Record(ctx, slog.LevelInfo, "connection_started", Details{})
	return ctx
}

func Record(ctx context.Context, level slog.Level, stage string, details Details) {
	a, ok := ctx.Value(attemptKey{}).(*attempt)
	message, known := messages[stage]
	if !ok || !known {
		return
	}
	details.Reason = safeReason(details.Reason)
	a.store.record(Event{
		Time: time.Now().UTC(), Level: level.String(), Message: message,
		Service: a.service, Language: a.language, Attempt: a.id, Stage: stage,
		ElapsedMS: time.Since(a.start).Milliseconds(), Reason: details.Reason,
		HTTPStatus: details.HTTPStatus, ProviderCode: details.ProviderCode, ProviderSubcode: details.ProviderSubcode,
		Pages: details.Pages, Matches: details.Matches,
	}, level)
}

func safeReason(reason string) string {
	switch reason {
	case "", "cancelled", "timeout", "network", "provider_rejected", "unreadable_response",
		"invalid_callback", "no_pages", "no_linked_profile", "username_mismatch",
		"missing_page_access", "credential_storage", "account_conflict", "configuration", "unknown":
		return reason
	default:
		return "unknown"
	}
}
