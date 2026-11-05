# Upload an approved episode

In **Publication**, open the draft for the platform, language and account label. Under **Platform upload**, connect your account. Compare the name and ID returned by the platform with your intended destination. Confirm the language and account, then use the action for that platform. Account labels organize drafts. The connected account ID determines the actual destination.

| Platform | Action | Result |
| --- | --- | --- |
| YouTube | Upload privately to YouTube | Sends the approved MP4 with its title, description, hashtags and language. YouTube returns a video ID. Open Studio to review, set audience and choose when to make it public. |
| TikTok | Send to TikTok inbox | Sends the MP4 as a draft. Open the TikTok inbox notification, paste the caption if desired, edit and publish in TikTok. |
| Instagram | Publish Reel on Instagram | Creates a Reel container, uploads the local file, checks processing and publishes to the connected professional account. The title, description and hashtags form the caption. |
| X | Open X composer | Opens the prepared text in the browser. Use **Copy upload file path** for the attachment, check the account and submit there. |

## Connect your accounts

You can connect from a saved draft before producing the first episode. Direct uploads require an approved delivery. The synthetic demo and streetlight pilot are always preview-only. Disconnect a destination to switch accounts. Authorization lasts for the running Publisher session or until the token expires. Each platform, language and account label has its own connection, reusable across episodes.

**YouTube.** Enable YouTube Data API v3 in your Google Cloud project, configure the OAuth consent screen and add your Google account as a test user if your project is in testing. Create a **Desktop app** OAuth client. Download its JSON and save it as `%APPDATA%\CATsoup Publisher\google-desktop.json` (create the folder if necessary). Click **Sign in with Google**. Choose the channel and grant upload and read-only channel access. Publisher uses the system browser and a temporary loopback listener. Uploads explicitly request private visibility. [Google desktop authorization](https://developers.google.com/youtube/data/guides/auth/installed-apps) · [Video uploads](https://developers.google.com/youtube/data/reference/rest/v3/videos/insert)

**TikTok.** Register a desktop app with Login Kit and the Content Posting API. Register `http://127.0.0.1:*/callback/` and obtain approval for `user.info.basic` and `video.upload`. Save your own app credentials in `%APPDATA%\CATsoup Publisher\tiktok-desktop.json` as a JSON object with `client_key` and `client_secret`. Click **Sign in with TikTok** and grant access. The app implements the inbox workflow, which requires you to finish posting in TikTok. [TikTok desktop setup](https://developers.tiktok.com/doc/login-kit-desktop/) · [Inbox upload requirements](https://developers.tiktok.com/docs/en/content-posting-api-get-started-upload-content)

**Instagram.** Use a Business or Creator account linked to a Facebook Page. Configure your Meta app for Instagram API with Facebook Login for Business, with access to that Page and Instagram account. Obtain an authorized **Page access token** for the linked Page with the permissions required by your app, including `instagram_basic`, `instagram_content_publish` and `pages_read_engagement`. Page discovery can also require `pages_show_list`. During development use an account with an app role and Page access. For other users, check Meta's access and review requirements. Paste the token into the password field in Publisher and click **Connect account**. The app reads the linked Instagram ID and username before enabling **Publish Reel on Instagram**. The token stays in memory. Publisher uses Graph v25.0 and Meta's local resumable upload flow. [Meta publishing guide](https://developers.facebook.com/docs/instagram-platform/instagram-api-with-facebook-login/content-publishing/) · [Meta's Instagram collection](https://www.postman.com/meta/instagram/)

**X.** The compositor runs in your normal browser. X charges for API usage, so this integration uses the web composer and file picker. [X API pricing](https://docs.x.com/x-api/getting-started/pricing)

## Check the result

Publisher revalidates the delivery and stages a temporary copy matching the approved hash before sending. SQLite records one attempt per platform, verified account ID and video hash. The same file can go to another account. A confirmed transfer cannot be repeated to the same destination through Publisher. Keep the app open while uploading. A receipt displays the returned reference.

If a response is lost, check the destination before doing anything else. Look in YouTube Studio (including private videos), the TikTok inbox or the professional Instagram profile. Only choose **Allow another attempt** after confirming the upload is absent. If Publisher exited during a transfer, an in-flight receipt becomes eligible for manual recovery after 30 minutes. **Refresh receipt** reads the local receipt. Publisher does not automatically retry, schedule posts or resume partial transfers.

Protocol and interface tests use local fixtures. A live transfer with an operator-owned account and an approved episode is still required to verify each provider end to end.
