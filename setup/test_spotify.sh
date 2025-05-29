#!/bin/bash

# Test Spotify API credentials

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

echo "Testing Spotify API credentials..."

# Source the credentials file if it exists
if [ -f "$HOME/.shadowbox_spotify" ]; then
    source "$HOME/.shadowbox_spotify"
    
    # Check if the credentials are set
    if [[ -z "$SPOTIFY_CLIENT_ID" || -z "$SPOTIFY_CLIENT_SECRET" ]]; then
        echo "❌ Spotify credentials not found in environment variables."
        echo "Try sourcing the credentials file manually:"
        echo "source \"$HOME/.shadowbox_spotify\""
    else
        echo "✅ Spotify credentials found in environment variables:"
        echo "SPOTIFY_CLIENT_ID: ${SPOTIFY_CLIENT_ID:0:4}...${SPOTIFY_CLIENT_ID: -4}"
        echo "SPOTIFY_CLIENT_SECRET: ${SPOTIFY_CLIENT_SECRET:0:4}...${SPOTIFY_CLIENT_SECRET: -4}"
        
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

        # Create a simple Python script to test the credentials
        TEMP_SCRIPT=$(mktemp)
        cat > "$TEMP_SCRIPT" << 'EOF'
import os
import sys
try:
    import spotipy
    from spotipy.oauth2 import SpotifyClientCredentials
except ImportError:
    print("❌ spotipy module not found. Please install it with: pip install spotipy")
    sys.exit(1)

client_id = os.environ.get('SPOTIFY_CLIENT_ID')
client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')

if not client_id or not client_secret:
    print("❌ Spotify credentials not found in Python environment.")
    sys.exit(1)

print(f"✅ Spotify credentials found in Python environment.")

try:
    # Initialize Spotify client
    sp = spotipy.Spotify(auth_manager=SpotifyClientCredentials(
        client_id=client_id,
        client_secret=client_secret
    ))
    
    # Test a simple API call
    results = sp.search(q="test", type="track", limit=1)
    if results and 'tracks' in results and 'items' in results['tracks']:
        print("✅ Successfully connected to Spotify API!")
    else:
        print("❌ Failed to get expected results from Spotify API.")
except Exception as e:
    print(f"❌ Error connecting to Spotify API: {e}")
EOF

        # Run the test script
        echo ""
        echo "Testing Spotify API connection..."
        echo "Using Python: $PYTHON"
        "$PYTHON" "$TEMP_SCRIPT"
        
        # Clean up
        rm "$TEMP_SCRIPT"
    fi
else
    echo "❌ Spotify credentials file not found."
    echo "Please run the setup script first:"
    echo "./setup/setup_spotify.sh"
fi