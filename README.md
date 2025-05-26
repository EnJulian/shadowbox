# lsfoobar

A Python application that downloads music from YouTube or Bandcamp, converts it to high-quality Opus format, adds metadata, and embeds album art from iTunes. It automatically organizes your music by artist in a customizable directory structure.

## Features

- 🎵 Download music from YouTube or Bandcamp
- 🔍 Search for songs by title and artist
- 🔄 Convert to high-quality Opus format (better quality than MP3)
- 📝 Add title, artist, and album metadata
- 🖼️ Fetch and embed album cover art from iTunes
- 📁 Automatically organize music by artist
- 🚀 Interactive user interface
- 📋 Batch download multiple songs
- ⚡ Fast downloads with aria2c (optional)

## Requirements

- Python 3.6 or higher
- [FFmpeg](https://ffmpeg.org/) (required for audio conversion)
- [aria2c](https://aria2.github.io/) (optional, for faster downloads)
- Internet connection

## Installation

### Automatic Installation (macOS/Linux)

1. Clone or download this repository
2. Run the installation script:
   ```bash
   ./install.sh
   ```

### Manual Installation

1. Clone or download this repository
2. Install FFmpeg if not already installed:
   - macOS: `brew install ffmpeg`
   - Ubuntu/Debian: `sudo apt install ffmpeg`
   - Windows: Download from [ffmpeg.org](https://ffmpeg.org/download.html)
3. Install aria2c (optional, for faster downloads):
   - macOS: `brew install aria2`
   - Ubuntu/Debian: `sudo apt install aria2`
   - Windows: Download from [aria2 GitHub releases](https://github.com/aria2/aria2/releases)
4. Create and activate a virtual environment (recommended):
   ```bash
   python3 -m venv .venv
   source .venv/bin/activate  # On Windows: .venv\Scripts\activate
   ```
5. Install the required Python packages:
   ```bash
   pip install -r requirements.txt
   ```

## Usage

### Interactive Mode (Recommended)

The easiest way to use the application is through the interactive interface:

```bash
./lsfoobar.sh  # On macOS/Linux
lsfoobar.bat   # On Windows
```

This will open a user-friendly menu where you can:
- Search and download songs
- Download from YouTube or Bandcamp URLs
- Batch download multiple songs
- Change settings
- View your downloaded music

### Command-line Mode

For quick downloads or scripting, you can use the command-line interface:

```bash
./run.sh -q "Imagine Dragons Believer"  # On macOS/Linux
run.bat -q "Imagine Dragons Believer"   # On Windows
```

Or download directly from a URL:

```bash
./run.sh -q "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

Options:
- `-q, --query`: Song title and artist, or a direct URL to YouTube or Bandcamp
- `-d, --directory`: Base music directory (default: ~/Music)
- `-o, --output`: Output file name (optional, rarely needed)

### Music Organization

The application automatically organizes your music by artist:

```
~/Music/
  ├── Imagine Dragons/
  │   ├── Believer.opus
  │   └── Believer_cover.jpg
  ├── Rick Astley/
  │   ├── Never Gonna Give You Up.opus
  │   └── Never Gonna Give You Up_cover.jpg
  └── ...
```

You can change the music directory in the interactive mode settings or with the `-d` option in command-line mode.

## Features in Detail

### Interactive Interface

The interactive interface provides a user-friendly way to:

- **Search and Download**: Enter a song title and artist to search and download
- **URL Download**: Directly download from YouTube or Bandcamp URLs
- **Batch Download**: Download multiple songs in one go
- **Settings**: Change your music directory
- **View Downloads**: Browse your downloaded music by artist

### Automatic Organization

- Songs are automatically organized by artist
- Cover art is stored alongside the music files
- Filenames are sanitized to work on all operating systems
- Duplicate songs are handled with unique filenames

### High-Quality Audio

- Downloads the best available audio quality
- Converts to Opus format (better quality than MP3 at the same bitrate)
- Embeds metadata (title, artist, album)
- Embeds album art from iTunes

## How It Works

1. The application searches YouTube for the requested song or uses the provided URL
2. It downloads the best audio quality available using yt-dlp (with aria2c if available)
3. FFmpeg converts the audio to Opus format (higher quality than MP3)
4. The application fetches album art from iTunes
5. Files are organized in artist folders with proper naming
6. Metadata and album art are embedded in the audio file

## Troubleshooting

- **FFmpeg not found**: Make sure FFmpeg is installed and in your system PATH
- **aria2c not found**: The application will still work, but downloads may be slower
- **Download fails**: Check your internet connection and try again
- **No album art**: The application will still work without album art
- **Permission denied**: Make sure you have write permissions to the music directory

## License

This project is open source and available under the MIT License.