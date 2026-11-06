# Local publication drafts

## Storage and identity

`data/publisher/drafts.sqlite` belongs to the selected workspace (`--root`), not
the executable directory. SQLite runs inside Go through pinned `modernc.org/sqlite`
and `database/sql`; no C toolchain, server or database process is required.
One connection per process, a 3-second busy timeout and FULL synchronous writes
keep this small desktop store simple. It uses SQLite's default rollback journal.

A draft is identified by episode ID, language, platform and a provider account key
(`account:<provider-id>`). Display names are not identity. Each connected account
has one current draft per episode; replacing a connection does not reuse another
account's text. Reconnecting the same ID restores its draft. Full edit history
and deletion are not implemented.

Stored fields are title, description, hashtags, delivery mode (`draft` or `publish`),
linked video ID/hash, manifest fingerprint, revision, approved revision and UTC update time. Every new
draft starts in draft mode; TikTok rejects public mode. Storage limits are
500/10,000/2,000 characters for title/description/hashtags and 100 for the account
key. They are not platform API limits. Credentials are stored separately and encrypted.

## Account approval

**Approve post** records `approved_revision` for exactly one episode/language/platform/account key. The bridge verifies the final video hash, manifest fingerprint, language, current draft revision and connected account identity. Approval never changes the production manifest. Text, mode and linked-source saves increment the revision and clear approval. A no-op UI Save draft does not write a new revision.

Remote transfers, YouTube visibility changes and browser file handoff require a current account approval and the contract's production approval. SQLite rechecks approval atomically when reserving transfers and promotions, so an in-memory approved draft cannot override a revoked decision. The source is revalidated before sending. Local preparation and local account approval are available without production approval; test deliveries remain non-publicable.

The review table and bridge remain supported for label-keyed records. They do not approve account-bound drafts. The workspace uses the per-account post action and does not display review notes or a changes-requested workflow.

## Saving and conflicts

The frontend debounces typing by 600ms and serializes writes. When text changes
during a save, the response updates the revision without replacing newer text,
then the next write saves that text. Navigation and normal native close requests
flush the queue. A save error leaves edits in memory and blocks leaving the editor.
An OS kill, crash or power loss can still lose changes not yet acknowledged as saved.

SQL statements use parameters. Inserts refuse duplicate keys; updates require the
expected revision. Two writers cannot silently overwrite each other's revision.
The UI's reload action requires a second explicit choice before discarding edits.
Reload is not automatic after a conflict. Users can select/copy unsaved text manually
before replacing it. Save draft remains visible and mode changes use the same
revision checks. Navigation is blocked while a mode write is in progress.
Unbound text is stored under `unassigned`; other label-keyed drafts remain intact.
Restore draft exposes explicit text recovery scoped to the selected platform, copying into
the connected account without changing the source or any existing receipts.
Drafts for disconnected provider IDs remain stored and reappear on reconnection.

Linking a new draft or choosing **Use selected video** revalidates the delivery and
captures its current fingerprint and video hash. Normal text autosaves only touch
SQLite; they do not hash/render/decode media. Refresh the delivery to detect changes
made outside the app. Existing metadata is retained when a source changes.

Copying reads the saved draft. Folder access accepts episode/video IDs and resolves
them through the existing validated reader before calling the native file manager.
No arbitrary paths or commands are exposed. Neither operation approves or publishes
anything; copy/reveal are also usable with non-publishable test material.

## Migrations, backup and recovery

Schema version 7 contains drafts with persisted delivery modes and approved revisions, language-specific
accounts, encrypted credentials, upload receipts, local reviews and publishing defaults.
Defaults are keyed by platform, language and connected provider account ID. They
initialize only an unsaved draft; changing a default never edits or approves a saved
post. A replacement account does not inherit the previous account's default.
TikTok defaults accept inbox mode only. Transactional
migration initializes approval to zero while preserving all text, revisions, links
and receipts. Restored credentials
require the Windows user and DPAPI keys that protected them. Unsupported newer versions and invalid/corrupt databases fail explicitly.
The app never silently deletes or resets a database. Future migrations must add
versioned transactional steps and tests preserving existing drafts.

For a backup, close every Publisher instance using that workspace and copy the
entire `data/publisher/` folder to a safe local backup location. Do not copy a live
database alone while writes are in progress. Git deliberately does not back it up.
Restore only with Publisher closed, preserving the original folder first. If the
schema is newer, use a compatible application build; do not edit the version number.
Keep a damaged store for diagnosis/recovery instead of overwriting it.

## Verification boundaries

Go tests cover reopening/migration, Unicode and SQL-looking text, destination
isolation, concurrent/stale writers, invalid input, future/corrupt DB preservation,
unchanged manifests, source rebinds and restricted clipboard/folder arguments.
Browser fixtures cover autosave, in-flight editing, language changes, copy/reveal
dispatch, conflicts, failed-save navigation/close protection and reloading.
Fixtures do not verify real clipboard/file-manager behavior or native close timing.

Receipts record private, restricted, inbox, published, sending, failed or check states. TikTok FAILED responses are definitive; transport failures stay uncertain. YouTube
promotion reserves publishing for the same remote ID; failures become check_publish.
That state cannot be cleared to permit a duplicate upload. An abandoned promotion
can be explicitly retried after 30 minutes. Older successful YouTube receipts are
interpreted as private uploads. No background polling or job framework is required.

## Individual and batch transfers

Individual TikTok Direct Post uses typed, validated per-send privacy and disclosure options and is excluded from batch transfer. Other entry points use the shared frontend transfer adapter and the same Go
publication methods, account checks and durable receipt reservations. A batch is
an explicit, foreground sequence for one language, not a background queue. Before
the first send it refreshes all connected account identities and approvals, plus every selected draft revision and receipt. The header action is blocked until all connected accounts in the current language have approved posts.
New account-bound keys are checked against the destination in Go as well as the UI.
All remote sends retain account approval, production approval and file-integrity validation in Go.
The sequence stops on error and cannot be dismissed while busy. It never rolls
back completed publications or retries automatically. Partial results are shown in
the dialog; remote transfer receipts survive restart. Local-only and X browser
handoff results are not remote publication receipts. Uncertain YouTube promotion
recovery remains an explicit individual action, outside batch sending.

## Post overview

`PostStatuses` joins saved drafts to receipts by platform, verified account ID and video hash, and to connected accounts by platform, language and ID. It exposes no credentials and makes no provider request. Search is literal and parameterized; filters are allowlisted and results are bounded to 50 rows with offset pagination. Disconnected drafts remain visible. Transfers abandoned for more than 30 minutes appear as Check result. The overview follows each draft's current file, rather than retaining immutable snapshots of every edited post.
