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
Complete **Review** and **Prepare**, then choose TikTok in the episode's **Publish** step. Keep **Keep as draft** selected, confirm the destination and select **Send to TikTok inbox**. Open the inbox notification in TikTok, paste the prepared caption if desired and finish posting there. **Publish now** is unavailable: Direct Post requires its own connector flow and TikTok approval. An inbox receipt does not mean the video is public.

[Desktop Login Kit](https://developers.tiktok.com/docs/en/login-kit-desktop) · [Inbox upload requirements](https://developers.tiktok.com/docs/en/content-posting-api-get-started-upload-content)
