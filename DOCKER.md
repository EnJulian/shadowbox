# Docker Setup for Shadowbox

This guide explains how to run Shadowbox using Docker, which provides a consistent environment across different machines.

## Prerequisites

- Docker and Docker Compose installed on your system
- Basic familiarity with Docker commands

## Quick Start

1. **Clone the repository** (if you haven't already):
   ```bash
   git clone https://github.com/lsnen/shadowbox.git
   cd shadowbox
   ```

2. **Set up environment variables** (optional but recommended):
   ```bash
   cp .env.example .env
   # Edit .env file with your API credentials
   ```

3. **Build and run the container**:
   ```bash
   docker-compose up --build
   ```

That's it! The application will start in interactive mode.

## Configuration

### API Credentials (Optional)

For enhanced metadata and album art, you can configure API access:

1. **Spotify API** (recommended):
   - Visit [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/)
   - Create a new app and get your Client ID and Client Secret
   - Add them to your `.env` file

2. **Genius API** (for lyrics):
   - Visit [Genius API Clients](https://genius.com/api-clients)
   - Create a new API client and get your access token
   - Add it to your `.env` file

### Environment Variables

Create a `.env` file in the project root:

```bash
# Spotify API (Optional)
SPOTIFY_CLIENT_ID=your_spotify_client_id_here
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret_here

# Genius API (Optional)
GENIUS_ACCESS_TOKEN=your_genius_access_token_here
```

## Usage Modes

### Interactive Mode (Recommended)

```bash
./docker-run.sh run
```

This starts the interactive terminal interface where you can:
- Search and download music
- Configure settings
- Browse downloaded files

**Note**: Use `./docker-run.sh run` instead of `docker-compose up` for better interactive terminal support.

### Alternative Interactive Mode

If you prefer using docker-compose directly:

```bash
docker-compose up
# or
./docker-run.sh run-up
```

**Troubleshooting Interactive Mode**: If you can't type or interact with the interface, use `./docker-run.sh run` instead.

### CLI Mode

For single downloads or scripting, modify the `docker-compose.yml` command:

```yaml
command: ["python", "/app/core/main.py", "-q", "Artist - Song Title"]
```

Or run directly:

```bash
docker-compose run --rm shadowbox python /app/core/main.py -q "Artist - Song Title"
```

### Background Mode

To keep the container running for manual commands:

```yaml
command: ["tail", "-f", "/dev/null"]
```

Then execute commands:

```bash
docker-compose exec shadowbox python /app/core/main.py -q "Artist - Song Title"
```

## File Management

### Download Locations

- **Downloads**: `./downloads/` - Temporary download location
- **Music**: `./music/` - Organized music library
- **Config**: `./config/` - Application settings and cache

These directories are automatically created and mounted as Docker volumes.

### Accessing Files

Downloaded files are available on your host machine in:
- `./downloads/` - Raw downloads
- `./music/` - Organized by artist/album

## Common Commands

### Build the image
```bash
docker-compose build
```

### Rebuild the image (no cache)
```bash
docker-compose build --no-cache
# Or use the helper script
./docker-run.sh rebuild
```

### Run in background
```bash
docker-compose up -d
```

### View logs
```bash
docker-compose logs -f
```

### Stop the container
```bash
docker-compose down
```

### Remove everything (including volumes)
```bash
docker-compose down -v
```

### Execute commands in running container
```bash
docker-compose exec shadowbox bash
```

### Update yt-dlp
```bash
docker-compose exec shadowbox pip install --upgrade yt-dlp
# Or use the helper script
./docker-run.sh update
```

### Check dependencies
```bash
# Check all dependencies in the container
./docker-run.sh check

# Or run specific checks
docker-compose exec shadowbox ./docker-deps-check.sh system  # System dependencies
docker-compose exec shadowbox ./docker-deps-check.sh python  # Python packages
docker-compose exec shadowbox ./docker-deps-check.sh test    # Functionality tests
```

## Troubleshooting

### Interactive Mode Issues

**Problem**: Can't type or interact with the application when using `docker-compose up`

**Solutions**:
1. **Use the recommended command** (best solution):
   ```bash
   ./docker-run.sh run
   ```

2. **Alternative**: Use docker-compose run directly:
   ```bash
   docker-compose run --rm shadowbox python /app/core/shbox.py
   ```

3. **Check terminal compatibility**:
   ```bash
   # Ensure your terminal supports TTY
   docker-compose run --rm shadowbox python -c "import sys; print('TTY:', sys.stdin.isatty())"
   ```

4. **Force TTY allocation**:
   ```bash
   docker-compose run --rm -it shadowbox python /app/core/shbox.py
   ```

### Permission Issues

If you encounter permission issues with mounted volumes:

```bash
# Fix ownership of mounted directories
sudo chown -R $USER:$USER downloads music config
```

### Container Won't Start

1. Check Docker logs:
   ```bash
   docker-compose logs shadowbox
   ```

2. Rebuild the image:
   ```bash
   docker-compose build --no-cache
   # Or use the helper script
   ./docker-run.sh rebuild
   ```

3. Check system requirements:
   - Ensure Docker has enough memory allocated (at least 2GB recommended)
   - Verify disk space availability

### Audio Download Issues

1. Check all dependencies first:
   ```bash
   ./docker-run.sh check
   ```

2. Update yt-dlp:
   ```bash
   ./docker-run.sh update
   ```

3. Check FFmpeg installation:
   ```bash
   docker-compose exec shadowbox ffmpeg -version
   ```

4. If issues persist, rebuild the container:
   ```bash
   ./docker-run.sh rebuild
   ```

### API Issues

1. Verify your API credentials in the `.env` file
2. Check API rate limits
3. Test API connectivity:
   ```bash
   docker-compose exec shadowbox python -c "import spotipy; print('Spotify OK')"
   ```

## Advanced Configuration

### Custom Audio Formats

Modify the command to specify audio format:

```bash
docker-compose run --rm shadowbox python /app/core/main.py -q "Artist - Song" --format mp3
```

### Custom Output Directory

Mount additional volumes in `docker-compose.yml`:

```yaml
volumes:
  - ./downloads:/app/downloads
  - ./music:/app/music
  - /path/to/your/music/library:/app/external_music
```

### Resource Limits

Add resource constraints to `docker-compose.yml`:

```yaml
deploy:
  resources:
    limits:
      memory: 1G
      cpus: '0.5'
```

## Dependencies

The Docker container includes all necessary dependencies:

### System Dependencies
- **Python 3.11** - Runtime environment
- **FFmpeg** - Audio processing and conversion
- **aria2** - Fast downloads (optional but recommended)
- **curl** - HTTP requests
- **wget** - File downloads
- **git** - Version control (for updates)
- **build-essential** - Compilation tools

### Python Dependencies
- **yt-dlp** - YouTube/video downloading
- **mutagen** - Audio metadata handling
- **requests** - HTTP library
- **pillow** - Image processing for album art
- **spotipy** - Spotify API integration
- **lyricsgenius** - Genius API for lyrics

### Dependency Verification
Run the dependency checker to verify all components:
```bash
./docker-run.sh check
```

This will verify:
- All system tools are installed and working
- Python packages are properly imported
- Application structure is correct
- Basic functionality tests pass
- Environment configuration

## Security Notes

- The container runs as a non-root user for security
- API credentials are passed as environment variables
- Downloaded files are owned by the container user
- No network ports are exposed by default (except for optional web interface)

## Development

For development purposes, you can mount the source code:

```yaml
volumes:
  - .:/app
  - ./downloads:/app/downloads
  - ./music:/app/music
```

This allows you to modify code without rebuilding the image.

## Support

If you encounter issues:

1. Check the main [README.md](README.md) for general troubleshooting
2. Review Docker logs: `docker-compose logs`
3. Ensure your system meets the requirements
4. Try rebuilding the image: `docker-compose build --no-cache`

For more help, visit the [GitHub Issues](https://github.com/lsnen/shadowbox/issues) page.