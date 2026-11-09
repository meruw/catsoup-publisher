# Licenses and attribution

- Original CATsoup Publisher source and included project artwork: GPL-3.0-only,
  copyright (C) 2026 CATsoup Publisher contributors. See LICENSE. No trademark
  rights or endorsement are granted by this license.
- Instrument Sans: SIL Open Font License 1.1. The font and original license are in
  `apps/publisher/frontend/public/fonts/`. Source: Google Fonts' Instrument Sans.
- Design guidance was informed by [Unslop](https://github.com/heydamir/unslop),
  by Damir Sokolovsky, licensed CC BY 4.0. Its skill is not bundled with the app.
  The design-system documentation retains attribution.
- UI icons: selected [Lucide SVGs](https://github.com/lucide-icons/lucide), ISC with MIT notices for Feather-derived icons. The original files and pinned upstream commit are in `apps/publisher/frontend/src/assets/icons/`; the complete notices ship in `apps/publisher/frontend/public/licenses/lucide.txt`. No icon package is loaded at runtime.
- Go and npm dependencies retain their respective licenses. Exact versions and
  transitive dependencies are recorded in `go.mod`, `go.sum` and `package-lock.json`.
  Their source/license files accompany dependency downloads, not this source tree.
- The Python `imageio-ffmpeg` package downloads/contains FFmpeg under its own
  distribution terms. Consult that installed package's licenses when redistributing
  its binary. This repository distributes neither FFmpeg nor Python environments.

No production video, voice model, social account or third-party content is bundled.
The optional demo is generated locally from FFmpeg test sources.

- Browser preparation uses [chromedp](https://github.com/chromedp/chromedp) v0.23.0
  and cdproto v0.159.0, MIT licensed. Their notices are bundled in
  `apps/publisher/frontend/public/licenses/chromedp.txt`. Chrome is installed and
  updated separately by the user; Publisher does not distribute a browser.
