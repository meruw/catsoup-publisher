# Upload an approved episode

Open **Accounts** in the sidebar. Select **ES** or **EN**, choose the platform and connect the intended account. In **Publication**, the draft's platform and language select that verified destination. Check its name and ID before confirming the upload. A draft label is for organizing text; editing it never switches accounts.

| Platform | Action | Result |
| --- | --- | --- |
| YouTube | Upload privately to YouTube | Sends the approved MP4 with its title, description, hashtags and language. YouTube returns a video ID. Open Studio to review, set audience and choose when to make it public. |
| TikTok | Send to TikTok inbox | Sends the MP4 as a draft. Open the TikTok inbox notification, paste the caption if desired, edit and publish in TikTok. |
| Instagram | Publish Reel on Instagram | Creates a Reel container, uploads the local file, checks processing and publishes to the connected professional account. The title, description and hashtags form the caption. |
| X | Open X composer | Opens the prepared text in the browser. Use **Copy upload file path** for the attachment, check the account and submit there. |

## Connect your accounts

Each workspace has one destination per platform and language. Connect both languages before creating an episode. **Change** replaces only the selected slot after successful authorization. **Remove** forgets its local credentials. Publisher prevents assigning one provider ID to both languages on the same platform. Old drafts retain their text and are routed using their language. The send screen shows the current assignment and requires confirmation.

Under **Developer setup** you can import the provider's client JSON once for both languages. The browser handles the provider's sign-in and returns to Publisher. Windows DPAPI protects saved user credentials for this Windows user and slot. Google and TikTok renew access using provider-issued refresh tokens when available. Reconnect from **Accounts** if a provider expires or revokes access. The synthetic demo and pilot are always preview-only.

**YouTube.** Enable YouTube Data API v3 in your Google Cloud project, configure the OAuth consent screen and add your Google account as a test user if your project is in testing. Create a **Desktop app** OAuth client. In Accounts → YouTube → Connect, use the Google Cloud setup buttons as needed. Import the downloaded JSON or enter the Desktop client ID and secret. Publisher saves the client for both languages. Click **Sign in with Google**. Choose the channel and grant upload and read-only channel access. Publisher uses the system browser and a temporary loopback listener. Uploads explicitly request private visibility. [Google desktop authorization](https://developers.google.com/youtube/data/guides/auth/installed-apps) · [Video uploads](https://developers.google.com/youtube/data/reference/rest/v3/videos/insert)

**TikTok.** Register a desktop app with Login Kit and the Content Posting API. Register `http://127.0.0.1:*/callback/` and obtain approval for `user.info.basic` and `video.upload`. Save your own app credentials in `%APPDATA%\CATsoup Publisher\tiktok-desktop.json` as a JSON object with `client_key` and `client_secret`. Click **Sign in with TikTok** and grant access. The app implements the inbox workflow, which requires you to finish posting in TikTok. [TikTok desktop setup](https://developers.tiktok.com/doc/login-kit-desktop/) · [Inbox upload requirements](https://developers.tiktok.com/docs/en/content-posting-api-get-started-upload-content)

**Instagram.** Configure Facebook Login for Business for your Meta app, using a user token configuration with `pages_show_list`, `pages_read_engagement`, `instagram_basic` and `instagram_content_publish`. For local development register the exact callback `http://localhost:48172/callback/` if your Meta app permits a localhost redirect. Under **Developer setup**, import a JSON file with `app_id`, `app_secret` and `config_id` from your app. Enter the Instagram username for the selected language, click **Sign in with Facebook** and grant access to its linked Page. Publisher selects the matching username from the authorized Pages. If your Meta configuration requires HTTPS or disallows loopback, use **Developer setup → Use an existing Facebook Page token** instead. Live permissions, app review and Page access depend on the Meta app. [Meta's Instagram collection](https://www.postman.com/meta/instagram/documentation/6yqw8pt/instagram-api)

**X.** Use **Set handle** for each language. This is a reminder for the manual web composer; check the active X handle in your browser. X charges for API usage, so this integration uses the web composer and file picker. [X API pricing](https://docs.x.com/x-api/getting-started/pricing)

## Check the result

Publisher revalidates the delivery and stages a temporary copy matching the approved hash before sending. SQLite records one attempt per platform, verified account ID and video hash. The same file can go to another account. A confirmed transfer cannot be repeated to the same destination through Publisher. Keep the app open while uploading. A receipt displays the returned reference.

If a response is lost, check the destination before doing anything else. Look in YouTube Studio (including private videos), the TikTok inbox or the professional Instagram profile. Only choose **Allow another attempt** after confirming the upload is absent. If Publisher exited during a transfer, an in-flight receipt becomes eligible for manual recovery after 30 minutes. **Refresh receipt** reads the local receipt. Publisher does not automatically retry, schedule posts or resume partial transfers.

Protocol and interface tests use local fixtures. A live transfer with an operator-owned account and an approved episode is still required to verify each provider end to end.
