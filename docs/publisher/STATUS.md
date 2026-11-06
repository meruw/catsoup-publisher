# Publisher status

## Available

- Windows x64 desktop workspace with ES/EN playback, custom window controls, an episode catalog and local SQLite drafts.
- Accounts manages destinations per language. Settings manages shared developer credentials and optional HTTPS callback hosting for Instagram.
- Console at the foot of the sidebar (Ctrl+Shift+D) opens a dark, collapsible diagnostic panel. Search stays visible; Filters expands the source/context/level controls, Copy groups full-result and time-window copying, and More contains refresh, temporary history hiding, export and the log folder. Clear restores all display controls to their initial state without deleting saved logs. Pause and active-filter indicators remain visible when folded. It can expand below the window bar; collapsing preserves filters and suspends live UI updates. Local slog/Lumberjack logs rotate at 10 MB with five backups and seven-day retention. Credentials and provider response bodies are excluded.
- The episode workspace separates Review, Prepare and Publish. Local reviews, notes and decisions persist per language/final version with optimistic concurrency. They never change the production contract. Current review and production approval are both required to send; changing the manifest invalidates a local review. Prepare retains autosave and close/navigation protection. Publish shows saved destinations and their receipts, with explicit draft/public modes. [Workflow and uploads](https://publisher.meruw.dev/guides/uploads/)
- Approved, reviewed episodes support private/public YouTube uploads, TikTok inbox transfers and Instagram Reels. A private YouTube upload can become public using the same ID, with its own durable lock and recovery state. X offers free browser handoff and optional paid API sign-in and posting per language. Instagram/X draft mode stays local. TikTok Direct Post is unavailable in this connector.
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

Have the owner check the new Review → Prepare → Publish workspace in the rebuilt native app, using the non-publicable PoC for local review and draft persistence. Live transfer testing remains separate and requires approved content and explicit destination confirmation. Do not release until the owner declares the PoC complete.

Workspace validation: strict frontend build, all Go tests, Go vet and Chrome fixture checks pass. Coverage includes review persistence and revision conflicts, changes requested, cross-language/source isolation, unchanged manifests, review and approval gates, paused preview when leaving Review, draft flush/recovery, private/public choice, existing-ID YouTube promotion and atomic duplicate prevention. Provider tests intercept HTTP locally; no real upload, paid X request, credential edit or production approval was performed. Native user interaction and live visibility changes remain unverified.

No runtime dependency was added. SQLite migrates transactionally to schema 4. The built executable is 18,533,888 bytes; app JavaScript is 30.01 kB gzip, shared JavaScript 28.61 kB and the lazy console 5.52 kB. YouTube sign-in includes `youtube.force-ssl` for updating an existing video's visibility; accounts without this grant may need an explicit reconnect. The Accounts ID separator remains an HTML entity.
