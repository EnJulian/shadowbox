#!/usr/bin/env python3
"""
Test script for Spotify metadata search with different combinations.
"""

from meta_ops.spotify_metadata import search_spotify_for_metadata

def test_search():
    """Test different search combinations."""
    # Test case similar to the one in the user's request
    print("\n=== Test Case 1: 'Fifty' by 'MODERNS MUSIC' ===")
    metadata = search_spotify_for_metadata("Fifty", "MODERNS MUSIC")
    
    # You can add more test cases here
    
    print("\nTest completed.")

if __name__ == "__main__":
    test_search()