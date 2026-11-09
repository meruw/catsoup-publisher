package uploads

import (
	"catsoup/publisher/internal/diagnostics"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
)

// Fixed messages exclude provider text and signed upload URL query parameters.
var tikTokErrorHints = map[string]string{
	"invalid_param":                                      "TikTok rejected the request parameters",
	"invalid_params":                                     "TikTok rejected the request parameters",
	"access_token_invalid":                               "Reconnect this TikTok account; its token is invalid or expired",
	"scope_not_authorized":                               "Reconnect this TikTok account with the required publishing permission",
	"rate_limit_exceeded":                                "TikTok's request limit was reached; wait before retrying",
	"spam_risk_too_many_pending_share":                   "TikTok has too many pending inbox shares; finish or remove them in TikTok",
	"spam_risk_too_many_posts":                           "TikTok's daily posting limit was reached",
	"spam_risk_user_banned_from_posting":                 "TikTok has restricted posting for this account",
	"reached_active_user_cap":                            "This TikTok developer app reached its active publishing user limit",
	"unaudited_client_can_only_post_to_private_accounts": "TikTok restricts this unaudited client to private posting; check the account and visibility requirements",
	"url_ownership_unverified":                           "TikTok could not verify the media URL's ownership",
	"privacy_level_option_mismatch":                      "TikTok's available visibility options changed; reopen the posting dialog",
}

func tikTokRejection(ctx context.Context, code string) error {
	return tikTokRejectionStatus(ctx, code, 200)
}

func tikTokRejectionStatus(ctx context.Context, code string, status int) error {
	hint, known := tikTokErrorHints[code]
	reason := code
	if !known {
		reason = "provider_rejected"
		hint = "TikTok rejected the request with an unrecognized error code"
	}
	safeCode := diagnostics.SafeProviderErrorCode(code)
	diagnostics.Record(ctx, slog.LevelError, "provider_failed", diagnostics.Details{Reason: reason, HTTPStatus: status, ProviderErrorCode: safeCode})
	if known {
		return fmt.Errorf("%s (%s)", hint, code)
	}
	if safeCode != "" {
		return fmt.Errorf("%s (%s)", hint, safeCode)
	}
	if code == "" {
		return errors.New("TikTok rejected the request without a readable error code")
	}
	return errors.New(hint)
}

// TikTok endpoints can use a nested error.code, OAuth-style error string or
// top-level code. Read only that field; messages and token-bearing URLs are ignored.
func tikTokResponseCode(raw []byte) string {
	var envelope struct {
		Error json.RawMessage
		Code  json.RawMessage
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return ""
	}
	var nested struct{ Code json.RawMessage }
	var code string
	value := envelope.Code
	if json.Unmarshal(envelope.Error, &nested) == nil && len(nested.Code) > 0 {
		value = nested.Code
	} else if json.Unmarshal(envelope.Error, &code) == nil && code != "" {
		return code
	}
	if json.Unmarshal(value, &code) == nil {
		return code
	}
	var number json.Number
	if json.Unmarshal(value, &number) == nil {
		return string(number)
	}
	return ""
}

var uploadHostname = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]{0,252}$`)

func tikTokSession(ctx context.Context, code, id, address string) error {
	if code != "ok" {
		return tikTokRejection(ctx, code)
	}
	if !safeID.MatchString(id) {
		diagnostics.Record(ctx, slog.LevelError, "upload_init_failed", diagnostics.Details{Reason: "invalid_publish_id"})
		return errors.New("TikTok returned an invalid publishing ID. No video data was sent")
	}
	if !safeSession(address, "open-upload.tiktokapis.com", "open-upload-us.tiktokapis.com", "open-upload-eu.tiktokapis.com", "open-upload-sg.tiktokapis.com") {
		parsed, err := url.Parse(address)
		if err == nil && parsed.Scheme == "https" && parsed.User == nil && parsed.Fragment == "" && parsed.Port() == "" && uploadHostname.MatchString(parsed.Hostname()) {
			diagnostics.Record(ctx, slog.LevelError, "upload_init_failed", diagnostics.Details{Reason: "unexpected_upload_host"})
			return fmt.Errorf("TikTok returned an unsupported upload host (%s). No video data was sent", parsed.Hostname())
		}
		diagnostics.Record(ctx, slog.LevelError, "upload_init_failed", diagnostics.Details{Reason: "invalid_upload_url"})
		return errors.New("TikTok returned an invalid upload address. No video data was sent")
	}
	return nil
}
