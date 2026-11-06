---
layout: guide
order: 1
title: Getting started
permalink: /guides/setup/
---
# Set up CATsoup Publisher

## 1. Install and launch on Windows

Install Git, Python 3.10 or newer, and the Microsoft WebView2 runtime on Windows x64. In PowerShell, open the folder where you keep your repositories and run:

```powershell
git clone https://github.com/meruw/catsoup-publisher.git
cd catsoup-publisher
python -m venv .venv
.venv/Scripts/python.exe -m pip install -r requirements.txt
powershell -ExecutionPolicy Bypass -File apps/publisher/setup.ps1
powershell -ExecutionPolicy Bypass -File apps/publisher/build.ps1
apps/publisher/bin/CATsoupPublisher.exe
```

An empty episode library is normal. To try playback, generate a synthetic ES/EN delivery from the repository root:

```powershell
.venv/Scripts/python.exe apps/publisher/support/create_demo.py
apps/publisher/bin/CATsoupPublisher.exe --check
```

Launch Publisher again to preview the demo. It is for playback only. For your own videos, place a [delivery package](https://github.com/meruw/catsoup-publisher/blob/main/contracts/delivery/CONTRACT.md) in `deliveries/<episode-id>/`. To open another compatible workspace, use `--root C:\path\to\workspace`. That workspace must include the validator and Python environment.

## 2. Configure the platforms you want

Open **Settings → Manage credentials** and choose a provider. One set of developer credentials serves both languages.

| Service guide | What you need |
| --- | --- |
| [YouTube](https://publisher.meruw.dev/guides/youtube/) | A Google Desktop OAuth client and access to your channels. |
| [TikTok](https://publisher.meruw.dev/guides/tiktok/) | Desktop Login Kit credentials and the target accounts. |
| [Meta](https://publisher.meruw.dev/guides/meta/) | A Meta app and professional Instagram accounts linked to Facebook Pages. |
| [X](https://publisher.meruw.dev/guides/x/) | The ES and EN handles and a browser session on X. |

**Return addresses on desktop.** Publisher opens the system browser for all three providers. Google returns to a free port on this computer. TikTok also returns locally and accepts a registered wildcard port. Only Instagram requires a choice of HTTPS return host in Publisher. [Google desktop OAuth](https://developers.google.com/identity/protocols/oauth2/native-app) · [TikTok Desktop Login Kit](https://developers.tiktok.com/docs/en/login-kit-desktop).

## 3. Choose the Instagram return page

Under **Settings → Manage credentials → Instagram**, expand **Callback hosting**:

- **Off** is the starting state. Google, TikTok and manual Page-token connections do not require this setting.
- **My HTTPS domain**: [host the return page](https://publisher.meruw.dev/guides/meta-callback/), enter the exact HTTPS directory URL with its trailing slash, then select **Save callback**.
- **CATsoup-hosted callback**: read the notice, check the agreement box and select **Save callback**. The host receives a temporary authorization code and state. Publisher exchanges the code on your PC and protects the saved user tokens there.

For browser sign-in, register the exact saved URL in Meta under **Redirect URI**. Registering it in the portal does not enable the setting on your computer. To turn it off later, select **Off** and save. This stops new browser sign-ins; existing connections can be removed separately in Accounts.

## 4. Connect and check each language

Open **Accounts → ES → Connect** for the provider. For Instagram, enter the Spanish Instagram username and click **Sign in with Facebook**. Check the account shown by the provider, finish authorization in the same browser tab on the same computer, and verify the destination back in Accounts. Repeat in **EN** with the English account. Connecting accounts does not publish a video.

If Accounts shows **Set up callback**, complete step 3. If Meta rejects the redirect URI, compare the scheme, host, path, port and final slash with Publisher. Restart sign-in from Accounts if the attempt expired. **Saved · unverified** means the credential fields exist. **Confirmed at sign-in** means the provider accepted them during a sign-in. See the [upload guide](https://publisher.meruw.dev/guides/uploads/) before sending an approved episode.
