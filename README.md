# lsnen-foobar

Music acquisition command line tool that rips audio from YouTube/Bandcamp, converts to Opus, and injects metadata + album art.

## Specification

- Python 3.6+
- FFmpeg
- aria2c (optional)

## Deployment

```bash
git clone <repo>
./install.sh  # or manual: pip install -r requirements.txt
```

## Execution

Interactive:
```bash
./lsfoobar.sh  # *nix
lsfoobar.bat   # win
```

CLI:
```bash
./run.sh -q "Artist Track"  # search
./run.sh -q "https://youtube.com/watch?v=dQw4w9WgXcQ"  # direct URL
```

## How It Works

1. Infiltrates YouTube/Bandcamp
2. Extracts highest quality audio stream
3. Transmutes to superior Opus format
4. Injects metadata + cover art from iTunes
5. Auto-organizes by artist

## Flags

- `-q` query/URL
- `-d` output directory
- `-o` filename override

Crafted with intent by [enjulian](https://github.com/enjulian).

MIT License
