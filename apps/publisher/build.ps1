param([switch]$Run)
$ErrorActionPreference = 'Stop'
$appRoot = $PSScriptRoot
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $appRoot '..\..'))
$go = Join-Path $repoRoot '.tools\go\bin\go.exe'
$wails = Join-Path $repoRoot '.tools\bin\wails3.exe'
if (-not (Test-Path -LiteralPath $go) -or -not (Test-Path -LiteralPath $wails)) {
    throw 'Build tools are missing. Follow docs/publisher/README.md first.'
}
$env:GOPATH = Join-Path $repoRoot '.tools\gopath'
$env:GOCACHE = Join-Path $repoRoot '.tools\gocache'
$env:GOTOOLCHAIN = 'local'
$env:npm_config_cache = Join-Path $repoRoot '.tools\npm-cache'
$env:PATH = (Join-Path $repoRoot '.tools\go\bin') + ';' + $env:PATH
Push-Location $appRoot
try {
    New-Item -ItemType Directory -Path 'bin' -Force | Out-Null
    & $wails generate icons -input frontend/public/brand/catsoup-rounded.png -windowsfilename bin/catsoup.ico '-macfilename=' -sizes '128,64,48,32,24,16'
    if ($LASTEXITCODE -ne 0) { throw 'Windows icon generation failed.' }
    & $wails generate syso -arch amd64 -icon bin/catsoup.ico -manifest resources/windows/app.manifest -out rsrc_windows_amd64.syso
    if ($LASTEXITCODE -ne 0) { throw 'Windows resource generation failed.' }
    & $go mod download
    if ($LASTEXITCODE -ne 0) { throw 'Go dependency download failed.' }
    New-Item -ItemType Directory -Path 'frontend\dist' -Force | Out-Null
    # embed requires a file before the first frontend build.
    if (-not (Test-Path -LiteralPath 'frontend\dist\index.html')) {
        Set-Content -LiteralPath 'frontend\dist\bootstrap.txt' -Value 'Build placeholder' -Encoding utf8
    }
    & $wails generate bindings -ts -i -d frontend\bindings
    if ($LASTEXITCODE -ne 0) { throw 'Bridge generation failed.' }
    # Normalize generated EOF whitespace without changing generated declarations.
    Get-ChildItem -LiteralPath 'frontend\bindings' -Recurse -File | ForEach-Object {
        $content = [System.IO.File]::ReadAllText($_.FullName).TrimEnd() + "`n"
        [System.IO.File]::WriteAllText($_.FullName, $content, [System.Text.UTF8Encoding]::new($false))
    }
    Push-Location frontend
    try {
        & npm.cmd ci --ignore-scripts --no-fund --no-audit
        if ($LASTEXITCODE -ne 0) { throw 'Frontend dependency installation failed.' }
        & npm.cmd run build
        if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed.' }
    } finally { Pop-Location }
    & $go test ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go tests failed.' }
    & $go build -tags production -ldflags '-s -w -H windowsgui' -o bin\CATsoupPublisher.exe .
    if ($LASTEXITCODE -ne 0) { throw 'Desktop build failed.' }
    Write-Output (Join-Path $appRoot 'bin\CATsoupPublisher.exe')
    if ($Run) { Start-Process -FilePath (Join-Path $appRoot 'bin\CATsoupPublisher.exe') -WorkingDirectory $appRoot }
} finally { Pop-Location }
