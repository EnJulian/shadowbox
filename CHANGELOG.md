# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.2] - 2025-06-01

### Added
- Enhanced YouTube download system with multiple anti-bot bypass strategies
- Automatic retry logic with delays between attempts to avoid rate limiting
- Built-in yt-dlp update functionality accessible from Settings menu
- Update scripts for easy yt-dlp maintenance (`update_ytdlp.sh`, `update_ytdlp.bat`, `update_ytdlp.py`)
- Comprehensive troubleshooting documentation for YouTube download issues
- Better error handling and user feedback for download failures

### Fixed
- YouTube "Sign in to confirm you're not a bot" error through multiple download strategies
- Enhanced user agents and headers to bypass YouTube's anti-bot detection
- Improved error messages with specific suggestions for different failure types

### Changed
- Download system now tries 4 different strategies automatically when YouTube blocks downloads:
  1. Enhanced Headers with realistic browser user agents
  2. Alternative Extractor using Android client
  3. Web Client with browser-like behavior  
  4. Original method with aria2c as fallback
- Settings menu expanded to include yt-dlp update option (now 6 options instead of 5)

## [1.0.1] - 2025-06-01

### Fixed
- Fixed Windows executable to show full menu interface instead of simple CLI prompt
- Updated PyInstaller entry point from `core/main.py` to `core/shbox.py`
- Fixed import issues in `core/shbox.py` for standalone execution

### Changed
- Updated project entry points: `shadowbox` and `shbox` now point to full menu interface
- Added `shadowbox-cli` entry point for simple command-line interface

## [1.0.0] - 2025-XX-XX

### Added
- GitHub Actions workflow for automated cross-platform releases
- PyInstaller configuration for standalone executables
- Build and release scripts for local development
- Support for Linux (x64), Windows (x64), macOS (Intel), and macOS (Apple Silicon)

### Changed
- Updated project structure for better packaging
- Improved documentation with installation options

## [0.1.0] - 2024-01-XX

### Added
- Initial release
- YouTube and Bandcamp audio downloading
- Spotify metadata integration
- iTunes fallback for cover art
- Support for multiple audio formats (opus, mp3, m4a, flac, wav, alac)
- Interactive and CLI modes
- Automatic file organization by artist
- Playlist support for YouTube
- Cross-platform support (Linux, macOS, Windows)

### Features
- High-quality audio extraction
- Metadata injection with cover art
- Spotify API integration for enhanced metadata
- Multiple download sources (YouTube, Bandcamp)
- Configurable output formats and directories
- Batch processing capabilities