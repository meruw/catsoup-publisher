# CATsoup Publisher

A lightweight, local Windows desktop workspace for reviewing bilingual video
deliveries and preparing them for publication. Built with Go, Wails and Vue.

**Early development:** delivery validation and ES/EN video previews work. Drafts,
account connections, uploads and scheduling are not implemented yet. No paid
service, hosted backend, Blender installation or AI subscription is required.

![Publisher with a synthetic test video](docs/images/publisher.png)

Browser-rendered preview using the generated demo; no private episode assets.

## What works

- Read-only delivery catalog with file hashes and technical validation.
- Spanish/English previews with playback, audio, seeking and version selection.
- Clear separation between a valid delivery and permission to publish it.
- Custom desktop title bar, keyboard focus and reusable CATsoup design tokens.

The Windows executable is about 11 MiB; it uses the installed WebView2 runtime.
That size excludes WebView2, Python, FFmpeg and development tools. RAM/CPU usage
has not yet been benchmarked. This is a source preview, not an installer release.

## Build and run (Windows x64)

Install Python 3.10 or newer, Node.js 24 LTS with npm, Git and Microsoft WebView2.
Run these commands in PowerShell from the repository root:

```powershell
python -m venv .venv
.venv/Scripts/python.exe -m pip install -r requirements.txt
powershell -ExecutionPolicy Bypass -File apps/publisher/setup.ps1
powershell -ExecutionPolicy Bypass -File apps/publisher/build.ps1
apps/publisher/bin/CATsoupPublisher.exe
```

Setup downloads pinned Go/Wails tools into `.tools/`; build installs locked npm
packages, generates native resources and bindings, type-checks and runs Go tests.
Initial setup requires internet. Review/import/playback run locally afterward.
The Python environment is used for the shared delivery validator, not a server.

Start with an empty library, or generate a synthetic sample (color bars and tones,
not a real episode; ignored by Git and explicitly not publishable):

```powershell
.venv/Scripts/python.exe apps/publisher/support/create_demo.py
apps/publisher/bin/CATsoupPublisher.exe --check
```

Place your own contract-v2 packages in `deliveries/<episode-id>/`. See
[the contract](contracts/delivery/CONTRACT.md). IDs currently use the `catsoup-`
prefix and supported languages are ES/EN. The contract's production provenance
fields describe the original producer; this application does not require that
producer or its source files. Keep all media inside its delivery folder.

The executable finds this workspace above its own directory. To review an
existing compatible workspace, use `--root C:\path\to\workspace`; that workspace
must have the validator and `.venv` layout too. Close the app before rebuilding.

## Development

- [Architecture](docs/publisher/ARCHITECTURE.md)
- [Design system](docs/publisher/DESIGN_SYSTEM.md)
- [Checks and contribution workflow](CONTRIBUTING.md)
- [Security boundaries](SECURITY.md)
- [Current status and next milestone](docs/publisher/STATUS.md)

This repository contains Publisher and the shared delivery contract only.
Production scripts, private episode data, videos, accounts and the original
workspace's Git history are not included. Trends remains a disabled planned item;
its implementation is a separate project.

## License and credits

Original project code is licensed under **GPL-3.0-only**. See [LICENSE](LICENSE).
Copyright (C) 2026 CATsoup Publisher contributors. Provided without warranty.
Third-party components keep their own licenses; see [NOTICES.md](NOTICES.md).
The license does not grant trademark rights or imply endorsement by CATsoup.
