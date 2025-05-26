#!/usr/bin/env python3
"""
Test script to verify that all components of the Music Downloader work correctly.
"""

import os
import sys
from cover_art import get_album_cover_url, download_cover_image

def test_cover_art():
    """Test the cover art functionality."""
    print("Testing cover art functionality...")
    
    # Test with a well-known song
    title = "Bohemian Rhapsody"
    artist = "Queen"
    
    print(f"Searching for cover art for '{title}' by '{artist}'...")
    image_url = get_album_cover_url(title, artist)
    
    if image_url:
        print(f"✅ Found cover art URL: {image_url}")
        
        # Test downloading the cover art
        cover_path = download_cover_image(image_url, "test_cover.jpg")
        if cover_path and os.path.exists(cover_path):
            print(f"✅ Successfully downloaded cover art to {cover_path}")
            # Clean up
            os.remove(cover_path)
            print(f"Removed test file: {cover_path}")
        else:
            print("❌ Failed to download cover art")
            return False
    else:
        print("❌ Failed to find cover art URL")
        return False
    
    return True

def main():
    """Run all tests."""
    print("Running Music Downloader tests...\n")
    
    # Test dependencies
    try:
        import yt_dlp
        print(f"✅ yt-dlp version: {yt_dlp.version.__version__}")
    except ImportError:
        print("❌ yt-dlp is not installed")
        return False
    
    try:
        import mutagen
        print(f"✅ mutagen version: {mutagen.version_string}")
    except ImportError:
        print("❌ mutagen is not installed")
        return False
    
    try:
        import requests
        print(f"✅ requests version: {requests.__version__}")
    except ImportError:
        print("❌ requests is not installed")
        return False
    
    # Test FFmpeg
    import subprocess
    try:
        result = subprocess.run(["ffmpeg", "-version"], 
                               stdout=subprocess.PIPE, 
                               stderr=subprocess.PIPE, 
                               text=True, 
                               check=True)
        print(f"✅ FFmpeg is installed")
    except (subprocess.SubprocessError, FileNotFoundError):
        print("❌ FFmpeg is not installed or not in PATH")
        return False
    
    print("\nTesting individual components:")
    
    # Test cover art functionality
    if not test_cover_art():
        return False
    
    print("\n✅ All tests passed! The Music Downloader should work correctly.")
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)