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

The lazily loaded Posts module groups local saved drafts and transfer outcomes into one row per video. It offers literal search and a separate platform/language/status filter dialog, 20-video pagination and a video detail page with language tabs and account details linking back to the correct editor. Private, inbox, limited-visibility, published, failed and uncertain outcomes remain distinct. It does not synchronize subsequent changes made on platforms or retain an immutable version history. Dropdown chevrons have consistent inset spacing.

Go tests cover account/grant isolation, persistent Direct Post drafts, creator constraints, explicit consent, UTF-16 caption limits, safe upload hosts, checkpoint ordering and definitive versus uncertain failures. Browser fixtures cover the overview and TikTok privacy, disabled interactions, commercial disclosures and private receipts. The Posts JavaScript chunk is 5.44 kB gzip and loads on demand. No dependency or schema migration was added. TikTok Direct Post is excluded from batch sends because it requires its individual privacy and disclosure confirmation; inbox transfer still supports batches.

Publishing defaults has its own Settings subpage, alongside Developer credentials, with a shortcut to TikTok Direct Post setup. Browser fixtures cover grouped video rows, filter application/reset, per-video language selection and account detail navigation. Go tests cover complete eight-destination summaries, same-destination filtering, catalog-title matches and pagination across 24 episodes. The Windows executable is 18,652,160 bytes; native owner review of this layout remains pending.

Status filters use four categories with exact receipt states retained in account details. Posts and detail views show small local thumbnails. Visible previews decode serially through the existing validated media endpoint, wait for a rendered frame, then release the decoder. An in-memory cache holds up to 64 images for one minute; unavailable media has a neutral placeholder. No additional bridge, runtime dependency or persistent media cache is introduced. Browser checks verify real thumbnail pixels, bounded dimensions, serial decoding, cache reuse and cancellation.

## Credential storage validation

Developer configurations use filename-bound Windows DPAPI and automatic migration of valid legacy JSON. Migration verifies exact plaintext recovery before replacement and preserves account tokens, client fingerprints, sign-in receipts and TikTok Direct Post preferences. Corrupt, swapped and oversized files fail closed. Temporary files contain ciphertext only and are removed after a failed replacement. Windows tests cover migration, repeat reads, client confirmation, token preservation and a locked destination file. Go tests, go vet and the Windows executable build pass.

The operator's three present developer configurations (Google, TikTok and Meta) were migrated and verified readable. The account/draft database remained byte-identical during repeat verification; client receipts and callback preferences were preserved. No live provider authorization or upload was performed for this change.

## Next action

Review the automatic TikTok observations in Posts during normal use. Monitoring resumes saved receipts without sending media. The PoC remains in progress; no release is planned.

## Publishing reliability and workspace update

The implementation includes Facebook Page Reels, background transfer queues, persistent upload history grouped by episode, searchable paginated episode navigation, native delivery import, per-language shared post text, and cached account-specific statistics for YouTube, Instagram and Facebook. X browser handoffs do not block automatic batch destinations. The application uses a single desktop instance.

Instagram prepares a temporary compatibility copy with AAC audio at 48 kHz stereo and 128 kb/s while preserving the video stream and approved source. Receipts retain the source hash, prepared hash and container. Check and finish publishes an already-ready container. For Meta's specific missing-file error with zero received bytes, it can retransmit the exact prepared file into the same saved container; it never creates a replacement for an ambiguous pending result. Recovery requires current approval, matching account, revision and hashes, and an atomic reservation. An interrupted publication remains locked for reconciliation.

TikTok handles regional upload hosts, valid chunk sizing and structured provider error identifiers. An inbox receipt records TikTok's SEND_TO_USER_INBOX response, not a published post or independently verified notification visibility. The owner confirmed that both delayed Spanish inbox deliveries eventually arrived on mobile. This confirms delayed delivery for that incident, not a failed transfer. Repeated uploads are not an automatic remedy.

Schema version 9 adds bounded statistics history and Instagram preparation checkpoints while retaining existing drafts, connections and receipts. No new runtime dependency is required for these changes. Operator records, credentials, real media, receipts and support diagnostics remain in ignored local storage.

Validation includes all Go tests and vet, strict frontend compilation, provider protocol fixtures, receipt/approval/concurrency and migration tests, media compatibility checks, and browser fixtures for navigation, imports, statistics, batch and background transfers. Browser fixtures do not replace native UI testing. The latest Windows build was installed, launched and responded; a real Instagram same-container recovery was verified successfully.

Next action: verify ES/EN browser sessions and the TikTok Studio preparation flow with an approved video that has no existing receipt. The existing inbox video remains locked against duplicate preparation; the owner confirmed receipt of both delayed Spanish deliveries. The application remains a proof of concept; no stable release is planned yet.


## TikTok browser preparation

Publisher uses pinned chromedp v0.23.0 with installed Chrome and separate persistent profiles for each language/account. The control connection uses an inherited pipe, without a debugging TCP port. Chrome retains its normal security and credential-storage defaults; passwords and cookies are not imported from the user's personal profile. Setup opens a visible browser for manual sign-in and verifies the own-profile menu username. Available creator-info access also binds that username to the OAuth account.

An individual Browser action in the TikTok send dialog verifies identity before attaching the approved, hash-checked temporary video and filling the combined caption. It runs in the existing background queue and reports progress. Publisher does not click Publish, change privacy automatically, or mark a web preparation as published. Browser preparations reserve the same durable account/hash record as API sends. Existing inbox/API receipts are preserved and block attachment. Uncertain browser outcomes do not unlock retries. An open prepared browser triggers a close warning.

Validation: isolated local Chrome fixtures cover wrong-account rejection before reservation, caption content, single attachment, and zero Publish clicks. SQLite tests cover approval, destination identity, API/web exclusion and preserving existing receipts. Native account login and the live TikTok caption editor still need operator verification; no real video was attached during development.

The browser-method UI fixture passes account setup, wrong-account feedback, verified preparation, and existing-inbox blocking at desktop and narrow widths. Go tests, vet, strict Vue/TypeScript build and govulncheck pass. The installed executable grows from 18,843,136 to 19,814,912 bytes (about 5.2%); the main app JavaScript bundle is 114,519 bytes before compression. Chrome starts only for explicit browser actions. The updated Windows executable was installed and its process responds; native login/preparation remains to be checked with the owner.

## Publication timestamps

SQLite schema 10 adds account/file-scoped receipt timestamps without changing existing upload locks, references, approvals or connection credentials. Terminal confirmations are recorded transactionally with receipts; matching repeated responses preserve the first confirmation. Provider dates enrich only ownership-checked published YouTube/Instagram receipts. Existing upload start times remain distinct and are not backfilled as publication times. Facebook video creation dates are excluded because they do not establish publication.

Posts shows each account's publication or delivery event time, with the local time zone and separate attempt/update times in details. Statistics reads persist supported remote dates using the existing request. No additional polling, dependency or paid API is introduced. Migration, confirmation preservation, account/reference isolation, public/private semantics and provider parsing tests pass. The full Go suite, go vet, strict frontend build and Chrome fixtures pass. Browser checks cover time-zone conversion, unknown dates, inbox/browser labels, live statistics enrichment and narrow layouts. The Posts chunk is 9.49 kB gzip. The Windows executable is 19,826,688 bytes (11,776 bytes more than the preceding local build); the installed build opens with a responding Windows window. Live statistics reads returned publication dates for existing YouTube and Instagram posts. A database backup and before/after comparison confirmed that drafts, approvals and upload receipts were preserved. Native interaction with the date display remains for owner review.

Schema 10 recovery: back up the SQLite database before upgrading. Earlier builds reject schema 10; downgrading requires restoring the matching pre-upgrade backup while Publisher is closed. Keep the upgraded database as well, since restoring an older backup loses subsequent activity.

## Pending episode navigation

The episode sidebar omits completed current deliveries while Posts retains the full catalog. Completion requires every declared language and all connected non-X accounts, with published receipts or TikTok inbox receipts matching the current draft video and delivery fingerprint. X is excluded in all modes. Missing drafts, private or restricted uploads, browser preparation, failures and unknown outcomes keep a delivery pending. Changed accounts or delivery versions reopen work. No new database schema, remote requests, dependency or background timer is introduced.

Go tests cover missing destination drafts, both languages, X exclusion, receipt states, changed hashes, account replacement, no connected destinations and mixed versions. Chrome fixtures cover bilingual filtering, version/invalid-delivery preservation, search and updates after completion. Go tests/vet and strict frontend production build pass. The main app chunk is 33.23 kB gzip (0.28 kB growth); the executable is 19,831,296 bytes (4,608 bytes growth). A read-only check against local receipts identifies seven completed episodes; it does not change approvals or upload state.

The PoC remains in progress; no release is created.

## Final language confirmation

Individual sends, batches and TikTok browser preparation require a final language acknowledgment before any transfer is queued. The review displays Spanish/ES or English/EN, the exact linked video with audio controls, account destinations and delivery modes. Back clears the acknowledgment without sending. Changes to language, video or destination invalidate the review; existing approval, receipt, revision and account checks still run before sending. The language label comes from delivery metadata, not speech detection. Batch draft selection also explicitly matches the selected language.

The isolated confirmation fixture covers no send before acknowledgment, Back/reset, EN destination routing, changed video, revoked approval, changes during draft flush, individual sends and browser preparation. The full browser regression suite and strict frontend production build pass. Go tests pass. Browser fixtures use synthetic media and mock providers; they do not publish or prove native interaction. The general harness now stubs the browser close guard and explicitly selects the TikTok API transport. No runtime dependency, database schema or remote polling is added. The main app chunk is 34.22 kB gzip; the executable is 19,836,928 bytes, 5,632 bytes more than the preceding build.

The tested executable was installed while Publisher was closed; its SHA-256 matches the staged build and the desktop process responds. Native interaction with the final confirmation remains for owner review.

Native review of the final language confirmation remains with the owner. No release is created.

## Google authorization diagnostics

Google OAuth token endpoint failures expose allowlisted identifiers and actionable local guidance. `invalid_grant` directs the user to reconnect the affected YouTube account; client configuration errors point to Settings. Unknown and malformed bodies remain generic. Raw OAuth descriptions, request bodies and credentials never enter diagnostics. The Uploads activity panel uses correctly encoded separators and punctuation.

Tests cover known and unknown errors, malformed bodies, one request without retry, UI-error redaction and the persisted diagnostic identifier/status. The full Go suite and strict frontend build pass. An initial sandbox-only temporary-workspace permission failure was resolved by running the same suite with filesystem access. A live authorization-only check confirmed `invalid_grant` for the affected connection without uploading, changing receipts or writing credentials. The YouTube guide documents reconnection and Google's seven-day Testing authorization limit. No dependency or schema changes.

The tested Windows executable was installed with Publisher closed; the installed SHA-256 matches the staged build and the opened desktop process responds. The owner confirmed reconnecting YouTube before installation. No uploads were started during installation.

Next action: verify the corrected Uploads labels in the desktop app during normal use. No release is created.

## Account-specific covers and Instagram recovery

Schema 11 stores content-addressed PNG/JPEG assets and per-video YouTube cover results. Draft selections are scoped by account/language; cover edits invalidate approval and changing videos clears the cover. The compact Thumbnail disclosure supports local image import or frame capture for YouTube, frame timestamps for Instagram and TikTok Direct Post, and explicit limitations for other connectors. No new runtime dependency or hosted-media service is required. The frontend main chunk is 36.04 kB gzip (about 1.8 kB growth); the staged Windows executable is 19,936,256 bytes. YouTube thumbnail failures retain the confirmed video receipt and can be retried independently.

Instagram's specific pending FILE_NOT_FOUND result with explicit zero bytes now permits one delayed retransmission to the same container. A stale Sending receipt can become Check result after 30 minutes without clearing its reference, prepared hash or history; publishing and confirmed receipts remain locked. Unit tests cover those boundaries, cover validation/persistence/approval, processing-before-thumbnail ordering and same-container recovery. Browser fixtures cover image import, real video-frame capture, account controls, approval invalidation and unsupported options. Full Go tests, vet, frontend build and existing UI/confirmation checks pass. The executable was installed with a matching checksum; the database migration preserved account credentials and draft metadata. Native cover interaction remains to be reviewed. Live Instagram recovery returned a provider-side failure, so fixture success must not be presented as confirmed live reliability.

## Instagram MP4 compatibility and container diagnostics

New Instagram transport copies omit edit lists and use nonnegative timestamps with the metadata atom before media data. The approved source and compressed video stream remain unchanged. Saved transfer preparation is versioned; an existing container still requires byte-identical recovery using its original preparation. No schema or dependency changes are required.

Container queries handle Meta's `100/2207065` rejection of `video_status` by requesting only supported status fields. Terminal `ERROR`/`EXPIRED` results return a processing failure with a sanitized numeric code; arbitrary status text and provider bodies are not logged. Regression checks cover legacy preparation, audio/silent clips, MP4 atom layout, hashes, full decoding, field fallback and preservation of unrelated authorization errors. Full Go tests, vet and the production Windows build pass.

Live validation resolved the pending Instagram ES publication through a separately authorized temporary HTTPS-ingestion experiment. Direct file transfer still failed with code `1363008` and zero reported received bytes, including a fully recoded compatibility copy and an HTTP/1.1 retransmission. The same compatible MP4 was then served through a temporary, single-file Cloudflare tunnel. Meta processed it successfully and the Reel was published, independently read back through the API, and saved as published in the local receipt. The tunnel was closed afterward. This does not establish the root cause of direct-ingest failures or add automatic external hosting to Publisher. The original delivery remained unchanged. No TikTok resend was performed by this recovery.

## Instagram processing after a recovered transfer

If a same-container retransmission returns an HTTP error but Meta still reports `IN_PROGRESS`, Publisher continues bounded read-only status polling instead of immediately returning Needs attention. A retained phase error `1363008` is tolerated only while the overall container is pending. Once ready, the caller publishes that same container; terminal errors, cancellation and uncertain publication retain their existing safeguards. No additional container or byte transfer is introduced by this wait.

Tests cover pending processing with and without a stale phase error, successful same-container publication, terminal failure, cancellation, and a bounded timeout. Full Go tests, vet and the production Windows build pass. The build was installed after the owner closed Publisher and a read-only database check found no active sending/publishing receipts. A residual windowless process was stopped before replacement. The installed SHA-256 matches the staged build (`d6be2b78a35c3954a186b27c0bdc6f358d708438eca2fe58293531148bd83140`); Publisher was restarted. Native upload behavior has not yet been verified. Installation did not initiate any upload. No frontend, dependency or schema change was made.


## TikTok inbox resends

Every account-specific acknowledged TikTok inbox receipt can be explicitly archived to enable another attempt. The action requires the exact current account, language, file hash, draft revision and receipt reference, plus explicit duplicate-arrival consent. It never sends bytes itself. Existing review, approval and final language confirmation remain required for the subsequent upload. Active transfers, uncertain results, Direct Post and other platforms remain locked.

Schema v12 adds durable inbox receipt history; archiving and unlocking are atomic, including confirmation timestamps. The send dialog opens Inbox / API for an existing inbox receipt and offers a collapsed history list. Tests cover reference/revision conflicts, language and account isolation, blocked states, repeated resets, persistence, consent, cancellation and no send before final confirmation. Go tests, vet, strict frontend build, browser resend and existing final-confirmation fixtures pass. Browser fixtures do not establish live TikTok notification delivery. Installed after Publisher closed and no active receipts remained. A SQLite backup precedes migration; schema v12 and executable SHA-256 `70d22a2cfcfeff89b75d1eb9e2059f934727996fb660a90d8c93dabd7b6a2f5f` were verified. Native window inspection confirmed that the app loaded. The prior manually archived inbox receipt was imported into the new history. No resend was performed during installation; native resend interaction remains owner-driven.


## Persistent TikTok inbox observations

Schema 13 adds account/reference-scoped provider observations and a durable polling schedule. A single cancellable worker resumes current and archived inbox receipts on startup, queries only the status endpoint, and preserves the latest known status on errors. Account identity joins, exact-reference reconciliation and a live-transfer grace period protect replacements, resends and concurrent completion writes. No dependencies, media upload, background Windows service or new publication permissions are added.

Posts details and the send dialog show provider status, last check, check failures and archived attempt results. Inbox acknowledgment explicitly differs from observed mobile receipt. Event-driven UI updates do not require frontend polling.

Migration preserves existing accounts, approvals, upload locks and receipts. Back up the SQLite database before installation; older builds reject schema 13. Downgrading requires restoring its matching backup while Publisher is closed.

Validation: provider status-only fixtures, migration/restart persistence, backoff, terminal stop, archive isolation, active-send protection and shutdown tests. Frontend fixtures exercise updates, failures and narrow layout. Native installation verified on November 12, 2026: the installed executable matches SHA-256 `290d6f5706c5640d5d234bb2b9698ec50a13d5166d5800522f5bcbd72f859bc8`, the window loaded, and the local database migrated to schema 13 after a SQLite backup. The worker saved 16 live provider observations without check errors during initial verification. Drafts, reviews, archived inbox receipts and upload identities/start times match the pre-installation backup. Installation and verification did not upload media. Provider inbox acknowledgment does not independently confirm notification visibility on the phone.

Next action: review TikTok account details in Posts during normal use. Keep Publisher open for automatic checks; closing it pauses the worker and reopening resumes the durable schedule.


## Instagram terminal-ingest recovery verification

On November 13, 2026, a new approved Spanish Reel failed during direct file ingestion: HTTP 400, then HTTP 500 during the single same-container retransmission. A fresh provider read confirmed terminal ERROR, code 2207082, an uploading-phase FILE_NOT_FOUND (1363008), and zero received bytes. The account media list confirmed no corresponding published Reel. The previous processing-wait correction does not resolve this terminal ingestion failure.

The owner authorized temporary HTTPS ingestion of this specific approved file and required exactly one publication. Recovery preserved a SQLite backup and the failed receipt, verified the approved source hash/review/account, prepared the existing compatibility copy, served only that file behind a random temporary URL, and reserved one replacement with an exclusive one-time marker. The replacement was published only after FINISHED. A separate account-media read verified exactly one matching Reel, and its provider timestamp was saved in Publisher. The temporary server and Cloudflare process were closed. No TikTok or other destination was sent.

This verifies recovery for this video, not a permanent fix to direct file ingestion. No app binary or automatic external hosting setting was changed. Next action: implement a deliberate, consented HTTPS transport option with durable recovery and duplicate prevention before claiming first-attempt Instagram reliability.
