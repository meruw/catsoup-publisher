# Publisher status

## Available

- Windows x64 desktop workspace with ES/EN playback, custom window controls, an episode catalog and local SQLite drafts.
- Accounts manages destinations per language. Settings manages shared developer credentials and optional HTTPS callback hosting for Instagram.
- Console at the foot of the sidebar (Ctrl+Shift+D) opens a dark, collapsible diagnostic panel. Search stays visible; Filters expands the source/context/level controls, Copy groups full-result and time-window copying, and More contains refresh, temporary history hiding, export and the log folder. Clear restores all display controls to their initial state without deleting saved logs. Pause and active-filter indicators remain visible when folded. It can expand below the window bar; collapsing preserves filters and suspends live UI updates. Local slog/Lumberjack logs rotate at 10 MB with five backups and seven-day retention. Credentials and provider response bodies are excluded.
- Approved episodes support private YouTube uploads, TikTok inbox transfers and Instagram Reels. X offers free browser handoff and optional paid API sign-in and posting per language. [Accounts and uploads](https://publisher.meruw.dev/guides/uploads/)
- Provider guides and the setup documentation are published on [publisher.meruw.dev](https://publisher.meruw.dev/guides/setup/). The site also serves the privacy policy, terms, optional Meta return page and TikTok verification file.

The seven service and setup guides are maintained in `docs/guides/` and mirrored to the GitHub wiki. See [Contributing](../../CONTRIBUTING.md) for the sync command.

## Proof of concept

GitHub Actions runs the Windows build and Go tests with a synthetic bilingual delivery on code changes. Browser fixtures cover playback, account setup, draft recovery, language routing and window controls. Go tests cover credential storage, callback validation and approved-file checks. The Windows build pins Go, Wails and Node; Node compiles the frontend and runs developer checks. The desktop runtime uses WebView2 and the delivery validator needs Python with imageio-ffmpeg.

The PoC is in progress. The project owner decides when it is complete. Release planning starts after that decision.

The owner confirmed native Instagram sign-in and assignment for both ES and EN with access restricted to CATsoup assets. The business login configuration includes `business_management` alongside the Instagram publishing and Page discovery permissions. This confirms account connection, not Reel publishing.

The owner confirmed the console controls work in the native app. The console also supports whole-result copying (including an attempt across pages), rolling 1/5/10/30-minute copy ranges and opening event details from any table cell. Copying keeps active filters, reports the copied count and leaves the clipboard intact for empty results. Go tests cover complete attempts, time windows and clipboard failures; browser fixtures cover the controls and row interaction.

The console layout uses progressive disclosure with compact text actions and more space for events. Browser checks cover Clear restoring live mode, filters, pagination, copy range, detail selection and panel size, along with menu dismissal using Escape.

X API pricing and protocol requirements were checked against official documentation on November 5, 2026. Browser mode remains the default, combining the validated video-path copy and composer launch. API mode requires an operator-owned Native App Client ID, an exact local callback and explicit cost consent at sign-in and before every send. The connector checks the expected username, encrypts tokens, binds them to the client, transfers bounded chunks and saves receipts without automatically retrying post creation. No dependency was added.

X protocol tests cover PKCE and state, refresh and client binding, occupied/rejected callbacks, identity lookup, chunk order/content, processing, receipts and failure paths. Backend tests reject missing paid or destination consent before accessing credentials. Browser fixtures cover free/API selection, X settings, cost acknowledgments, combined browser handoff and the successful receipt blocking a repeat send. Build, Go tests, Go vet and browser checks pass; these do not verify live X authorization, paid access or posting. No real X request or charge was made. The executable is 18,484,224 bytes; app JS is 26.82 kB gzip.

For subsequent release validation, test the packaged validator, media preview and browser sign-in on a clean Windows machine without build tools. Verify ES/EN routing and receipts with operator-owned accounts and approved content, including recovery after an interrupted upload. Supply a versioned artifact, prerequisites and a checksum. No installer or signed release is available.

## Next action

Restart the updated desktop build and verify the free X handoff for ES and EN. Live X API validation is optional and requires the owner's explicit decision to use paid access plus separately approved content. The production demo remains non-publicable; there is no release while the PoC is incomplete.
