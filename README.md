# lsnen-foobar

Music acquisition command line tool that rips audio from YouTube/Bandcamp, converts to Opus, and injects metadata + album art.

## Specification

- Python 3.6+
- FFmpeg
- aria2c (optional)

## Deployment

```bash
git clone <repo>
cd <repo>

# Set up project structure (first time only)
./setup/organize.sh  # Linux/macOS
setup\organize.bat   # Windows (may require admin privileges)

# Install dependencies
./setup/install.sh   # Linux/macOS
setup\install.bat    # Windows
```

## Execution

Interactive:
```bash
./lsfoobar.sh        # Linux/macOS
lsfoobar.bat         # Windows
```

CLI:
```bash
./run.sh -q "Artist Track"  # search (Linux/macOS)
run.bat -q "Artist Track"   # search (Windows)
./run.sh -q "https://youtube.com/watch?v=dQw4w9WgXcQ"  # direct URL
```

## How It Works

1. Infiltrates YouTube/Bandcamp
2. Extracts highest quality audio stream
3. Transmutes to the superior Opus format
4. Injects metadata + cover art from iTunes
5. Auto-organizes by artist

## Flags

- `-q` query/URL
- `-d` output directory
- `-o` filename override

## Project Structure

- `core/` - Main application files
- `meta_ops/` - Metadata and download operations
- `exec_bin/` - Executable scripts
- `tests/` - Test files
- `setup/` - Installation and setup files

## Credits
Crafted with intent by [enjulian](https://github.com/enjulian).

## License
MIT License
