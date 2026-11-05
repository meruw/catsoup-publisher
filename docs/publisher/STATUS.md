# Publisher status

## Implemented

- Accounts module with independent ES/EN destinations, system-browser sign-in,
  protected Windows credentials and verified upload routing across restarts.
  X supports a manual handle per language. See [account setup](UPLOADS.md).

- Private YouTube uploads, TikTok inbox transfers, Instagram Reel publishing and
  X web composition. Account verification, approved-file snapshots and durable
  receipts protect against sending the wrong file or repeating an uncertain upload.
  See [connection setup and upload workflows](UPLOADS.md).

- Local Go/Wails/Vue desktop app, custom title bar and CATsoup design system.
- Read-only contract-v2 validation, ES/EN playback, seeking and version selection.
- Path containment, hash validation, bounded validator process and restricted CSP.
- Standalone minimal Python requirements and a generated synthetic sample.
- Public source packaging with GPL-3.0-only, contribution notes and font attribution.
- Local SQLite drafts for episode/language/platform/account labels; debounced
  autosave, manual save, revision conflicts, source-change warnings and explicit rebind.
- Saved-title/caption copying and validated video-folder access. Native close guard
  flushes pending edits; failures retain the form. No platform posting or approval.
- Manual upload preparation copies the linked final video's path after an account
  check and fresh validation of the delivery, approval, language and source hashes.

## Verification

The original owner confirmed native ES/EN playback, audio, seeking, window dragging,
resizing, minimizing, maximizing/restoring and closing. Browser fixture checks are
separate from native checks.

The standalone export passed strict frontend checks/build, Go unit tests, real
validator/media integration using the generated demo, `go vet`, ES/EN browser
playback/seek/variants/empty-state checks, and the design-system checks. The demo
validates as exported and not publishable. A tracked-file scan found no matches
for the tested secret/private-key patterns or the original user's absolute path;
this is not an exhaustive security audit.

Verification used a new Python environment with only imageio-ffmpeg. Fresh Go CLI
dependency downloads timed out, so the existing pinned Wails CLI and verified Go
module cache were reused for the successful build. A fully uncached installation
still needs confirmation on another machine. No production data was required.

Draft milestone checks passed: strict frontend build, Go tests and vet, SQLite
restart/migration and concurrent-writer checks, service boundary checks, and browser
fixtures covering in-flight edits, failures, conflicts, navigation, close and metadata
copy/reveal. Layout inspected at 1200/820px; design checks passed. Executable:
15,338,496 bytes (14.63 MiB); main JS about 125 kB (44 kB gzip). No new frontend packages.
`govulncheck v1.9.0` reported no known vulnerabilities on 2026-11-04. SQLite v1.61.1
release/security notes were reviewed and libc is pinned to its required v1.79.0.

Manual handoff passed the Go tests, strict frontend build and browser fixture
for account confirmation and the approved final. The fixture uses a separate
in-memory episode. Selecting an MP4 in a live platform file picker remains a
native check for the first approved episode.

## Limits

Native accessibility inspection confirms the actual Publication section and the
account field accepts edits. Native click/keyboard automation is still unreliable;
the owner confirmed that drafts persist on reopening the current app. Native
clipboard, folder access and close flushing still need separate confirmation.
The initial owner attempt opened the older private-workspace executable;
the independent build has now been opened and its header includes a direct
Publication button. Browser fixtures do not establish those native outcomes.

Windows x64 only; WebView2 and Python/imageio-ffmpeg required. No installer/signing,
background queue or scheduling. No RAM/CPU benchmark or independent
security audit. Trends is a disabled placeholder. The demo is not publishable.
Legacy PoC handling remains for compatibility, but private PoC assets are not included.

## Next action

Import the Google desktop client in Accounts, connect the ES and EN channels, and send the first
owner-approved episode privately to that channel. The Go protocol tests and Vue
browser fixtures use simulated providers. Real account permissions and live
uploads still require verification on each platform.
