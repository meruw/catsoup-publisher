# Security

This is an early local Windows application, not a sandbox for hostile media or
processes running as the same operating-system user. No independent security audit
or signed installer is available. Only current main is maintained at this stage.

The app uses a narrow typed bridge, canonical path boundaries, file integrity
checks, a bounded validator process and a restricted content security policy.
It never loads social sites into the privileged application webview.

Do not commit tokens, credentials, private media or real account data. Future
authentication must use the system browser and OS credential store. Preview/import
must not create a review or approval; any future posting must recheck current hashes.

Please avoid posting credentials or exploit details in public issues. Use GitHub's
private vulnerability reporting for this repository when available; otherwise open
an issue requesting a private reporting channel without disclosing the vulnerability.
