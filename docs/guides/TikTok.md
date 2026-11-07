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
Review the video and edit its **TikTok** post, then choose **Approve post** and the send action above the editor. Confirm the destination and select **Send to TikTok inbox**. Open the inbox notification in TikTok, paste the prepared caption if desired and finish posting there. An inbox receipt does not mean the video is public. Finish in the TikTok mobile app using the destination account. Inbox transfer sends the video without pre-filling its caption; copy the prepared text to your phone and paste it in the editor. The browser session does not select the upload destination; the saved account authorization does.

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

Local videos use `FILE_UPLOAD`; the documentation site does not host the MP4. Direct Post has protocol and browser-fixture coverage in Publisher. Authorization and an inbox transfer with approved content have been validated for the operator's client. Public Direct Post still requires TikTok's audit.

See TikTok's [Direct Post setup](https://developers.tiktok.com/docs/en/content-posting-api-get-started), [creator-info API](https://developers.tiktok.com/docs/en/content-posting-api-reference-query-creator-info) and [Content Sharing Guidelines](https://developers.tiktok.com/docs/en/content-sharing-guidelines).

## Prepare the public Direct Post review

Complete the product and its end-to-end posting demonstration before submission. TikTok's review guidelines exclude apps still in development/testing and apps intended only for private use. A public repository by itself is insufficient.

1. Open the app in the [Developer Portal](https://developers.tiktok.com/) and review its Production configuration. Use Desktop Login Kit and Content Posting API; enable Direct Post and request only the scopes actually used: `user.info.basic`, `video.upload` and `video.publish`.
2. Provide the complete product website, visible working privacy/terms links, app identity and instructions for reviewers. The site's content must describe the functioning product; do not present the unfinished PoC as a released application.
3. In App review, explain each scope and attach a demonstration of login, account identity, the video preview, editable caption, explicit privacy, interactions, disclosures, consent and the final outcome. Demonstrate inbox transfer too if requesting `video.upload`. TikTok requests sandbox demonstrations for a first review.
4. Follow the [Content Posting API audit request](https://developers.tiktok.com/application/content-posting-api) linked by TikTok's sharing guidelines. General app review, the scope grant and the audit for public posting are distinct requirements; an enabled toggle does not prove audit approval.
5. When TikTok confirms the production client's eligibility, configure that client's credentials in Publisher and reconnect ES and EN if switching from sandbox. Approval belongs to the API client; it does not extend to other users' separately registered clients.

Suggested scope explanation, to adapt to the actual submitted product:

> CATsoup Publisher is a local desktop publishing application for creators reviewing and publishing their own original videos. Login Kit identifies the account receiving each post. user.info.basic displays its identity; video.upload sends a video for completion in TikTok; video.publish supports a user-confirmed Direct Post with editable text, privacy selection and content disclosures. Publisher stores authorizations protected on the user's Windows PC and shows the provider's processing result.

Do not claim public availability, user numbers or successful public Direct Post unless demonstrated. Review and audit acceptance are decided by TikTok.

### Developer clients and distribution

Publisher currently uses the developer credentials configured by each operator. An audit for one client does not approve other independently registered clients. Do not put the maintainer's client secret in a downloadable executable, repository or setup guide. A shared-client service would require separate confidential secret handling and an appropriate service design; the optional static Meta callback does not provide it.

For a review demonstration, use the exact submitted build and a separately approved original test clip. Show account sign-in, editing and approval, the Direct Post confirmation controls, the provider-confirmed outcome and inbox completion if requesting `video.upload`. Do not resend an existing public episode to create a demonstration. Keep the recording consistent with the product's actual state; fixture tests are not evidence of a successful live post.

See the [App Review Guidelines](https://developers.tiktok.com/docs/en/app-review-guidelines) and [Content Sharing Guidelines](https://developers.tiktok.com/docs/en/content-sharing-guidelines).
