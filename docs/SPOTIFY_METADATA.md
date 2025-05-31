# Spotify Metadata Integration

This document explains how to use the Spotify metadata integration feature in shadowbox.

## Overview

shadowbox can now find and bind Spotify metadata to music files downloaded from YouTube, rather than extracting metadata from the YouTube URL itself. This feature allows you to:

1. Download a YouTube video and automatically tag it with accurate metadata from Spotify
2. Enhance existing music files with Spotify metadata
3. Process multiple files in a directory to add Spotify metadata

## Prerequisites

Before using this feature, you need to set up Spotify API credentials:

1. Create a Spotify Developer account at [developer.spotify.com](https://developer.spotify.com/dashboard/)
2. Create a new application to get your Client ID and Client Secret
3. Run the setup script:
   ```bash
   ./setup/setup_spotify.sh  # Linux/macOS
   ```
   ```batch
   setup\setup_spotify.bat  # Windows
   ```
4. Enter your Client ID and Client Secret when prompted

## Usage

### 1. Using the main shadowbox command with Spotify metadata

Add the `-s` or `--spotify` flag to use Spotify metadata:

```bash
# Linux/macOS
./run.sh -q "https://www.youtube.com/watch?v=dQw4w9WgXcQ" -s
./run.sh -q "Rick Astley Never Gonna Give You Up" -s
```

```batch
# Windows
run.bat -q "https://www.youtube.com/watch?v=dQw4w9WgXcQ" -s
run.bat -q "Rick Astley Never Gonna Give You Up" -s
```

### 2. Using the interactive app with Spotify metadata

You can enable Spotify metadata in the interactive app:

1. Launch the interactive app:
   ```bash
   ./shbox.sh  # Linux/macOS
   shbox.bat   # Windows
   ```

2. Go to Settings (option 4)
3. Select "Toggle Spotify metadata" (option 2)
4. Return to the main menu and download songs as usual

The app will remember your preference for future sessions.

### 3. Using the dedicated Spotify tag utility

For more control, use the dedicated Spotify tag utility:

```bash
# Linux/macOS
./exec_bin/spotify_tag.sh https://www.youtube.com/watch?v=dQw4w9WgXcQ
./exec_bin/spotify_tag.sh -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ
./exec_bin/spotify_tag.sh -f existing_song.opus -t "Never Gonna Give You Up" -a "Rick Astley"
./exec_bin/spotify_tag.sh "Rick Astley Never Gonna Give You Up"
```

```batch
# Windows
exec_bin\spotify_tag.bat https://www.youtube.com/watch?v=dQw4w9WgXcQ
exec_bin\spotify_tag.bat -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ
exec_bin\spotify_tag.bat -f existing_song.opus -t "Never Gonna Give You Up" -a "Rick Astley"
exec_bin\spotify_tag.bat "Rick Astley Never Gonna Give You Up"
```

### 4. Enhancing multiple files in a directory

To process multiple files in a directory:

```bash
# Linux/macOS
./exec_bin/enhance_files.py /path/to/music/directory
./exec_bin/enhance_files.py -r /path/to/music/directory  # Recursive
./exec_bin/enhance_files.py -e .mp3,.flac,.opus /path/to/music/directory  # Specific extensions
./exec_bin/enhance_files.py -d /path/to/music/directory  # Dry run
```

```batch
# Windows
exec_bin\enhance_files.bat C:\path\to\music\directory
exec_bin\enhance_files.bat -r C:\path\to\music\directory  # Recursive
exec_bin\enhance_files.bat -e .mp3,.flac,.opus C:\path\to\music\directory  # Specific extensions
exec_bin\enhance_files.bat -d C:\path\to\music\directory  # Dry run
```

## How It Works

1. When downloading from YouTube:
   - The system extracts the title and artist from the YouTube video title
   - It searches Spotify for matching tracks
   - It downloads the audio from YouTube
   - It applies the Spotify metadata (title, artist, album) to the downloaded file
   - It embeds album cover art from Spotify

2. When enhancing existing files:
   - The system extracts the title and artist from the filename
   - It searches Spotify for matching tracks
   - It applies the Spotify metadata to the file
   - It embeds album cover art from Spotify

## Troubleshooting

If you encounter issues:

1. Verify your Spotify credentials are working:
   ```bash
   ./setup/test_spotify.sh  # Linux/macOS
   ```

2. Make sure your environment variables are set:
   ```bash
   source ~/.shadowbox_spotify  # Linux/macOS
   ```

3. If Spotify can't find a match, try specifying the title and artist manually:
   ```bash
   ./exec_bin/spotify_tag.sh -f your_file.opus -t "Exact Song Title" -a "Exact Artist Name"
   ```

4. For files with unusual naming patterns, the automatic extraction might not work. In this case, use the manual approach above.