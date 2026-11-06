# Publisher status

## Available

- Windows x64 desktop workspace with ES/EN playback, custom window controls, an episode catalog and local SQLite drafts.
- Accounts manages destinations per language. Settings manages shared developer credentials, optional HTTPS callback hosting for Instagram and per-account publishing defaults for ES/EN. Defaults initialize new drafts only; saved posts and approvals stay unchanged.
- Console at the foot of the sidebar (Ctrl+Shift+D) opens a dark, collapsible diagnostic panel. Search stays visible; Filters expands the source/context/level controls, Copy groups full-result and time-window copying, and More contains refresh, temporary history hiding, export and the log folder. Clear restores all display controls to their initial state without deleting saved logs. Pause and active-filter indicators remain visible when folded. It can expand below the window bar; collapsing preserves filters and suspends live UI updates. Local slog/Lumberjack logs rotate at 10 MB with five backups and seven-day retention. Credentials and provider response bodies are excluded.
- The episode composer places version selection under the video and account-scoped text alongside it. Save draft confirms saving beside the button; Restore draft reloads this account's saved text or explicitly recovers platform-scoped unassigned text. Approve post confirms the exact saved revision, account, mode and final video. Text, mode and source edits revoke it. Only the selected platform has a send action, named for its actual behavior. Publish all stays in the sticky header and requires all connected accounts in the selected language to have current approvals; hovering or focusing the control explains blockers without a separate information button. The separate production approval gate remains visible below the video. [Workflow and uploads](https://publisher.meruw.dev/guides/uploads/)
- Approved, reviewed episodes support private/public YouTube uploads, TikTok inbox transfers and Instagram Reels. A private YouTube upload can become public using the same ID, with its own durable lock and recovery state. X offers free browser handoff and optional paid API sign-in and posting per language. Instagram/X draft mode stays local. TikTok Direct Post is optional, off by default, and requires a per-account video.publish grant and its individual privacy/disclosure confirmation.
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

Go migration tests preserve drafts from schemas 1, 4, 5 and 6. Publishing-default tests cover restart, replacement accounts, language isolation, rejected modes and preserving saved posts and approvals. Account-approval tests cover restart, text/mode/source invalidation, stale revisions, disconnected/replaced accounts and revocation before atomic upload or promotion reservation. Service tests confirm local post approval cannot change the production manifest or unlock test material.

Strict frontend build, Go tests and Chrome fixtures cover playback, autosave/recovery, ES/EN identity isolation, all-account approval gating, disabled-mode tooltips, visible save feedback, sticky Publish all, individual modes, paid consent, receipts, preflight and stopping after partial failure. Revoking an approval while batch confirmation is open prevents all sends. The tests use synthetic media and fixture accounts; no live upload, paid request or production approval was performed. Native user interaction with the updated controls remains to be confirmed.

SQLite schema 7 stores publishing defaults separately from saved posts and approvals. The composer groups account/mode/send controls above the text and save/restore/approval below it. Browser checks cover failed preference saves restoring the stored selection and disabled controls showing tooltips on hover or keyboard focus. Six statically bundled Lucide SVGs replace the reload, window, close and console-chevron symbols, with complete ISC/MIT notices. No runtime dependency was added. The runtime still uses the system WebView2, Go and SQLite.

TikTok Direct Post requirements were checked against the official documentation on November 6, 2026 and recorded in the service guide. The integration requires `video.publish`, compliant posting controls and an audit for public visibility. TikTok excludes utilities limited to developer/team accounts; open-source availability alone does not guarantee eligibility. Publisher supports inbox transfer and optional Direct Post. Creator limits, explicit visibility, interaction defaults, commercial disclosures, AI marking, linked policy consent and processing results are validated. Public posting requires the operator’s TikTok client approval; live Direct Post remains untested.

## Posts and Direct Post validation

The lazily loaded Posts module lists local saved drafts and transfer outcomes by verified account and platform. It offers literal search, folded platform/language/status filters, 50-row pagination and a detail dialog linking back to the correct account's editor. Private, inbox, limited-visibility, published, failed and uncertain outcomes remain distinct. It does not synchronize subsequent changes made on platforms or retain an immutable version history. Dropdown chevrons have consistent inset spacing.

Go tests cover account/grant isolation, persistent Direct Post drafts, creator constraints, explicit consent, UTF-16 caption limits, safe upload hosts, checkpoint ordering and definitive versus uncertain failures. Browser fixtures cover the overview and TikTok privacy, disabled interactions, commercial disclosures and private receipts. The Posts JavaScript chunk is 2.87 kB gzip and loads on demand. No dependency or schema migration was added. TikTok Direct Post is excluded from batch sends because it requires its individual privacy and disclosure confirmation; inbox transfer still supports batches.

## Next action

Check Posts in the Windows app with the owner, then configure video.publish in the operator's TikTok developer client and reconnect ES and EN. Validate Direct Post with explicitly approved content and the visibility permitted by TikTok. The PoC remains in progress; no release until the owner declares it complete.
