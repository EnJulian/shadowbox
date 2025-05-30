#!/bin/bash

# Spotify API Setup Script

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Check if virtual environment exists and activate it
if [ -d "$PROJECT_ROOT/.venv" ]; then
    source "$PROJECT_ROOT/.venv/bin/activate"
    PYTHON="$PROJECT_ROOT/.venv/bin/python"
    
    # Check if spotipy is installed
    if ! "$PYTHON" -c "import spotipy" 2>/dev/null; then
        echo "Installing spotipy module..."
        "$PYTHON" -m pip install spotipy
    fi
fi

echo "Setting up Spotify API integration for shadowbox..."
echo "This script will help you configure your Spotify API credentials."
echo ""
echo "You need to create a Spotify Developer account and register an application to get your credentials."
echo "Visit: https://developer.spotify.com/dashboard/ to create your app."
echo ""

# Check if credentials file exists
CREDENTIALS_FILE="$HOME/.shadowbox_spotify"

# Ask for client ID and secret
read -p "Enter your Spotify Client ID: " CLIENT_ID
read -p "Enter your Spotify Client Secret: " CLIENT_SECRET

# Validate input
if [[ -z "$CLIENT_ID" || -z "$CLIENT_SECRET" ]]; then
    echo "❌ Error: Both Client ID and Client Secret are required."
    exit 1
fi

# Save credentials to file
echo "export SPOTIFY_CLIENT_ID=\"$CLIENT_ID\"" > "$CREDENTIALS_FILE"
echo "export SPOTIFY_CLIENT_SECRET=\"$CLIENT_SECRET\"" >> "$CREDENTIALS_FILE"
chmod 600 "$CREDENTIALS_FILE"

# Source the credentials file
source "$CREDENTIALS_FILE"

# Export the variables for the current session
export SPOTIFY_CLIENT_ID="$CLIENT_ID"
export SPOTIFY_CLIENT_SECRET="$CLIENT_SECRET"

# Verify that the variables are set
if [[ -z "$SPOTIFY_CLIENT_ID" || -z "$SPOTIFY_CLIENT_SECRET" ]]; then
    echo "❌ Warning: Failed to set environment variables. Please try sourcing the file manually:"
    echo "source \"$CREDENTIALS_FILE\""
else
    echo "✅ Spotify credentials successfully loaded into current session."
fi

echo ""
echo "Credentials saved to $CREDENTIALS_FILE"
echo ""
echo "For future sessions, add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
echo "source \"$CREDENTIALS_FILE\""
echo ""
echo "Setup complete!"