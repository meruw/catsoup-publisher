---
layout: guide
order: 13
section: Services
title: X
permalink: /guides/x/
---
# Connect X

Choose a publishing mode for each language in **Accounts → X**. Browser mode is the default. Saving developer credentials does not change an account's mode.

| Mode | Setup | Posting |
| --- | --- | --- |
| **Browser · Free** | Save your @username. | Publisher prepares the text and file path; attach the video and post in X. |
| **API · Paid** | Configure your own X Native App and authorize the account. | Publisher uploads the video, creates the post and saves a receipt. X bills your developer project. |

## Browser posting

In **Accounts → ES → X**, choose **Set handle**, keep **Browser · Free** and enter the Spanish @username. Repeat under **EN** for English. A saved handle is a destination reminder, not an authenticated connection.

You can keep both X accounts signed in within one browser and use [X's account switcher](https://help.x.com/en/managing-your-account/managing-multiple-x-accounts). Separate browser profiles are optional. Publisher opens the system's default browser and cannot inspect or select its active X account.

1. Switch to the intended account in X. In Publisher, complete **Review**, prepare the X draft, then select **Publish → Publish now**. Check its language and text.
2. Confirm the browser account and choose **Copy video path and open X**. This copies the approved file's path and opens the composer with the title, description and hashtags.
3. In X, attach a video, paste the copied path into the file picker and confirm the selection. Review the account, text and video, then post.

This mode needs no API credentials or paid API access. The composer cannot attach a local file through its URL. Browser posts do not create API receipts in Publisher.

## Optional API setup

1. Open the [X Developer Console](https://console.x.com/) and create or select an app. Enable **OAuth 2.0** and select **Native App**, a public client. Publisher uses the **Client ID**, not an app-only bearer token or a client secret.
2. Register `http://127.0.0.1:48173/callback/` as the callback. Use the same exact value, including the port and trailing slash, in Publisher.
3. In **Settings → Developer credentials → X**, enter the Client ID and callback, then save. These settings are shared by ES and EN.
4. Review [X API pricing](https://docs.x.com/x-api/getting-started/pricing), access eligibility and billing controls in your developer project. X charges API usage; Publisher adds no posting fee. Do not treat the price of one post-creation request as the total cost of uploading a video. Identity checks, media requests and other calls can affect usage.
5. In **Accounts → ES → X → API · Paid**, enter the expected Spanish @username, acknowledge API charges and choose **Sign in with X**. Check the account in the browser before authorizing. Publisher rejects a different username. Repeat in **EN** with the English account.

Publisher requests `tweet.read`, `users.read`, `tweet.write`, `media.write` and `offline.access`. Sign-in uses [OAuth 2.0 with PKCE](https://docs.x.com/fundamentals/authentication/oauth-2-0/authorization-code). Access and refresh tokens stay encrypted on this Windows PC; no CATsoup-hosted callback is involved. Settings displays a confirmation after a successful authorization-code exchange; it does not certify available billing credit or video eligibility.

If the callback port is occupied, Publisher stops before opening the browser. Close the other sign-in or choose another port between 1024 and 65535 in **both** X and Publisher. The host must remain `127.0.0.1` and the path `/callback/`. Changing the Client ID requires reconnecting existing API accounts so tokens are not silently associated with another developer project.

## Publish through the API

Complete **Review** and **Prepare** for the approved episode, then select the X destination in **Publish** and choose **Publish now**. The destination comes from its language's connected account, not the editable draft label. Confirm the destination and accept X API usage charges for this attempt, then choose **Publish on X · Paid API**. This creates a post visible according to the account's settings; it is not a private staging upload. **Keep as draft** keeps the metadata on this PC without an X request.

Publisher validates the approved file, uploads it in bounded chunks, waits for processing and creates the post. The receipt contains the post URL. Keep Publisher open until the result appears. The implementation follows X's [chunked media workflow](https://docs.x.com/x-api/media/quickstart/media-upload-chunked).

Publisher currently accepts MP4 files up to 512 MiB and text up to a conservative 280-character weighted bound. Wide characters count twice; long URLs and combined emoji may be overcounted. X applies its own format, duration and account limits and can still reject an upload.

## Recover or change mode

An interrupted request can have an uncertain outcome. Check the X account before allowing another attempt; retries may cost more. Publisher does not automatically retry post creation. A successful receipt blocks another upload of the same video to the same account through Publisher. A `media:` reference identifies an upload session, not a published post.

To return to free posting, choose **Accounts → X → Change → Browser · Free** and save the handle. This removes the API token from that local account slot. To revoke the authorization itself, remove Publisher in X's connected-app settings. Posting manually is not tracked by the API receipt system.
