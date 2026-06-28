# Installing Shadowbox on Windows

Shadowbox needs **yt-dlp** and **ffmpeg** on your `PATH` (`aria2` is optional).
Run `shadowbox doctor` after installing to verify.

## Scoop (recommended)

Installs Shadowbox plus its runtime dependencies in one step:

```powershell
scoop bucket add shadowbox https://github.com/EnJulian/scoop-shadowbox
scoop install shadowbox
```

Optional faster downloads:

```powershell
scoop install aria2
```

Upgrade later with `scoop update shadowbox`.

## PowerShell (no Scoop)

Downloads the latest release from GitHub, verifies the SHA256 checksum, and
installs `shadowbox.exe` to `%LOCALAPPDATA%\Programs\Shadowbox`:

```powershell
irm https://raw.githubusercontent.com/EnJulian/shadowbox/main/scripts/install.ps1 | iex
```

Install a specific version:

```powershell
$env:SHADOWBOX_VERSION = 'v1.3.0'
irm https://raw.githubusercontent.com/EnJulian/shadowbox/main/scripts/install.ps1 | iex
```

Or download the script and pass `-Version` directly:

```powershell
iwr -useb https://raw.githubusercontent.com/EnJulian/shadowbox/main/scripts/install.ps1 -OutFile install.ps1
.\install.ps1 -Version v1.3.0
```

You must install **yt-dlp** and **ffmpeg** yourself (e.g. from
[yt-dlp releases](https://github.com/yt-dlp/yt-dlp/releases) and
[ffmpeg builds](https://www.gyan.dev/ffmpeg/builds/)), or switch to Scoop.

## Manual

Download `shadowbox-windows-amd64.zip` from
[GitHub Releases](https://github.com/EnJulian/shadowbox/releases), extract
`shadowbox.exe`, and place it on your `PATH`.

Verify the download with the signed `checksums.txt` on the release page — see
[RELEASING.md](RELEASING.md#verifying-downloads).

## From source

See [INSTALL_FROM_SOURCE.md](INSTALL_FROM_SOURCE.md) if you want to build with Go.
