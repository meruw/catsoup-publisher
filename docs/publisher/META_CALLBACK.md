# Choose a Meta callback

Instagram browser sign-in starts **off**. Saving developer credentials does not enable it. Upgrades from the old CATsoup default require a new choice; existing independent callback URLs keep working.

To use the optional shared page, open **Settings → Developer credentials → Instagram → Callback hosting**, select **CATsoup-hosted callback**, read the notice, check the agreement box and click **Save callback**. Register `https://publisher.meruw.dev/auth/meta/` in your own Meta app. GitHub Pages receives the temporary authorization code and state in the return URL and serves the script. The desktop app exchanges that code for tokens on the user's PC. New sign-ins depend on the shared host.

For independent hosting:

1. Copy [index.html](../auth/meta/index.html) and [return.js](../auth/meta/return.js) into one directory on a static HTTPS site you control. The stylesheet is optional: copy [site.css](../site.css) and adjust its relative link if needed. Serve the directory itself as the index at its final public URL, without redirects. For example, `https://studio.example.org/tools/meta/`. No build, server API or edits to the script are required.
2. In Publisher, open **Settings → Developer credentials → Instagram → Callback hosting**. Select **My HTTPS domain**, enter the full HTTPS directory URL including its trailing slash and click **Save callback**. Use the exact URL displayed after saving in the Meta portal. Do not enter secrets, query parameters or fragments in it.
3. Register that URL as a redirect in **your** Meta app, then sign in from **Accounts** in the same browser tab on the same computer. A change applies to the next attempt. Select **Off → Save callback** to disable future browser sign-ins. To revoke an existing grant, remove the local connection in Accounts and revoke the app in Meta.

The hosting operator serves JavaScript during authorization and receives the short-lived authorization code and state in the HTTPS request. Choose a host you trust, keep the return page free of analytics and other scripts, and avoid recording callback query strings in hosting logs where you control them. Tokens and the app secret are handled by Publisher locally. A hosting outage or expired certificate prevents new sign-ins using that URL. Self-hosting shifts availability, updates and trust to your host. Google and TikTok desktop sign-in continue using local callbacks.
