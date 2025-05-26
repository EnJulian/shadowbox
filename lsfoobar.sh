#!/bin/bash

# lsnen's Foobar2000 Music Downloader Interactive Launcher

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

# Make the lsfoobar.py script executable
chmod +x lsfoobar.py

# Run the interactive application
python lsfoobar.py