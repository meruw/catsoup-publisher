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
label. They are storage bounds, not platform API limits. No credential storage.

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
before replacing it. Account labels in the destination controls do not rename drafts;
**Open draft** opens or prepares a separate destination.

Linking a new draft or choosing **Use selected video** revalidates the delivery and
captures its current fingerprint and video hash. Normal text autosaves only touch
SQLite; they do not hash/render/decode media. Refresh the delivery to detect changes
made outside the app. Existing metadata is retained when a source changes.

Copying reads the saved draft. Folder access accepts episode/video IDs and resolves
them through the existing validated reader before calling the native file manager.
No arbitrary paths or commands are exposed. Neither operation approves or publishes
anything; copy/reveal are also usable with non-publishable test material.

## Migrations, backup and recovery

Schema version 2 adds upload receipts in a transaction and preserves existing
version-1 drafts. Unsupported newer versions and invalid/corrupt databases fail explicitly.
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
