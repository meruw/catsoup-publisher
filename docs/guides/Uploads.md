---
layout: guide
order: 3
title: Accounts and uploads
permalink: /guides/uploads/
---
# Edit and publish an episode

Open **Accounts** in the sidebar. Select **ES** or **EN**, choose the platform and connect the intended account. Each draft belongs to one episode, language, platform and provider account ID. Check the connected account before confirming a transfer. Replacing an account opens its own draft, even when its display name matches another account.

## Work on an episode

Select an episode and **ES** or **EN**. The video and post editor share one workspace. Choose YouTube, TikTok, Instagram or X in the destination selector to open that platform's draft. Connected account names appear in the selector; **Connect** opens Accounts when no destination is assigned. You can prepare unassigned text before connecting an account; explicitly copy it into the connected account's draft through **More → Recover unassigned text**.

Watch and listen to the final video, open **Video review** below the player, then confirm **Mark reviewed**. **Notes** opens optional review notes and **Request changes**. Decisions belong to the current manifest and video hashes; opening a preview never records a review. Clean versions remain available to preview, while review decisions apply to final versions. **Video details** contains version selection, technical checks and production approval information.

Write the title, description and hashtags beside the video. Text saves locally after typing pauses. Switching platforms, languages or modules flushes edits first; a failed save preserves the current draft. **Save draft** stays visible beside the destination selector. **More** opens copy actions and the video folder. **Recover unassigned text**, when available, copies text from an unbound draft for this platform into the current account; the original is retained. There is no combined selector of all accounts' drafts.

Each destination row shows its account and saved delivery mode: **Private video**, **TikTok inbox**, **Local draft**, **Publish now** or **Finish in X**. Modes persist per account and episode; all start in draft mode. Choose **Publish…** on a platform row to open its confirmation dialog. Opening it does not send anything. Close it or press Escape to return to editing; it cannot be dismissed during a transfer. A blocked delivery shows its reason; **Review video** returns to an unfinished review.

**Publish all…** opens a confirmation for the selected language's connected accounts and saved drafts. Check each destination and mode, uncheck any destination you want to omit, supply YouTube's audience and acknowledge any paid X request. The app checks all selected account IDs, draft revisions and local receipts before starting. Transfers run sequentially and stop on the first error; completed transfers remain completed and subsequent destinations show **Not sent**. There is no automatic retry. Already-sent and unresolved destinations are excluded from the batch. Resolve an uncertain attempt through that platform's individual dialog. Local draft mode keeps the text on this PC; X browser mode still requires finishing in X.

Local editorial review does not grant production approval or modify the delivery package. The production owner must provide an approved delivery under the [delivery contract](https://github.com/meruw/catsoup-publisher/blob/main/contracts/delivery/CONTRACT.md). Both that approval and a current local review are required for remote transfers and browser file handoff. Local preparation works before approval. The demo and pilot cannot be sent.

| Destination | Keep as draft | Publish now |
| --- | --- | --- |
| [YouTube](https://publisher.meruw.dev/guides/youtube/) | Private video on the selected channel | Public upload, or publish the recorded private video using its existing ID |
| [TikTok](https://publisher.meruw.dev/guides/tiktok/) | Video in the creator's inbox; finish posting in TikTok | Unavailable in this connector; Direct Post requires its own integration and provider approval |
| [Meta](https://publisher.meruw.dev/guides/meta/) | Local draft on this PC | Reel on the connected Instagram profile |
| [X](https://publisher.meruw.dev/guides/x/) | Local draft on this PC | Free browser handoff to finish in X, or direct posting with paid API access |

## Connect your accounts

Each workspace has one destination per platform and language. Connect both languages before creating an episode. **Change** replaces only the selected slot after successful authorization. **Remove** forgets its local credentials. Publisher prevents assigning one provider ID to both languages on the same platform. Drafts are routed using their language and bound account ID. Reconnecting the same provider ID restores that account's draft; connecting a different ID does not reuse its text. The send screen shows the current assignment and requires confirmation.

In **Settings → Developer credentials**, select the provider. When credentials are missing, **Set up credentials** in Accounts takes you directly to that provider. Enter the developer credentials once for both languages. YouTube also accepts a Google desktop client JSON. The browser handles the provider's sign-in and returns to Publisher. Windows DPAPI protects saved user credentials for this Windows user and slot. Google and TikTok renew access using provider-issued refresh tokens when available. Reconnect from **Accounts** if a provider expires or revokes access. The synthetic demo and pilot are always preview-only.

## Check the result

Publisher revalidates the delivery and stages a temporary copy matching the approved hash before sending. SQLite records one attempt per platform, verified account ID and video hash. The same file can go to another account. A confirmed transfer cannot be repeated to the same destination through Publisher. Keep the app open while uploading. A receipt displays the returned reference.

If a response is lost, check the destination before doing anything else. Look in YouTube Studio (including private videos), the TikTok inbox, the professional Instagram profile or the connected X account. Only choose **Allow another attempt** after confirming the upload is absent. If Publisher exited during a transfer, an in-flight receipt becomes eligible for manual recovery after 30 minutes. **Refresh receipt** reads the local receipt. Publisher does not automatically retry, schedule posts or resume partial transfers. X API publishing requires a separate cost acknowledgment for every attempt; failed or retried requests can also affect usage. Free browser handoffs do not create transfer receipts.

Receipts distinguish **Private video uploaded**, **Sent to inbox**, **Published** and unresolved attempts. Sending a video to TikTok's inbox does not mean it is published. For an uncertain YouTube visibility change, check Studio, then explicitly retry publishing the same ID. This does not unlock another upload; an interrupted visibility request can be retried after 30 minutes. A successful receipt records the provider's response at that time, not continuous monitoring of later manual changes.

Protocol and interface tests use local fixtures. A live transfer with an operator-owned account and an approved episode is still required to verify each provider end to end.
