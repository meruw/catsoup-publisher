# Publisher desktop app

Go + Wails + Vue/TypeScript. Start with [the Publisher docs](../../README.md).

Build: `powershell -ExecutionPolicy Bypass -File apps/publisher/build.ps1`

Run: `apps/publisher/bin/CATsoupPublisher.exe` (finds the workspace above the executable).

Current milestone: read-only validated delivery catalog and native-window video preview.
Draft storage, connectors and publishing are not implemented yet.
