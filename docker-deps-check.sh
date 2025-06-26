#!/bin/bash

# Docker Dependencies Check Script
# This script verifies all required dependencies are properly installed in the Docker container

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[CHECK]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Function to check if running in Docker
check_docker_environment() {
    if [ -f /.dockerenv ]; then
        print_status "Running inside Docker container"
        return 0
    else
        print_error "This script should be run inside the Docker container"
        echo "Run: docker-compose exec shadowbox bash"
        echo "Then: ./docker-deps-check.sh"
        exit 1
    fi
}

# Function to check system dependencies
check_system_deps() {
    print_header "System Dependencies"
    
    # Check FFmpeg
    if command -v ffmpeg >/dev/null 2>&1; then
        FFMPEG_VERSION=$(ffmpeg -version 2>/dev/null | head -n1 | cut -d' ' -f3)
        print_status "FFmpeg: $FFMPEG_VERSION"
    else
        print_error "FFmpeg not found"
        return 1
    fi
    
    # Check aria2c
    if command -v aria2c >/dev/null 2>&1; then
        ARIA2_VERSION=$(aria2c --version 2>/dev/null | head -n1 | cut -d' ' -f3)
        print_status "aria2c: $ARIA2_VERSION"
    else
        print_warning "aria2c not found (optional but recommended)"
    fi
    
    # Check curl
    if command -v curl >/dev/null 2>&1; then
        CURL_VERSION=$(curl --version 2>/dev/null | head -n1 | cut -d' ' -f2)
        print_status "curl: $CURL_VERSION"
    else
        print_error "curl not found"
        return 1
    fi
    
    # Check git
    if command -v git >/dev/null 2>&1; then
        GIT_VERSION=$(git --version 2>/dev/null | cut -d' ' -f3)
        print_status "git: $GIT_VERSION"
    else
        print_warning "git not found"
    fi
}

# Function to check Python dependencies
check_python_deps() {
    print_header "Python Dependencies"
    
    # Check Python version
    PYTHON_VERSION=$(python --version 2>&1 | cut -d' ' -f2)
    print_status "Python: $PYTHON_VERSION"
    
    # Check pip
    PIP_VERSION=$(pip --version 2>/dev/null | cut -d' ' -f2)
    print_status "pip: $PIP_VERSION"
    
    # Check required Python packages
    local packages=(
        "yt_dlp:yt-dlp"
        "mutagen:mutagen"
        "requests:requests"
        "PIL:pillow"
        "spotipy:spotipy"
        "lyricsgenius:lyricsgenius"
    )
    
    for package_info in "${packages[@]}"; do
        IFS=':' read -r import_name package_name <<< "$package_info"
        
        if python -c "import $import_name" 2>/dev/null; then
            # Try to get version
            if [ "$import_name" = "yt_dlp" ]; then
                VERSION=$(python -c "import yt_dlp; print(yt_dlp.version.__version__)" 2>/dev/null || echo "unknown")
            elif [ "$import_name" = "PIL" ]; then
                VERSION=$(python -c "from PIL import Image; print(Image.__version__)" 2>/dev/null || echo "unknown")
            else
                VERSION=$(python -c "import $import_name; print(getattr($import_name, '__version__', 'unknown'))" 2>/dev/null || echo "unknown")
            fi
            print_status "$package_name: $VERSION"
        else
            print_error "$package_name ($import_name) not found"
            return 1
        fi
    done
}

# Function to check application structure
check_app_structure() {
    print_header "Application Structure"
    
    local required_dirs=(
        "/app/core"
        "/app/meta_ops"
        "/app/setup"
        "/app/downloads"
        "/app/music"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            print_status "Directory exists: $dir"
        else
            print_error "Missing directory: $dir"
            return 1
        fi
    done
    
    local required_files=(
        "/app/core/main.py"
        "/app/core/shbox.py"
        "/app/meta_ops/downloader.py"
        "/app/meta_ops/metadata.py"
        "/app/meta_ops/spotify_metadata.py"
        "/app/setup/requirements.txt"
    )
    
    for file in "${required_files[@]}"; do
        if [ -f "$file" ]; then
            print_status "File exists: $file"
        else
            print_error "Missing file: $file"
            return 1
        fi
    done
}

# Function to test basic functionality
test_basic_functionality() {
    print_header "Basic Functionality Tests"
    
    # Test Python imports
    if python -c "
import sys
sys.path.insert(0, '/app')
from meta_ops.downloader import is_url
from meta_ops.metadata import extract_metadata
from meta_ops.spotify_metadata import search_spotify_for_metadata
print('All core imports successful')
" 2>/dev/null; then
        print_status "Core module imports: OK"
    else
        print_error "Core module imports failed"
        return 1
    fi
    
    # Test yt-dlp basic functionality
    if python -c "
import yt_dlp
ydl = yt_dlp.YoutubeDL({'quiet': True})
print('yt-dlp initialization: OK')
" 2>/dev/null; then
        print_status "yt-dlp initialization: OK"
    else
        print_error "yt-dlp initialization failed"
        return 1
    fi
    
    # Test FFmpeg integration
    if ffmpeg -version >/dev/null 2>&1; then
        print_status "FFmpeg integration: OK"
    else
        print_error "FFmpeg integration failed"
        return 1
    fi
}

# Function to check environment variables
check_environment() {
    print_header "Environment Configuration"
    
    # Check optional API credentials
    if [ -n "$SPOTIFY_CLIENT_ID" ] && [ -n "$SPOTIFY_CLIENT_SECRET" ]; then
        print_status "Spotify API credentials: Configured"
    else
        print_warning "Spotify API credentials: Not configured (optional)"
    fi
    
    if [ -n "$GENIUS_ACCESS_TOKEN" ]; then
        print_status "Genius API credentials: Configured"
    else
        print_warning "Genius API credentials: Not configured (optional)"
    fi
    
    # Check Python path
    if echo "$PYTHONPATH" | grep -q "/app"; then
        print_status "PYTHONPATH: Configured"
    else
        print_warning "PYTHONPATH may not include /app"
    fi
}

# Function to run all checks
run_all_checks() {
    local failed=0
    
    echo "Starting comprehensive dependency check..."
    echo
    
    check_docker_environment || failed=1
    echo
    
    check_system_deps || failed=1
    echo
    
    check_python_deps || failed=1
    echo
    
    check_app_structure || failed=1
    echo
    
    test_basic_functionality || failed=1
    echo
    
    check_environment
    echo
    
    if [ $failed -eq 0 ]; then
        print_status "All dependency checks passed! ✅"
        echo
        echo "Your Shadowbox Docker container is ready to use."
        echo "You can now run music downloads with confidence."
    else
        print_error "Some dependency checks failed! ❌"
        echo
        echo "Please rebuild the Docker image to fix missing dependencies:"
        echo "docker-compose build --no-cache"
        exit 1
    fi
}

# Main execution
if [ "${1:-all}" = "all" ]; then
    run_all_checks
elif [ "$1" = "system" ]; then
    check_system_deps
elif [ "$1" = "python" ]; then
    check_python_deps
elif [ "$1" = "structure" ]; then
    check_app_structure
elif [ "$1" = "test" ]; then
    test_basic_functionality
elif [ "$1" = "env" ]; then
    check_environment
else
    echo "Usage: $0 [all|system|python|structure|test|env]"
    echo
    echo "  all       - Run all checks (default)"
    echo "  system    - Check system dependencies"
    echo "  python    - Check Python packages"
    echo "  structure - Check application structure"
    echo "  test      - Test basic functionality"
    echo "  env       - Check environment variables"
    exit 1
fi