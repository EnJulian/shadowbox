#Requires -Version 5.1
<#
.SYNOPSIS
  Install Shadowbox on Windows from GitHub Releases.

.DESCRIPTION
  Downloads shadowbox-windows-amd64.zip, verifies its SHA256 against the
  release checksums.txt, installs shadowbox.exe under %LOCALAPPDATA%\Programs\Shadowbox,
  and adds that folder to your user PATH.

  Shadowbox still needs yt-dlp and ffmpeg on your PATH — run shadowbox doctor
  after installing. Scoop users get those automatically via scoop install shadowbox.

.PARAMETER Version
  Release tag to install (e.g. v1.3.0 or 1.3.0). Defaults to the latest release.

.PARAMETER InstallDir
  Directory for shadowbox.exe. Defaults to %LOCALAPPDATA%\Programs\Shadowbox.

.EXAMPLE
  irm https://raw.githubusercontent.com/EnJulian/shadowbox/main/scripts/install.ps1 | iex

.EXAMPLE
  .\install.ps1 -Version v1.3.0
#>
[CmdletBinding()]
param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\Shadowbox"
)

$ErrorActionPreference = 'Stop'

if (-not $Version -and $env:SHADOWBOX_VERSION) {
    $Version = $env:SHADOWBOX_VERSION
}

$Repo = 'EnJulian/shadowbox'
$AssetName = 'shadowbox-windows-amd64.zip'

function Resolve-ReleaseTag {
    param([string]$InputVersion)

    if (-not $InputVersion) {
        return ""
    }
    if ($InputVersion -match '^v') {
        return $InputVersion
    }
    return "v$InputVersion"
}

function Get-GitHubRelease {
    param([string]$Tag)

    if ($Tag) {
        $uri = "https://api.github.com/repos/$Repo/releases/tags/$Tag"
    } else {
        $uri = "https://api.github.com/repos/$Repo/releases/latest"
    }

    return Invoke-RestMethod -Uri $uri -Headers @{ 'User-Agent' = 'shadowbox-installer' }
}

function Get-ExpectedSha256 {
    param(
        [string]$ChecksumsPath,
        [string]$FileName
    )

    foreach ($line in Get-Content -Path $ChecksumsPath) {
        if ($line -match '^\s*([a-f0-9]{64})\s+(.+)\s*$') {
            if ($Matches[2].Trim() -eq $FileName) {
                return $Matches[1].ToLowerInvariant()
            }
        }
    }

    throw "checksums.txt does not list $FileName."
}

$tag = Resolve-ReleaseTag -InputVersion $Version
Write-Host "Fetching release metadata..."
$release = Get-GitHubRelease -Tag $tag
$tag = $release.tag_name
Write-Host "Installing Shadowbox $tag"

$asset = $release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1
if (-not $asset) {
    throw "Release $tag has no $AssetName asset."
}

$checksumsAsset = $release.assets | Where-Object { $_.name -eq 'checksums.txt' } | Select-Object -First 1
if (-not $checksumsAsset) {
    throw "Release $tag has no checksums.txt asset."
}

$tempRoot = Join-Path $env:TEMP ("shadowbox-install-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tempRoot -Force | Out-Null

try {
    $zipPath = Join-Path $tempRoot $AssetName
    $checksumsPath = Join-Path $tempRoot 'checksums.txt'

    Write-Host "Downloading $AssetName..."
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipPath -UseBasicParsing

    Write-Host "Verifying checksum..."
    Invoke-WebRequest -Uri $checksumsAsset.browser_download_url -OutFile $checksumsPath -UseBasicParsing
    $expected = Get-ExpectedSha256 -ChecksumsPath $checksumsPath -FileName $AssetName
    $actual = (Get-FileHash -Path $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        throw "Checksum mismatch for $AssetName."
    }

    Write-Host "Extracting..."
    $extractDir = Join-Path $tempRoot 'extract'
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $exe = Get-ChildItem -Path $extractDir -Filter 'shadowbox.exe' -Recurse | Select-Object -First 1
    if (-not $exe) {
        throw 'Archive does not contain shadowbox.exe.'
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $target = Join-Path $InstallDir 'shadowbox.exe'
    Copy-Item -Path $exe.FullName -Destination $target -Force

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -notlike "*$InstallDir*") {
        $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added $InstallDir to your user PATH."
        Write-Host "Open a new terminal if shadowbox is not found yet."
    }

    Write-Host ""
    Write-Host "Installed to $target"
    & $target version
    Write-Host ""
    Write-Host "Run 'shadowbox doctor' to check for yt-dlp and ffmpeg."
    Write-Host "See https://github.com/EnJulian/shadowbox/blob/main/docs/INSTALL_WINDOWS.md"
} finally {
    Remove-Item -Path $tempRoot -Recurse -Force -ErrorAction SilentlyContinue
}
