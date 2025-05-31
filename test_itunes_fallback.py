#!/usr/bin/env python3
"""
Test script to verify iTunes fallback functionality for genre metadata.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import search_metadata_with_fallback, search_itunes_for_metadata

def test_itunes_search():
    """Test iTunes metadata search directly."""
    print("Testing iTunes metadata search...")
    print("=" * 50)
    
    # Test with a popular song that should have genre info
    title = "Bohemian Rhapsody"
    artist = "Queen"
    
    print(f"Searching iTunes for: {title} by {artist}")
    metadata = search_itunes_for_metadata(title, artist)
    
    if metadata:
        print(f"✓ Found iTunes metadata:")
        print(f"  Title: {metadata.get('title')}")
        print(f"  Artist: {metadata.get('artist')}")
        print(f"  Album: {metadata.get('album')}")
        print(f"  Genre: {metadata.get('genre')}")
        print(f"  Release Date: {metadata.get('release_date')}")
        return True
    else:
        print("✗ No iTunes metadata found")
        return False

def test_fallback_functionality():
    """Test the combined Spotify + iTunes fallback functionality."""
    print("\nTesting Spotify + iTunes fallback...")
    print("=" * 50)
    
    # Test with a song that might not have genre in Spotify
    title = "Never Gonna Give You Up"
    artist = "Rick Astley"
    
    print(f"Searching with fallback for: {title} by {artist}")
    metadata = search_metadata_with_fallback(title, artist)
    
    if metadata:
        print(f"✓ Found enhanced metadata:")
        print(f"  Title: {metadata.get('title')}")
        print(f"  Artist: {metadata.get('artist')}")
        print(f"  Album: {metadata.get('album')}")
        print(f"  Genre: {metadata.get('genre')}")
        print(f"  Release Date: {metadata.get('release_date')}")
        print(f"  Composer: {metadata.get('composer')}")
        print(f"  Performer: {metadata.get('performer')}")
        
        if metadata.get('genre'):
            print(f"✓ Genre information available: {metadata['genre']}")
        else:
            print("⚠ No genre information found")
        
        return True
    else:
        print("✗ No enhanced metadata found")
        return False

def main():
    """Run all tests."""
    print("iTunes Fallback Functionality Test")
    print("=" * 60)
    
    # Test iTunes search directly
    itunes_success = test_itunes_search()
    
    # Test combined functionality
    fallback_success = test_fallback_functionality()
    
    print("\n" + "=" * 60)
    print("Test Results:")
    print(f"iTunes Search: {'✓ PASS' if itunes_success else '✗ FAIL'}")
    print(f"Fallback Functionality: {'✓ PASS' if fallback_success else '✗ FAIL'}")
    
    if itunes_success and fallback_success:
        print("\n🎉 All tests passed! iTunes fallback is working correctly.")
        return True
    else:
        print("\n❌ Some tests failed. Please check the implementation.")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)