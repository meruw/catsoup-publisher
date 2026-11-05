---
layout: guide
order: 3
title: Accounts and uploads
permalink: /guides/uploads/
---
# Upload an approved episode

Open **Accounts** in the sidebar. Select **ES** or **EN**, choose the platform and connect the intended account. In **Publication**, the draft's platform and language select that verified destination. Check its name and ID before confirming the upload. A draft label is for organizing text; editing it never switches accounts.

| Destination | What happens when you send |
| --- | --- |
| [YouTube](https://publisher.meruw.dev/guides/youtube/) | Upload privately. Review the video in YouTube Studio. |
| [TikTok](https://publisher.meruw.dev/guides/tiktok/) | Send a draft to the creator's inbox. Finish posting in TikTok. |
| [Meta](https://publisher.meruw.dev/guides/meta/) | Publish a Reel on the connected Instagram profile. |
| [X](https://publisher.meruw.dev/guides/x/) | Open the composer. Attach the video and submit in the browser. |

## Connect your accounts

Each workspace has one destination per platform and language. Connect both languages before creating an episode. **Change** replaces only the selected slot after successful authorization. **Remove** forgets its local credentials. Publisher prevents assigning one provider ID to both languages on the same platform. Drafts are routed using their language. The send screen shows the current assignment and requires confirmation.

In **Settings → Developer credentials**, select the provider. When credentials are missing, **Set up credentials** in Accounts takes you directly to that provider. Enter the developer credentials once for both languages. YouTube also accepts a Google desktop client JSON. The browser handles the provider's sign-in and returns to Publisher. Windows DPAPI protects saved user credentials for this Windows user and slot. Google and TikTok renew access using provider-issued refresh tokens when available. Reconnect from **Accounts** if a provider expires or revokes access. The synthetic demo and pilot are always preview-only.

## Check the result

Publisher revalidates the delivery and stages a temporary copy matching the approved hash before sending. SQLite records one attempt per platform, verified account ID and video hash. The same file can go to another account. A confirmed transfer cannot be repeated to the same destination through Publisher. Keep the app open while uploading. A receipt displays the returned reference.

If a response is lost, check the destination before doing anything else. Look in YouTube Studio (including private videos), the TikTok inbox or the professional Instagram profile. Only choose **Allow another attempt** after confirming the upload is absent. If Publisher exited during a transfer, an in-flight receipt becomes eligible for manual recovery after 30 minutes. **Refresh receipt** reads the local receipt. Publisher does not automatically retry, schedule posts or resume partial transfers.

Protocol and interface tests use local fixtures. A live transfer with an operator-owned account and an approved episode is still required to verify each provider end to end.
