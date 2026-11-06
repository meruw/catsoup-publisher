# Contributing

Discuss substantial behavior or architecture changes in an issue before a PR.
Keep changes focused and describe the problem, resulting behavior and checks run.
Contributions are made under the project's GPL-3.0-only license.

## Checks

`apps/publisher/build.ps1` runs strict frontend checking, production bundling and
Go tests. To run Go checks separately after setup (from the repository root):

```powershell
$env:GOPATH = "$PWD\.tools\gopath"
$env:GOCACHE = "$PWD\.tools\gocache"
$env:GOTOOLCHAIN = 'local'
Push-Location apps/publisher
try {
    & ..\..\.tools\go\bin\go.exe test ./...
    & ..\..\.tools\go\bin\go.exe vet ./...
} finally { Pop-Location }
```

For the optional real-validator/media integration test, generate the demo first,
then set `$env:CATSOUP_TEST_ROOT = "$PWD"` before running Go tests. The test reads
the synthetic sample and never approves it.

After dependency changes, review upstream release/security notes and run a pinned
Go vulnerability scanner (development tool only):

```powershell
$env:GOBIN = "$PWD\.tools\bin"
& .tools/go/bin/go.exe install golang.org/x/vuln/cmd/govulncheck@v1.9.0
$env:PATH = "$PWD\.tools\go\bin;" + $env:PATH
Push-Location apps/publisher
try { & ..\..\.tools\bin\govulncheck.exe ./... } finally { Pop-Location }
```

For browser checks, install Google Chrome and the development-only harness:

```powershell
$env:PATH = "$PWD\.tools\node-v24.22.2-win-x64;" + $env:PATH
npm.cmd install --prefix .tools/visual-qa --save-exact --ignore-scripts --no-audit --no-fund playwright@1.64.1
.venv/Scripts/python.exe apps/publisher/support/create_demo.py
node apps/publisher/support/check_ui.mjs
node apps/publisher/support/check_design.mjs
node apps/publisher/support/check_thumbnails.mjs
```

Reports/screenshots go into ignored `work/`. Browser checks use a fixture bridge;
they do not prove native Wails behavior. Also check the desktop app when changing
window controls, keyboard navigation or playback. Never commit generated media.

## Publish documentation

Edit the guide in `docs/guides/`. Its Markdown filename is its GitHub wiki page name. Its Jekyll permalink is its address on the Publisher site.

From the repository root, update a clean wiki checkout (clone it once), then review the generated changes:

```powershell
git clone https://github.com/meruw/catsoup-publisher.wiki.git work/wiki-publication
git -C work/wiki-publication pull --ff-only
python scripts/sync_wiki.py work/wiki-publication
python scripts/sync_wiki.py work/wiki-publication --check
git -C work/wiki-publication status --short
git -C work/wiki-publication diff
```

Commit the updated guide files in the wiki checkout and push them to publish. The command copies only the guide bodies. Maintain the wiki home, sidebar and footer directly in the wiki.

## Structure and boundaries

- `apps/publisher/`: desktop application, build scripts and checks.
- `contracts/delivery/`: contract v2 and shared validator. Preserve compatibility;
  contract changes need explicit versioning and review.
- `docs/guides/`: setup, provider and upload guides for the website and wiki.
- `docs/publisher/`: architecture, design, draft storage and project status.
- `deliveries/`, `data/`, `work/`: local ignored files.

Do not add background polling, a routing framework or an icon library for basic
UI tasks. Record bundle/executable growth for desktop-shell changes. Any future
stored drafts need a migration and recovery strategy; credentials belong in the
OS credential store, never frontend storage or SQLite.
