---
layout: guide
order: 10
section: Services
title: YouTube
permalink: /guides/youtube/
---
# Connect a YouTube channel

Publisher can upload an approved video privately or publicly to the selected channel. A recorded private upload can be made public without uploading its video file again.

## Set up Google
1. In [Google Cloud](https://console.cloud.google.com/), select the project for Publisher and enable **YouTube Data API v3**.
2. Configure the OAuth consent screen. In testing, add the Google accounts that manage your Spanish and English channels as test users.
3. Create one OAuth client of type **Desktop app**. In **Settings → Manage credentials → YouTube**, import the downloaded JSON or enter the client ID and secret. Both channels share the client.

## Connect ES and EN
In **Accounts → ES → YouTube**, select **Connect → Sign in with Google**. Select the account and channel for Spanish and verify the returned channel name. Repeat under **EN** for the English channel. If the channels have similar names, open **Account details** to compare their IDs.

## Send an approved video
Review the video, select **YouTube** above the post editor and write its text. Set the YouTube row to **Private video** or **Publish now**, then choose **Approve post** and the send action above the editor. Select the audience (made for kids or not), indicate realistic altered/synthetic content when applicable, and confirm the account and action. The delivery must already have production approval.

After a private upload, **Publish now → Publish existing YouTube video** applies the saved text and public visibility to that same ID. Publisher verifies its channel, preserves other editable metadata and never uploads a duplicate. If the visibility change is uncertain, check YouTube Studio before retrying it.

Google sign-in requests upload, channel reading and `youtube.force-ssl` permissions. Updating an existing video's visibility needs the management permission; reconnect in **Accounts** if the saved authorization does not include it. [Video update authorization](https://developers.google.com/youtube/v3/docs/videos/update)

Google returns the authorization to an available local port on your computer. [Desktop OAuth reference](https://developers.google.com/identity/protocols/oauth2/native-app) · [Video uploads](https://developers.google.com/youtube/data/guides/uploading_a_video)

## Expired authorization

If Google returns `invalid_grant`, reconnect the affected channel through **Accounts → ES or EN → YouTube → Change**. Verify the returned channel before retrying the intended post. A token-renewal failure happens before a new video transfer; it does not require deleting an earlier receipt or uploading to another account. Do not remove existing published receipts.

For external OAuth apps in **Testing**, Google normally expires refresh tokens after seven days when YouTube permissions are requested. Reconnecting restores access, but Testing can require another authorization later. Review the project's publishing status and Google's production requirements before moving out of Testing. `invalid_grant` can also mean revoked or otherwise invalid authorization; it does not establish the cause by itself. [Google refresh-token expiration](https://developers.google.com/identity/protocols/oauth2#expiration)

Errors such as `invalid_client` instead point to the Google Desktop app configuration in Settings. Publisher records only known OAuth error identifiers and its own guidance; raw provider descriptions and tokens are excluded from diagnostics.

## Thumbnails

Open **Thumbnail** in the account editor to choose a frame or import a PNG/JPG up to 2 MiB. Save and approve the post. Publisher uploads the video first and waits for YouTube processing before applying the image. If the image is not confirmed, use **Apply thumbnail** to retry only that image, without uploading another video. This also allows changing a confirmed video's thumbnail after approving the new draft revision. Custom-thumbnail eligibility and display behavior remain controlled by YouTube. [Thumbnail API](https://developers.google.com/youtube/v3/docs/thumbnails/set)
