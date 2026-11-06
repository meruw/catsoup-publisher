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
linked video ID/hash, manifest fingerprint, revision and UTC update time. Every new
draft starts in draft mode; TikTok rejects public mode. Storage limits are
500/10,000/2,000 characters for title/description/hashtags and 100 for the account
key. They are not platform API limits. Credentials are stored separately and encrypted.

## Editorial review

The workspace shows video review and post editing together for ES/EN. Publishing
options open in a native browser dialog over the editor. Each final
video has a local review keyed by episode, language and video ID, with its manifest
fingerprint, video hash, notes, decision, revision and timestamp. Decisions are
pending, changes_requested or reviewed. Only an explicit watched/listened
confirmation records reviewed; typing notes and leaving saves them as pending.
The bridge revalidates the current delivery and refuses stale sources and other
languages. Review writes use optimistic revision checks and never modify manifests.

API transfers, YouTube visibility changes and browser file handoff require both a
matching reviewed record and the delivery contract's existing owner approval.
SQLite checks the review again when reserving a transfer or visibility change.
Changing the manifest invalidates a review even when its video bytes are unchanged.
Notes remain available for the next review. Local metadata preparation does not
require approval. The production demo remains non-publicable.

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
More exposes explicit text recovery scoped to the selected platform, copying into
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

Schema version 5 contains drafts with persisted delivery modes, language-specific
accounts, encrypted credentials, upload receipts and local reviews. Transactional
migration adds the draft mode default while preserving all text, revisions, links
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

Receipts record private, inbox, published, sending or check states. YouTube
promotion reserves publishing for the same remote ID; failures become check_publish.
That state cannot be cleared to permit a duplicate upload. An abandoned promotion
can be explicitly retried after 30 minutes. Older successful YouTube receipts are
interpreted as private uploads. No background polling or job framework is required.

## Individual and batch transfers

Both entry points use the shared frontend transfer adapter and the same Go
publication methods, account checks and durable receipt reservations. A batch is
an explicit, foreground sequence for one language, not a background queue. Before
the first send it refreshes every selected account, draft revision and receipt.
New account-bound keys are checked against the destination in Go as well as the UI.
All remote sends retain approval, review and file-integrity validation in Go.
The sequence stops on error and cannot be dismissed while busy. It never rolls
back completed publications or retries automatically. Partial results are shown in
the dialog; remote transfer receipts survive restart. Local-only and X browser
handoff results are not remote publication receipts. Uncertain YouTube promotion
recovery remains an explicit individual action, outside batch sending.
