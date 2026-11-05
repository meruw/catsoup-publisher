# Publisher status

## Implemented

- Local Go/Wails/Vue desktop app, custom title bar and CATsoup design system.
- Read-only contract-v2 validation, ES/EN playback, seeking and version selection.
- Path containment, hash validation, bounded validator process and restricted CSP.
- Standalone minimal Python requirements and a generated synthetic sample.
- Public source packaging with GPL-3.0-only, contribution notes and font attribution.

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

## Limits

Windows x64 only; WebView2 and Python/imageio-ffmpeg required. No installer/signing,
SQLite drafts, OAuth, uploads, queue or scheduling. No RAM/CPU benchmark or independent
security audit. Trends is a disabled placeholder. The demo is not publishable.
Legacy PoC handling remains for compatibility, but private PoC assets are not included.

## Next action

Implement local SQLite-backed publication drafts by episode, language and platform,
with account labels, editable metadata, autosave and persistence/migration checks.
Manual handoff comes before external account connections. Do not publish test media.
