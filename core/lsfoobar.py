#!/usr/bin/env python3
"""
lsnen's Foobar2000 Music Downloader - Interactive Mode

A user-friendly interactive interface for the lsnen's Foobar2000 Music Downloader application.
"""

import os
import sys
import time
import glob
from main import run, sanitize_filename

def clear_screen():
    """Clear the terminal screen."""
    os.system('cls' if os.name == 'nt' else 'clear')

def print_header():
    """Print the application header."""
    clear_screen()
    print("=" * 70)
    print("\033[32m[SYSTEM]\033[0m  LSNEN'S FOOBAR2000 MUSIC DOWNLOADER  \033[32m[SYSTEM]\033[0m")
    print("=" * 70)
    print("Download music from YouTube or Bandcamp with metadata and cover art")
    print("=" * 70)
    print()

def print_menu():
    """Print the main menu."""
    print("\nMAIN MENU:")
    print("1. Search and download a song")
    print("2. Download from URL (YouTube or Bandcamp)")
    print("3. Batch download from a list")
    print("4. Settings")
    print("5. View downloaded songs")
    print("6. Exit")
    print()

def get_music_directory():
    """Get or create the music directory."""
    # Default is ~/Music
    default_dir = os.path.join(os.path.expanduser('~'), 'Music')
    
    # Check if there's a saved preference
    config_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.lsnen_foobar_config')
    if os.path.exists(config_file):
        with open(config_file, 'r') as f:
            saved_dir = f.read().strip()
            if saved_dir and os.path.exists(saved_dir):
                return saved_dir
    
    # If no saved preference or it doesn't exist, use default
    if not os.path.exists(default_dir):
        try:
            os.makedirs(default_dir, exist_ok=True)
        except:
            # If can't create default, use current directory
            default_dir = os.path.dirname(os.path.abspath(__file__))
    
    return default_dir

def save_music_directory(directory):
    """Save the music directory preference."""
    config_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.lsnen_foobar_config')
    with open(config_file, 'w') as f:
        f.write(directory)

def search_and_download():
    """Search for and download a song."""
    print_header()
    print("\033[37m[SCAN]\033[0m SEARCH AND DOWNLOAD")
    print("=" * 60)
    print("\nEnter the song details (e.g., 'Imagine Dragons Believer')")
    print("Or type 'back' to return to the main menu")
    print()
    
    query = input("\033[32m[INPUT]\033[0m Song to search: ").strip()
    if query.lower() == 'back':
        return
    
    if not query:
        print("\n\033[31m[ERROR]\033[0m Please enter a search query")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    print(f"\n\033[32m[DIR]\033[0m Music will be saved to: {music_dir}")
    print("\nDownloading... (this may take a minute)")
    print()
    
    success = run(query, music_dir=music_dir)
    
    if success:
        print("\n\033[32m[SUCCESS]\033[0m Operation Success!")
    else:
        print("\n\033[31m[FAIL]\033[0m Download failed. Please try again.")
    
    input("\nPress Enter to continue...")

def download_from_url():
    """Download a song from a URL."""
    print_header()
    print("\033[32m[URL]\033[0m DOWNLOAD FROM URL")
    print("=" * 60)
    print("\nEnter a YouTube or Bandcamp URL")
    print("Or type 'back' to return to the main menu")
    print()
    
    url = input("\033[32m[INPUT]\033[0m URL: ").strip()
    if url.lower() == 'back':
        return
    
    if not url.startswith(('http://', 'https://')):
        print("\n\033[31m[ERROR]\033[0m Please enter a valid URL (starting with http:// or https://)")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    print(f"\n\033[32m[DIR]\033[0m Music will be saved to: {music_dir}")
    print("\nDownloading... (this may take a minute)")
    print()
    
    success = run(url, music_dir=music_dir)
    
    if success:
        print("\n\033[32m[SUCCESS]\033[0m Operation Success!")
    else:
        print("\n\033[31m[FAIL]\033[0m Download failed. Please try again.")
    
    input("\nPress Enter to continue...")

def batch_download():
    """Download multiple songs from a list."""
    print_header()
    print("\033[32m[BATCH]\033[0m BATCH DOWNLOAD")
    print("=" * 60)
    print("\nEnter one song or URL per line.")
    print("When finished, enter an empty line.")
    print("Or type 'back' to return to the main menu")
    print()
    
    songs = []
    while True:
        line = input(f"Song/URL #{len(songs) + 1} (or empty to finish): ").strip()
        if not line:
            break
        if line.lower() == 'back':
            return
        songs.append(line)
    
    if not songs:
        print("\n\033[31m[ERROR]\033[0m No songs entered")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    print(f"\n\033[32m[DIR]\033[0m Music will be saved to: {music_dir}")
    print(f"\nDownloading {len(songs)} songs... (this may take a while)")
    print()
    
    successful = 0
    for i, song in enumerate(songs):
        print(f"\n[{i+1}/{len(songs)}] Downloading: {song}")
        if run(song, music_dir=music_dir):
            successful += 1
            print(f"\033[32m[SUCCESS]\033[0m Downloaded: {song}")
        else:
            print(f"\033[31m[FAIL]\033[0m Failed: {song}")
    
    print(f"\n\033[32m[COMPLETE]\033[0m Downloaded {successful} out of {len(songs)} songs")
    input("\nPress Enter to continue...")

def settings():
    """Change application settings."""
    while True:
        print_header()
        print("\033[32m[CONFIG]\033[0m  SETTINGS")
        print("=" * 60)
        
        current_dir = get_music_directory()
        print(f"\nCurrent music directory: {current_dir}")
        print("\n1. Change music directory")
        print("2. Back to main menu")
        print()
        
        choice = input("Select an option (1-2): ").strip()
        
        if choice == '1':
            print("\nEnter the new music directory path")
            print("(Use ~ for home directory on macOS/Linux)")
            print("Or press Enter to keep the current directory")
            
            new_dir = input("\nNew directory: ").strip()
            if new_dir:
                # Expand ~ to home directory
                if new_dir.startswith('~'):
                    new_dir = os.path.expanduser(new_dir)
                
                # Create directory if it doesn't exist
                if not os.path.exists(new_dir):
                    try:
                        os.makedirs(new_dir, exist_ok=True)
                        print(f"\n\033[32m[SUCCESS]\033[0m Created directory: {new_dir}")
                    except Exception as e:
                        print(f"\n\033[31m[ERROR]\033[0m Error creating directory: {e}")
                        input("\nPress Enter to continue...")
                        continue
                
                # Save the new directory
                save_music_directory(new_dir)
                print(f"\n\033[32m[SUCCESS]\033[0m Music directory changed to: {new_dir}")
                input("\nPress Enter to continue...")
        
        elif choice == '2':
            break
        
        else:
            print("\n\033[31m[ERROR]\033[0m Invalid choice. Please try again.")
            time.sleep(1)

def view_downloads():
    """View downloaded songs."""
    print_header()
    print("\033[32m[LIST]\033[0m VIEW DOWNLOADED SONGS")
    print("=" * 60)
    
    music_dir = get_music_directory()
    print(f"\nMusic directory: {music_dir}")
    
    # Get all artist directories
    artist_dirs = [d for d in os.listdir(music_dir) 
                  if os.path.isdir(os.path.join(music_dir, d))]
    
    if not artist_dirs:
        print("\nNo music found. Download some songs first!")
        input("\nPress Enter to continue...")
        return
    
    # Sort alphabetically
    artist_dirs.sort()
    
    print(f"\nFound {len(artist_dirs)} artists:")
    for i, artist in enumerate(artist_dirs):
        # Count songs for this artist
        artist_path = os.path.join(music_dir, artist)
        songs = glob.glob(os.path.join(artist_path, "*.opus")) + \
                glob.glob(os.path.join(artist_path, "*.mp3"))
        print(f"{i+1}. {artist} ({len(songs)} songs)")
    
    print("\nEnter an artist number to see their songs, or 'back' to return")
    choice = input("\nSelect artist: ").strip()
    
    if choice.lower() == 'back':
        return
    
    try:
        idx = int(choice) - 1
        if 0 <= idx < len(artist_dirs):
            view_artist_songs(music_dir, artist_dirs[idx])
        else:
            print("\n\033[31m[ERROR]\033[0m Invalid choice")
            time.sleep(1)
    except ValueError:
        print("\n\033[31m[ERROR]\033[0m Please enter a number")
        time.sleep(1)

def view_artist_songs(music_dir, artist):
    """View songs for a specific artist."""
    artist_path = os.path.join(music_dir, artist)
    
    # Get all songs
    songs = glob.glob(os.path.join(artist_path, "*.opus")) + \
            glob.glob(os.path.join(artist_path, "*.mp3"))
    
    # Sort by modification time (newest first)
    songs.sort(key=os.path.getmtime, reverse=True)
    
    print_header()
    print(f"\033[32m[ARTIST]\033[0m SONGS BY {artist.upper()}")
    print("=" * 60)
    
    if not songs:
        print("\nNo songs found for this artist.")
    else:
        print(f"\nFound {len(songs)} songs:")
        for i, song in enumerate(songs):
            # Get file size in MB
            size_mb = os.path.getsize(song) / (1024 * 1024)
            # Get modification time
            mod_time = time.strftime("%Y-%m-%d %H:%M", time.localtime(os.path.getmtime(song)))
            # Get filename without extension
            filename = os.path.splitext(os.path.basename(song))[0]
            print(f"{i+1}. {filename} ({size_mb:.1f} MB, {mod_time})")
    
    input("\nPress Enter to continue...")

def main():
    """Main application loop."""
    while True:
        print_header()
        print_menu()
        
        choice = input("Select an option (1-6): ").strip()
        
        if choice == '1':
            search_and_download()
        elif choice == '2':
            download_from_url()
        elif choice == '3':
            batch_download()
        elif choice == '4':
            settings()
        elif choice == '5':
            view_downloads()
        elif choice == '6':
            print("\n\033[32m[SYSTEM]\033[0m Application shut down.")
            sys.exit(0)
        else:
            print("\n\033[31m[ERROR]\033[0m Invalid choice. Please try again.")
            time.sleep(1)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n\033[31m[INTERRUPT]\033[0m Program interrupted. Exiting...")
        sys.exit(0)