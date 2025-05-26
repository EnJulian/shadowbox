#!/bin/bash

# lsnen's Foobar2000 Music Downloader Run Script

# Check if virtual environment exists
if [ ! -d ".venv" ]; then
    echo "❌ Virtual environment not found. Please run ./install.sh first."
    exit 1
fi

# Activate virtual environment
source .venv/bin/activate

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
echo "🎵 Starting lsnen's Foobar2000 Music Downloader..."
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

# Check if arguments were provided
if [ $# -eq 0 ]; then
    # No arguments, run in interactive mode
    python main.py
else
    # Pass all arguments to the script
    python main.py "$@"
fi