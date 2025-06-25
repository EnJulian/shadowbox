#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Testing Genius API Connection ===${NC}"
echo ""

# Check if .env file exists and load it
if [ -f ".env" ]; then
    echo -e "${YELLOW}Loading environment variables from .env...${NC}"
    export $(grep -v '^#' .env | xargs)
else
    echo -e "${RED}Error: .env file not found. Please run setup_genius.sh first.${NC}"
    exit 1
fi

# Check if GENIUS_ACCESS_TOKEN is set
if [ -z "$GENIUS_ACCESS_TOKEN" ]; then
    echo -e "${RED}Error: GENIUS_ACCESS_TOKEN not found in environment.${NC}"
    echo -e "${YELLOW}Please run setup_genius.sh to configure your Genius API credentials.${NC}"
    exit 1
fi

echo -e "${GREEN}Found Genius API token${NC}"
echo ""

# Test the connection using Python
python3 -c "
import os
import sys

# Add the current directory to Python path
sys.path.insert(0, '.')

try:
    from meta_ops.lyrics import get_lyrics_from_genius, search_lyrics_with_fallbacks
    
    print('🎵 Testing Genius API with multiple songs...')
    print('')
    
    # Test songs
    test_songs = [
        ('Bohemian Rhapsody', 'Queen'),
        ('Imagine', 'John Lennon'),
        ('Hotel California', 'Eagles'),
        ('Billie Jean', 'Michael Jackson')
    ]
    
    success_count = 0
    
    for title, artist in test_songs:
        print(f'Testing: \"{title}\" by {artist}')
        lyrics = get_lyrics_from_genius(title, artist)
        
        if lyrics:
            print(f'  ✅ SUCCESS: Found lyrics ({len(lyrics)} characters)')
            success_count += 1
        else:
            print(f'  ❌ FAILED: No lyrics found')
        print('')
    
    print(f'Results: {success_count}/{len(test_songs)} songs found')
    
    if success_count > 0:
        print('')
        print('🎉 Genius API is working correctly!')
        print('   Lyrics functionality is ready to use.')
    else:
        print('')
        print('⚠️  Warning: No lyrics were found for any test songs.')
        print('   This could indicate an issue with the API token or network connection.')
        
except ImportError as e:
    print('❌ ERROR: Missing required library')
    print('   Please run: pip3 install lyricsgenius')
    sys.exit(1)
except Exception as e:
    print(f'❌ ERROR: {e}')
    sys.exit(1)
"

echo ""
echo -e "${BLUE}Test completed.${NC}"