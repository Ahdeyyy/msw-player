$ErrorActionPreference = 'Stop'
# The progress bar slows Invoke-WebRequest to a crawl in Windows PowerShell.
$ProgressPreference = 'SilentlyContinue'
$root = Split-Path $PSScriptRoot -Parent
$destination = Join-Path $root 'runtime'
New-Item -ItemType Directory -Path $destination -Force | Out-Null
# CI passes its token: unauthenticated API calls from shared runners hit the rate limit.
$headers = @{}
if ($env:GITHUB_TOKEN) { $headers.Authorization = "Bearer $env:GITHUB_TOKEN" }
$release = Invoke-RestMethod 'https://api.github.com/repos/shinchiro/mpv-winbuild-cmake/releases/tags/20261102' -Headers $headers
$asset = $release.assets | Where-Object { $_.name -eq 'mpv-x86_64-20261102-git-2137d21b2e.7z' } | Select-Object -First 1
if (-not $asset) { throw 'Pinned mpv build is unavailable.' }
$archive = Join-Path $destination $asset.name
Invoke-WebRequest $asset.browser_download_url -OutFile $archive -UseBasicParsing
if ($asset.digest -and $asset.digest.StartsWith('sha256:')) {
    $hash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($hash -ne $asset.digest.Substring(7)) { throw 'mpv download checksum does not match.' }
}
$sevenZip = Get-Command 7z -ErrorAction SilentlyContinue
$extractor = if ($sevenZip) { $sevenZip.Source } else { Join-Path $env:ProgramFiles '7-Zip/7z.exe' }
if (-not (Test-Path -LiteralPath $extractor)) { throw 'Install 7-Zip, then run this script again.' }
& $extractor x $archive "-o$destination" -y | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'mpv extraction failed.' }
Remove-Item -LiteralPath $archive
Write-Output "mpv installed in $destination"
