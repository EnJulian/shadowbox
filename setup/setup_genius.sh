#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Genius API Setup ===${NC}"
echo ""
echo -e "${YELLOW}This script will help you set up Genius API credentials for lyrics functionality.${NC}"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}Creating .env file...${NC}"
    touch .env
fi

echo -e "${BLUE}Step 1: Get your Genius API Access Token${NC}"
echo ""
echo "1. Go to https://genius.com/api-clients"
echo "2. Sign in or create a Genius account"
echo "3. Click 'New API Client'"
echo "4. Fill in the form:"
echo "   - App Name: ShadowBox (or any name you prefer)"
echo "   - App Website URL: http://localhost (or your website)"
echo "   - Redirect URI: http://localhost (can be anything for this use case)"
echo "5. Click 'Save'"
echo "6. Copy the 'Client Access Token' (NOT the Client ID or Client Secret)"
echo ""
echo -e "${YELLOW}Note: The Client Access Token is all you need for lyrics fetching.${NC}"
echo ""

# Prompt for access token
read -p "Enter your Genius Client Access Token: " genius_token

if [ -z "$genius_token" ]; then
    echo -e "${RED}Error: No access token provided. Exiting.${NC}"
    exit 1
fi

# Check if GENIUS_ACCESS_TOKEN already exists in .env
if grep -q "GENIUS_ACCESS_TOKEN=" .env; then
    # Update existing entry
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/GENIUS_ACCESS_TOKEN=.*/GENIUS_ACCESS_TOKEN=$genius_token/" .env
    else
        # Linux
        sed -i "s/GENIUS_ACCESS_TOKEN=.*/GENIUS_ACCESS_TOKEN=$genius_token/" .env
    fi
    echo -e "${GREEN}Updated existing GENIUS_ACCESS_TOKEN in .env${NC}"
else
    # Add new entry
    echo "GENIUS_ACCESS_TOKEN=$genius_token" >> .env
    echo -e "${GREEN}Added GENIUS_ACCESS_TOKEN to .env${NC}"
fi

echo ""
echo -e "${BLUE}Step 2: Testing the connection...${NC}"

# Export the token for this session
export GENIUS_ACCESS_TOKEN="$genius_token"

# Test the connection using Python
python3 -c "
import os
import sys

# Add the current directory to Python path
sys.path.insert(0, '.')

try:
    from meta_ops.lyrics import get_lyrics_from_genius
    
    print('Testing Genius API connection...')
    
    # Test with a well-known song
    test_lyrics = get_lyrics_from_genius('Bohemian Rhapsody', 'Queen')
    
    if test_lyrics:
        print('✅ SUCCESS: Genius API connection working!')
        print(f'   Retrieved lyrics ({len(test_lyrics)} characters)')
    else:
        print('⚠️  WARNING: Could not retrieve test lyrics, but API token seems valid')
        print('   This might be due to the specific test song not being found')
        
except ImportError as e:
    print('❌ ERROR: Missing required library')
    print('   Please run: pip3 install lyricsgenius')
    sys.exit(1)
except Exception as e:
    print(f'❌ ERROR: {e}')
    print('   Please check your access token and try again')
    sys.exit(1)
"

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}=== Setup Complete! ===${NC}"
    echo ""
    echo -e "${YELLOW}Your Genius API credentials have been saved to .env${NC}"
    echo -e "${YELLOW}Lyrics will now be automatically fetched and embedded when downloading music.${NC}"
    echo ""
    echo -e "${BLUE}Usage:${NC}"
    echo "  The lyrics functionality is now integrated into the main download process."
    echo "  When you download music, lyrics will be automatically:"
    echo "  • Searched for using the song title and artist"
    echo "  • Embedded into the audio file metadata"
    echo "  • Available in music players that support lyrics display"
    echo ""
    echo -e "${YELLOW}Note: Make sure to source your environment or restart your terminal${NC}"
    echo -e "${YELLOW}to ensure the new environment variables are loaded.${NC}"
else
    echo ""
    echo -e "${RED}=== Setup Failed ===${NC}"
    echo ""
    echo "Please check your access token and try running this script again."
    echo "Make sure you copied the 'Client Access Token' and not the Client ID or Secret."
fi