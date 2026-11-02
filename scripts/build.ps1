$ErrorActionPreference = 'Stop'
$root = Split-Path $PSScriptRoot -Parent
Push-Location $root
try {
    if (-not (Test-Path 'runtime/mpv.exe')) { & "$PSScriptRoot/setup-mpv.ps1" }
    $env:CGO_ENABLED = '0'
    go tool mygo build -platform windows/amd64 -skip-installer
    if ($LASTEXITCODE -ne 0) { throw 'Build failed.' }
    go run ./scripts/icon-design embed build/windows-amd64/mswplayr.exe resources/icon.ico
    if ($LASTEXITCODE -ne 0) { throw 'Icon resource embedding failed.' }
    go run ./scripts/icon-design verify build/windows-amd64/mswplayr.exe
    if ($LASTEXITCODE -ne 0) { throw 'Icon verification failed.' }
    Copy-Item build/windows-amd64/mswplayr.exe build/mswplayr.exe -Force
    Copy-Item resources/icon.ico,resources/icon.png build/ -Force
    Copy-Item runtime/mpv.exe,runtime/d3dcompiler_43.dll build/ -Force
    Copy-Item README.md build/ -Force
    Write-Output "Player ready: $root/build/mswplayr.exe"
} finally { Pop-Location }
