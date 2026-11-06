---
layout: guide
order: 13
section: Services
title: X
permalink: /guides/x/
---
# Set up X for manual posting

Publisher opens the X composer with the text from your draft. You choose the account, attach the video and publish in the browser.

## Assign your handles
In **Accounts → ES → X**, choose **Set handle** and enter your Spanish @username. Repeat under **EN** for English. The saved handle identifies your intended destination. X uses the session in your browser and Publisher does not authenticate or verify the handle.

## Open the composer
1. In **Publication**, select an approved episode, its language and the X draft. Review the title, description and hashtags.
2. Choose **Open X composer**. Compare the X account shown in the browser with the handle saved in Publisher. Switch accounts if they differ.
3. Use **Copy upload file path** in Publisher and paste it into the browser's file picker. Review the video and text on X, then submit the post.

This workflow requires no X API credentials, callback or paid API plan.

## Direct API integration and cost

Publisher's X workflow uses the browser composer. OAuth sign-in and automatic video uploads are not implemented.

X supports native-app authorization using [OAuth 2.0 with PKCE](https://docs.x.com/fundamentals/authentication/oauth-2-0/authorization-code). A future integration would need a developer app, a registered callback, separate authorization and local token storage for each language, and the `tweet.read`, `users.read`, `tweet.write`, `media.write` and `offline.access` scopes. Video transfer uses X's [chunked upload flow](https://docs.x.com/x-api/media/quickstart/media-upload-chunked), followed by post creation and a saved receipt.

As checked on November 5, 2026, X uses prepaid, pay-per-use API credits. Its [pricing page](https://docs.x.com/x-api/getting-started/pricing) lists US$0.015 for a post-creation request and US$0.200 for one containing a URL. These are post-creation charges, not an end-to-end video-upload estimate; other API calls may add costs. [Promotional credits](https://docs.x.com/x-api/getting-started/free-credits) require an eligible purchase and are not a permanent free publishing allowance.

For a zero-cost setup, assign both handles in Publisher and finish posting in the browser. API integration remains an optional future feature requiring a separate implementation and budget decision.
