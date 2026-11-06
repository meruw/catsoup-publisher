---
layout: guide
order: 11
section: Services
title: TikTok
permalink: /guides/tiktok/
---
# Connect a TikTok account

Publisher sends an approved MP4 to the TikTok inbox. The creator edits the draft and finishes posting in TikTok.

## Set up the desktop app
1. In [TikTok for Developers](https://developers.tiktok.com/), configure **Login Kit** for Desktop and the **Content Posting API** for inbox uploads.
2. Register `http://127.0.0.1:*/callback/` and request `user.info.basic` and `video.upload`. The wildcard lets Publisher choose an available local port.
3. For sandbox testing, add both accounts as target users. In **Settings → Manage credentials → TikTok**, paste the **Client key** and **Client secret** from that same sandbox. Save once for both languages.

## Connect each language
Select **Accounts → ES → TikTok → Connect → Sign in with TikTok**. TikTok can reuse the account active in your browser. Check the @username before granting access; use **Change account** if the wrong account is active. Repeat under **EN**. The ES/EN selector in Publisher controls where the resulting connection is stored.

## Send to the inbox
Review the video and edit its **TikTok** post, then choose **Approve post** and the send action above the editor. Confirm the destination and select **Send to TikTok inbox**. Open the inbox notification in TikTok, paste the prepared caption if desired and finish posting there. The dialog offers only an inbox transfer; Direct Post requires its own connector flow and TikTok approval. An inbox receipt does not mean the video is public.

[Desktop Login Kit](https://developers.tiktok.com/docs/en/login-kit-desktop) · [Inbox upload requirements](https://developers.tiktok.com/docs/en/content-posting-api-get-started-upload-content)

## Direct Post requirements

Direct Post is not implemented in Publisher. Enabling the switch in the TikTok portal alone does not add it to the app.

- Enable Direct Post for the Content Posting API, obtain approval for `video.publish`, and ask connected users to authorize that scope. Inbox permission (`video.upload`) does not grant direct publishing.
- Implement creator-info lookup, a posting screen, Direct Post initialization, local-file transfer and processing-status checks. The screen must show the destination and preview, let the creator edit the caption and choose an available privacy option without a preset selection, and respect interaction limits. Include the required commercial-content disclosures and music consent.
- An unaudited client is restricted to private visibility. Public posting requires TikTok's audit. TikTok explicitly excludes utilities limited to uploading to accounts managed by the developer or their team. A public open-source repository alone does not establish eligibility or guarantee approval; a broadly available creator product must satisfy the review requirements.

Local videos can use `FILE_UPLOAD`; hosting the MP4 on the documentation site is unnecessary. Until a compliant Direct Post integration is implemented and approved, use the inbox flow and complete posting in TikTok.

See TikTok's [Direct Post setup](https://developers.tiktok.com/docs/en/content-posting-api-get-started) and [Content Sharing Guidelines](https://developers.tiktok.com/docs/en/content-sharing-guidelines).
