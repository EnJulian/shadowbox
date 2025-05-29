#!/bin/bash

# Shadowbox Music Downloader Run Script

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$SCRIPT_DIR"

# Check if virtual environment exists
if [ ! -d "$PROJECT_ROOT/.venv" ]; then
    echo "❌ Virtual environment not found. Please run ./setup/install.sh first."
    exit 1
fi

# Activate virtual environment
source "$PROJECT_ROOT/.venv/bin/activate"

# Source Spotify credentials if available
if [ -f "$HOME/.shadowbox_spotify" ]; then
    echo "🎵 Loading Spotify credentials..."
    source "$HOME/.shadowbox_spotify"
fi

# Check if aria2c is installed
if ! command -v aria2c &> /dev/null; then
    echo "⚠️ aria2c is not installed. For better download performance, install aria2c:"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "brew install aria2"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "sudo apt install aria2"
    fi
    echo "Continuing without aria2c..."
fi

# Run the application
echo "🎵 Starting lsnen's shadowbox..."
echo "------------------------------------"
echo "Usage examples:"
echo "1. Download by search: ./run.sh -q \"Imagine Dragons Believer\""
echo "2. Download by URL: ./run.sh -q \"https://www.youtube.com/watch?v=dQw4w9WgXcQ\""
echo "3. Specify music directory: ./run.sh -q \"Imagine Dragons Believer\" -d \"~/Music/MyCollection\""
echo "4. Interactive mode: ./run.sh"
echo "------------------------------------"
echo "Options:"
echo "-q, --query: Song title and artist or URL"
echo "-d, --directory: Base music directory (default: ~/Music)"
echo "-o, --output: Output file name (optional, rarely needed)"
echo "------------------------------------"

# Get the Python executable from the virtual environment
PYTHON="$PROJECT_ROOT/.venv/bin/python"

# Check if arguments were provided
if [ $# -eq 0 ]; then
    # No arguments, run in interactive mode
    "$PYTHON" "$PROJECT_ROOT/core/main.py"
else
    # Pass all arguments to the script
    "$PYTHON" "$PROJECT_ROOT/core/main.py" "$@"
fi