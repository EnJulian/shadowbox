```
                ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
                ██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝
                ███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝ 
                ╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗ 
                ███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗
                ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝
```

Single-binary music downloader. Rips audio from various sources, tags it with metadata, cover art, and lyrics, and organizes it under `~/Music/Artist/Album/`.

Needs `yt-dlp` and `ffmpeg` on your `PATH` (`aria2` optional). Run
`shadowbox doctor` to check.

## Install

```bash
# macOS / Linux
brew tap EnJulian/shadowbox && brew install shadowbox

# From source
go install github.com/EnJulian/shadowbox/cmd/shadowbox@latest
```

```powershell
# Windows (Scoop — includes ffmpeg and yt-dlp)
scoop bucket add shadowbox https://github.com/EnJulian/scoop-shadowbox
scoop install shadowbox

# Windows (PowerShell — no Scoop)
irm https://raw.githubusercontent.com/EnJulian/shadowbox/main/scripts/install.ps1 | iex
```

See [docs/INSTALL_WINDOWS.md](docs/INSTALL_WINDOWS.md) for version pinning and manual install.



## Usage

```bash
shadowbox                                   # interactive interface
shadowbox download -q "High Speed Chasing BØRNS"
shadowbox download -q "https://youtu.be/..." -f opus -d ~/Music
shadowbox tag -f track.opus -t "Cocktails in Space" -a "Pharma"
shadowbox enhance ~/Music/Unsorted -r
```

Commands: `download`, `tag`, `enhance`, `config`, `doctor`, `version`.
Run `shadowbox <command> --help` for flags.

## Configure

```bash
shadowbox config set genius.access_token YOUR_TOKEN
```

Metadata and cover art use MusicBrainz, Cover Art Archive, iTunes, and Last.fm
with no API keys required. Genius is optional for lyrics.

Stored at `~/.config/shadowbox/config.yaml`; `GENIUS_*` env vars override it.

## Docs

- [Install on Windows](docs/INSTALL_WINDOWS.md)
- [Install from source](docs/INSTALL_FROM_SOURCE.md)
- [Testing & releasing](docs/RELEASING.md)
- [Security overview (beginner's guide)](docs/SECURITY_OVERVIEW.md)
- [Security policy](SECURITY.md)

## Verifying downloads

Release checksums are signed with Sigstore cosign (keyless) and build provenance
is published as GitHub attestations. After downloading `checksums.txt` from a
[GitHub Release](https://github.com/EnJulian/shadowbox/releases):

```bash
# Verify the checksum signature (requires cosign: https://docs.sigstore.dev)
cosign verify-blob checksums.txt \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp 'https://github.com/EnJulian/shadowbox/.github/workflows/release.yml@refs/tags/.*' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com

# Verify a release archive attestation (requires gh CLI)
gh attestation verify shadowbox-linux-amd64.tar.gz \
  --owner EnJulian --repo shadowbox
```

See [docs/RELEASING.md](docs/RELEASING.md#verifying-downloads) for full details.

## License

MIT — see [LICENSE](LICENSE). Shadowbox invokes [yt-dlp](https://github.com/yt-dlp/yt-dlp),
[FFmpeg](https://ffmpeg.org), and [aria2](https://aria2.github.io) as external
programs; each keeps its own license.