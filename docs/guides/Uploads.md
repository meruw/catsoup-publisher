---
layout: guide
order: 3
title: Accounts and uploads
permalink: /guides/uploads/
---
# Review, prepare and publish an episode

Open **Accounts** in the sidebar. Select **ES** or **EN**, choose the platform and connect the intended account. In the episode workspace, the draft's platform and language select that verified destination. Check its name and ID before confirming the upload. A draft label is for organizing text; editing it never switches accounts.

## Work through the episode

1. **Review:** watch and listen to the final video in each language. Keep notes, request changes or confirm **Mark reviewed**. Notes and decisions stay on this PC and belong to the current manifest and video hashes. Opening a preview never records a review. Changing the delivery requires a fresh review. Clean versions remain available to preview; review decisions apply to final versions.
2. **Prepare:** choose a platform and account label, then write the title, description and hashtags. Changes save locally after typing pauses. Switching steps, languages or modules flushes edits first; a failed save keeps you in place. Copy actions and the video folder are available here.
3. **Publish:** select a saved destination from the status table, check its text and connected account, then choose **Keep as draft** or **Publish now**. Confirm each transfer separately. Keep as draft is selected initially; its description explains whether the result stays local or goes to the platform.

Local editorial review does not grant production approval or modify the delivery package. The production owner must provide an approved delivery under the [delivery contract](https://github.com/meruw/catsoup-publisher/blob/main/contracts/delivery/CONTRACT.md). Both that approval and a current local review are required for remote transfers and browser file handoff. Local preparation works before approval. The demo and pilot cannot be sent.

| Destination | Keep as draft | Publish now |
| --- | --- | --- |
| [YouTube](https://publisher.meruw.dev/guides/youtube/) | Private video on the selected channel | Public upload, or publish the recorded private video using its existing ID |
| [TikTok](https://publisher.meruw.dev/guides/tiktok/) | Video in the creator's inbox; finish posting in TikTok | Unavailable in this connector; Direct Post requires its own integration and provider approval |
| [Meta](https://publisher.meruw.dev/guides/meta/) | Local draft on this PC | Reel on the connected Instagram profile |
| [X](https://publisher.meruw.dev/guides/x/) | Local draft on this PC | Free browser handoff to finish in X, or direct posting with paid API access |

## Connect your accounts

Each workspace has one destination per platform and language. Connect both languages before creating an episode. **Change** replaces only the selected slot after successful authorization. **Remove** forgets its local credentials. Publisher prevents assigning one provider ID to both languages on the same platform. Drafts are routed using their language. The send screen shows the current assignment and requires confirmation.

In **Settings → Developer credentials**, select the provider. When credentials are missing, **Set up credentials** in Accounts takes you directly to that provider. Enter the developer credentials once for both languages. YouTube also accepts a Google desktop client JSON. The browser handles the provider's sign-in and returns to Publisher. Windows DPAPI protects saved user credentials for this Windows user and slot. Google and TikTok renew access using provider-issued refresh tokens when available. Reconnect from **Accounts** if a provider expires or revokes access. The synthetic demo and pilot are always preview-only.

## Check the result

Publisher revalidates the delivery and stages a temporary copy matching the approved hash before sending. SQLite records one attempt per platform, verified account ID and video hash. The same file can go to another account. A confirmed transfer cannot be repeated to the same destination through Publisher. Keep the app open while uploading. A receipt displays the returned reference.

If a response is lost, check the destination before doing anything else. Look in YouTube Studio (including private videos), the TikTok inbox, the professional Instagram profile or the connected X account. Only choose **Allow another attempt** after confirming the upload is absent. If Publisher exited during a transfer, an in-flight receipt becomes eligible for manual recovery after 30 minutes. **Refresh receipt** reads the local receipt. Publisher does not automatically retry, schedule posts or resume partial transfers. X API publishing requires a separate cost acknowledgment for every attempt; failed or retried requests can also affect usage. Free browser handoffs do not create transfer receipts.

Receipts distinguish **Private video uploaded**, **Sent to inbox**, **Published** and unresolved attempts. Sending a video to TikTok's inbox does not mean it is published. For an uncertain YouTube visibility change, check Studio, then explicitly retry publishing the same ID. This does not unlock another upload; an interrupted visibility request can be retried after 30 minutes. A successful receipt records the provider's response at that time, not continuous monitoring of later manual changes.

Protocol and interface tests use local fixtures. A live transfer with an operator-owned account and an approved episode is still required to verify each provider end to end.
