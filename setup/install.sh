#!/bin/bash

# Music Downloader Installation Script

echo "Installing shadowbox..."

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

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." &> /dev/null && pwd )"

# Create virtual environment if it doesn't exist
if [ ! -d "$PROJECT_ROOT/.venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv "$PROJECT_ROOT/.venv"
fi

# Activate virtual environment
echo "Activating virtual environment..."
source "$PROJECT_ROOT/.venv/bin/activate"

# Make scripts executable
echo "Making scripts executable..."
find "$PROJECT_ROOT" -name "*.py" -type f -exec chmod +x {} \;
find "$PROJECT_ROOT" -name "*.sh" -type f -exec chmod +x {} \;
chmod +x "$PROJECT_ROOT/run.sh" "$PROJECT_ROOT/shbox.sh"
chmod +x "$SCRIPT_DIR/"*.sh

# Install dependencies
echo "Installing dependencies..."
pip install -r "$SCRIPT_DIR/requirements.txt"

# Explicitly check and install spotipy
echo "Ensuring Spotify API module is installed..."
# Force installation of spotipy regardless of whether it's already installed
echo "Installing spotipy module..."
pip install --force-reinstall spotipy>=2.23.0

# Verify the installation
if ! "$PROJECT_ROOT/.venv/bin/python" -c "import spotipy" 2>/dev/null; then
    echo "First installation attempt failed. Trying alternative method..."
    pip install --upgrade pip
    pip install --force-reinstall spotipy>=2.23.0
    
    # Final verification
    if ! "$PROJECT_ROOT/.venv/bin/python" -c "import spotipy" 2>/dev/null; then
        echo "⚠️ Warning: Could not verify spotipy installation. Spotify features may not work."
    else
        echo "Spotipy successfully installed!"
    fi
else
    echo "Spotipy successfully installed!"
fi

# Ensure meta_ops is a proper Python package
echo "Ensuring meta_ops is a proper Python package..."
if [ ! -f "$PROJECT_ROOT/meta_ops/__init__.py" ]; then
    echo "# meta_ops package initialization" > "$PROJECT_ROOT/meta_ops/__init__.py"
    echo "Created meta_ops/__init__.py"
fi

echo "Installation complete!"

# Ask if user wants to set up Spotify integration
echo ""
echo "Spotify Integration Setup"
echo "------------------------"
echo "Shadowbox can use Spotify as the primary source for album covers."
echo "Would you like to set up Spotify integration now? [y/n]"
read -p "> " setup_spotify

if [[ "$setup_spotify" == "y" || "$setup_spotify" == "Y" ]]; then
    # Run the Spotify setup script
    "$SCRIPT_DIR/setup_spotify.sh"
else
    echo ""
    echo "You can set up Spotify integration later by running:"
    echo "$SCRIPT_DIR/setup_spotify.sh"
fi

echo ""
echo "To run the application, use:"
echo "source $PROJECT_ROOT/.venv/bin/activate && $PROJECT_ROOT/shbox.sh"
echo "Or for CLI mode:"
echo "source $PROJECT_ROOT/.venv/bin/activate && $PROJECT_ROOT/run.sh -q \"Artist Track\""