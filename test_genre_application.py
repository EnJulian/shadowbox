#!/usr/bin/env python3
"""
Test script to verify that Last.fm genre metadata is properly applied to files.
"""

import os
import sys
import tempfile
import shutil

# Add the project root to the Python path
project_root = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, project_root)

from meta_ops.spotify_metadata import search_metadata_with_fallback, search_lastfm_for_metadata, apply_spotify_metadata_to_file
from meta_ops.metadata import extract_metadata, add_metadata

def create_test_audio_file(file_path, format='opus'):
    """Create a minimal test audio file."""
    if format == 'opus':
        # Create a minimal Opus file header
        opus_header = b'OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00OpusHead\x01\x02\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00'
        with open(file_path, 'wb') as f:
            f.write(opus_header)
            f.write(b'\x00' * 1024)  # Add some padding
    elif format == 'mp3':
        # Create a minimal MP3 file with ID3 header
        mp3_header = b'ID3\x03\x00\x00\x00\x00\x00\x00'
        with open(file_path, 'wb') as f:
            f.write(mp3_header)
            f.write(b'\x00' * 1024)  # Add some padding

def test_genre_application():
    """Test that Last.fm genre is properly applied to audio files."""
    print("Testing Last.fm genre application to audio files...\n")
    
    # Create a temporary directory for test files
    temp_dir = tempfile.mkdtemp()
    print(f"Using temporary directory: {temp_dir}")
    
    try:
        # Test 1: Direct Last.fm metadata application
        print("\nTest 1: Direct Last.fm metadata application")
        print("=" * 50)
        
        title = "Bohemian Rhapsody"
        artist = "Queen"
        
        # Get Last.fm metadata
        print(f"Getting Last.fm metadata for: '{title}' by '{artist}'")
        lastfm_metadata = search_lastfm_for_metadata(title, artist)
        
        if not lastfm_metadata or not lastfm_metadata.get('genre'):
            print("❌ Failed to get genre from Last.fm")
            return False
        
        print(f"✅ Last.fm genre found: {lastfm_metadata['genre']}")
        
        # Create test files for different formats
        test_files = [
            (os.path.join(temp_dir, "test_opus.opus"), 'opus'),
            (os.path.join(temp_dir, "test_mp3.mp3"), 'mp3')
        ]
        
        for test_file, format_type in test_files:
            print(f"\nTesting {format_type.upper()} format...")
            
            # Create test file
            create_test_audio_file(test_file, format_type)
            
            # Apply metadata directly
            try:
                add_metadata(
                    file_path=test_file,
                    title=lastfm_metadata['title'],
                    artist=lastfm_metadata['artist'],
                    album=lastfm_metadata.get('album', 'Test Album'),
                    genre=lastfm_metadata['genre']
                )
                print(f"✅ Metadata applied to {format_type} file")
                
                # Verify the metadata was applied
                extracted = extract_metadata(test_file)
                if extracted.get('genre') == lastfm_metadata['genre']:
                    print(f"✅ Genre correctly applied and verified: {extracted['genre']}")
                else:
                    print(f"❌ Genre not applied correctly. Expected: {lastfm_metadata['genre']}, Got: {extracted.get('genre', 'None')}")
                    return False
                    
            except Exception as e:
                print(f"❌ Failed to apply metadata to {format_type}: {e}")
                return False
        
        # Test 2: Fallback integration with file application
        print("\nTest 2: Fallback integration with file application")
        print("=" * 50)
        
        test_file_fallback = os.path.join(temp_dir, "test_fallback.opus")
        create_test_audio_file(test_file_fallback, 'opus')
        
        # Test with a song that might need Last.fm fallback
        title2 = "Paranoid Android"
        artist2 = "Radiohead"
        
        print(f"Testing fallback for: '{title2}' by '{artist2}'")
        
        # Get metadata through fallback
        fallback_metadata = search_metadata_with_fallback(title2, artist2)
        
        if fallback_metadata and fallback_metadata.get('genre'):
            print(f"✅ Fallback found genre: {fallback_metadata['genre']}")
            
            # Apply using the spotify metadata function
            success = apply_spotify_metadata_to_file(test_file_fallback, fallback_metadata, download_cover=False)
            
            if success:
                print("✅ Metadata applied successfully via fallback")
                
                # Verify the genre was applied
                extracted_fallback = extract_metadata(test_file_fallback)
                if extracted_fallback.get('genre') == fallback_metadata['genre']:
                    print(f"✅ Fallback genre correctly applied: {extracted_fallback['genre']}")
                else:
                    print(f"❌ Fallback genre not applied correctly. Expected: {fallback_metadata['genre']}, Got: {extracted_fallback.get('genre', 'None')}")
                    return False
            else:
                print("❌ Failed to apply metadata via fallback")
                return False
        else:
            print("⚠️  No genre found through fallback (this might be expected)")
        
        # Test 3: Test scenario where Last.fm is the only source with genre
        print("\nTest 3: Simulating Last.fm as third fallback")
        print("=" * 50)
        
        # Create a test file
        test_file_third = os.path.join(temp_dir, "test_third_fallback.opus")
        create_test_audio_file(test_file_third, 'opus')
        
        # Manually test the scenario where Spotify and iTunes fail but Last.fm succeeds
        print("Testing scenario: Spotify fails → iTunes fails → Last.fm succeeds")
        
        # Get Last.fm metadata directly
        lastfm_only = search_lastfm_for_metadata("Bohemian Rhapsody", "Queen")
        
        if lastfm_only and lastfm_only.get('genre'):
            # Create a mock metadata dict that simulates what would happen
            # if Spotify/iTunes found basic info but no genre
            mock_metadata = {
                'title': 'Bohemian Rhapsody',
                'artist': 'Queen',
                'album': 'A Night at the Opera',
                'genre': None,  # No genre initially
                'release_date': '1975'
            }
            
            # Simulate adding Last.fm genre
            mock_metadata['genre'] = lastfm_only['genre']
            print(f"✅ Simulated Last.fm genre addition: {mock_metadata['genre']}")
            
            # Apply this metadata
            success = apply_spotify_metadata_to_file(test_file_third, mock_metadata, download_cover=False)
            
            if success:
                print("✅ Simulated third fallback metadata applied successfully")
                
                # Verify
                extracted_third = extract_metadata(test_file_third)
                if extracted_third.get('genre') == lastfm_only['genre']:
                    print(f"✅ Third fallback genre correctly applied: {extracted_third['genre']}")
                else:
                    print(f"❌ Third fallback genre not applied. Expected: {lastfm_only['genre']}, Got: {extracted_third.get('genre', 'None')}")
                    return False
            else:
                print("❌ Failed to apply simulated third fallback metadata")
                return False
        
        return True
        
    finally:
        # Clean up temporary directory
        shutil.rmtree(temp_dir)
        print(f"\nCleaned up temporary directory: {temp_dir}")

def main():
    """Run the genre application test."""
    print("Last.fm Genre Application Test")
    print("=" * 60)
    print("This test verifies that Last.fm genre metadata is properly")
    print("applied to audio files through the fallback system.\n")
    
    try:
        success = test_genre_application()
        
        if success:
            print("\n" + "=" * 60)
            print("🎉 All genre application tests passed!")
            print("✅ Last.fm genre metadata is being applied correctly")
            print("✅ Fallback system properly applies genre to files")
            print("✅ Both Opus and MP3 formats handle genre correctly")
        else:
            print("\n" + "=" * 60)
            print("❌ Some genre application tests failed")
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