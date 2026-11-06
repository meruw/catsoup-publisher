# Publisher status

## Available

- Windows x64 desktop workspace with ES/EN playback, custom window controls, an episode catalog and local SQLite drafts.
- Accounts manages destinations per language. Settings manages shared developer credentials and optional HTTPS callback hosting for Instagram.
- Console at the foot of the sidebar (Ctrl+Shift+D) opens a dark, collapsible diagnostic panel. Search stays visible; Filters expands the source/context/level controls, Copy groups full-result and time-window copying, and More contains refresh, temporary history hiding, export and the log folder. Clear restores all display controls to their initial state without deleting saved logs. Pause and active-filter indicators remain visible when folded. It can expand below the window bar; collapsing preserves filters and suspends live UI updates. Local slog/Lumberjack logs rotate at 10 MB with five backups and seven-day retention. Credentials and provider response bodies are excluded.
- The episode composer shows video, destination, visible Save draft and title/description/hashtags together. Drafts are bound to provider account IDs with delivery modes persisted per account and episode. Destination rows expose each mode and individual Publish action; Publish all confirms and sends the current language's selected destinations sequentially, stopping on error. Review and file details are folded below the player; More holds copy tools and additional drafts. Publish… opens a native browser dialog with account, applicable mode and explicit confirmation. A blocked delivery shows one reason. Both local review and production approval are required to send. Autosave and navigation/close guards preserve unsaved edits. [Workflow and uploads](https://publisher.meruw.dev/guides/uploads/)
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

X protocol tests cover PKCE and state, refresh and client binding, occupied/rejected callbacks, identity lookup, chunk order/content, processing, receipts and failure paths. Backend tests reject missing paid or destination consent before accessing credentials. Browser fixtures cover free/API selection, X settings, cost acknowledgments, combined browser handoff and the successful receipt blocking a repeat send. Build, Go tests, Go vet and browser checks pass; these do not verify live X authorization, paid access or posting. No real X request or charge was made.

For subsequent release validation, test the packaged validator, media preview and browser sign-in on a clean Windows machine without build tools. Verify ES/EN routing and receipts with operator-owned accounts and approved content, including recovery after an interrupted upload. Supply a versioned artifact, prerequisites and a checksum. No installer or signed release is available.

## Workspace validation

Go migration tests preserve text/revisions from schemas 1 and 4 and persist delivery modes after restart. Browser fixtures verify account-ID isolation with identical names, batch language routing, duplicate exclusion, all-account preflight and stopping after partial failure. Strict frontend build, Go tests and Chrome fixture checks cover the composer, playback, local review, autosave/recovery, ES/EN isolation, approved-file gates and provider transfer modes. Dialog checks cover focus restoration, keyboard help, Escape, returning to unfinished review and blocking dismissal during an in-flight transfer. TikTok presents only its inbox action. Private YouTube promotion and duplicate protection retain their existing backend checks. No real upload, paid X request, credential change or production approval was performed.

Shared-control checks also pass, including keyboard interaction and text contrast (all checked pairs exceed 4.5:1).

The Windows executable at `apps/publisher/bin/CATsoupPublisher.exe` is updated (18,546,176 bytes). The app entry bundle is 31.41 kB gzip; the shared Vue/style bundle is 28.46 kB gzip and the lazy console is 5.52 kB gzip. No runtime dependency was added. SQLite schema 5 adds the delivery mode with a transactional migration; unbound text remains available for explicit recovery. The review and production approval gates remain separate. Native user interaction with the redesigned composer remains to be confirmed.

## Next action

Check account-scoped drafts, delivery modes and individual/batch confirmation in the updated Windows executable with the project owner. The video-state overview is a separate next feature. The PoC remains in progress; no release until the owner declares it complete.
