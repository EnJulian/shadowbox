#!/usr/bin/env python3
"""
Debug script to test Last.fm genre integration without creating audio files.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import search_metadata_with_fallback, search_lastfm_for_metadata

def test_genre_metadata_flow():
    """Test the metadata flow to ensure genre is being passed correctly."""
    print("Testing Last.fm genre metadata flow...\n")
    
    # Test 1: Direct Last.fm search
    print("Test 1: Direct Last.fm search")
    print("=" * 40)
    
    title = "Bohemian Rhapsody"
    artist = "Queen"
    
    print(f"Searching Last.fm for: '{title}' by '{artist}'")
    lastfm_result = search_lastfm_for_metadata(title, artist)
    
    if lastfm_result:
        print("✅ Last.fm search successful")
        print(f"   Title: {lastfm_result.get('title', 'N/A')}")
        print(f"   Artist: {lastfm_result.get('artist', 'N/A')}")
        print(f"   Album: {lastfm_result.get('album', 'N/A')}")
        print(f"   Genre: {lastfm_result.get('genre', 'N/A')}")
        
        if lastfm_result.get('genre'):
            print(f"✅ Genre found: '{lastfm_result['genre']}'")
        else:
            print("❌ No genre in Last.fm result")
            return False
    else:
        print("❌ Last.fm search failed")
        return False
    
    # Test 2: Fallback integration
    print("\nTest 2: Fallback integration")
    print("=" * 40)
    
    print(f"Testing fallback for: '{title}' by '{artist}'")
    fallback_result = search_metadata_with_fallback(title, artist)
    
    if fallback_result:
        print("✅ Fallback search successful")
        print(f"   Title: {fallback_result.get('title', 'N/A')}")
        print(f"   Artist: {fallback_result.get('artist', 'N/A')}")
        print(f"   Album: {fallback_result.get('album', 'N/A')}")
        print(f"   Genre: {fallback_result.get('genre', 'N/A')}")
        
        if fallback_result.get('genre'):
            print(f"✅ Genre found through fallback: '{fallback_result['genre']}'")
        else:
            print("❌ No genre in fallback result")
            return False
    else:
        print("❌ Fallback search failed")
        return False
    
    # Test 3: Test with a song that might only have genre on Last.fm
    print("\nTest 3: Testing with different song")
    print("=" * 40)
    
    title2 = "Creep"
    artist2 = "Radiohead"
    
    print(f"Testing Last.fm for: '{title2}' by '{artist2}'")
    lastfm_result2 = search_lastfm_for_metadata(title2, artist2)
    
    if lastfm_result2 and lastfm_result2.get('genre'):
        print(f"✅ Last.fm found genre for {title2}: '{lastfm_result2['genre']}'")
    else:
        print(f"⚠️  Last.fm didn't find genre for {title2}")
    
    print(f"Testing fallback for: '{title2}' by '{artist2}'")
    fallback_result2 = search_metadata_with_fallback(title2, artist2)
    
    if fallback_result2:
        print("✅ Fallback search successful")
        if fallback_result2.get('genre'):
            print(f"✅ Genre found through fallback: '{fallback_result2['genre']}'")
        else:
            print("⚠️  No genre found through fallback")
    else:
        print("❌ Fallback search failed")
    
    # Test 4: Test scenario where we force Last.fm as third fallback
    print("\nTest 4: Simulating third fallback scenario")
    print("=" * 40)
    
    # Test with a song that might not be on Spotify but is on Last.fm
    title3 = "Black"
    artist3 = "Pearl Jam"
    
    print(f"Testing Last.fm for: '{title3}' by '{artist3}'")
    lastfm_result3 = search_lastfm_for_metadata(title3, artist3)
    
    if lastfm_result3 and lastfm_result3.get('genre'):
        print(f"✅ Last.fm has genre: '{lastfm_result3['genre']}'")
        
        # Now test the full fallback
        print(f"Testing full fallback for: '{title3}' by '{artist3}'")
        fallback_result3 = search_metadata_with_fallback(title3, artist3)
        
        if fallback_result3:
            if fallback_result3.get('genre'):
                print(f"✅ Fallback preserved genre: '{fallback_result3['genre']}'")
                
                # Check if it's the same genre from Last.fm
                if fallback_result3['genre'] == lastfm_result3['genre']:
                    print("✅ Genre matches Last.fm source")
                else:
                    print(f"⚠️  Genre differs - Last.fm: '{lastfm_result3['genre']}', Fallback: '{fallback_result3['genre']}'")
            else:
                print("❌ Fallback lost the genre")
                return False
        else:
            print("❌ Fallback failed completely")
    else:
        print(f"⚠️  Last.fm doesn't have genre for {title3}")
    
    return True

def test_metadata_dict_structure():
    """Test that the metadata dictionary structure is correct."""
    print("\nTesting metadata dictionary structure...")
    print("=" * 50)
    
    title = "Bohemian Rhapsody"
    artist = "Queen"
    
    # Get metadata from Last.fm
    lastfm_metadata = search_lastfm_for_metadata(title, artist)
    
    if lastfm_metadata:
        print("✅ Last.fm metadata structure:")
        for key, value in lastfm_metadata.items():
            print(f"   {key}: {repr(value)}")
        
        # Check if genre is a string and not empty
        genre = lastfm_metadata.get('genre')
        if isinstance(genre, str) and genre.strip():
            print(f"✅ Genre is valid string: '{genre}'")
        else:
            print(f"❌ Genre is invalid: {repr(genre)}")
            return False
    else:
        print("❌ No Last.fm metadata to check")
        return False
    
    # Get metadata from fallback
    fallback_metadata = search_metadata_with_fallback(title, artist)
    
    if fallback_metadata:
        print("\n✅ Fallback metadata structure:")
        for key, value in fallback_metadata.items():
            print(f"   {key}: {repr(value)}")
        
        # Check if genre is preserved
        genre = fallback_metadata.get('genre')
        if isinstance(genre, str) and genre.strip():
            print(f"✅ Fallback genre is valid string: '{genre}'")
        else:
            print(f"❌ Fallback genre is invalid: {repr(genre)}")
            return False
    else:
        print("❌ No fallback metadata to check")
        return False
    
    return True

def main():
    """Run the debug tests."""
    print("Last.fm Genre Debug Test")
    print("=" * 60)
    print("This test debugs the Last.fm genre integration to identify")
    print("where the issue might be in the metadata flow.\n")
    
    try:
        # Test the metadata flow
        if not test_genre_metadata_flow():
            print("\n❌ Metadata flow test failed")
            return False
        
        # Test the metadata structure
        if not test_metadata_dict_structure():
            print("\n❌ Metadata structure test failed")
            return False
        
        print("\n" + "=" * 60)
        print("🎉 All debug tests passed!")
        print("✅ Last.fm genre metadata is flowing correctly")
        print("✅ Fallback system is preserving genre information")
        print("✅ Metadata dictionary structure is correct")
        print("\nIf genre is not being applied to files, the issue is likely")
        print("in the file writing process, not the metadata retrieval.")
        
    except Exception as e:
        print(f"\n❌ Debug test failed with error: {e}")
        import traceback
        traceback.print_exc()
        return False
    
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)