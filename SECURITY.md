# Security

This is an early local Windows application, not a sandbox for hostile media or
processes running as the same operating-system user. No independent security audit
or signed installer is available. Only current main is maintained at this stage.

The app uses a narrow typed bridge, canonical path boundaries, file integrity
checks, a bounded validator process and a restricted content security policy.
It never loads social sites into the privileged application webview.

Keep tokens, client credentials, private media and account data outside Git.
Google and TikTok use system-browser OAuth with state and PKCE. The Meta
development callback checks the loopback host, path and state. Windows DPAPI
protects stored user tokens and binds them to the platform, language and ID.
The UI receives account identities; refresh credentials stay in Go. Uploads revalidate approval and stage the exact hashed
bytes; the HTTP client restricts transfer hosts and rejects redirects. Local
receipts preserve uncertain outcomes across restarts. See [uploads](docs/publisher/UPLOADS.md).

Please avoid posting credentials or exploit details in public issues. Use GitHub's
private vulnerability reporting for this repository when available; otherwise open
an issue requesting a private reporting channel without disclosing the vulnerability.
