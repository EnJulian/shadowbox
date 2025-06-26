#!/usr/bin/env python3
"""
Test script to verify Last.fm returns at least 2 genres.
"""

import sys
import os
sys.path.insert(0, '/Users/julian/PycharmProjects/shadowbox')

from meta_ops.spotify_metadata import search_lastfm_for_metadata

def test_lastfm_genres():
    """Test that Last.fm returns at least 2 genres when available."""
    print("Testing Last.fm genre retrieval...\n")
    
    test_cases = [
        # Classic rock tracks
        ("Bohemian Rhapsody", "Queen"),
        ("Stairway to Heaven", "Led Zeppelin"),
        ("Hotel California", "Eagles"),
        ("Sweet Child O' Mine", "Guns N' Roses"),
        ("Purple Haze", "Jimi Hendrix"),
        
        # Different genres
        ("Smells Like Teen Spirit", "Nirvana"),  # Grunge
        ("Billie Jean", "Michael Jackson"),      # Pop
        ("Like a Rolling Stone", "Bob Dylan"),   # Folk
        ("Imagine", "John Lennon"),              # Singer-songwriter
        ("Hey Jude", "The Beatles"),             # Pop rock
        
        # More diverse genres
        ("What's Going On", "Marvin Gaye"),      # Soul/R&B
        ("Respect", "Aretha Franklin"),          # Soul
        ("Good Vibrations", "The Beach Boys"),   # Pop/Surf rock
        ("Born to Run", "Bruce Springsteen"),   # Rock
        ("Superstition", "Stevie Wonder")        # Funk/Soul
    ]
    
    total_tests = len(test_cases)
    successful_two_genres = 0
    successful_one_genre = 0
    failed_tests = 0
    
    for title, artist in test_cases:
        print(f"Testing: '{title}' by '{artist}'")
        result = search_lastfm_for_metadata(title, artist)
        
        if result and result.get('genre'):
            genre = result['genre']
            genre_count = len(genre.split(', ')) if genre else 0
            print(f"  ✅ Found {genre_count} genre(s): {genre}")
            
            if genre_count >= 2:
                successful_two_genres += 1
            elif genre_count == 1:
                successful_one_genre += 1
                print(f"  ⚠️  Only 1 genre found, expected at least 2")
        else:
            print(f"  ❌ No genre found")
            failed_tests += 1
        
        print()
    
    # Print summary
    print("=" * 60)
    print("SUMMARY:")
    print(f"Total tests: {total_tests}")
    print(f"✅ Tests with 2+ genres: {successful_two_genres} ({successful_two_genres/total_tests*100:.1f}%)")
    print(f"⚠️  Tests with 1 genre: {successful_one_genre} ({successful_one_genre/total_tests*100:.1f}%)")
    print(f"❌ Failed tests: {failed_tests} ({failed_tests/total_tests*100:.1f}%)")
    
    if successful_two_genres >= total_tests * 0.8:  # 80% success rate for 2+ genres
        print("\n🎉 EXCELLENT: Last.fm is successfully providing at least 2 genres for most tracks!")
    elif successful_two_genres + successful_one_genre >= total_tests * 0.9:  # 90% success rate for any genre
        print("\n✅ GOOD: Last.fm is providing genre information for most tracks.")
    else:
        print("\n⚠️  NEEDS IMPROVEMENT: Last.fm genre coverage could be better.")

if __name__ == "__main__":
    test_lastfm_genres()