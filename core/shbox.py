#!/usr/bin/env python3
"""
Shadowbox Music Downloader - Interactive Mode

A user-friendly interactive interface for the Shadowbox Music Downloader application.
"""

import os
import sys
import time
import glob
from main import run, run_with_spotify, sanitize_filename

def clear_screen():
    """Clear the terminal screen."""
    os.system('cls' if os.name == 'nt' else 'clear')

def print_header():
    """Print the application header."""
    clear_screen()
    print("=" * 70)
    print("\033[32m[SYSTEM]\033[0m  SHADOWBOX  \033[32m[SYSTEM]\033[0m")
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
    config_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.shadowbox_config')
    if os.path.exists(config_file):
        with open(config_file, 'r') as f:
            config_lines = f.readlines()
            for line in config_lines:
                if line.startswith('music_dir='):
                    saved_dir = line.replace('music_dir=', '').strip()
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

def get_use_spotify():
    """Get whether to use Spotify for metadata."""
    # Check if there's a saved preference
    config_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.shadowbox_config')
    if os.path.exists(config_file):
        with open(config_file, 'r') as f:
            config_lines = f.readlines()
            for line in config_lines:
                if line.startswith('use_spotify='):
                    return line.replace('use_spotify=', '').strip().lower() == 'true'
    
    # Default is False
    return False

def get_audio_format():
    """Get the preferred audio format."""
    # Check if there's a saved preference
    config_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.shadowbox_config')
    if os.path.exists(config_file):
        with open(config_file, 'r') as f:
            config_lines = f.readlines()
            for line in config_lines:
                if line.startswith('audio_format='):
                    format_value = line.replace('audio_format=', '').strip()
                    if format_value:
                        return format_value
    
    # Default is opus
    return 'opus'

def save_config(music_dir=None, use_spotify=None, audio_format=None):
    """Save the configuration preferences."""
    config_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), '.shadowbox_config')
    
    # Read existing config
    config = {}
    if os.path.exists(config_file):
        with open(config_file, 'r') as f:
            config_lines = f.readlines()
            for line in config_lines:
                if '=' in line:
                    key, value = line.strip().split('=', 1)
                    config[key] = value
    
    # Update config
    if music_dir is not None:
        config['music_dir'] = music_dir
    if use_spotify is not None:
        config['use_spotify'] = 'true' if use_spotify else 'false'
    if audio_format is not None:
        config['audio_format'] = audio_format
    
    # Write config
    with open(config_file, 'w') as f:
        for key, value in config.items():
            f.write(f"{key}={value}\n")

def save_music_directory(directory):
    """Save the music directory preference."""
    save_config(music_dir=directory)

def save_audio_format(format_name):
    """Save the audio format preference."""
    save_config(audio_format=format_name)

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
    use_spotify = get_use_spotify()
    audio_format = get_audio_format()
    
    print(f"\n\033[32m[DIR]\033[0m Music will be saved to: {music_dir}")
    print(f"\033[32m[FORMAT]\033[0m Using audio format: {audio_format}")
    if use_spotify:
        print("\033[32m[SPOTIFY]\033[0m Using Spotify for metadata")
    print("\nDownloading... (this may take a minute)")
    print()
    
    if use_spotify:
        success = run_with_spotify(query, music_dir=music_dir, audio_format=audio_format)
    else:
        success = run(query, music_dir=music_dir, audio_format=audio_format)
    
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
    use_spotify = get_use_spotify()
    audio_format = get_audio_format()
    
    print(f"\n\033[32m[DIR]\033[0m Music will be saved to: {music_dir}")
    print(f"\033[32m[FORMAT]\033[0m Using audio format: {audio_format}")
    if use_spotify:
        print("\033[32m[SPOTIFY]\033[0m Using Spotify for metadata")
    print("\nDownloading... (this may take a minute)")
    print()
    
    if use_spotify:
        success = run_with_spotify(url, music_dir=music_dir, audio_format=audio_format)
    else:
        success = run(url, music_dir=music_dir, audio_format=audio_format)
    
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
    use_spotify = get_use_spotify()
    audio_format = get_audio_format()
    
    print(f"\n\033[32m[DIR]\033[0m Music will be saved to: {music_dir}")
    print(f"\033[32m[FORMAT]\033[0m Using audio format: {audio_format}")
    if use_spotify:
        print("\033[32m[SPOTIFY]\033[0m Using Spotify for metadata")
    print(f"\nDownloading {len(songs)} songs... (this may take a while)")
    print()
    
    successful = 0
    for i, song in enumerate(songs):
        print(f"\n[{i+1}/{len(songs)}] Downloading: {song}")
        if use_spotify:
            success = run_with_spotify(song, music_dir=music_dir, audio_format=audio_format)
        else:
            success = run(song, music_dir=music_dir, audio_format=audio_format)
            
        if success:
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
        use_spotify = get_use_spotify()
        audio_format = get_audio_format()
        
        print(f"\nCurrent music directory: {current_dir}")
        print(f"Use Spotify for metadata: {'Yes' if use_spotify else 'No'}")
        print(f"Audio format: {audio_format}")
        
        print("\n1. Change music directory")
        print("2. Toggle Spotify metadata")
        print("3. Change audio format")
        print("4. Back to main menu")
        print()
        
        choice = input("Select an option (1-4): ").strip()
        
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
            # Check if Spotify credentials are set
            spotify_client_id = os.environ.get('SPOTIFY_CLIENT_ID')
            spotify_client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')
            
            if not spotify_client_id or not spotify_client_secret:
                print("\n\033[33m[WARNING]\033[0m Spotify credentials not found.")
                print("You need to set up Spotify credentials first.")
                print("Run ./setup/setup_spotify.sh (Linux/macOS) or setup\\setup_spotify.bat (Windows)")
                input("\nPress Enter to continue...")
                continue
            
            # Toggle Spotify setting
            new_setting = not use_spotify
            save_config(use_spotify=new_setting)
            
            if new_setting:
                print("\n\033[32m[SUCCESS]\033[0m Spotify metadata enabled")
            else:
                print("\n\033[32m[SUCCESS]\033[0m Spotify metadata disabled")
            
            input("\nPress Enter to continue...")
        
        elif choice == '3':
            print_header()
            print("\033[32m[FORMAT]\033[0m CHANGE AUDIO FORMAT")
            print("=" * 60)
            
            print(f"\nCurrent audio format: {audio_format}")
            print("\nAvailable audio formats:")
            formats = [
                ('alac', 'Apple lossless Audio Codec, highest quality (recommended)'),
                ('opus', 'High quality, small file size'),
                ('m4a', 'Good quality, compatible with Apple devices'),
                ('mp3', 'Universal compatibility, decent quality'),
                ('flac', 'Lossless audio, large file size'),
                ('wav', 'Uncompressed audio, very large file size'),
                ('aac', 'Good quality, small file size')
            ]
            
            for i, (fmt, desc) in enumerate(formats, 1):
                print(f"{i}. {fmt} - {desc}")
            
            print("7. Custom format")
            print()
            
            format_choice = input("Select an option (1-7) or press Enter to keep current: ").strip()
            
            if not format_choice:
                continue
                
            try:
                choice_num = int(format_choice)
                if 1 <= choice_num <= 6:
                    # Standard format
                    new_format = formats[choice_num - 1][0]
                    save_audio_format(new_format)
                    print(f"\n\033[32m[SUCCESS]\033[0m Audio format changed to: {new_format}")
                elif choice_num == 7:
                    # Custom format
                    custom_format = input("\nEnter custom audio format: ").strip()
                    if custom_format:
                        save_audio_format(custom_format)
                        print(f"\n\033[32m[SUCCESS]\033[0m Audio format changed to: {custom_format}")
                else:
                    print("\n\033[31m[ERROR]\033[0m Invalid choice")
            except ValueError:
                print("\n\033[31m[ERROR]\033[0m Please enter a number")
            
            input("\nPress Enter to continue...")
        
        elif choice == '4':
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
        songs = []
        for ext in ['alac', 'opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']:
            songs.extend(glob.glob(os.path.join(artist_path, f"*.{ext}")))
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
    
    # Get all songs with common audio formats
    songs = []
    for ext in ['opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']:
        songs.extend(glob.glob(os.path.join(artist_path, f"*.{ext}")))
    
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
            # Get filename without extension and the extension
            filename = os.path.splitext(os.path.basename(song))[0]
            ext = os.path.splitext(song)[1][1:]  # Remove the dot
            print(f"{i+1}. {filename} ({ext.upper()}, {size_mb:.1f} MB, {mod_time})")
    
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