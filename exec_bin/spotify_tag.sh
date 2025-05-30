#!/bin/bash

# Script to download YouTube audio and tag it with Spotify metadata

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if virtual environment exists and activate it
if [ -d "$PROJECT_ROOT/.venv" ]; then
    source "$PROJECT_ROOT/.venv/bin/activate"
    PYTHON="$PROJECT_ROOT/.venv/bin/python"
else
    # Try to use system Python if venv doesn't exist
    PYTHON=$(which python3 || which python)
    if [ -z "$PYTHON" ]; then
        echo "❌ Python not found. Please install Python or run the setup script first."
        exit 1
    fi
fi

# Check if Spotify credentials are set
if [ -f "$HOME/.shadowbox_spotify" ]; then
    source "$HOME/.shadowbox_spotify"
    
    # Check if the credentials are set
    if [[ -z "$SPOTIFY_CLIENT_ID" || -z "$SPOTIFY_CLIENT_SECRET" ]]; then
        echo "❌ Spotify credentials not found in environment variables."
        echo "Try sourcing the credentials file manually:"
        echo "source \"$HOME/.shadowbox_spotify\""
        exit 1
    fi
else
    echo "❌ Spotify credentials file not found."
    echo "Please run the setup script first:"
    echo "./setup/setup_spotify.sh"
    exit 1
fi

# Check if spotipy is installed
if ! "$PYTHON" -c "import spotipy" 2>/dev/null; then
    echo "❌ spotipy module not found. Attempting to install it..."
    "$PYTHON" -m pip install spotipy
    if [ $? -ne 0 ]; then
        echo "❌ Failed to install spotipy. Please install it manually with: pip install spotipy"
        exit 1
    fi
    echo "✅ spotipy module installed successfully."
fi

# Function to show usage
show_usage() {
    echo "Usage: spotify_tag.sh [OPTIONS] <youtube_url or search_query>"
    echo ""
    echo "Options:"
    echo "  -h, --help                 Show this help message"
    echo "  -o, --output FILE          Specify output file [default: auto-generated]"
    echo "  -f, --file FILE            Tag an existing file instead of downloading"
    echo "  -t, --title TITLE          Specify title for search [with -f]"
    echo "  -a, --artist ARTIST        Specify artist for search [with -f]"
    echo ""
    echo "Examples:"
    echo "  spotify_tag.sh https://www.youtube.com/watch?v=dQw4w9WgXcQ"
    echo "  spotify_tag.sh -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ"
    echo "  spotify_tag.sh -f existing_song.opus -t \"Never Gonna Give You Up\" -a \"Rick Astley\""
    echo "  spotify_tag.sh \"Rick Astley Never Gonna Give You Up\""
    echo ""
}

# Parse command line arguments
OUTPUT_FILE=""
EXISTING_FILE=""
TITLE=""
ARTIST=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -f|--file)
            EXISTING_FILE="$2"
            shift 2
            ;;
        -t|--title)
            TITLE="$2"
            shift 2
            ;;
        -a|--artist)
            ARTIST="$2"
            shift 2
            ;;
        *)
            QUERY="$1"
            shift
            ;;
    esac
done

# Check if we have a query or existing file
if [[ -z "$QUERY" && -z "$EXISTING_FILE" ]]; then
    echo "❌ Error: No YouTube URL, search query, or existing file specified."
    show_usage
    exit 1
fi

# Create a simple Python script to handle the request
TEMP_SCRIPT=$(mktemp)
cat > "$TEMP_SCRIPT" << 'EOF'
import sys
import os

# Add project root to Python path
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)

from meta_ops.downloader import download_audio
from meta_ops.spotify_metadata import enhance_existing_file_with_spotify

def main():
    args = sys.argv[1:]
    
    # Parse arguments
    mode = args[0]
    
    if mode == "download":
        query = args[1]
        output_file = args[2] if len(args) > 2 and args[2] else None
        
        # Download and tag
        success = download_audio(query, output_file or '%(title)s.%(ext)s', use_spotify_metadata=True)
        sys.exit(0 if success else 1)
    
    elif mode == "tag":
        file_path = args[1]
        title = args[2] if len(args) > 2 else None
        artist = args[3] if len(args) > 3 else None
        
        # Tag existing file
        success = enhance_existing_file_with_spotify(file_path, title, artist)
        sys.exit(0 if success else 1)
    
    else:
        print(f"Unknown mode: {mode}")
        sys.exit(1)

if __name__ == "__main__":
    main()
EOF

# Run the Python script with appropriate arguments
if [[ -n "$EXISTING_FILE" ]]; then
    # Tag existing file
    "$PYTHON" "$TEMP_SCRIPT" "tag" "$EXISTING_FILE" "$TITLE" "$ARTIST"
else
    # Download and tag
    "$PYTHON" "$TEMP_SCRIPT" "download" "$QUERY" "$OUTPUT_FILE"
fi

# Clean up
rm "$TEMP_SCRIPT"