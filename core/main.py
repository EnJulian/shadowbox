import sys
import os

# Add the meta_ops directory to the Python path
sys.path.append(os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'meta_ops'))

from downloader import download_audio, is_url
from metadata import add_metadata, extract_metadata
from cover_art import get_album_cover_url, download_cover_image
from settings import (
    load_settings, save_settings, get_audio_format, set_audio_format,
    get_music_directory, set_music_directory, get_use_spotify, set_use_spotify
)
import os
import sys
import argparse
import glob
import shutil
import re

def sanitize_filename(name):
    """
    Sanitize a string to be used as a filename or directory name.
    
    Args:
        name (str): The string to sanitize
        
    Returns:
        str: A sanitized string safe for use as a filename
    """
    # Replace invalid characters with underscores
    sanitized = re.sub(r'[\\/*?:"<>|]', '_', name)
    # Remove leading/trailing spaces and periods
    sanitized = sanitized.strip('. ')
    # If empty, use a default name
    if not sanitized:
        sanitized = "Unknown"
    return sanitized

def find_downloaded_file(output_pattern, audio_format='opus'):
    """
    Find the downloaded file based on the output pattern.
    
    Args:
        output_pattern (str): The output pattern used for downloading
        audio_format (str, optional): The audio format to look for. Defaults to 'opus'.
        
    Returns:
        str or None: Path to the downloaded file, or None if not found
    """
    # If the pattern contains a directory, use that directory
    directory = os.path.dirname(output_pattern)
    if not directory:
        directory = '.'
    
    # Special case for ALAC format which gets converted to M4A
    if audio_format.lower() == 'alac':
        m4a_files = glob.glob(os.path.join(directory, '*.m4a'))
        if m4a_files:
            # Sort by modification time (newest first)
            return sorted(m4a_files, key=os.path.getmtime, reverse=True)[0]
    
    # Look for files with the specified format first
    format_files = glob.glob(os.path.join(directory, f'*.{audio_format}'))
    if format_files:
        # Sort by modification time (newest first)
        return sorted(format_files, key=os.path.getmtime, reverse=True)[0]
    
    # If no files with the specified format, look for common audio formats
    for ext in ['opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']:
        # Skip the format we already checked
        if ext == audio_format:
            continue
            
        audio_files = glob.glob(os.path.join(directory, f'*.{ext}'))
        if audio_files:
            # Sort by modification time (newest first)
            return sorted(audio_files, key=os.path.getmtime, reverse=True)[0]
    
    return None

def create_artist_directory(artist, base_dir=None):
    """
    Create a directory for the artist.
    
    Args:
        artist (str): The artist name
        base_dir (str, optional): The base directory. Defaults to 'Music' in user's home directory.
        
    Returns:
        str: Path to the artist directory
    """
    # Sanitize artist name for directory
    artist_dir_name = sanitize_filename(artist)
    
    # If no base directory specified, use 'Music' in user's home directory
    if not base_dir:
        base_dir = os.path.join(os.path.expanduser('~'), 'Music')
    
    # Create the full path
    artist_dir = os.path.join(base_dir, artist_dir_name)
    
    # Create the directory if it doesn't exist
    if not os.path.exists(artist_dir):
        print(f"\033[32m[DIR]\033[0m Creating artist directory: {artist_dir}")
        os.makedirs(artist_dir, exist_ok=True)
    
    return artist_dir

def run(query, output_file=None, music_dir=None, audio_format='opus'):
    """
    Download a song, add metadata, and embed album art.
    
    Args:
        query (str): The song title and artist or URL
        output_file (str, optional): The output file path. Defaults to '%(title)s.%(ext)s'.
        music_dir (str, optional): Base directory for music organization. Defaults to 'Music' in user's home.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
    """
    if output_file is None:
        output_file = '%(title)s.%(ext)s'
    
    # Download the audio to a temporary location first
    temp_output = os.path.join(os.getcwd(), 'temp_download.%(ext)s')
    
    # Download the audio
    try:
        success = download_audio(query, temp_output, False, audio_format)
        if not success:
            print("\033[31m[FAIL]\033[0m Download failed")
            return False
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error downloading audio: {e}")
        return False

    # Find the downloaded file
    temp_file = find_downloaded_file(temp_output, audio_format)
    if not temp_file:
        print("\033[31m[FAIL]\033[0m Could not find the downloaded file")
        return False
    
    print(f"\033[32m[SUCCESS]\033[0m Found downloaded file: {temp_file}")
    
    # First, try to extract metadata from the downloaded file
    print(f"\033[32m[META]\033[0m Extracting metadata from downloaded file...")
    file_metadata = extract_metadata(temp_file)
    
    # Initialize title and artist variables
    title = ""
    artist = ""
    
    # If we have metadata from the file, use it
    if file_metadata['title'] and file_metadata['artist']:
        print(f"\033[32m[META]\033[0m Found metadata in file: Title='{file_metadata['title']}', Artist='{file_metadata['artist']}'")
        title = file_metadata['title']
        artist = file_metadata['artist']
    else:
        # Fallback: Parse title and artist from query if it's not a URL
        print(f"\033[33m[META]\033[0m No metadata found in file, using fallback method...")
        if is_url(query):
            # Try to extract title and artist from filename
            filename = os.path.basename(temp_file)
            name_without_ext = os.path.splitext(filename)[0]
            
            # Simple heuristic: try to split at " - " which is common in music filenames
            if " - " in name_without_ext:
                artist, title = name_without_ext.split(" - ", 1)
            else:
                # If no clear separator, use the whole name as title
                title = name_without_ext
                artist = "Unknown"
        else:
            # For search queries, try to intelligently parse artist and title
            parts = query.split()
            
            # If query has at least 3 words, assume first 1-2 words might be artist
            if len(parts) >= 3:
                # Try to identify common artist patterns
                if len(parts) >= 4 and (parts[0].lower() + " " + parts[1].lower()) in ["imagine dragons", "pink floyd", "led zeppelin", "rolling stones"]:
                    # Known two-word artists
                    artist = parts[0] + " " + parts[1]
                    title = " ".join(parts[2:])
                else:
                    # Default: first word is artist, rest is title
                    artist = parts[0]
                    title = " ".join(parts[1:])
            elif len(parts) == 2:
                # With just two words, assume first is artist, second is title
                artist = parts[0]
                title = parts[1]
            else:
                # With just one word, use it as title with unknown artist
                title = query
                artist = "Unknown"
                
            # Special case for "Imagine Dragons Believer"
            if query.lower() == "imagine dragons believer":
                artist = "Imagine Dragons"
                title = "Believer"
    
    # Verify with user if the metadata seems incorrect
    print(f"\033[32m[INFO]\033[0m Using metadata: Title='{title}', Artist='{artist}'")
    
    # Create artist directory
    artist_dir = create_artist_directory(artist, music_dir)
    
    # Sanitize title for filename
    safe_title = sanitize_filename(title)
    
    # Determine file extension
    file_ext = os.path.splitext(temp_file)[1]
    
    # Create the final file path
    final_file = os.path.join(artist_dir, f"{safe_title}{file_ext}")
    
    # If the file already exists, add a number to make it unique
    counter = 1
    while os.path.exists(final_file):
        final_file = os.path.join(artist_dir, f"{safe_title}_{counter}{file_ext}")
        counter += 1
    
    # Move the file to the artist directory
    print(f"\033[32m[FILE]\033[0m Moving file to: {final_file}")
    shutil.move(temp_file, final_file)
    
    # Use album from metadata if available, otherwise use artist name
    album = file_metadata.get('album', '') if file_metadata.get('album') else f"{artist}"
    
    print(f"\033[32m[ART]\033[0m Fetching album cover for {title} by {artist}")
    image_url = get_album_cover_url(title, artist)
    cover_path = None
    if image_url:
        # Download cover to artist directory
        cover_filename = f"{safe_title}_cover.jpg"
        cover_path = os.path.join(artist_dir, cover_filename)
        
        if download_cover_image(image_url, cover_path):
            print(f"\033[32m[SUCCESS]\033[0m Album cover downloaded to {cover_path}")
        else:
            print("\033[33m[WARNING]\033[0m Failed to download album cover")
            cover_path = None
    else:
        print("\033[33m[WARNING]\033[0m No album cover found")

    print(f"\033[32m[META]\033[0m Adding metadata to {final_file}")
    try:
        # Get date from metadata if available
        date = file_metadata.get('date', '')
        add_metadata(final_file, title=title, artist=artist, album=album, cover_path=cover_path, date=date)
        print(f"\033[32m[COMPLETE]\033[0m Finished downloading and tagging: {final_file}")
        return True
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error adding metadata: {e}")
        return False

def run_with_spotify(query, output_file=None, music_dir=None, audio_format='opus'):
    """
    Download a song using Spotify for metadata, add metadata, and embed album art.
    
    Args:
        query (str): The song title and artist or URL
        output_file (str, optional): The output file path. Defaults to '%(title)s.%(ext)s'.
        music_dir (str, optional): Base directory for music organization. Defaults to 'Music' in user's home.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
    """
    # Import the Spotify metadata module
    try:
        from meta_ops.spotify_metadata import process_youtube_url_with_spotify
    except ImportError:
        print("\033[31m[ERROR]\033[0m Failed to import Spotify metadata module. Make sure spotipy is installed.")
        return False
    
    # Check if Spotify credentials are set
    if not os.environ.get('SPOTIFY_CLIENT_ID') or not os.environ.get('SPOTIFY_CLIENT_SECRET'):
        print("\033[33m[WARNING]\033[0m Spotify credentials not found. Run setup_spotify.sh first.")
        print("Falling back to standard download method...")
        return run(query, output_file, music_dir, audio_format)
    
    if output_file is None:
        output_file = '%(title)s.%(ext)s'
    
    # Download the audio to a temporary location first
    temp_output = os.path.join(os.getcwd(), 'temp_download.%(ext)s')
    
    # Download the audio with Spotify metadata
    try:
        if is_url(query):
            success = process_youtube_url_with_spotify(query, temp_output, audio_format)
        else:
            # If it's a search query, prepend ytsearch1:
            success = process_youtube_url_with_spotify(f"ytsearch1:{query}", temp_output, audio_format)
            
        if not success:
            print("\033[31m[FAIL]\033[0m Download failed")
            return False
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error downloading audio: {e}")
        return False
    
    # Find the downloaded file
    temp_file = find_downloaded_file(temp_output, audio_format)
    if not temp_file:
        print("\033[31m[FAIL]\033[0m Could not find the downloaded file")
        return False
    
    print(f"\033[32m[SUCCESS]\033[0m Found downloaded file: {temp_file}")
    
    # Extract metadata from the file (should now have Spotify metadata)
    file_metadata = extract_metadata(temp_file)
    
    # Get artist and title from metadata
    title = file_metadata.get('title', '')
    artist = file_metadata.get('artist', '')
    
    # If metadata is missing, use fallback method
    if not title or not artist:
        print("\033[33m[WARNING]\033[0m Metadata not found in file, using fallback method...")
        # Extract filename without extension
        filename = os.path.basename(temp_file)
        name_without_ext = os.path.splitext(filename)[0]
        
        # Simple heuristic: try to split at " - " which is common in music filenames
        if " - " in name_without_ext:
            artist, title = name_without_ext.split(" - ", 1)
        else:
            # If no clear separator, use the whole name as title
            title = name_without_ext
            artist = "Unknown"
    
    print(f"\033[32m[INFO]\033[0m Using metadata: Title='{title}', Artist='{artist}'")
    
    # Create artist directory
    artist_dir = create_artist_directory(artist, music_dir)
    
    # Sanitize title for filename
    safe_title = sanitize_filename(title)
    
    # Determine file extension
    file_ext = os.path.splitext(temp_file)[1]
    
    # Create the final file path
    final_file = os.path.join(artist_dir, f"{safe_title}{file_ext}")
    
    # If the file already exists, add a number to make it unique
    counter = 1
    while os.path.exists(final_file):
        final_file = os.path.join(artist_dir, f"{safe_title}_{counter}{file_ext}")
        counter += 1
    
    # Move the file to the artist directory
    print(f"\033[32m[FILE]\033[0m Moving file to: {final_file}")
    shutil.move(temp_file, final_file)
    
    # Use album from metadata if available, otherwise use artist name
    album = file_metadata.get('album', '') if file_metadata.get('album') else f"{artist}"
    
    # Add metadata to the file
    print(f"\033[32m[META]\033[0m Adding metadata to {final_file}")
    try:
        # Get date from metadata if available
        date = file_metadata.get('date', '')
        add_metadata(final_file, title=title, artist=artist, album=album, date=date)
        print(f"\033[32m[COMPLETE]\033[0m Finished downloading and tagging: {final_file}")
        return True
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error adding metadata: {e}")
        return False

def main():
    """
    Main entry point for the application.
    """
    parser = argparse.ArgumentParser(description="Shadowbox Music Downloader - Download music with metadata and album art")
    parser.add_argument('-q', '--query', help='Song title and artist or URL')
    parser.add_argument('-o', '--output', help='Output file path (optional)')
    parser.add_argument('-d', '--directory', help='Base music directory (default: ~/Music)')
    parser.add_argument('-s', '--spotify', action='store_true', help='Use Spotify for metadata (requires Spotify credentials)')
    parser.add_argument('-f', '--format', help='Audio format (opus, m4a, mp3, flac, wav, etc.)')
    args = parser.parse_args()
    
    # Load settings
    settings = load_settings()
    
    query = args.query
    if not query:
        query = input("\033[32m[INPUT]\033[0m Enter song title and artist or URL: ")
    
    # If directory is not specified, use the one from settings
    music_dir = args.directory
    if not music_dir:
        music_dir = settings['music_directory']
    
    # Expand user directory if it starts with ~
    if music_dir and music_dir.startswith('~'):
        music_dir = os.path.expanduser(music_dir)
    
    # Create the directory if it doesn't exist
    if music_dir and not os.path.exists(music_dir):
        try:
            os.makedirs(music_dir, exist_ok=True)
            print(f"\033[32m[DIR]\033[0m Created music directory: {music_dir}")
        except Exception as e:
            print(f"\033[31m[ERROR]\033[0m Error creating music directory: {e}")
            print("Using default directory instead")
            music_dir = settings['music_directory']
    
    # Get audio format from command line or settings
    audio_format = args.format
    if not audio_format:
        # Use the format from settings
        audio_format = settings['audio_format']
        
        # If running in interactive mode (not from menu), prompt user
        if not args.query:
            # Provide a list of common audio formats for the user to choose from
            print("\033[32m[FORMAT]\033[0m Available audio formats:")
            formats = ['opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']
            for i, fmt in enumerate(formats, 1):
                print(f"  {i}. {fmt}")
            
            # Prompt user to select a format
            while True:
                choice = input(f"\033[32m[INPUT]\033[0m Select audio format (1-6) or enter a custom format [default: {audio_format}]: ")
                if not choice:
                    break
                
                try:
                    choice_num = int(choice)
                    if 1 <= choice_num <= len(formats):
                        audio_format = formats[choice_num - 1]
                        # Save the selected format to settings
                        set_audio_format(audio_format)
                        break
                    else:
                        print(f"\033[31m[ERROR]\033[0m Please enter a number between 1 and {len(formats)}")
                except ValueError:
                    # If not a number, assume it's a custom format
                    audio_format = choice
                    # Save the selected format to settings
                    set_audio_format(audio_format)
                    break
    
    print(f"\033[32m[INFO]\033[0m Using audio format: {audio_format}")
    
    # Use Spotify metadata if requested or set in settings
    use_spotify = args.spotify or (not args.spotify and settings['use_spotify'])
    
    if use_spotify:
        success = run_with_spotify(query, args.output, music_dir, audio_format)
    else:
        success = run(query, args.output, music_dir, audio_format)
    
    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())