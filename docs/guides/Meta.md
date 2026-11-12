---
layout: guide
order: 12
section: Services
title: Meta
permalink: /guides/meta/
---
# Connect Instagram and Facebook through Meta

Publisher publishes Reels to a professional Instagram account using Facebook Login. The connected Facebook Page identifies the Instagram destination. Facebook Page Reels use a separate Facebook destination, with their own draft, approval and receipt. Both destinations share the same Meta app credentials and callback.

## Prepare the accounts
For each language, link the professional Instagram account (Creator or Business) to the corresponding Facebook Page you administer. In the Page's linked accounts, verify the Instagram username before continuing.

## Configure your Meta app
1. In [Meta for Developers](https://developers.facebook.com/apps/), select the Instagram content-management and Page-management use cases for your app.
2. Under the use cases, enable `pages_show_list`, `pages_read_engagement`, `instagram_basic` and `instagram_content_publish`. For Pages managed through a business portfolio, also include `business_management`. Create a Facebook Login for Business configuration with a **user access token**, select these permissions in the configuration and copy its **configuration ID**.
3. In **Settings → Manage credentials → Meta**, save your **App ID**, **App secret** and **Configuration ID**. Keep the secret on your PC.
4. Select [callback hosting](https://publisher.meruw.dev/guides/meta-callback/) and register the exact saved HTTPS URL in Meta, including the final slash. Browser sign-in starts **Off**. You can host your own page or explicitly opt in to CATsoup hosting after reading the notice.

## Connect and publish
Select **Accounts → ES → Instagram**, enter the Instagram username and sign in with Facebook. Use the same browser tab on the same computer. Verify the Instagram account that Publisher assigns, then repeat under **EN**.

In Facebook's authorization screen, grant access to the business portfolios, Pages and Instagram accounts needed for your destinations. Access to every portfolio or future account is not required. Both CATsoup language accounts have been connected with access restricted to CATsoup and `business_management` included in the configuration.

## If sign-in finishes but the account is not connected

Open **Console** at the foot of Publisher's sidebar and inspect the connection attempt. `no_pages` means the Page lookup returned no accessible Pages; a successful code exchange alone does not establish the destination.

Check that the Facebook profile has access to the correct Page, that the Page links to the requested Instagram account, and that the selected business portfolio, Page and Instagram account are included in the grant. For portfolio-managed Pages, check `business_management` in both the app's available permissions and the login configuration. Save the configuration and sign in again to grant the updated permissions. Keep access limited to the assets you intend to use.

## Publish a Reel

Review the approved video and edit its **Instagram** post. Set the Instagram row to **Publish now**, choose **Approve post** and the send action above the editor, confirm the connected account and select **Publish Reel on Instagram**. **Local draft** retains a local draft on this PC; it does not create a draft in Instagram. Check the receipt and the Instagram profile before retrying an uncertain transfer. For manual sign-in, **Other sign-in options** accepts an existing Facebook Page access token.

[Meta's Instagram API guide](https://www.postman.com/meta/instagram/documentation/6yqw8pt/instagram-api)

## Connect a Facebook Page

1. In the Page-management use case, add `pages_manage_posts`. Include it in the same Facebook Login for Business configuration, together with `pages_show_list`, `pages_read_engagement` and the existing Instagram permissions. Keep `business_management` for portfolio-managed assets.
2. In Publisher, open **Accounts → ES → Facebook → Connect**. Enter the Page's numeric ID or the username of its linked Instagram account. Publisher pre-fills the connected Instagram username when available.
3. Choose **Sign in with Facebook** and authorize the intended Page. Publisher checks the publishing grant and content-creation task before saving a separate encrypted Page connection. Check the resulting Page name and ID, then repeat for EN.
4. In the episode, select Facebook, prepare its post, choose **Publish now**, save and approve this account's post. Confirm **Publish Reel on Facebook**. It can also participate in the selected language's batch.

**Local draft** stays on this PC. A Facebook send streams the approved local file, requests public publication and waits for Meta's processing and publication confirmation. An uncertain result retains the Reel ID; check the Page before enabling another attempt. Publishing an Instagram Reel does not automatically create the Facebook Reel.

A browser session is needed for authorization, not for each upload. App roles and Meta's access-level/review requirements still apply. The local connector does not bypass Meta approval for third-party users.

Reference: [Meta's official Reels publishing collection](https://www.postman.com/meta/facebook/request/juhnm3q/4-publish-reel).

## Reel cover

In the Instagram account editor, open **Thumbnail → Choose frame**, pause the preview and choose **Use this frame** before approving and uploading. Publisher sends that frame's timestamp with the Reel container. The cover cannot be changed here once its upload exists. This connector does not host local PNG/JPG images for custom covers. Facebook cover selection is not implemented.
