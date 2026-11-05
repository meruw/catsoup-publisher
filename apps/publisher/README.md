# Publisher desktop app

Go + Wails + Vue/TypeScript. Start with [the Publisher docs](../../README.md).

Build: `powershell -ExecutionPolicy Bypass -File apps/publisher/build.ps1`

Run: `apps/publisher/bin/CATsoupPublisher.exe` (finds the workspace above the executable).

Current milestone: validated delivery review and local SQLite publication drafts.
Account connections, uploads and scheduling are not implemented yet.
