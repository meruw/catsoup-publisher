# Local publication drafts

## Storage and identity

`data/publisher/drafts.sqlite` belongs to the selected workspace (`--root`), not
the executable directory. SQLite runs inside Go through pinned `modernc.org/sqlite`
and `database/sql`; no C toolchain, server or database process is required.
One connection per process, a 3-second busy timeout and FULL synchronous writes
keep this small desktop store simple. It uses SQLite's default rollback journal.

A draft is identified by episode ID, language, platform and account label. Labels
are case-sensitive and trimmed when entered. They are not credentials or verified
platform accounts. Platform choices are YouTube, TikTok, Instagram and X.
Each key has one current draft; full edit history and deletion are not implemented.

Stored fields are title, description, hashtags, linked video ID/hash, delivery
manifest fingerprint, monotonic revision and UTC update time. Local limits are
500/10,000/2,000 characters for title/description/hashtags and 100 for the account
label. They are storage bounds, not platform API limits. Draft text contains no credentials; the separate accounts table stores encrypted connection data.

## Editorial review

The workspace shows video review and post editing together for ES/EN. Publishing
options open beside the same preview. Each final
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
before replacing it. Platform buttons open a draft using the connected account label
or a local platform/language label. Existing drafts remain accessible under
**Saved drafts**. Account labels there do not rename drafts; **Open draft** opens
or prepares a separate labeled draft without changing the connected destination.

Linking a new draft or choosing **Use selected video** revalidates the delivery and
captures its current fingerprint and video hash. Normal text autosaves only touch
SQLite; they do not hash/render/decode media. Refresh the delivery to detect changes
made outside the app. Existing metadata is retained when a source changes.

Copying reads the saved draft. Folder access accepts episode/video IDs and resolves
them through the existing validated reader before calling the native file manager.
No arbitrary paths or commands are exposed. Neither operation approves or publishes
anything; copy/reveal are also usable with non-publishable test material.

## Migrations, backup and recovery

Schema version 4 contains drafts, language-specific accounts, encrypted credentials,
upload receipts and local reviews. Transactional migration preserves existing data. Restored credentials
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
