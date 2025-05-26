#!/bin/bash

# Music Downloader Installation Script

echo "🎵 Installing lsnen's shadowbox..."

# Check if Python 3 is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed. Please install Python 3 and try again."
    exit 1
fi

# Check if FFmpeg is installed
if ! command -v ffmpeg &> /dev/null; then
    echo "❌ FFmpeg is not installed. Please install FFmpeg and try again."
    
    # Suggest installation commands based on OS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "To install FFmpeg on macOS using Homebrew, run:"
        echo "brew install ffmpeg"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "To install FFmpeg on Ubuntu/Debian, run:"
        echo "sudo apt update && sudo apt install ffmpeg"
        echo "To install FFmpeg on Fedora, run:"
        echo "sudo dnf install ffmpeg"
    elif [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* ]]; then
        echo "To install FFmpeg on Windows, download from:"
        echo "https://ffmpeg.org/download.html"
    fi
    
    exit 1
fi

# Check if aria2c is installed
if ! command -v aria2c &> /dev/null; then
    echo "❌ aria2c is not installed. Please install aria2c and try again."
    
    # Suggest installation commands based on OS
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "To install aria2c on macOS using Homebrew, run:"
        echo "brew install aria2"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "To install aria2c on Ubuntu/Debian, run:"
        echo "sudo apt update && sudo apt install aria2"
        echo "To install aria2c on Fedora, run:"
        echo "sudo dnf install aria2"
    elif [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* ]]; then
        echo "To install aria2c on Windows, download from:"
        echo "https://github.com/aria2/aria2/releases"
    fi
    
    exit 1
fi

# Create virtual environment if it doesn't exist
if [ ! -d "../.venv" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv ../.venv
fi

# Activate virtual environment
echo "🔌 Activating virtual environment..."
source ../.venv/bin/activate

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." &> /dev/null && pwd )"

# Make scripts executable
echo "🔧 Making scripts executable..."
chmod +x "$PROJECT_ROOT/core/"*.py
chmod +x "$PROJECT_ROOT/"*.sh
chmod +x "$PROJECT_ROOT/run.sh" "$PROJECT_ROOT/shbox.sh"
chmod +x "$SCRIPT_DIR/"*.sh

# Install dependencies
echo "📥 Installing dependencies..."
pip install -r "$SCRIPT_DIR/requirements.txt"

echo "✅ Installation complete!"
echo "To run the application, use:"
echo "source $PROJECT_ROOT/.venv/bin/activate && $PROJECT_ROOT/shbox.sh"
echo "Or for CLI mode:"
echo "source $PROJECT_ROOT/.venv/bin/activate && $PROJECT_ROOT/run.sh -q \"Artist Track\""