package uploads

import "encoding/json"

// Only known identifiers and our own guidance leave this parser. OAuth response
// descriptions can contain user data and must never reach the log or UI.
func googleAuthorizationError(raw []byte) (string, string) {
	var out struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(raw, &out)
	switch out.Error {
	case "invalid_grant":
		return out.Error, "Google no longer accepts this authorization. Reconnect the YouTube account using Accounts > the selected language > YouTube > Change. External apps in Testing can expire authorizations after seven days."
	case "invalid_client", "unauthorized_client", "deleted_client", "disabled_client":
		return out.Error, "Check the Google Desktop app client in Settings > Developer credentials > YouTube, then reconnect the account."
	case "invalid_scope":
		return out.Error, "Reconnect YouTube and grant the requested permissions."
	case "invalid_request", "unsupported_grant_type":
		return out.Error, "Google rejected the authorization request. Check the developer client configuration."
	case "temporarily_unavailable", "server_error":
		return out.Error, "Google authorization is temporarily unavailable. Try again later."
	default:
		return "google_oauth_unknown", "Google rejected authorization. Check the account connection and developer client configuration."
	}
}
