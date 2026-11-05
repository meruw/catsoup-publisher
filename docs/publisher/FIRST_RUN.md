---
layout: guide
order: 1
title: Getting started
permalink: /guides/setup/
---
# Set up CATsoup Publisher

## 1. Install and launch on Windows

Install Git, Python 3.10 or newer, Node.js 24 LTS with npm, and the Microsoft WebView2 runtime on Windows x64. In PowerShell, open the folder where you keep your repositories and run:

```powershell
git clone https://github.com/meruw/catsoup-publisher.git
cd catsoup-publisher
python -m venv .venv
.venv/Scripts/python.exe -m pip install -r requirements.txt
powershell -ExecutionPolicy Bypass -File apps/publisher/setup.ps1
powershell -ExecutionPolicy Bypass -File apps/publisher/build.ps1
apps/publisher/bin/CATsoupPublisher.exe
```

An empty episode library is normal. The [optional synthetic demo](https://github.com/meruw/catsoup-publisher#build-and-run-windows-x64) lets you test playback without publishing.

## 2. Configure the platforms you want

Open **Settings → Manage credentials** and choose a provider. One set of developer credentials serves both languages.

| Platform | Create or select in the provider portal | Save in Publisher |
|---|---|---|
| YouTube | Enable YouTube Data API v3, configure the consent screen and test users, and create a Desktop OAuth client. | Import the Google client JSON or enter its ID and secret. |
| TikTok | Configure a desktop app with Login Kit and Content Posting API. For sandbox testing, add both target users. Register `http://127.0.0.1:*/callback/` and request `user.info.basic` and `video.upload`. | Paste the client key and secret from the same environment. |
| Instagram | Link each professional Instagram to its Facebook Page. In the Meta app, configure Facebook Login for Business with a user access token. Enable `pages_show_list`, `pages_read_engagement`, `instagram_basic` and `instagram_content_publish` on the use cases and in the login configuration. | Enter the Meta App ID, app secret and configuration ID. Continue to step 3. |
| X | Publisher uses the browser composer for manual posting. | In Accounts, save the handle for each language. |

## 3. Choose the Instagram return page

Under **Settings → Manage credentials → Instagram**, expand **Callback hosting**:

- **Off** is the starting state. Google, TikTok and manual Page-token connections do not require this setting.
- **My HTTPS domain**: [host the return page](https://publisher.meruw.dev/guides/meta-callback/), enter the exact HTTPS directory URL with its trailing slash, then select **Save callback**.
- **CATsoup-hosted callback**: read the notice, check the agreement box and select **Save callback**. The host receives a temporary authorization code and state. Publisher exchanges the code on your PC and protects the saved user tokens there.

For browser sign-in, register the exact saved URL in Meta under **Redirect URI**. Registering it in the portal does not enable the setting on your computer. To turn it off later, select **Off** and save. This stops new browser sign-ins; existing connections can be removed separately in Accounts.

## 4. Connect and check each language

Open **Accounts → ES → Connect** for the provider. For Instagram, enter the Spanish Instagram username and click **Sign in with Facebook**. Check the account shown by the provider, finish authorization in the same browser tab on the same computer, and verify the destination back in Accounts. Repeat in **EN** with the English account. Connecting accounts does not publish a video.

If Accounts shows **Set up callback**, complete step 3. If Meta rejects the redirect URI, compare the scheme, host, path, port and final slash with Publisher. Restart sign-in from Accounts if the attempt expired. **Saved · unverified** means the credential fields exist. **Confirmed at sign-in** means the provider accepted them during a sign-in. See the [upload guide](https://publisher.meruw.dev/guides/uploads/) before sending an approved episode.
