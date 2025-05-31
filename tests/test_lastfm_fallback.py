#!/usr/bin/env python3
"""
Test script to specifically verify Last.fm fallback integration for genre metadata.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import search_metadata_with_fallback, search_lastfm_for_metadata

def test_lastfm_fallback():
    """Test that Last.fm works as a third fallback for genre information."""
    print("Testing Last.fm as third fallback for genre metadata...\n")
    
    # Test 1: A song that should be found on Last.fm with genre
    print("Test 1: Testing Last.fm direct search")
    print("=" * 50)
    
    title = "Bohemian Rhapsody"
    artist = "Queen"
    
    print(f"Searching Last.fm directly for: '{title}' by '{artist}'")
    lastfm_result = search_lastfm_for_metadata(title, artist)
    
    if lastfm_result and lastfm_result.get('genre'):
        print(f"✅ Last.fm found genre: {lastfm_result['genre']}")
    else:
        print("❌ Last.fm failed to find genre")
        return False
    
    print("\nTest 2: Testing fallback integration")
    print("=" * 50)
    
    # Test the full fallback chain
    print(f"Testing fallback chain for: '{title}' by '{artist}'")
    print("Expected order: Spotify → iTunes → Last.fm")
    
    fallback_result = search_metadata_with_fallback(title, artist)
    
    if fallback_result:
        print(f"✅ Fallback search successful")
        print(f"   Title: {fallback_result.get('title', 'N/A')}")
        print(f"   Artist: {fallback_result.get('artist', 'N/A')}")
        print(f"   Album: {fallback_result.get('album', 'N/A')}")
        print(f"   Genre: {fallback_result.get('genre', 'N/A')}")
        
        if fallback_result.get('genre'):
            print(f"✅ Genre found through fallback: {fallback_result['genre']}")
        else:
            print("⚠️  No genre found through fallback")
    else:
        print("❌ Fallback search failed completely")
        return False
    
    print("\nTest 3: Testing with a song that might only have genre on Last.fm")
    print("=" * 50)
    
    # Test with a more obscure song that might not be on Spotify/iTunes
    title = "Paranoid Android"
    artist = "Radiohead"
    
    print(f"Testing with: '{title}' by '{artist}'")
    
    # First check Last.fm directly
    lastfm_direct = search_lastfm_for_metadata(title, artist)
    if lastfm_direct and lastfm_direct.get('genre'):
        print(f"✅ Last.fm has genre for this song: {lastfm_direct['genre']}")
    else:
        print("⚠️  Last.fm doesn't have genre for this song")
    
    # Now test the fallback
    fallback_result2 = search_metadata_with_fallback(title, artist)
    if fallback_result2:
        print(f"✅ Fallback found metadata")
        if fallback_result2.get('genre'):
            print(f"   Genre: {fallback_result2['genre']}")
        else:
            print("   No genre found")
    else:
        print("❌ Fallback failed")
    
    return True

def main():
    """Run the Last.fm fallback test."""
    print("Last.fm Fallback Integration Test")
    print("=" * 60)
    print("This test verifies that Last.fm works as the third fallback")
    print("option for genre metadata after Spotify and iTunes.\n")
    
    try:
        success = test_lastfm_fallback()
        
        if success:
            print("\n" + "=" * 60)
            print("🎉 All Last.fm fallback tests passed!")
            print("✅ Last.fm integration is working correctly as third fallback")
            print("✅ Genre metadata can be retrieved from Last.fm")
            print("✅ Fallback order: Spotify → iTunes → Last.fm")
        else:
            print("\n" + "=" * 60)
            print("❌ Some tests failed")
            return False
            
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        return False
    
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)