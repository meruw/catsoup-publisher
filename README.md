# CATsoup Publisher

CATsoup Publisher is a local Windows desktop app for reviewing Spanish and English videos and sending each approved edit to the right account. Its episode workspace puts the video beside its post editor, with account-specific drafts and approvals, saved delivery modes and transfer receipts. Send to one platform or confirm a batch for the selected language. Upload privately or publicly to YouTube, send a video to the TikTok inbox or use optional Direct Post, publish an Instagram or Facebook Page Reel or post to X through the browser or an optional paid API connection.

![Publisher previewing a synthetic bilingual delivery](docs/images/publisher.png)

## Build and run (Windows x64)

Publisher is a proof of concept available to build from source. Follow the [Windows setup guide](https://publisher.meruw.dev/guides/setup/) to build the app and connect your accounts. Setup installs the project's pinned Go and Node tools locally. The desktop app uses WebView2 for its interface and Python for delivery validation.

Start with the synthetic demo or import your own [delivery package](contracts/delivery/CONTRACT.md). Publisher checks the files, previews each language and keeps publication drafts in your workspace. Uploads require approval tied to the exact video.

## Connect a service

| Service | Where the video goes |
| --- | --- |
| [YouTube](https://publisher.meruw.dev/guides/youtube/) | Private or public upload; publish a recorded private video using the same ID |
| [TikTok](https://publisher.meruw.dev/guides/tiktok/) | Inbox uploads or optional Direct Post with explicit privacy |
| [Meta](https://publisher.meruw.dev/guides/meta/) | A Reel on the connected Instagram profile or Facebook Page |
| [X](https://publisher.meruw.dev/guides/x/) | Free browser handoff or direct posting using your paid X API access |

Read the [upload guide](https://publisher.meruw.dev/guides/uploads/) for account selection, receipts and retries. All guides are also available in the [GitHub wiki](https://github.com/meruw/catsoup-publisher/wiki).

## Development

Start with [Contributing](CONTRIBUTING.md) for build checks and documentation maintenance. See the [architecture](docs/publisher/ARCHITECTURE.md), [design system](docs/publisher/DESIGN_SYSTEM.md), [draft storage](docs/publisher/DRAFTS.md) and [project status](docs/publisher/STATUS.md) for implementation details. Report security issues through the [security policy](SECURITY.md).

## License and credits

Original project code is licensed under [GPL-3.0-only](LICENSE). See [third-party notices and artwork attribution](NOTICES.md).
