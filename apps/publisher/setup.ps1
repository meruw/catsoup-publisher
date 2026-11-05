param()
$ErrorActionPreference = 'Stop'
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..'))
$python = Join-Path $repoRoot '.venv\Scripts\python.exe'
if (-not (Test-Path -LiteralPath $python)) { throw 'Create .venv and install requirements.txt first; see the root README.' }
if (-not (Get-Command npm.cmd -ErrorAction SilentlyContinue)) { throw 'Install Node.js 24 LTS first.' }
& $python -I -B (Join-Path $PSScriptRoot 'support\setup_tools.py')
if ($LASTEXITCODE -ne 0) { throw 'Go setup failed.' }
$env:GOPATH = Join-Path $repoRoot '.tools\gopath'
$env:GOCACHE = Join-Path $repoRoot '.tools\gocache'
$env:GOBIN = Join-Path $repoRoot '.tools\bin'
$env:GOTOOLCHAIN = 'local'
& (Join-Path $repoRoot '.tools\go\bin\go.exe') install github.com/wailsapp/wails/v3/cmd/wails3@v3.1.0
if ($LASTEXITCODE -ne 0) { throw 'Wails setup failed.' }
