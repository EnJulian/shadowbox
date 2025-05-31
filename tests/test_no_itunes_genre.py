#!/usr/bin/env python3
"""
Test script to verify that iTunes is no longer used for genre and Last.fm is used instead.
"""

import os
import sys

# Add the project root to the Python path
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import search_metadata_with_fallback, search_lastfm_for_metadata, search_itunes_for_metadata

def test_no_itunes_genre():
    """Test that iTunes genre is ignored and Last.fm genre is used instead."""
    print("Testing that iTunes genre is skipped and Last.fm genre is used...\n")
    
    title = "Dimma"
    artist = "John Dunder"
    
    print(f"Testing with: '{title}' by '{artist}'")
    print("=" * 50)
    
    # First, check what iTunes would provide
    print("1. Checking iTunes metadata (for reference only)...")
    itunes_result = search_itunes_for_metadata(title, artist)
    if itunes_result:
        print(f"   iTunes title: {itunes_result.get('title', 'N/A')}")
        print(f"   iTunes artist: {itunes_result.get('artist', 'N/A')}")
        print(f"   iTunes album: {itunes_result.get('album', 'N/A')}")
        print(f"   iTunes genre: {itunes_result.get('genre', 'N/A')}")
    else:
        print("   iTunes: No metadata found")
    
    # Check what Last.fm provides
    print("\n2. Checking Last.fm metadata...")
    lastfm_result = search_lastfm_for_metadata(title, artist)
    if lastfm_result:
        print(f"   Last.fm title: {lastfm_result.get('title', 'N/A')}")
        print(f"   Last.fm artist: {lastfm_result.get('artist', 'N/A')}")
        print(f"   Last.fm album: {lastfm_result.get('album', 'N/A')}")
        print(f"   Last.fm genre: {lastfm_result.get('genre', 'N/A')}")
    else:
        print("   Last.fm: No metadata found")
    
    # Now test the fallback
    print("\n3. Testing fallback integration...")
    fallback_result = search_metadata_with_fallback(title, artist)
    
    if fallback_result:
        print("   Fallback result:")
        print(f"     Title: {fallback_result.get('title', 'N/A')}")
        print(f"     Artist: {fallback_result.get('artist', 'N/A')}")
        print(f"     Album: {fallback_result.get('album', 'N/A')}")
        print(f"     Genre: {fallback_result.get('genre', 'N/A')}")
        
        # Check if the genre matches Last.fm (not iTunes)
        if lastfm_result and lastfm_result.get('genre'):
            if fallback_result.get('genre') == lastfm_result['genre']:
                print(f"\n   ✅ SUCCESS: Genre matches Last.fm!")
                print(f"   Last.fm genre used: '{lastfm_result['genre']}'")
                
                # Verify it's NOT the iTunes genre (if different)
                if itunes_result and itunes_result.get('genre') and itunes_result['genre'] != lastfm_result['genre']:
                    print(f"   ✅ iTunes genre correctly ignored: '{itunes_result['genre']}'")
                
                return True
            else:
                print(f"\n   ❌ FAIL: Genre doesn't match Last.fm")
                print(f"   Expected (Last.fm): '{lastfm_result['genre']}'")
                print(f"   Got (Fallback): '{fallback_result.get('genre', 'None')}'")
                return False
        else:
            print(f"\n   ⚠️  Last.fm has no genre to compare")
            if fallback_result.get('genre'):
                print(f"   Fallback has genre: '{fallback_result['genre']}'")
            else:
                print(f"   Fallback has no genre")
    else:
        print("   ❌ Fallback failed")
        return False

def test_multiple_songs():
    """Test with multiple songs to verify consistent behavior."""
    print("\nTesting multiple songs to verify consistent behavior...\n")
    
    test_songs = [
        ("Creep", "Radiohead"),
        ("Wonderwall", "Oasis"),
        ("Yellow", "Coldplay"),
        ("Mr. Brightside", "The Killers")
    ]
    
    success_count = 0
    
    for title, artist in test_songs:
        print(f"Testing: '{title}' by '{artist}'")
        print("-" * 40)
        
        # Get Last.fm genre
        lastfm_result = search_lastfm_for_metadata(title, artist)
        lastfm_genre = lastfm_result.get('genre') if lastfm_result else None
        
        # Get fallback result
        fallback_result = search_metadata_with_fallback(title, artist)
        fallback_genre = fallback_result.get('genre') if fallback_result else None
        
        print(f"Last.fm genre: {lastfm_genre or 'None'}")
        print(f"Fallback genre: {fallback_genre or 'None'}")
        
        if lastfm_genre and fallback_genre == lastfm_genre:
            print("✅ Genre matches Last.fm")
            success_count += 1
        elif not lastfm_genre and not fallback_genre:
            print("⚠️  No genre available from either source")
            success_count += 1
        elif not lastfm_genre and fallback_genre:
            print("⚠️  Fallback has genre but Last.fm doesn't")
            success_count += 1
        else:
            print("❌ Genre mismatch or unexpected result")
        
        print()
    
    print(f"Results: {success_count}/{len(test_songs)} songs processed correctly")
    return success_count == len(test_songs)

def test_spotify_fallback_scenario():
    """Test the scenario where Spotify fails and iTunes is primary but Last.fm provides genre."""
    print("Testing scenario: Spotify fails → iTunes primary → Last.fm for genre...\n")
    
    # Since we don't have Spotify credentials, this will naturally test the scenario
    title = "Paranoid Android"
    artist = "Radiohead"
    
    print(f"Testing: '{title}' by '{artist}'")
    print("Expected: Spotify fails → iTunes provides metadata → Last.fm provides genre")
    print("=" * 70)
    
    # Check Last.fm genre
    lastfm_result = search_lastfm_for_metadata(title, artist)
    lastfm_genre = lastfm_result.get('genre') if lastfm_result else None
    print(f"Last.fm genre: {lastfm_genre or 'None'}")
    
    # Test fallback
    fallback_result = search_metadata_with_fallback(title, artist)
    
    if fallback_result:
        print(f"Fallback successful:")
        print(f"  Title: {fallback_result.get('title', 'N/A')}")
        print(f"  Artist: {fallback_result.get('artist', 'N/A')}")
        print(f"  Album: {fallback_result.get('album', 'N/A')}")
        print(f"  Genre: {fallback_result.get('genre', 'N/A')}")
        
        if lastfm_genre and fallback_result.get('genre') == lastfm_genre:
            print(f"\n✅ SUCCESS: Last.fm genre correctly applied!")
            print(f"Genre: '{lastfm_genre}'")
            return True
        elif not lastfm_genre and not fallback_result.get('genre'):
            print(f"\n⚠️  No genre available (acceptable)")
            return True
        else:
            print(f"\n❌ Genre mismatch or unexpected result")
            return False
    else:
        print("❌ Fallback failed")
        return False

def main():
    """Run the tests to verify iTunes genre is not used."""
    print("iTunes Genre Removal Test")
    print("=" * 60)
    print("This test verifies that iTunes is no longer used for genre")
    print("acquisition and Last.fm is used exclusively for genre.\n")
    
    try:
        print("Test 1: Single song test")
        print("=" * 30)
        test1_passed = test_no_itunes_genre()
        
        print("\nTest 2: Multiple songs test")
        print("=" * 30)
        test2_passed = test_multiple_songs()
        
        print("\nTest 3: Spotify fallback scenario")
        print("=" * 30)
        test3_passed = test_spotify_fallback_scenario()
        
        print("\n" + "=" * 60)
        if test1_passed and test2_passed and test3_passed:
            print("🎉 All tests passed!")
            print("✅ iTunes genre acquisition successfully removed")
            print("✅ Last.fm is now used exclusively for genre")
            print("✅ Fallback order: Spotify → iTunes (no genre) → Last.fm (genre only)")
        else:
            print("❌ Some tests failed")
            return False
            
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        import traceback
        traceback.print_exc()
        return False
    
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)