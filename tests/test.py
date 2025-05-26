#!/usr/bin/env python3
"""
Test script to verify that all components of the Music Downloader work correctly.
"""

import os
import sys

# Add the meta_ops directory to the Python path
sys.path.append(os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'meta_ops'))

from cover_art import get_album_cover_url, download_cover_image
from metadata import extract_metadata, add_metadata

def test_cover_art():
    """Test the cover art functionality."""
    print("Testing cover art functionality...")
    
    # Test with a well-known song
    title = "Bohemian Rhapsody"
    artist = "Queen"
    
    print(f"Searching for cover art for '{title}' by '{artist}'...")
    image_url = get_album_cover_url(title, artist)
    
    if image_url:
        print(f"\033[32m[SUCCESS]\033[0m Found cover art URL: {image_url}")
        
        # Test downloading the cover art
        cover_path = download_cover_image(image_url, "test_cover.jpg")
        if cover_path and os.path.exists(cover_path):
            print(f"\033[32m[SUCCESS]\033[0m Successfully downloaded cover art to {cover_path}")
            # Clean up
            os.remove(cover_path)
            print(f"Removed test file: {cover_path}")
        else:
            print("\033[31m[FAIL]\033[0m Failed to download cover art")
            return False
    else:
        print("\033[31m[FAIL]\033[0m Failed to find cover art URL")
        return False
    
    # Test with a more challenging case
    print("\nTesting fallback cover art search...")
    title = "Obscure Song Title That Probably Doesn't Exist"
    artist = "Famous Artist"  # Using a famous artist as fallback
    
    print(f"Searching for cover art for '{title}' by '{artist}'...")
    image_url = get_album_cover_url(title, artist)
    
    if image_url:
        print(f"\033[32m[SUCCESS]\033[0m Fallback search worked! Found cover art URL: {image_url}")
        # We don't need to download this one
    else:
        print("\033[33m[WARNING]\033[0m Fallback search didn't find a cover, but this is acceptable for the test")
    
    return True

def test_metadata():
    """Test the metadata functionality."""
    print("Testing metadata functionality...")
    
    # Create a test file with metadata
    test_file = "test_metadata.opus"
    
    # First, check if the file already exists
    if os.path.exists(test_file):
        os.remove(test_file)
    
    # Create an empty file
    with open(test_file, "wb") as f:
        f.write(b"\x00" * 1024)  # Write some dummy data
    
    # Test metadata extraction on an empty file
    print("Testing metadata extraction on an empty file...")
    metadata = extract_metadata(test_file)
    
    if metadata['title'] == '' and metadata['artist'] == '' and metadata['album'] == '':
        print("\033[32m[SUCCESS]\033[0m Correctly handled empty file metadata")
    else:
        print("\033[31m[FAIL]\033[0m Unexpected metadata in empty file")
        os.remove(test_file)
        return False
    
    # Clean up
    os.remove(test_file)
    print(f"Removed test file: {test_file}")
    
    return True

def main():
    """Run all tests."""
    print("Running Music Downloader tests...\n")
    
    # Test dependencies
    try:
        import yt_dlp
        print(f"\033[32m[SUCCESS]\033[0m yt-dlp version: {yt_dlp.version.__version__}")
    except ImportError:
        print("\033[31m[FAIL]\033[0m yt-dlp is not installed")
        return False
    
    try:
        import mutagen
        print(f"\033[32m[SUCCESS]\033[0m mutagen version: {mutagen.version_string}")
    except ImportError:
        print("\033[31m[FAIL]\033[0m mutagen is not installed")
        return False
    
    try:
        import requests
        print(f"\033[32m[SUCCESS]\033[0m requests version: {requests.__version__}")
    except ImportError:
        print("\033[31m[FAIL]\033[0m requests is not installed")
        return False
    
    # Test FFmpeg
    import subprocess
    try:
        result = subprocess.run(["ffmpeg", "-version"], 
                               stdout=subprocess.PIPE, 
                               stderr=subprocess.PIPE, 
                               text=True, 
                               check=True)
        print(f"\033[32m[SUCCESS]\033[0m FFmpeg is installed")
    except (subprocess.SubprocessError, FileNotFoundError):
        print("\033[31m[FAIL]\033[0m FFmpeg is not installed or not in PATH")
        return False
    
    print("\nTesting individual components:")
    
    # Test cover art functionality
    if not test_cover_art():
        return False
    
    # Test metadata functionality
    if not test_metadata():
        return False
    
    print("\n\033[32m[COMPLETE]\033[0m All tests passed! The Music Downloader should work correctly.")
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)