---
layout: guide
order: 11
section: Services
title: TikTok
permalink: /guides/tiktok/
---
# Connect a TikTok account

Publisher can send an approved MP4 to the TikTok inbox or use optional Direct Post. Inbox uploads are completed in TikTok; Direct Post uses the privacy chosen in Publisher.

## Set up the desktop app
1. In [TikTok for Developers](https://developers.tiktok.com/), configure **Login Kit** for Desktop and the **Content Posting API** for inbox uploads.
2. Register `http://127.0.0.1:*/callback/` and request `user.info.basic` and `video.upload`. The wildcard lets Publisher choose an available local port.
3. For sandbox testing, add both accounts as target users. In **Settings → Manage credentials → TikTok**, paste the **Client key** and **Client secret** from that same sandbox. Save once for both languages.

## Connect each language
Select **Accounts → ES → TikTok → Connect → Sign in with TikTok**. TikTok can reuse the account active in your browser. Check the @username before granting access; use **Change account** if the wrong account is active. Repeat under **EN**. The ES/EN selector in Publisher controls where the resulting connection is stored.

## Send to the inbox
Review the video and edit its **TikTok** post, then choose **Approve post** and the send action above the editor. Confirm the destination and select **Send to TikTok inbox**. Open the inbox notification in TikTok, paste the prepared caption if desired and finish posting there. An inbox receipt does not mean the video is public.

[Desktop Login Kit](https://developers.tiktok.com/docs/en/login-kit-desktop) · [Inbox upload requirements](https://developers.tiktok.com/docs/en/content-posting-api-get-started-upload-content)

## Enable Direct Post

1. Enable **Direct Post** under **Content Posting API** in your TikTok developer app and obtain access to `video.publish` for the environment you are testing.
2. In **Settings → Manage credentials → TikTok**, turn on **Enable Direct Post**. It is off by default and uses the same client credentials and desktop callback as inbox uploads.
3. Reconnect each TikTok account in **Accounts**, separately under ES and EN, and grant `video.publish`. The account shows **Inbox + Direct Post** only when its saved grant includes that permission. An existing inbox grant is insufficient.
4. Open the episode, select the account's TikTok post and choose **Direct Post** above the editor. Edit the caption, save and approve that post. Choose **Post on TikTok…**.
5. Check the destination nickname and video preview. Choose visibility explicitly; there is no preset. Comments, Duet and Stitch start unchecked and respect the account's current limits. Declare commercial and AI-generated content when applicable, accept the linked posting terms and confirm the account and visibility.

Publisher refreshes creator information before sending, checks the permitted duration and caption length, uploads the local file and polls processing while the app stays open. A private receipt means **Only me**, not public publication. Friends/followers visibility is recorded separately. Processing failures are distinguished from uncertain network results. Use **Posts** to inspect the saved result; later changes made in TikTok are not monitored.

Direct Post is confirmed through the individual TikTok dialog, where the required privacy and disclosures can be reviewed. **Publish all** excludes that direct-post destination and explains why; inbox transfers can participate in the batch. The default for new TikTok drafts remains inbox mode. Disabling Direct Post does not rewrite saved drafts or approvals.

## TikTok review and visibility

An unaudited client is limited to private accounts and `SELF_ONLY` posts. Public visibility requires TikTok's audit. TikTok excludes utilities limited to uploading to accounts managed by the developer or their team. A public open-source repository alone does not establish eligibility or guarantee approval; a broadly available creator product must satisfy the review requirements.

Local videos use `FILE_UPLOAD`; the documentation site does not host the MP4. Direct Post has protocol and browser-fixture coverage in Publisher. Live authorization and a transfer with approved content still need validation for your developer client.

See TikTok's [Direct Post setup](https://developers.tiktok.com/docs/en/content-posting-api-get-started), [creator-info API](https://developers.tiktok.com/docs/en/content-posting-api-reference-query-creator-info) and [Content Sharing Guidelines](https://developers.tiktok.com/docs/en/content-sharing-guidelines).
