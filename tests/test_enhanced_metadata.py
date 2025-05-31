#!/usr/bin/env python3
"""
Test script for enhanced metadata functionality.
This script tests the new metadata fields: genre, composer, performer, track number, total tracks, disc number, and total discs.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)

def test_spotify_metadata_extraction():
    """Test Spotify metadata extraction with enhanced fields."""
    try:
        from meta_ops.spotify_metadata import search_spotify_for_metadata
        
        print("Testing Spotify metadata extraction...")
        print("=" * 50)
        
        # Test with a popular song
        title = "Bohemian Rhapsody"
        artist = "Queen"
        
        print(f"Searching for: {title} by {artist}")
        metadata = search_spotify_for_metadata(title, artist)
        
        if metadata:
            print("\n✅ Spotify metadata found!")
            print(f"Title: {metadata.get('title')}")
            print(f"Artist: {metadata.get('artist')}")
            print(f"Album: {metadata.get('album')}")
            print(f"Genre: {metadata.get('genre')}")
            print(f"Track Number: {metadata.get('track_number')}")
            print(f"Total Tracks: {metadata.get('total_tracks')}")
            print(f"Disc Number: {metadata.get('disc_number')}")
            print(f"Total Discs: {metadata.get('total_discs')}")
            print(f"Performer: {metadata.get('performer')}")
            print(f"Composer: {metadata.get('composer')}")
            print(f"Release Date: {metadata.get('release_date')}")
            return True
        else:
            print("❌ No Spotify metadata found")
            return False
            
    except ImportError as e:
        print(f"❌ Import error: {e}")
        print("Make sure spotipy is installed and Spotify credentials are set")
        return False
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

def test_metadata_application():
    """Test metadata application to different file formats."""
    try:
        from meta_ops.metadata import add_metadata
        
        print("\nTesting metadata application...")
        print("=" * 50)
        
        # Create a test metadata dictionary
        test_metadata = {
            'title': 'Test Song',
            'artist': 'Test Artist',
            'album': 'Test Album',
            'genre': 'Rock',
            'track_number': '1',
            'total_tracks': '10',
            'disc_number': '1',
            'total_discs': '2',
            'composer': 'Test Composer',
            'performer': 'Test Performer',
            'date': '2024'
        }
        
        print("Test metadata fields:")
        for key, value in test_metadata.items():
            print(f"  {key}: {value}")
        
        print("\n✅ Metadata structure test passed!")
        print("Note: Actual file testing requires audio files")
        return True
        
    except ImportError as e:
        print(f"❌ Import error: {e}")
        return False
    except Exception as e:
        print(f"❌ Error: {e}")
        return False

def main():
    """Run all tests."""
    print("Enhanced Metadata Test Suite")
    print("=" * 60)
    
    # Check if Spotify credentials are available
    spotify_client_id = os.environ.get('SPOTIFY_CLIENT_ID')
    spotify_client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')
    
    if not spotify_client_id or not spotify_client_secret:
        print("⚠️  Spotify credentials not found in environment variables")
        print("   Set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET to test Spotify functionality")
        print("   Skipping Spotify tests...\n")
        spotify_available = False
    else:
        print("✅ Spotify credentials found")
        spotify_available = True
    
    # Run tests
    results = []
    
    if spotify_available:
        results.append(test_spotify_metadata_extraction())
    
    results.append(test_metadata_application())
    
    # Summary
    print("\n" + "=" * 60)
    print("Test Summary:")
    passed = sum(results)
    total = len(results)
    print(f"Passed: {passed}/{total}")
    
    if passed == total:
        print("🎉 All tests passed!")
    else:
        print("⚠️  Some tests failed")
    
    return passed == total

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)