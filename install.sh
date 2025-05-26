#!/bin/bash

# Music Downloader Installation Script

echo "🎵 Installing lsnen's Foobar2000 Music Downloader..."

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
if [ ! -d ".venv" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv .venv
fi

# Activate virtual environment
echo "🔌 Activating virtual environment..."
source .venv/bin/activate

# Install dependencies
echo "📥 Installing dependencies..."
pip install -r requirements.txt

echo "✅ Installation complete!"
echo "To run the application, use:"
echo "source .venv/bin/activate && python3 main.py"