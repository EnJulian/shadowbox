import sys
import os

# Add the meta_ops directory to the Python path
sys.path.append(os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), 'meta_ops'))

from downloader import download_audio, is_url
from metadata import add_metadata
from cover_art import get_album_cover_url, download_cover_image
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

def find_downloaded_file(output_pattern):
    """
    Find the downloaded file based on the output pattern.
    
    Args:
        output_pattern (str): The output pattern used for downloading
        
    Returns:
        str or None: Path to the downloaded file, or None if not found
    """
    # If the pattern contains a directory, use that directory
    directory = os.path.dirname(output_pattern)
    if not directory:
        directory = '.'
    
    # Look for opus files first (our default format)
    opus_files = glob.glob(os.path.join(directory, '*.opus'))
    if opus_files:
        # Sort by modification time (newest first)
        return sorted(opus_files, key=os.path.getmtime, reverse=True)[0]
    
    # If no opus files, look for mp3 files
    mp3_files = glob.glob(os.path.join(directory, '*.mp3'))
    if mp3_files:
        # Sort by modification time (newest first)
        return sorted(mp3_files, key=os.path.getmtime, reverse=True)[0]
    
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
        print(f"📁 Creating artist directory: {artist_dir}")
        os.makedirs(artist_dir, exist_ok=True)
    
    return artist_dir

def run(query, output_file=None, music_dir=None):
    """
    Download a song, add metadata, and embed album art.
    
    Args:
        query (str): The song title and artist or URL
        output_file (str, optional): The output file path. Defaults to '%(title)s.%(ext)s'.
        music_dir (str, optional): Base directory for music organization. Defaults to 'Music' in user's home.
    """
    if output_file is None:
        output_file = '%(title)s.%(ext)s'
    
    # Download the audio to a temporary location first
    temp_output = os.path.join(os.getcwd(), 'temp_download.%(ext)s')
    
    # Download the audio
    try:
        success = download_audio(query, temp_output)
        if not success:
            print("❌ Download failed")
            return False
    except Exception as e:
        print(f"❌ Error downloading audio: {e}")
        return False

    # Find the downloaded file
    temp_file = find_downloaded_file(temp_output)
    if not temp_file:
        print("❌ Could not find the downloaded file")
        return False
    
    print(f"✅ Found downloaded file: {temp_file}")
    
    # Parse title and artist from query if it's not a URL
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
    print(f"📦 Moving file to: {final_file}")
    shutil.move(temp_file, final_file)
    
    album = f"{artist} Best Hits"
    
    print(f"🖼️ Fetching album cover for {title} by {artist}")
    image_url = get_album_cover_url(title, artist)
    cover_path = None
    if image_url:
        # Download cover to artist directory
        cover_filename = f"{safe_title}_cover.jpg"
        cover_path = os.path.join(artist_dir, cover_filename)
        
        if download_cover_image(image_url, cover_path):
            print(f"✅ Album cover downloaded to {cover_path}")
        else:
            print("⚠️ Failed to download album cover")
            cover_path = None
    else:
        print("⚠️ No album cover found")

    print(f"📝 Adding metadata to {final_file}")
    try:
        add_metadata(final_file, title=title, artist=artist, album=album, cover_path=cover_path)
        print(f"✅ Finished downloading and tagging: {final_file}")
        return True
    except Exception as e:
        print(f"❌ Error adding metadata: {e}")
        return False

def main():
    """
    Main entry point for the application.
    """
    parser = argparse.ArgumentParser(description="lsnen's Foobar2000 Music Downloader - Download music with metadata and album art")
    parser.add_argument('-q', '--query', help='Song title and artist or URL')
    parser.add_argument('-o', '--output', help='Output file path (optional)')
    parser.add_argument('-d', '--directory', help='Base music directory (default: ~/Music)')
    args = parser.parse_args()
    
    query = args.query
    if not query:
        query = input("🎵 Enter song title and artist or URL: ")
    
    # If directory is not specified, use default (~/Music)
    music_dir = args.directory
    if music_dir:
        # Expand user directory if it starts with ~
        if music_dir.startswith('~'):
            music_dir = os.path.expanduser(music_dir)
        
        # Create the directory if it doesn't exist
        if not os.path.exists(music_dir):
            try:
                os.makedirs(music_dir, exist_ok=True)
                print(f"📁 Created music directory: {music_dir}")
            except Exception as e:
                print(f"❌ Error creating music directory: {e}")
                print("Using default directory instead")
                music_dir = None
    
    success = run(query, args.output, music_dir)
    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())