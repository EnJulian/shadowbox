#!/usr/bin/env python3
"""
Shadowbox Music Downloader - Interactive Mode

A user-friendly interactive interface for the Shadowbox Music Downloader application.
Enhanced with hacker-style terminal UI for maximum readability and style.
"""

import os
import sys
import time
import glob
from main import run, run_with_spotify, run_playlist, run_playlist_with_spotify, sanitize_filename
from downloader import is_youtube_playlist
from terminal_ui import ui, clear_screen, print_header, print_menu, success, error, warning, info, system, exit_animation

# Enhanced terminal UI functions are imported from terminal_ui module

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
    ui.hacker_banner("AUDIO SEARCH & ACQUISITION")
    ui.section_divider()
    
    info("Enter the song details (e.g., 'Imagine Dragons Believer')")
    info("Type 'back' to return to the main menu")
    print()
    
    query = ui.input_prompt("Song to search", "SEARCH").strip()
    if query.lower() == 'back':
        return
    
    if not query:
        error("Please enter a search query")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    use_spotify = get_use_spotify()
    audio_format = get_audio_format()
    
    ui.directory(f"Music will be saved to: {music_dir}")
    system(f"Using audio format: {audio_format}")
    if use_spotify:
        ui.api("Using Spotify for metadata enhancement")
    
    ui.loading_spinner("Initializing download sequence", 1.0)
    
    if use_spotify:
        success = run_with_spotify(query, music_dir=music_dir, audio_format=audio_format)
    else:
        success = run(query, music_dir=music_dir, audio_format=audio_format)
    
    if success:
        success("Operation completed successfully!", "MISSION_SUCCESS")
    else:
        error("Download failed. Please try again.", "MISSION_FAILED")
    
    ui.input_prompt("Press Enter to continue", "CONTINUE")

def download_from_url():
    """Download a song from a URL."""
    print_header()
    ui.hacker_banner("URL EXTRACTION")
    ui.section_divider()
    
    ui.info("Enter a YouTube or Bandcamp URL")
    ui.info("Type 'back' to return to the main menu")
    print()
    
    url = ui.input_prompt("Target URL", "URL").strip()
    if url.lower() == 'back':
        return
    
    if not url.startswith(('http://', 'https://')):
        error("Please enter a valid URL (starting with http:// or https://)")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    use_spotify = get_use_spotify()
    audio_format = get_audio_format()
    
    ui.directory(f"Music will be saved to: {music_dir}")
    system(f"Using audio format: {audio_format}")
    if use_spotify:
        ui.api("Using Spotify for metadata enhancement")
    
    ui.loading_spinner("Initiating URL extraction", 1.0)
    
    if use_spotify:
        success = run_with_spotify(url, music_dir=music_dir, audio_format=audio_format)
    else:
        success = run(url, music_dir=music_dir, audio_format=audio_format)
    
    if success:
        success("URL extraction completed successfully!", "EXTRACTION_SUCCESS")
    else:
        error("URL extraction failed. Please try again.", "EXTRACTION_FAILED")
    
    ui.input_prompt("Press Enter to continue", "CONTINUE")

def batch_download():
    """Download multiple songs from a list."""
    print_header()
    ui.hacker_banner("BATCH PROCESSING")
    ui.section_divider()
    
    ui.info("Enter one song or URL per line")
    ui.info("When finished, enter an empty line")
    ui.info("Type 'back' to return to the main menu")
    print()
    
    songs = []
    while True:
        line = ui.input_prompt(f"Song/URL #{len(songs) + 1} (or empty to finish)", "BATCH").strip()
        if not line:
            break
        if line.lower() == 'back':
            return
        songs.append(line)
    
    if not songs:
        error("No songs entered")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    use_spotify = get_use_spotify()
    audio_format = get_audio_format()
    
    ui.directory(f"Music will be saved to: {music_dir}")
    system(f"Using audio format: {audio_format}")
    if use_spotify:
        ui.api("Using Spotify for metadata enhancement")
    
    ui.loading_spinner(f"Initializing batch processing for {len(songs)} targets", 1.5)
    
    successful = 0
    for i, song in enumerate(songs):
        ui.progress_bar(i, len(songs), f"Processing batch", f"({i}/{len(songs)})")
        ui.download(f"Target: {song}", f"BATCH_{i+1}")
        
        if use_spotify:
            success = run_with_spotify(song, music_dir=music_dir, audio_format=audio_format)
        else:
            success = run(song, music_dir=music_dir, audio_format=audio_format)
            
        if success:
            successful += 1
            success(f"Acquired: {song}", f"TARGET_{i+1}")
        else:
            error(f"Failed: {song}", f"TARGET_{i+1}")
    
    ui.progress_bar(len(songs), len(songs), "Batch processing", "Complete")
    success(f"Batch completed: {successful}/{len(songs)} targets acquired", "BATCH_COMPLETE")
    ui.input_prompt("Press Enter to continue", "CONTINUE")

def settings():
    """Change application settings."""
    while True:
        print_header()
        
        current_dir = get_music_directory()
        use_spotify = get_use_spotify()
        audio_format = get_audio_format()
        
        # Use the new enhanced settings menu
        ui.print_settings_menu(current_dir, use_spotify, audio_format)
        
        choice = ui.input_prompt("Select an option (1-5)", "SETTINGS").strip()
        
        if choice == '1':
            ui.info("Enter the new music directory path")
            ui.info("(Use ~ for home directory on macOS/Linux)")
            ui.info("Or press Enter to keep the current directory")
            
            new_dir = ui.input_prompt("New directory", "DIR_PATH").strip()
            if new_dir:
                # Expand ~ to home directory
                if new_dir.startswith('~'):
                    new_dir = os.path.expanduser(new_dir)
                
                # Create directory if it doesn't exist
                if not os.path.exists(new_dir):
                    try:
                        os.makedirs(new_dir, exist_ok=True)
                        ui.success(f"Created directory: {new_dir}")
                    except Exception as e:
                        ui.error(f"Error creating directory: {e}")
                        ui.input_prompt("Press Enter to continue", "CONTINUE")
                        continue
                
                # Save the new directory
                save_music_directory(new_dir)
                ui.success(f"Music directory changed to: {new_dir}")
                ui.input_prompt("Press Enter to continue", "CONTINUE")
        
        elif choice == '2':
            # Check if Spotify credentials are set
            spotify_client_id = os.environ.get('SPOTIFY_CLIENT_ID')
            spotify_client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')
            
            if not spotify_client_id or not spotify_client_secret:
                ui.warning("Spotify credentials not found.")
                ui.info("You need to set up Spotify credentials first.")
                ui.info("Run ./setup/setup_spotify.sh (Linux/macOS) or setup\\setup_spotify.bat (Windows)")
                ui.input_prompt("Press Enter to continue", "CONTINUE")
                continue
            
            # Toggle Spotify setting
            new_setting = not use_spotify
            save_config(use_spotify=new_setting)
            
            if new_setting:
                ui.success("Spotify metadata enabled")
            else:
                ui.success("Spotify metadata disabled")
            
            ui.input_prompt("Press Enter to continue", "CONTINUE")
        
        elif choice == '3':
            print_header()
            
            # Use the new enhanced audio format menu
            ui.print_audio_format_menu(audio_format)
            
            format_choice = ui.input_prompt("Select an option (1-9)", "FORMAT").strip()
            
            if not format_choice:
                continue
                
            try:
                choice_num = int(format_choice)
                formats = [
                    ('alac', 'Apple lossless Audio Codec, highest quality (recommended)'),
                    ('opus', 'High quality, small file size'),
                    ('m4a', 'Good quality, compatible with Apple devices'),
                    ('mp3', 'Universal compatibility, decent quality'),
                    ('flac', 'Lossless audio, large file size'),
                    ('wav', 'Uncompressed audio, very large file size'),
                    ('aac', 'Good quality, small file size')
                ]
                
                if 1 <= choice_num <= 7:
                    # Standard format
                    new_format = formats[choice_num - 1][0]
                    save_audio_format(new_format)
                    ui.success(f"Audio format changed to: {new_format}")
                elif choice_num == 8:
                    # Custom format
                    custom_format = ui.input_prompt("Enter custom audio format", "CUSTOM").strip()
                    if custom_format:
                        save_audio_format(custom_format)
                        ui.success(f"Audio format changed to: {custom_format}")
                elif choice_num == 9:
                    # Keep current format
                    continue
                else:
                    ui.error("Invalid choice")
            except ValueError:
                ui.error("Please enter a number")
            
            ui.input_prompt("Press Enter to continue", "CONTINUE")
        
        elif choice == '4':
            # Theme selection
            selected_theme = ui.theme_selection_menu()
            if selected_theme:
                ui.input_prompt("Press Enter to continue", "CONTINUE")
        
        elif choice == '5':
            break
        
        else:
            ui.error("Invalid choice. Please try again.")
            time.sleep(1)

def view_downloads():
    """View downloaded songs."""
    print_header()
    
    music_dir = get_music_directory()
    ui.info(f"Music directory: {music_dir}", "DIR")
    
    # Get all artist directories
    artist_dirs = [d for d in os.listdir(music_dir) 
                  if os.path.isdir(os.path.join(music_dir, d))]
    
    if not artist_dirs:
        ui.warning("No music found. Download some songs first!")
        ui.input_prompt("Press Enter to continue", "CONTINUE")
        return
    
    # Sort alphabetically
    artist_dirs.sort()
    
    # Prepare artist data with song counts
    artists_data = []
    for artist in artist_dirs:
        # Count songs for this artist
        artist_path = os.path.join(music_dir, artist)
        songs = []
        for ext in ['alac', 'opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']:
            songs.extend(glob.glob(os.path.join(artist_path, f"*.{ext}")))
        artists_data.append((artist, len(songs)))
    
    # Use the new enhanced artist list menu
    ui.print_artist_list_menu(artists_data)
    
    choice = ui.input_prompt("Select artist or 'back' to return", "ARTIST").strip()
    
    if choice.lower() == 'back':
        return
    
    try:
        idx = int(choice) - 1
        if 0 <= idx < len(artist_dirs):
            view_artist_songs(music_dir, artist_dirs[idx])
        elif idx == len(artist_dirs):
            # This is the "Back to main menu" option (last option in the menu)
            return
        else:
            ui.error("Invalid choice")
            time.sleep(1)
    except ValueError:
        ui.error("Please enter a number")
        time.sleep(1)

def view_artist_songs(music_dir, artist):
    """View songs for a specific artist."""
    artist_path = os.path.join(music_dir, artist)
    
    # Get all songs with common audio formats
    songs = []
    for ext in ['alac', 'opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']:
        songs.extend(glob.glob(os.path.join(artist_path, f"*.{ext}")))
    
    # Sort by modification time (newest first)
    songs.sort(key=os.path.getmtime, reverse=True)
    
    print_header()
    
    # Prepare songs data for the enhanced display
    songs_data = []
    for song in songs:
        # Get file size in MB
        size_mb = os.path.getsize(song) / (1024 * 1024)
        # Get modification time
        mod_time = time.strftime("%Y-%m-%d %H:%M", time.localtime(os.path.getmtime(song)))
        # Get filename without extension and the extension
        filename = os.path.splitext(os.path.basename(song))[0]
        ext = os.path.splitext(song)[1][1:]  # Remove the dot
        songs_data.append((filename, ext, size_mb, mod_time))
    
    # Use the new enhanced artist songs display
    ui.print_artist_songs_display(artist, songs_data)
    
    ui.input_prompt("Press Enter to continue", "CONTINUE")

def download_playlist():
    """Download a YouTube playlist."""
    print_header()
    ui.hacker_banner("DOWNLOAD YOUTUBE PLAYLIST")
    ui.section_divider()
    
    ui.info("Enter a YouTube playlist URL")
    ui.info("Or type 'back' to return to the main menu")
    
    url = ui.input_prompt("Playlist URL", "PLAYLIST").strip()
    if url.lower() == 'back':
        return
    
    if not url.startswith(('http://', 'https://')):
        ui.error("Please enter a valid URL (starting with http:// or https://)")
        time.sleep(2)
        return
    
    # Verify it's a YouTube playlist URL
    if not is_youtube_playlist(url):
        ui.error("The URL does not appear to be a YouTube playlist.")
        ui.info("Playlist URLs typically contain 'playlist' or 'list=' in the URL.")
        time.sleep(2)
        return
    
    music_dir = get_music_directory()
    audio_format = get_audio_format()
    
    # Ask if user wants to use Spotify metadata using the new UI
    ui.info("Do you want to use Spotify for enhanced metadata?")
    ui.info("This will add genre, composer, performer, track numbers, and disc information.")
    
    ui.print_playlist_options_menu()
    
    while True:
        choice = ui.input_prompt("Enter your choice (1-2)", "METADATA").strip()
        if choice in ['1', '2']:
            break
        ui.error("Please enter 1 or 2")
    
    use_spotify = choice == '1'
    
    ui.info(f"Music will be saved to: {music_dir}", "DIR")
    ui.info(f"Using audio format: {audio_format}", "FORMAT")
    if use_spotify:
        ui.info("Using Spotify for enhanced metadata", "METADATA")
    else:
        ui.info("Using basic metadata only", "METADATA")
    
    ui.info("Downloading playlist... (this may take several minutes)", "PLAYLIST")
    
    # Use the appropriate function based on user choice
    if use_spotify:
        success = run_playlist_with_spotify(url, music_dir=music_dir, audio_format=audio_format)
    else:
        success = run_playlist(url, music_dir=music_dir, audio_format=audio_format)
    
    if success:
        ui.success("Playlist download complete!")
    else:
        ui.error("Playlist download failed. Please try again.")
    
    ui.input_prompt("Press Enter to continue", "CONTINUE")

def main():
    """Main application loop."""
    first_run = True
    while True:
        # Show startup animation only on first run
        print_header(with_startup_animation=first_run)
        print_menu()
        first_run = False
        
        choice = ui.input_prompt("Select an option (1-7)", "COMMAND").strip()
        
        if choice == '1':
            search_and_download()
        elif choice == '2':
            download_from_url()
        elif choice == '3':
            download_playlist()
        elif choice == '4':
            batch_download()
        elif choice == '5':
            settings()
        elif choice == '6':
            view_downloads()
        elif choice == '7':
            exit_animation()
            sys.exit(0)
        else:
            error("Invalid choice. Please try again.")
            time.sleep(1)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print()  # New line after ^C
        error("Program interrupted. Exiting...", "INTERRUPT")
        exit_animation()
        sys.exit(0)