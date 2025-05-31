# Shadowbox Dependency Audit

This document compares the dependencies used in the source installation vs. the release builds to ensure all packages are properly accounted for.

## Source Installation Dependencies

### From `setup/requirements.txt`:
```
yt-dlp>=2025.5.22
mutagen>=1.47.0
requests>=2.32.3
pillow>=10.0.0
spotipy>=2.23.0
```

### From `pyproject.toml`:
```
yt-dlp>=2025.5.22
mutagen>=1.47.0
requests>=2.32.3
pillow>=10.0.0
spotipy>=2.23.0
```

### System Dependencies (from install scripts):
- **FFmpeg** - Required for audio processing
- **aria2** - Optional, for faster downloads
- **Python 3.8+** - Runtime environment

## Code Analysis - Actual Imports Used

### Core Application (`core/main.py`):
```python
import sys, os, argparse, glob, shutil, re
from downloader import download_audio, is_url, is_youtube_playlist
from metadata import add_metadata, extract_metadata
from cover_art import get_album_cover_url, download_cover_image
from settings import load_settings, save_settings, ...
```

### Metadata Operations (`meta_ops/`):

#### `downloader.py`:
```python
from yt_dlp import YoutubeDL
import os, logging, re, subprocess
from meta_ops.spotify_metadata import process_youtube_url_with_spotify
```

#### `metadata.py`:
```python
from mutagen.easyid3 import EasyID3
from mutagen.mp3 import MP3
from mutagen.id3 import APIC, ID3
from mutagen.oggopus import OggOpus
from mutagen.flac import Picture
from mutagen.mp4 import MP4, MP4Cover
import os, base64
```

#### `cover_art.py`:
```python
import requests, os, urllib.parse, time
import spotipy
from spotipy.oauth2 import SpotifyClientCredentials
```

#### `spotify_metadata.py`:
```python
import os, re, subprocess, json
import spotipy
from spotipy.oauth2 import SpotifyClientCredentials
from meta_ops.metadata import add_metadata
from meta_ops.cover_art import download_cover_image
```

## PyInstaller Configuration Analysis

### Current `shadowbox.spec` includes:

#### Data Files:
- yt_dlp data files
- mutagen data files
- spotipy data files
- requests data files
- certifi data files
- Project modules (meta_ops, core)

#### Hidden Imports:
- All yt_dlp submodules
- All mutagen submodules
- All spotipy submodules
- Core dependencies: requests, PIL, spotipy
- Standard library modules
- Network/HTTP modules
- System modules

## Potential Issues and Solutions

### 1. Missing Transitive Dependencies

**Issue**: Some dependencies of our main packages might not be automatically detected.

**Solution**: Added comprehensive hidden imports including:
- `urllib3`, `certifi`, `charset_normalizer`, `idna` (requests dependencies)
- `PIL.Image`, `PIL.ImageFile` (Pillow submodules)
- `spotipy.oauth2` (Spotify authentication)

### 2. System Dependencies in Release Builds

**Issue**: Release builds need FFmpeg and aria2 but users might not have them.

**Current Status**: 
- ✅ GitHub Actions installs FFmpeg and aria2 during build
- ❌ End users still need to install these separately

**Recommendation**: Document this clearly in release notes and README.

### 3. Platform-Specific Code

**Issue**: Some dependencies (especially PIL) have platform-specific code.

**Solution**: PyInstaller handles this automatically, but we've added comprehensive imports to be safe.

### 4. Dynamic Imports

**Issue**: Code uses dynamic imports that PyInstaller might miss.

**Found**: 
- `subprocess` calls to external tools (yt-dlp, ffmpeg)
- JSON parsing for API responses
- File system operations

**Solution**: Added all standard library modules that might be dynamically imported.

## Release Build Verification Checklist

### Before Each Release:

1. **Test Local Build**:
   ```bash
   python build.py
   ./dist/shadowbox --help
   ```

2. **Verify Dependencies**:
   ```bash
   # Check if all imports work
   python -c "
   import yt_dlp
   import mutagen
   import requests
   import PIL
   import spotipy
   print('All dependencies imported successfully')
   "
   ```

3. **Test Core Functionality**:
   ```bash
   # Test with a simple YouTube URL
   ./dist/shadowbox "https://youtube.com/watch?v=dQw4w9WgXcQ"
   ```

4. **Check File Size**:
   - Linux/macOS: Should be ~50-100MB
   - Windows: Should be ~50-100MB
   - If larger, review excludes in shadowbox.spec

### After GitHub Actions Build:

1. **Download and Test Each Platform**:
   - Extract archive
   - Run `./shadowbox --help`
   - Test basic download functionality

2. **Verify System Dependencies**:
   - Ensure FFmpeg is mentioned in release notes
   - Ensure aria2 is mentioned as optional

## Dependency Synchronization

### Files to Keep in Sync:

1. **`setup/requirements.txt`** ← Source of truth
2. **`pyproject.toml`** dependencies section
3. **`shadowbox.spec`** hiddenimports section
4. **`.github/workflows/release.yml`** pip install commands

### Update Process:

When adding a new dependency:

1. Add to `setup/requirements.txt`
2. Add to `pyproject.toml` dependencies
3. Add to `shadowbox.spec` hiddenimports if needed
4. Test local build
5. Update this audit document

## Current Status: ✅ SYNCHRONIZED

All dependency files are currently synchronized and include:

- **Core Python packages**: All accounted for
- **System dependencies**: Documented and handled in CI
- **Hidden imports**: Comprehensive list in PyInstaller spec
- **Data files**: All necessary data files included

## Recommendations

1. **Regular Audits**: Run this comparison monthly or when adding dependencies
2. **Automated Testing**: Consider adding dependency verification to CI
3. **Documentation**: Keep release notes updated with system requirements
4. **Size Monitoring**: Track executable sizes to catch bloat early

## External Tool Dependencies

### Required by Users:
- **FFmpeg**: Audio processing (conversion, metadata)
- **aria2**: Optional, faster downloads

### Bundled in Executable:
- All Python packages and dependencies
- Python runtime
- Standard library modules

### Not Bundled (User Must Install):
- FFmpeg binary
- aria2 binary (optional)
- Operating system libraries (handled by OS)