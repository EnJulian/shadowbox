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

# Windows
winget install EnJulian.shadowbox

# From source
go install github.com/EnJulian/shadowbox/cmd/shadowbox@latest
```



## Usage

```bash
shadowbox                                   # interactive interface
shadowbox download -q "High Speed Chasing BØRNS"
shadowbox download -q "https://youtu.be/..." -s -f opus -d ~/Music
shadowbox tag -f track.opus -t "Cocktails in Space" -a "Pharma"
shadowbox enhance ~/Music/Unsorted -r
```

Commands: `download`, `tag`, `enhance`, `config`, `doctor`, `version`.
Run `shadowbox <command> --help` for flags.

## Configure

```bash
shadowbox config set spotify.client_id     YOUR_ID
shadowbox config set spotify.client_secret YOUR_SECRET
shadowbox config set genius.access_token   YOUR_TOKEN
```

Stored at `~/.config/shadowbox/config.yaml`; `SPOTIFY_*` / `GENIUS_*` env vars
override it. Optional — Shadowbox falls back to iTunes and Last.fm without them.

## Docs

- [Install from source](docs/INSTALL_FROM_SOURCE.md)
- [Testing & releasing](docs/RELEASING.md)



## License

MIT — see [LICENSE](LICENSE). Shadowbox invokes [yt-dlp](https://github.com/yt-dlp/yt-dlp),
[FFmpeg](https://ffmpeg.org), and [aria2](https://aria2.github.io) as external
programs; each keeps its own license.