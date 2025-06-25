#!/usr/bin/env python3
"""
Utility script to check if lyrics have been embedded in audio files.
"""

import os
import sys
from meta_ops.lyrics import extract_lyrics_from_file

def check_lyrics_in_file(file_path):
    """
    Check if a file contains embedded lyrics and display the information.
    
    Args:
        file_path (str): Path to the audio file
    """
    if not os.path.exists(file_path):
        print(f"❌ File not found: {file_path}")
        return False
    
    # Get file extension
    file_ext = os.path.splitext(file_path)[1].lower()
    supported_formats = ['.mp3', '.opus', '.m4a']
    
    if file_ext not in supported_formats:
        print(f"❌ Unsupported file format: {file_ext}")
        print(f"   Supported formats: {', '.join(supported_formats)}")
        return False
    
    print(f"🎵 Checking: {os.path.basename(file_path)}")
    print(f"   Format: {file_ext.upper()}")
    
    # Extract lyrics
    lyrics = extract_lyrics_from_file(file_path)
    
    if lyrics:
        print(f"✅ Lyrics found! ({len(lyrics)} characters)")
        print(f"   Preview: {lyrics[:100]}{'...' if len(lyrics) > 100 else ''}")
        return True
    else:
        print("❌ No lyrics found in file")
        return False

def main():
    """Main function to handle command line arguments."""
    if len(sys.argv) < 2:
        print("Usage: python3 check_lyrics.py <audio_file_or_directory>")
        print("Examples:")
        print("  python3 check_lyrics.py song.mp3")
        print("  python3 check_lyrics.py ~/Music/")
        sys.exit(1)
    
    path = sys.argv[1]
    
    if os.path.isfile(path):
        # Single file
        check_lyrics_in_file(path)
    elif os.path.isdir(path):
        # Directory - check all audio files
        audio_files = []
        for root, dirs, files in os.walk(path):
            for file in files:
                if file.lower().endswith(('.mp3', '.opus', '.m4a')):
                    audio_files.append(os.path.join(root, file))
        
        if not audio_files:
            print(f"❌ No audio files found in directory: {path}")
            sys.exit(1)
        
        print(f"Found {len(audio_files)} audio files to check:\n")
        
        files_with_lyrics = 0
        for file_path in sorted(audio_files):
            if check_lyrics_in_file(file_path):
                files_with_lyrics += 1
            print()  # Empty line for readability
        
        print(f"📊 Summary: {files_with_lyrics}/{len(audio_files)} files have embedded lyrics")
    else:
        print(f"❌ Path not found: {path}")
        sys.exit(1)

if __name__ == "__main__":
    main()