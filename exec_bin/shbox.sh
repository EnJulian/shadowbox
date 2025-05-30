#!/bin/bash

# Shadowbox Music Downloader Interactive Launcher

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

# Make the shbox.py script executable
chmod +x "$PROJECT_ROOT/core/shbox.py"

# Get the Python executable from the virtual environment
PYTHON="$PROJECT_ROOT/.venv/bin/python"

# Set PYTHONPATH to include the project root
export PYTHONPATH="$PROJECT_ROOT:$PYTHONPATH"

# Run the interactive application
"$PYTHON" "$PROJECT_ROOT/core/shbox.py"