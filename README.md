```
           ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
           ██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝
           ███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝ 
           ╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗ 
           ███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗
           ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝
```

Shadowbox is a fast, single-binary music acquisition tool. It downloads audio
from YouTube and Bandcamp, converts it to Opus (or your chosen format), and
injects rich metadata, cover art, and lyrics — then files everything neatly under
`~/Music/Artist/Album/`.

It ships as one statically-linked Go binary with no runtime to install. You only
need a couple of external tools on your `PATH`.

## Features

- Download from YouTube videos, YouTube playlists, and Bandcamp.
- Robust downloading with three `yt-dlp` strategies (aria2 acceleration, a
standard pass, and a browser-simulating fallback for anti-bot situations).
- Metadata enrichment from Spotify, with a Last.fm genre fallback.
- Cover art from Spotify, falling back to the iTunes Search API.
- Lyrics embedding via Genius.
- Pure-Go tag writing for Opus, MP3, M4A, and FLAC (cover art and lyrics
included) — no `ffmpeg` round-trip for tagging.
- An interactive terminal interface with five themes, settings, and a library
browser, plus a fully scriptable CLI.



## Requirements

Shadowbox shells out to a few well-known tools:


| Tool                                         | Required | Purpose                            |
| -------------------------------------------- | -------- | ---------------------------------- |
| `[yt-dlp](https://github.com/yt-dlp/yt-dlp)` | yes      | downloading audio                  |
| `[ffmpeg](https://ffmpeg.org/)`              | yes      | audio extraction and conversion    |
| `[aria2](https://aria2.github.io/)`          | optional | faster, multi-connection downloads |


Run `shadowbox doctor` at any time to check what is installed and configured.

## Installation



### Homebrew (macOS / Linux)

```bash
brew tap EnJulian/shadowbox
brew install shadowbox
```

(Or as a one-liner: `brew install EnJulian/shadowbox/shadowbox`.)

This pulls in `ffmpeg` and `yt-dlp` automatically. Add `aria2` for faster
downloads: `brew install aria2`.

### WinGet (Windows)

```bash
winget install EnJulian.shadowbox
```

Install the dependencies separately, for example with WinGet:

```bash
winget install yt-dlp.yt-dlp Gyan.FFmpeg aria2.aria2
```



### Manual

Download the archive for your platform from the
[latest release](https://github.com/EnJulian/shadowbox/releases/latest),
extract it, and place the `shadowbox` binary on your `PATH`.

### From source

```bash
git clone https://github.com/EnJulian/shadowbox.git
cd shadowbox
make build      # produces ./shadowbox
```



## Usage

Run Shadowbox with no arguments to launch the interactive interface:

```bash
shadowbox
```

Or use the CLI directly:

```bash
# Search and download
shadowbox download -q "Imagine Dragons Believer"

# Download a specific URL (YouTube or Bandcamp)
shadowbox download -q "https://youtu.be/7wtfhZwyrcc"

# Download a whole playlist
shadowbox download -q "https://www.youtube.com/playlist?list=..."

# Use Spotify metadata, choose a format and output directory
shadowbox download -q "Adele Hello" -s -f opus -d ~/Music

# Tag an existing file in place
shadowbox tag -f track.opus -t "Hello" -a "Adele"

# Batch-enhance a directory of files
shadowbox enhance ~/Music/Unsorted -r
```



### Commands


| Command                   | Description                                              |
| ------------------------- | -------------------------------------------------------- |
| `shadowbox`               | Launch the interactive interface                         |
| `shadowbox download`      | Download a track or playlist and tag it                  |
| `shadowbox tag`           | Tag an existing file, or download using Spotify metadata |
| `shadowbox enhance <dir>` | Batch-enhance existing audio files                       |
| `shadowbox config`        | View and edit configuration                              |
| `shadowbox doctor`        | Check external tools and credentials                     |
| `shadowbox version`       | Print version information                                |




### `download` flags


| Flag              | Description                                       |
| ----------------- | ------------------------------------------------- |
| `-q, --query`     | Song title and artist, or a URL                   |
| `-d, --directory` | Base music directory (default `~/Music`)          |
| `-o, --output`    | Output filename override (no extension)           |
| `-f, --format`    | Audio format: `opus`, `m4a`, `mp3`, `flac`, `wav` |
| `-s, --spotify`   | Use Spotify for metadata                          |
| `-v, --verbose`   | Verbose logging                                   |




## Configuration

Configuration lives at `~/.config/shadowbox/config.yaml` (or the platform
equivalent) and can be managed with the `config` command:

```bash
shadowbox config set spotify.client_id     YOUR_ID
shadowbox config set spotify.client_secret YOUR_SECRET
shadowbox config set genius.access_token   YOUR_TOKEN
shadowbox config set audio_format          opus
shadowbox config list
```

Credentials may also be supplied through environment variables, which take
precedence over the config file:

- `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`
- `GENIUS_ACCESS_TOKEN`
- Any `SHADOWBOX_*` equivalent (e.g. `SHADOWBOX_AUDIO_FORMAT`).

On first run, an existing `~/.shadowbox_settings.json` from the previous Python
version is imported automatically.

### Getting API credentials

- **Spotify**: create an app at the
[Spotify Developer Dashboard](https://developer.spotify.com/dashboard) to get a
client ID and secret. Used for metadata and cover art.
- **Genius**: create a client at
[Genius API](https://genius.com/api-clients) to get an access token. Used for
lyrics.

Shadowbox works without these — it falls back to iTunes for cover art and
Last.fm for genres — but Spotify and Genius produce the best results.

## Project layout

```
cmd/shadowbox/        program entry point
internal/
  cmd/                Cobra commands
  config/             Viper config + legacy JSON migration
  app/                download-and-tag pipeline orchestration
  download/           yt-dlp wrapper, strategies, URL routing
  apis/               Spotify, iTunes, Last.fm, Genius clients
  cover/              cover-art resolution cascade
  tag/                pure-Go tag writers (opus, mp3, m4a, flac)
  organize/           filename/dir sanitisation and layout
  ui/                 Bubble Tea interactive interface and themes
  log/                styled, tagged output
packaging/            Homebrew and WinGet manifest references
.goreleaser.yaml      release configuration
```



## Development

```bash
make test     # go test -race ./...
make lint     # golangci-lint run
make build    # build the binary
make snapshot # cross-platform build via GoReleaser
```



## License

MIT — see [LICENSE](LICENSE).