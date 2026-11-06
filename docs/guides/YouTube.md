---
layout: guide
order: 10
section: Services
title: YouTube
permalink: /guides/youtube/
---
# Connect a YouTube channel

Publisher uploads an approved video privately to the channel you select. Review it in YouTube Studio before making it public.

## Set up Google
1. In [Google Cloud](https://console.cloud.google.com/), select the project for Publisher and enable **YouTube Data API v3**.
2. Configure the OAuth consent screen. In testing, add the Google accounts that manage your Spanish and English channels as test users.
3. Create one OAuth client of type **Desktop app**. In **Settings → Manage credentials → YouTube**, import the downloaded JSON or enter the client ID and secret. Both channels share the client.

## Connect ES and EN
In **Accounts → ES → YouTube**, select **Connect → Sign in with Google**. Select the account and channel for Spanish and verify the returned channel name. Repeat under **EN** for the English channel. If the channels have similar names, open **Account details** to compare their IDs.

## Send an approved video
In **Publication**, choose the language, select YouTube and review the destination. **Upload privately to YouTube** returns a video ID. Check that video in Studio before retrying an interrupted upload.

Google returns the authorization to an available local port on your computer. [Desktop OAuth reference](https://developers.google.com/identity/protocols/oauth2/native-app) · [Video uploads](https://developers.google.com/youtube/data/guides/uploading_a_video)
