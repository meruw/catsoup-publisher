---
layout: guide
order: 12
section: Services
title: Meta
permalink: /guides/meta/
---
# Connect Instagram through Meta

Publisher publishes Reels to a professional Instagram account using Facebook Login. The connected Facebook Page identifies the Instagram destination. Facebook Page Reels are a separate publishing destination outside this workflow.

## Prepare the accounts
For each language, link the professional Instagram account (Creator or Business) to the corresponding Facebook Page you administer. In the Page's linked accounts, verify the Instagram username before continuing.

## Configure your Meta app
1. In [Meta for Developers](https://developers.facebook.com/apps/), select the Instagram content-management and Page-management use cases for your app.
2. Under the use cases, enable `pages_show_list`, `pages_read_engagement`, `instagram_basic` and `instagram_content_publish`. Create a Facebook Login for Business configuration with a **user access token** and those four permissions. Copy the **configuration ID**.
3. In **Settings → Manage credentials → Instagram**, save your **App ID**, **App secret** and **Configuration ID**. Keep the secret on your PC.
4. Select [callback hosting](https://publisher.meruw.dev/guides/meta-callback/) and register the exact saved HTTPS URL in Meta, including the final slash. Browser sign-in starts **Off**. You can host your own page or explicitly opt in to CATsoup hosting after reading the notice.

## Connect and publish
Select **Accounts → ES → Instagram**, enter the Instagram username and sign in with Facebook. Use the same browser tab on the same computer. Verify the Instagram account that Publisher assigns, then repeat under **EN**.

For an approved episode choose **Publish Reel on Instagram** and confirm the destination. Check the receipt and the Instagram profile before retrying an uncertain transfer. For manual sign-in, **Other sign-in options** accepts an existing Facebook Page access token.

[Meta's Instagram API guide](https://www.postman.com/meta/instagram/documentation/6yqw8pt/instagram-api)
