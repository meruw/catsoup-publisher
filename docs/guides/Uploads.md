---
layout: guide
order: 3
title: Accounts and uploads
permalink: /guides/uploads/
---
# Edit and publish an episode

Open **Accounts** in the sidebar. Select **ES** or **EN**, choose the platform and connect the intended account. Each draft belongs to one episode, language, platform and provider account ID. Check the connected account before confirming a transfer. Replacing an account opens its own draft, even when its display name matches another account.

## Work on an episode

Select an episode and **ES** or **EN**. The video and post editor share one workspace. Choose a platform above the editor to open its connected account's draft. The **Version** selector stays below the player; **Video details** expands only technical information.

Write the title, description and hashtags. Edits save locally after typing pauses; **Save draft** gives visible confirmation beside the button. Switching platforms, languages or modules flushes edits first. A failed save retains your text and blocks navigation.

**Restore draft**, beside Save draft, reloads this account's last saved text. Autosave means this is not an edit history or undo function. It also offers explicit recovery of unassigned text for the selected platform, when present. Recovery copies the text into the connected account and preserves its source. Drafts from different accounts never share a selector.

Set the default for each account in **Settings â†’ Manage publishing**, on the Publishing defaults page under ES or EN. Defaults apply only to new drafts; existing text, modes and approvals stay unchanged. The mode selector above the post editor controls that individual post. TikTok offers **TikTok inbox** and, after optional setup and authorization, **Direct Post**. The **Set up TikTok Direct Post** link opens its developer credentials. Hover or focus a disabled selector for its explanation. Instagram and X use a local draft when their mode is **Local draft**.

Watch and listen to the final video, check the text, destination and mode, then choose **Approve post** below the text fields, alongside Save draft and Restore draft. This confirms only that account's saved revision. Editing text, changing the mode or linking another video removes approval. **Undo approval** removes it explicitly. Opening the video or saving a draft does not approve it.

Only the selected platform has an individual send button. Its label describes the action: **Upload privatelyâ€¦**, **Send to TikTok inboxâ€¦**, **Publish on Instagramâ€¦**, or the applicable provider action. It opens a confirmation dialog and sends nothing until you confirm there. Hover or focus a disabled action to see its blocker.

**Publish allâ€¦** stays in the workspace header while scrolling. It is disabled until every connected account in the selected language has an approved current post and the delivery passes the sending checks. Hover or focus the button to see missing approvals or other blockers. Disconnected platforms are not targets. In the dialog, check the saved modes, omit any destinations you do not want to send, provide YouTube's audience and acknowledge paid X usage if applicable. All connected accounts must still have current approvals, even if a destination is omitted.

Every individual send, batch and TikTok browser preparation ends with a separate **language confirmation**. Check the prominent **Spanish / ES** or **English / EN** label, play the exact linked video with audio, and review the account and delivery mode for each destination. Select **I confirm this is the English/Spanish version for these accounts**, then **Send English/Spanish video** (or **Prepare** for the browser). Opening this review sends nothing. **Back** returns to the options without sending and clears the final acknowledgment. Changes to the video, language or destination invalidate the review. The label comes from the delivery metadata; Publisher does not detect the spoken language automatically.

The app rechecks account identities, approvals, selected revisions and receipts before the first send. Platforms run independently, with sequential transfers within each platform. A failed destination does not stop other approved sends; completed transfers stay completed. There is no automatic retry. Already-sent and unresolved destinations are excluded from the batch; resolve uncertain attempts through the individual platform dialog. A local draft stays on this PC. X browser mode requires finishing in X. TikTok Direct Post requires its individual confirmation for privacy and disclosure choices and is excluded from the batch.

**Production approval missing** below the video means the source delivery is blocked independently of account approval. Hover or focus that status to see the reason. The production owner must approve the exact delivery through the [delivery contract](https://github.com/meruw/catsoup-publisher/blob/main/contracts/delivery/CONTRACT.md), then refresh Publisher. Publisher's **Approve post** never changes that manifest. Local preparation works before production approval; the demo and pilot cannot be sent.

| Destination | Keep as draft | Publish now |
| --- | --- | --- |
| [YouTube](https://publisher.meruw.dev/guides/youtube/) | Private video on the selected channel | Public upload, or publish the recorded private video using its existing ID |
| [TikTok](https://publisher.meruw.dev/guides/tiktok/) | Video in the creator's inbox; finish posting in TikTok | Optional Direct Post with explicit visibility and TikTok authorization |
| [Meta](https://publisher.meruw.dev/guides/meta/) | Local draft on this PC | Reel on the selected Instagram profile or Facebook Page |
| [X](https://publisher.meruw.dev/guides/x/) | Local draft on this PC | Free browser handoff to finish in X, or direct posting with paid API access |

## Connect your accounts

Each workspace has one destination per platform and language. Connect both languages before creating an episode. **Change** replaces only the selected slot after successful authorization. **Remove** forgets its local credentials. Publisher prevents assigning one provider ID to both languages on the same platform. Drafts are routed using their language and bound account ID. Reconnecting the same provider ID restores that account's draft; connecting a different ID does not reuse its text. The send screen shows the current assignment and requires confirmation.

In **Settings â†’ Developer credentials**, select the provider. When credentials are missing, **Set up credentials** in Accounts takes you directly to that provider. Enter the developer credentials once for both languages. YouTube also accepts a Google desktop client JSON. The browser handles the provider's sign-in and returns to Publisher. Windows DPAPI protects saved account tokens for this Windows user and slot, and developer configurations for this Windows user and provider. Existing developer files migrate automatically without reconnecting accounts. Imported source JSON files elsewhere on disk are not changed. Google and TikTok renew access using provider-issued refresh tokens when available. Reconnect from **Accounts** if a provider expires or revokes access. The synthetic demo and pilot are always preview-only.

## Posts overview

Open **Posts** in the sidebar for one row per saved video, with a small video thumbnail, its languages, platform count and a status summary. Search videos or accounts. **Filters** opens platform, language and status choices; select **Apply filters** to use them. Active filters remain visible beside **Clear filters**. Status choices are **Not sent** (drafts and approved posts), **In progress**, **Sent** (public, private, limited-visibility and inbox deliveries), and **Needs attention** (failed or uncertain results). Account details retain the precise outcome.

Select any video row to see its destinations. The language tabs narrow that list; selecting an account opens its status, version and transfer reference. **Open post** returns to that account's editor. An Instagram result that needs checking also offers **Review Instagram upload**, which opens the existing-upload confirmation directly. Opening it does not send or publish anything. Navigation errors appear inside the detail dialog. Disconnected accounts remain readable.

The list has 20 videos per page. Grouping happens before pagination, so a video's destinations stay together. Filters select videos with a matching destination; their summaries still count every destination. A summary such as **2 of 8 published** does not imply the other accounts have been published. Destination lists use 50 rows per page when needed.

Statuses distinguish local **Draft**, **Approved**, **Private video**, **TikTok inbox**, **Limited visibility**, **Published**, **Failed** and **Check result**. The last label means the result is uncertain, not that publishing failed. The view shows the file currently linked to each saved post and its local receipt. **Refresh posts** reloads local records; it does not query social networks or reconstruct remote history. Browser handoffs to X cannot establish a published status.

## Publication dates and times

Each destination in **Posts → video → Posts** shows its recorded date and time below the status. Open an account for the publication date, latest attempt start and last local update. Dates use this PC's local time zone, shown above the destination list and in the detail dialog; storage uses UTC.

- **Published** uses the publication timestamp returned by YouTube for a public video or Instagram for the published media. Opening statistics or choosing **Refresh all** retrieves and saves these dates for existing posts, using the same read request as the counters.
- **Publication confirmed** records when Publisher received a successful publishing result. It is a confirmation time, not an independently reported platform publication timestamp. Facebook and optional direct API posts use this record; YouTube and Instagram can replace the displayed fallback with their provider date when statistics load.
- **Private upload confirmed**, **Sent to inbox** and **Prepared in browser** identify their actual events. A TikTok inbox confirmation is not the arrival time of the mobile notification or the time you publish from your phone.
- **Latest attempt started** is a transfer start, not publication. Draft updates and private YouTube upload dates are never used as public publication dates.

Dates survive app restarts. Re-reading a receipt does not reset its confirmation time. Older receipts without a recorded confirmation remain undated unless a supported provider returns their publication date. Facebook video creation dates are not substituted for publication; manual TikTok/X publishing cannot be dated from an inbox or browser handoff. **Publication time unavailable** states that limit explicitly. Refreshing statistics does not send or republish videos.

## Choose a thumbnail

Open **Thumbnail** in the selected account's editor before approving the post. The selection belongs to that account and language, saves with the draft, and changing it revokes approval. Changing the linked video clears its cover. Images stay in the local database; previews load only when needed.

| Destination | Options in Publisher |
| --- | --- |
| YouTube | Choose a video frame or import PNG/JPG, up to 2 MiB and 4096 pixels per dimension |
| Instagram Reels | Choose a video frame before the upload starts |
| TikTok Direct Post | Choose a video frame before the upload starts |
| TikTok Inbox | Choose the cover in TikTok's mobile editor; the inbox API does not accept it |
| Facebook / X | Custom covers are not implemented in these connectors |

For **Choose frame**, pause the preview and select **Use this frame**. For **Choose image**, use the app button to select a PNG/JPG. YouTube receives the selected image after processing the video. If that step fails or times out, the video stays published/private and **Apply thumbnail** retries only the image. A changed thumbnail must be saved and approved again. The channel must be eligible for custom thumbnails. Provider surfaces may display thumbnails differently.

Instagram and TikTok Direct Post receive a frame timestamp. Their existing uploads cannot change cover through Publisher. Local image files are not uploaded to an external hosting service for Instagram.

## Read post statistics

Open **Posts**, choose a video, then **Statistics**. Counts are grouped by platform and separated by account and language. They load automatically when this view or an individual post detail opens. Successful results are reused for five minutes within the app session; **Refresh statistics** requests fresh values for one account. **Refresh all** updates the eligible accounts displayed for the current video and language filter, with two requests at a time. Its progress reports individual failures without stopping the remaining accounts; failed reads keep the last successful counts and timestamp. There is no continuous background polling or paid statistics API. Click outside a detail dialog, press Escape or use Close to dismiss it. Dialogs cannot be dismissed during a blocking operation.

| Platform | Available in Publisher |
| --- | --- |
| YouTube | Views, likes and comments returned by the video API |
| Instagram | Likes and comments |
| Facebook Page Reels | Likes and comments; likes exclude other reactions |
| TikTok | Not implemented; needs `video.list` and the published video ID, which an inbox receipt does not provide |
| X | Not queried, to avoid paid API usage |

Counts use the connected account and the saved receipt for this exact video file. Missing or hidden counters are omitted, not replaced with zero. The timestamp describes when counts were fetched, not when the provider last updated them. Providers may lag. If a refresh fails, the last successful counts remain visibly dated. Recent responses are cached in memory for up to 100 post identities and reused for five minutes; successful observations are also saved locally for the chart. Closing a view does not clear this session cache. Reconnecting an account can be necessary if permissions expire.

Select a metric above the chart to see its observed totals, and choose the last 7, 28 or 90 days or the last year. Hover over the chart or focus it and use the arrow keys to inspect exact measurements. History starts with this PC's first successful read; it does not reconstruct activity before that date. A single observation shows a first-measurement message. Missing metrics stay absent rather than becoming zero.

Measurements are stored locally in the workspace database, separated by platform, account and published video. The latest successful observation in each five-minute interval is retained, up to 2,000 observations per post and one year of history. Old samples are pruned on successful reads. Opening the view or pressing Refresh records a measurement; Publisher does not poll in the background. These are cumulative counters observed at those times, not daily views or a retention curve.

These are individual platform counters, not a cross-platform reach total. Retention, revenue, reach and watch time are not included. YouTube's [video statistics](https://developers.google.com/youtube/data/reference/rest/v3/videos#statistics) use that platform's definitions. The request consumes normal API quota; Publisher does not buy quota or enable billing.

## Uploads in the background


Confirm an individual send or a batch to add its API transfers to **Uploads**, at the foot of the sidebar. The confirmation closes and you can switch videos, languages or modules while Publisher sends. **In progress** groups active work by video. Expand a row to see its accounts, language, mode, elapsed time and result. Completed video groups leave this view automatically; repeated attempts do not add more permanent rows. Platforms run independently, with one active transfer per platform and at most five overall. A processing wait on Instagram does not delay YouTube, Facebook or TikTok. Transfers within the same platform run in order; one failed destination does not stop the rest.

**Waiting** means the file has not been sent. You can cancel a waiting entry. **Sending / awaiting platform confirmation** covers validation, transfer and provider processing; it is not a percentage or proof of publication. Finished entries show the actual outcome, including private and inbox transfers. **Needs attention** includes uncertain results: read Details and check the destination before trying again. Posts refreshes local results when a transfer finishes.

The queue keeps the confirmed account and draft revision. Editing or revoking a queued post's approval causes the backend's checks to reject that changed revision before sending. Duplicate active transfers for the same account and file are blocked. Free X browser handoff remains an immediate clipboard/browser action.

Keep Publisher running; minimizing is supported. Closing the window during an active or waiting transfer opens Uploads instead. This is a session queue, not a Windows background service: quitting forcibly, rebooting or a crash stops work. Unsent queue entries are not restored; existing transfer receipts remain available for manual reconciliation. The queue does not automatically repeat failed jobs. The Instagram connector has the bounded recovery described below.

**History** lists the latest saved transfer results by video, including previous sessions. Search by saved title or account, and use 20-video pagination. Select a row to open its destinations in Posts. This reads existing local receipts; it neither queries platforms nor restarts uploads. It is not a permanent journal of every attempt: cancelled, unsent queue entries are not recorded, and the current draft/file determines the displayed receipt.

## Find an episode

The sidebar shows **Pending episodes**, up to six per page, with search by title or identifier. An episode leaves this list when every language in its current delivery has at least one connected non-X account and every such account has a successful receipt for its linked video: **Published**, or **TikTok inbox** for TikTok. X never blocks completion, including paid API accounts. It does not matter whether sends were individual or part of Publish all.

Private uploads, limited visibility, browser preparation, failures and uncertain results remain pending. Missing account drafts also remain pending. The check uses the current delivery fingerprint, account identity and video hash: a new delivery version or replacement account can bring the episode back. Receipt changes update the list after a transfer completes, without remote polling. If the completion lookup fails, episodes stay visible.

Completed episodes are still available in **Posts**, including their editor, statistics and receipts. No files or drafts are deleted or archived. Finishing a transfer does not forcibly close the currently open editor. Selecting Publisher or refreshing chooses an available pending episode; if all are complete, the workspace offers **View Posts**.

## Check the result

Publisher revalidates the delivery and stages a temporary copy matching the approved hash before sending. SQLite records one attempt per platform, verified account ID and video hash. The same file can go to another account. A confirmed transfer cannot be repeated to the same destination through Publisher. Keep the app open while uploading. A receipt displays the returned reference.

If a response is lost, check the destination before doing anything else. Look in YouTube Studio (including private videos), the TikTok inbox, the professional Instagram profile, the Facebook Page or the connected X account. Only choose **Allow another attempt** after confirming the upload is absent. If Publisher exited during a transfer, an in-flight receipt becomes eligible for manual recovery after 30 minutes. **Refresh receipt** reads the local receipt. Publisher does not schedule posts or resume partial binary transfers. Automatic replacement is limited to the confirmed Instagram failure described below. X API publishing requires a separate cost acknowledgment for every attempt; failed or retried requests can also affect usage. Free browser handoffs do not create transfer receipts.

Receipts distinguish **Private video uploaded**, **Sent to inbox**, **Published** and unresolved attempts. Sending a video to TikTok's inbox does not mean it is published. For an uncertain YouTube visibility change, check Studio, then explicitly retry publishing the same ID. This does not unlock another upload; an interrupted visibility request can be retried after 30 minutes. A successful receipt records the provider's response at that time, not continuous monitoring of later manual changes.

For an Instagram receipt marked **Failed** or **Check result**, open that account's post and choose **Check and finish on Instagram**. Publisher checks the saved container with Meta. If it is ready, Publisher publishes that same container without uploading the video again. A pending, published or unreadable result blocks a replacement. Only a confirmed failed or expired container can be replaced. Approval, file integrity and account identity are checked again; an interrupted publication remains locked for reconciliation.

If Meta reports a missing file (error `1363008`, zero transferred bytes and zero source size) while the container remains pending, that same action can retransmit the original prepared file to the **same container**. The prepared hash must match the saved checkpoint. Concurrent recovery is blocked, no new container is created, and publishing requires Meta to confirm readiness. An uncertain publish response remains locked; it does not permit another upload. If the retransmission returns an HTTP error while Meta still reports `IN_PROGRESS`, Publisher continues checking the same container for up to ten minutes. A stale missing-file phase error during processing does not immediately end the job. Readiness allows publication of that container; a terminal failure or timeout retains its receipt for reconciliation.

Instagram sends use a temporary compatibility copy: the video stream is copied unchanged and audio is encoded as AAC, 48 kHz stereo, targeting 128 kb/s. The MP4 places its metadata before media data and omits edit lists, as required by Meta. Timestamp shifting preserves relative audio/video timing. Resuming an existing transfer reproduces its saved preparation version and still requires an identical hash. The original delivery is never overwritten. Publisher verifies the approved source hash, checks that the video stream is identical, verifies the prepared file hash and records both hashes with the upload container. The temporary copy is removed when the send ends. This uses the local FFmpeg dependency already installed with Publisher, without a hosting service or paid conversion API.

For error `1363008` with explicit zero transferred bytes and zero source size, Publisher waits 15 seconds and checks again. If the same container still reports that empty-transfer failure, it retransmits once to that container. Only a terminal failed, empty container permits one replacement. Other pending containers are polled; a final `FINISHED` status takes precedence over stale phase errors. A saved Instagram transfer left in Sending for more than 30 minutes becomes Check result when its receipt opens, retaining its container and hashes. An ambiguous publishing result stays locked. Keep Publisher open during processing. A network error or interrupted publication never triggers an automatic replacement. Details appear in Console.

Protocol and interface tests use local fixtures. A live transfer with an operator-owned account and an approved episode is still required to verify each provider end to end.

If Meta rejects the `video_status` field with `100/2207065`, Publisher reads the supported container status separately. A terminal `ERROR` or `EXPIRED` is shown as a processing failure, with its numeric code when available; it is not left as an unexplained pending upload. Removing edit lists improves format compliance but does not guarantee that Meta will accept a transfer.

An inbox receipt confirms TikTok's API acknowledgment, not that a mobile notification was observed. Notifications can be delayed. Repeated sends can produce delayed duplicates, so keep the receipt and avoid resending an acknowledged transfer solely because its notification has not appeared.


### Resend a TikTok inbox transfer

For any Spanish or English video with an acknowledged inbox transfer, open its TikTok post and choose **Inbox / API → Resend to inbox…**. Confirm the account and language and acknowledge that the earlier copy may still arrive. **Enable another send** archives that exact receipt and unlocks one new attempt; it does not upload anything. Continue through the normal approval and final language confirmation to send.

**Previous inbox receipts** keeps the earlier references and confirmation times on this PC, including after restarting Publisher. An active transfer, Direct Post or another platform cannot be unlocked by this action. A resend does not cancel the original transfer or guarantee faster notification. If both copies arrive, publish only the intended one.
