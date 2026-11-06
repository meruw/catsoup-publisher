# Security

This is an early local Windows application, not a sandbox for hostile media or
processes running as the same operating-system user. No independent security audit
or signed installer is available. Only current main is maintained at this stage.

The app uses a narrow typed bridge, canonical path boundaries, file integrity
checks, a bounded validator process and a restricted content security policy.
It never loads social sites into the privileged application webview.

Keep tokens, client credentials, private media and account data outside Git.
Google and TikTok use system-browser OAuth with state and PKCE. The Meta
HTTPS return page pairs a browser tab with a short-lived state and AES-GCM key. The key arrives in the launch fragment. Only an authenticated reply is accepted on the random IPv4 loopback port. The selected HTTPS host (GitHub Pages only after explicit opt-in) receives the short-lived authorization code in the callback URL and serves the browser script. The desktop app alone exchanges the code with Meta. Changes of callback host require an explicit local setting, independent of imported Meta credentials. Windows DPAPI
protects stored user tokens and binds them to the platform, language and ID.
The UI receives account identities; refresh credentials stay in Go. Uploads revalidate approval and stage the exact hashed
bytes; the HTTP client restricts transfer hosts and rejects redirects. Local
receipts preserve uncertain outcomes across restarts. See [uploads](docs/guides/Uploads.md).

Report vulnerabilities through [GitHub private vulnerability reporting](https://github.com/meruw/catsoup-publisher/security/advisories/new). Include the affected version, reproduction steps and expected impact. Remove credentials and account data from any examples.
