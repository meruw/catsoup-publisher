// Package diagnostics records a fixed vocabulary of account and transfer events.
// Raw provider messages, URLs, usernames and credentials never enter this API.
package diagnostics

import (
	"context"
	"crypto/rand"
	"log/slog"
	"regexp"
	"time"
)

var messages = map[string]string{
	"web_attaching":             "TikTok browser account verified; attaching approved video",
	"web_attention":             "TikTok browser preparation needs attention; no automatic retry",
	"web_ready":                 "TikTok browser is ready for manual publishing",
	"transfer_started":          "Video transfer started",
	"transfer_finished":         "Video transfer finished",
	"transfer_failed":           "Video transfer failed",
	"upload_init_started":       "Requesting a video upload session",
	"upload_init_failed":        "Video upload session rejected",
	"upload_init_finished":      "Video upload session validated",
	"upload_bytes_started":      "Sending the video file",
	"upload_bytes_finished":     "Video file sent",
	"upload_processing":         "Waiting for video processing",
	"upload_processing_pending": "Meta still reports processing with a temporary phase error; keeping the same container",
	"upload_processing_failed":  "Provider confirmed a video upload or processing failure",
	"upload_reconciled":         "Upload confirmed by the saved container after a transfer error",
	"upload_status_check":       "Checking the saved upload container",
	"upload_preparing":          "Preparing Instagram-compatible audio; keeping the video stream unchanged",
	"upload_prepared":           "Instagram upload copy prepared and verified",
	"upload_retry_wait":         "Meta could not receive the video; waiting five minutes before one recovery attempt",
	"upload_retry_started":      "Meta confirmed the empty upload failed; starting one replacement upload",
	"publish_started":           "Publishing the processed video",
	"connection_started":        "Account connection started",
	"browser_opening":           "Opening the provider in your browser",
	"callback_received":         "Browser returned a matching authorization response",
	"callback_rejected":         "Rejected an unmatched browser response",
	"exchange_started":          "Exchanging authorization code",
	"exchange_finished":         "Authorization code exchange finished",
	"pages_started":             "Looking up authorized Facebook Pages",
	"pages_received":            "Received a page of Facebook results",
	"profile_matched":           "Found the requested Instagram profile",
	"identity_started":          "Checking the destination account",
	"identity_finished":         "Destination account identified",
	"credential_saving":         "Saving the connection on this PC",
	"connection_finished":       "Account assigned to the selected language",
	"connection_failed":         "Account connection failed",
	"connection_removed":        "Saved connection removed",
	"provider_response":         "Provider request completed",
	"provider_failed":           "Provider request failed",
	"sign_in_failed":            "Browser sign-in failed",
	"pages_failed":              "Could not resolve the Instagram profile",
}

type Details struct {
	ProviderErrorCode                                         string
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
	return start(ctx, store, service, language, "connection_started")
}
func StartTransfer(ctx context.Context, store *Store, service, language string) context.Context {
	return start(ctx, store, service, language, "transfer_started")
}
func start(ctx context.Context, store *Store, service, language, stage string) context.Context {
	if store == nil || (language != "es" && language != "en") {
		return ctx
	}
	switch service {
	case "youtube", "tiktok", "instagram", "facebook", "x":
	default:
		return ctx
	}
	a := &attempt{store, rand.Text(), service, language, time.Now()}
	ctx = context.WithValue(ctx, attemptKey{}, a)
	Record(ctx, slog.LevelInfo, stage, Details{})
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
		ProviderErrorCode: SafeProviderErrorCode(details.ProviderErrorCode),
		Pages:             details.Pages, Matches: details.Matches,
	}, level)
}

// Accept only bounded symbolic or numeric identifiers, never provider messages,
// URLs, mixed-case bearer tokens or arbitrarily long values.
var providerErrorCode = regexp.MustCompile(`^(?:[a-z]{2,24}(?:_[a-z0-9]{1,24}){1,7}|[0-9]{1,12})$`)

func SafeProviderErrorCode(code string) string {
	if len(code) > 96 || !providerErrorCode.MatchString(code) {
		return ""
	}
	return code
}

func safeReason(reason string) string {
	switch reason {
	case "", "cancelled", "timeout", "network", "provider_rejected", "unreadable_response",
		"invalid_callback", "no_pages", "no_linked_profile", "username_mismatch",
		"missing_page_access", "credential_storage", "account_conflict", "configuration", "unknown",
		"invalid_publish_id", "unexpected_upload_host", "invalid_upload_url",
		"invalid_param", "invalid_params", "access_token_invalid", "scope_not_authorized", "rate_limit_exceeded",
		"spam_risk_too_many_pending_share", "spam_risk_too_many_posts", "spam_risk_user_banned_from_posting",
		"reached_active_user_cap", "unaudited_client_can_only_post_to_private_accounts",
		"url_ownership_unverified", "privacy_level_option_mismatch":
		return reason
	default:
		return "unknown"
	}
}
