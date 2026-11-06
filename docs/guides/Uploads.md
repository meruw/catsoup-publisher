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

Set each account's mode in **Delivery modes**. TikTok's selector is fixed to **TikTok inbox**; its **i** explains that Direct Post is not implemented and posting must be completed in TikTok. Instagram and X use a local draft when their mode is **Local draft**.

Watch and listen to the final video, check the text, destination and mode, then choose **Approve post** beside the selected account's send button. This confirms only that account's saved revision. Editing text, changing the mode or linking another video removes approval. **Undo approval** removes it explicitly. Opening the video or saving a draft does not approve it.

Only the selected platform has an individual send button. Its label describes the action: **Upload privately…**, **Send to TikTok inbox…**, **Publish on Instagram…**, or the applicable provider action. It opens a confirmation dialog and sends nothing until you confirm there. The **i** beside a disabled action explains its blocker.

**Publish all…** stays in the workspace header while scrolling. It is disabled until every connected account in the selected language has an approved current post and the delivery passes the sending checks. Its **i** lists missing approvals or other blockers. Disconnected platforms are not targets. In the dialog, check the saved modes, omit any destinations you do not want to send, provide YouTube's audience and acknowledge paid X usage if applicable. All connected accounts must still have current approvals, even if a destination is omitted.

The app rechecks account identities, approvals, selected revisions and receipts before the first send. Transfers run sequentially and stop on the first error. Completed transfers stay completed; later destinations show **Not sent**. There is no automatic retry. Already-sent and unresolved destinations are excluded from the batch; resolve uncertain attempts through the individual platform dialog. A local draft stays on this PC. X browser mode requires finishing in X.

**Production approval missing** below the video means the source delivery is blocked independently of account approval. Its **i** explains the reason. The production owner must approve the exact delivery through the [delivery contract](https://github.com/meruw/catsoup-publisher/blob/main/contracts/delivery/CONTRACT.md), then refresh Publisher. Publisher's **Approve post** never changes that manifest. Local preparation works before production approval; the demo and pilot cannot be sent.

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
