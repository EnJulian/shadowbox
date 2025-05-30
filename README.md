# shadowbox

Music acquisition command line tool that rips audio from YouTube/Bandcamp, converts to Opus, and injects metadata + album art.

## Specification

- Python 3.6+
- FFmpeg
- aria2c (optional)

## Deployment

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


## License
MIT License

Crafted with intent by [enjulian](https://github.com/enjulian).
