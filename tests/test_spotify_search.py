#!/usr/bin/env python3
"""
Test script for Spotify metadata search with different combinations.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import search_spotify_for_metadata

def test_search():
    """Test different search combinations."""
    print("Spotify Search Test Suite")
    print("=" * 60)
    
    # Check if Spotify credentials are available
    spotify_client_id = os.environ.get('SPOTIFY_CLIENT_ID')
    spotify_client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')
    
    if not spotify_client_id or not spotify_client_secret:
        print("⚠️  Spotify credentials not found in environment variables")
        print("   Set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET to test Spotify functionality")
        print("   Skipping Spotify tests...")
        return False
    
    print("✅ Spotify credentials found")
    
    # Test case similar to the one in the user's request
    print("\n=== Test Case 1: 'Fifty' by 'MODERNS MUSIC' ===")
    try:
        metadata = search_spotify_for_metadata("Fifty", "MODERNS MUSIC")
        if metadata:
            print("✅ Search successful!")
            print(f"Found: {metadata.get('title')} by {metadata.get('artist')}")
            print(f"Album: {metadata.get('album')}")
        else:
            print("⚠️  No results found for this search")
    except Exception as e:
        print(f"❌ Error during search: {e}")
        return False
    
    # Test case with a popular song
    print("\n=== Test Case 2: 'Bohemian Rhapsody' by 'Queen' ===")
    try:
        metadata = search_spotify_for_metadata("Bohemian Rhapsody", "Queen")
        if metadata:
            print("✅ Search successful!")
            print(f"Found: {metadata.get('title')} by {metadata.get('artist')}")
            print(f"Album: {metadata.get('album')}")
        else:
            print("❌ No results found for this popular song (unexpected)")
            return False
    except Exception as e:
        print(f"❌ Error during search: {e}")
        return False
    
    # Test case with title only
    print("\n=== Test Case 3: 'Imagine' (no artist) ===")
    try:
        metadata = search_spotify_for_metadata("Imagine")
        if metadata:
            print("✅ Search successful!")
            print(f"Found: {metadata.get('title')} by {metadata.get('artist')}")
            print(f"Album: {metadata.get('album')}")
        else:
            print("⚠️  No results found for title-only search")
    except Exception as e:
        print(f"❌ Error during search: {e}")
        return False
    
    print("\n✅ Spotify search tests completed successfully.")
    return True

def main():
    """Run all Spotify search tests."""
    success = test_search()
    
    print("\n" + "=" * 60)
    if success:
        print("🎉 All Spotify search tests passed!")
    else:
        print("⚠️  Some Spotify search tests failed or were skipped")
    
    return success

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)