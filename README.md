<div align="center">

```
                                                                   ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
                                                                   ██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝
                                                                   ███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝ 
                                                                   ╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗ 
                                                                   ███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗
                                                                   ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝
```

</div>

Music acquisition command line tool that rips audio from YouTube/Bandcamp, converts to Opus, and injects metadata + album art.

## Specification

- Python 3.6+
- FFmpeg
- aria2c (optional)

## Installation

### Option 1: Download Pre-built Binaries (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/lsnen/shadowbox/releases):

- **Linux (x64)**: `shadowbox-linux-x64.tar.gz`
- **Windows (x64)**: `shadowbox-windows-x64.zip`
- **macOS (Intel)**: `shadowbox-macos-x64.tar.gz`
- **macOS (Apple Silicon)**: `shadowbox-macos-arm64.tar.gz`

**Prerequisites:**
- FFmpeg (for audio processing)
- aria2 (optional, for faster downloads)

**Installation steps:**
1. Download the appropriate file for your platform
2. Extract the archive
3. Run the executable directly

### Option 2: Install from Source

```bash
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
./shbox.sh        # Linux/macOS
shbox.bat         # Windows
```
In the interactive app, you can enable Spotify metadata in the Settings menu.

CLI:
```bash
./run.sh -q "Artist Track"  # search (Linux/macOS)
run.bat -q "Artist Track"   # search (Windows)
./run.sh -q "https://youtube.com/watch?v=dQw4w9WgXcQ"  # direct URL
./run.sh -q "Artist Track" -s  # search with Spotify metadata
```

## How It Works

1. Infiltrates YouTube/Bandcamp
2. Extracts highest quality audio stream
3. Transmutes to the superior Opus format
4. Injects metadata + cover art from Spotify (primary) or iTunes (fallback)
5. Auto-organizes by artist

## Flags

- `-q` query/URL
- `-d` output directory
- `-o` filename override
- `-s` use Spotify for metadata

## Project Structure

- `core/` - Main application files
- `meta_ops/` - Metadata and download operations
- `exec_bin/` - Executable scripts
- `tests/` - Test files
- `setup/` - Installation and setup files


## Spotify Integration

shadowbox uses Spotify for both album covers and metadata enrichment, with iTunes as a fallback for cover art.

**During installation**, you'll be prompted to set up Spotify integration. If you choose to set it up:

1. You'll need a Spotify Developer account at [developer.spotify.com](https://developer.spotify.com/dashboard/)
2. Create a new application to get your Client ID and Client Secret
3. Enter these credentials when prompted during installation

**To set up later or update credentials**:
```bash
./setup/setup_spotify.sh
```

**To verify your Spotify credentials are working**:
```bash
./setup/test_spotify.sh
```

**For future sessions**, you may need to load your credentials:
```bash
source ~/.shadowbox_spotify
```
(Consider adding this line to your shell profile for convenience)

### Using Spotify Metadata for YouTube Downloads

shadowbox can now find and bind Spotify metadata to music files downloaded from YouTube:

**Download a YouTube video and tag it with Spotify metadata**:
```bash
./exec_bin/spotify_tag.sh https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

**Specify an output file**:
```bash
./exec_bin/spotify_tag.sh -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ
```

**Tag an existing file with Spotify metadata**:
```bash
./exec_bin/spotify_tag.sh -f existing_song.opus -t "Never Gonna Give You Up" -a "Rick Astley"
```

**Search YouTube and download with Spotify metadata**:
```bash
./exec_bin/spotify_tag.sh "Rick Astley Never Gonna Give You Up"
```

**Enhance multiple files in a directory**:
```bash
./exec_bin/enhance_files.py /path/to/music/directory
```

If Spotify credentials are not configured, the application will automatically fall back to using iTunes for album covers and will not be able to use Spotify metadata.

### Tools and Libraries
This project wouldn't be possible without the following tools:

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - A youtube-dl fork with additional features and fixes
- [FFmpeg](https://ffmpeg.org/) - A complete, cross-platform solution to record, convert and stream audio and video
- [mutagen](https://github.com/quodlibet/mutagen) - A Python module to handle audio metadata
- [requests](https://github.com/psf/requests) - A simple, yet elegant, HTTP library for Python
- [Pillow](https://python-pillow.org/) - Python Imaging Library fork
- [aria2](https://aria2.github.io/) - A lightweight multi-protocol & multi-source command-line download utility
- [spotipy](https://spotipy.readthedocs.io/) - A lightweight Python library for the Spotify Web API

## Troubleshooting

### YouTube Download Issues

If you encounter the error `"Sign in to confirm you're not a bot"` when downloading from YouTube, this is due to YouTube's anti-bot detection measures. Shadowbox includes several solutions:

#### Quick Fix: Update yt-dlp

**Using Shadowbox Interface:**
1. Run Shadowbox
2. Go to Settings (option 4)
3. Select "Update yt-dlp" (option 5)
4. Follow the prompts

**Using Command Line:**
```bash
# Windows
setup\update_ytdlp.bat

# Linux/macOS
./setup/update_ytdlp.sh

# Python (cross-platform)
python3 setup/update_ytdlp.py
```

#### Enhanced Download System

Shadowbox v1.0.1+ includes an enhanced download system that automatically tries multiple strategies when YouTube blocks downloads:

1. **Enhanced Headers**: Uses realistic browser headers and user agents
2. **Alternative Extractor**: Uses Android client for extraction  
3. **Web Client**: Uses web client with browser-like behavior
4. **Original Method**: Falls back to the original aria2c method

These strategies are applied automatically with delays between attempts to avoid rate limiting.

#### Additional Solutions

- **Try different search terms**: Use various combinations of artist and song names
- **Use direct URLs**: Paste YouTube URLs instead of search queries
- **Wait and retry**: YouTube's anti-bot measures are sometimes temporary
- **Check video availability**: Some videos may be region-locked or private

For detailed troubleshooting information, see [docs/YOUTUBE_DOWNLOAD_ISSUES.md](docs/YOUTUBE_DOWNLOAD_ISSUES.md).

### Common Issues

**FFmpeg not found:**
- Install FFmpeg from [ffmpeg.org](https://ffmpeg.org/download.html)
- Ensure FFmpeg is in your system PATH

**Permission errors:**
- Run setup scripts with appropriate permissions
- On Windows, you may need to run as Administrator
- On Linux/macOS, ensure scripts are executable: `chmod +x script_name.sh`

**Spotify metadata not working:**
- Run the Spotify setup: `./setup/setup_spotify.sh` (Linux/macOS) or `setup\setup_spotify.bat` (Windows)
- Verify credentials are loaded: `./setup/test_spotify.sh`
- Check that environment variables are set correctly

**Virtual environment issues:**
- Activate the virtual environment: `source .venv/bin/activate` (Linux/macOS) or `.venv\Scripts\activate` (Windows)
- Reinstall dependencies: `pip3 install -r setup/requirements.txt`

