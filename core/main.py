import sys
import os
import argparse
import glob
import shutil
import re

# Add the meta_ops directory to the Python path for development
current_dir = os.path.dirname(os.path.abspath(__file__))
parent_dir = os.path.dirname(current_dir)
meta_ops_dir = os.path.join(parent_dir, 'meta_ops')

if os.path.exists(meta_ops_dir) and meta_ops_dir not in sys.path:
    sys.path.insert(0, meta_ops_dir)

# Import enhanced terminal UI
try:
    from terminal_ui import ui, success, error, warning, info, system, audio, download, scan, directory, api
except ImportError:
    # Fallback to basic print if terminal_ui is not available
    def success(msg, tag="SUCCESS"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def error(msg, tag="ERROR"): 
        print(f"[{tag}] {msg}")  # Always show errors
    def warning(msg, tag="WARNING"): 
        print(f"[{tag}] {msg}")  # Always show warnings
    def info(msg, tag="INFO"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def system(msg, tag="SYSTEM"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def audio(msg, tag="AUDIO"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def download(msg, tag="DOWNLOAD"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def scan(msg, tag="SCAN"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def directory(msg, tag="DIR"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    def api(msg, tag="API"): 
        try:
            if get_verbose_logging(): print(f"[{tag}] {msg}")
        except:
            pass  # Silently ignore if settings not available
    ui = None

# Try absolute imports first (for PyInstaller), then relative imports (for development)
try:
    from meta_ops.downloader import download_audio, is_url, is_youtube_playlist
    from meta_ops.metadata import add_metadata, extract_metadata
    from meta_ops.cover_art import get_album_cover_url, download_cover_image
    from meta_ops.settings import (
        load_settings, save_settings, get_audio_format, set_audio_format,
        get_music_directory, set_music_directory, get_use_spotify, set_use_spotify,
        get_verbose_logging, set_verbose_logging, get_use_genius, set_use_genius,
        get_genius_access_token, set_genius_access_token,
        get_spotify_client_id, set_spotify_client_id,
        get_spotify_client_secret, set_spotify_client_secret
    )
except ImportError:
    # Fallback to direct imports (for development when meta_ops is in path)
    from downloader import download_audio, is_url, is_youtube_playlist
    from metadata import add_metadata, extract_metadata
    from cover_art import get_album_cover_url, download_cover_image
    from settings import (
        load_settings, save_settings, get_audio_format, set_audio_format,
        get_music_directory, set_music_directory, get_use_spotify, set_use_spotify,
        get_verbose_logging, set_verbose_logging, get_use_genius, set_use_genius,
        get_genius_access_token, set_genius_access_token,
        get_spotify_client_id, set_spotify_client_id,
        get_spotify_client_secret, set_spotify_client_secret
    )

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
    # Use only the first artist name if multiple artists are separated by commas
    first_artist = artist.split(',')[0].strip()
    
    # Sanitize artist name for directory
    artist_dir_name = sanitize_filename(first_artist)
    
    # If no base directory specified, use 'Music' in user's home directory
    if not base_dir:
        base_dir = os.path.join(os.path.expanduser('~'), 'Music')
    
    # Create the full path
    artist_dir = os.path.join(base_dir, artist_dir_name)
    
    # Create the directory if it doesn't exist
    if not os.path.exists(artist_dir):
        directory(f"Creating artist directory: {artist_dir}")
        os.makedirs(artist_dir, exist_ok=True)
    
    return artist_dir

def create_album_directory(artist_dir, album_name, title=None):
    """
    Create a directory for the album inside the artist directory.
    
    Args:
        artist_dir (str): Path to the artist directory
        album_name (str): The album name
        title (str, optional): The track title, used as fallback if album_name is empty
        
    Returns:
        str: Path to the album directory
    """
    # If album name is empty or None, use the track title as the album name
    if not album_name and title:
        album_name = title
        info(f"No album name found, using track title as album folder: '{title}'")
    elif not album_name and not title:
        album_name = "Unknown Album"
        info("No album name or title found, using 'Unknown Album' as folder name")
    
    # Remove "Album - " prefix if present
    if album_name and album_name.startswith("Album - "):
        album_name = album_name[8:]  # Remove the first 8 characters ("Album - ")
        info(f"Removed 'Album - ' prefix from album name: '{album_name}'")
    
    # Sanitize album name for directory
    album_dir_name = sanitize_filename(album_name)
    
    # Create the full path
    album_dir = os.path.join(artist_dir, album_dir_name)
    
    # Create the directory if it doesn't exist
    if not os.path.exists(album_dir):
        directory(f"Creating album directory: {album_dir}")
        os.makedirs(album_dir, exist_ok=True)
    
    return album_dir

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
        download_success = download_audio(query, temp_output, False, audio_format)
        if not download_success:
            error("Download failed", "FAIL")
            return False
    except Exception as e:
        error(f"Error downloading audio: {e}")
        return False

    # Find the downloaded file
    temp_file = find_downloaded_file(temp_output, audio_format)
    if not temp_file:
        error("Could not find the downloaded file", "FAIL")
        return False
    
    success(f"Found downloaded file: {temp_file}")
    
    # First, try to extract metadata from the downloaded file
    system("Extracting metadata from downloaded file...", "META")
    file_metadata = extract_metadata(temp_file)
    
    # Initialize title and artist variables
    title = ""
    artist = ""
    
    # If we have metadata from the file, use it
    if file_metadata['title'] and file_metadata['artist']:
        system(f"Found metadata in file: Title='{file_metadata['title']}', Artist='{file_metadata['artist']}'", "META")
        title = file_metadata['title']
        artist = file_metadata['artist']
    else:
        # Fallback: Parse title and artist from query if it's not a URL
        warning("No metadata found in file, using fallback method...", "META")
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
    
    # Try to enhance metadata with Spotify search before proceeding
    enhanced_metadata = None
    try:
        from meta_ops.spotify_metadata import search_spotify_for_metadata
        system(f"Searching Spotify for enhanced metadata: {title} by {artist}", "SPOTIFY")
        enhanced_metadata = search_spotify_for_metadata(title, artist)
        
        if enhanced_metadata:
            # Use the enhanced metadata from Spotify
            original_title = title
            original_artist = artist
            title = enhanced_metadata['title']
            artist = enhanced_metadata['artist']
            success(f"Enhanced metadata from Spotify: '{title}' by '{artist}'", "SPOTIFY")
            
            # If the metadata changed significantly, inform the user
            if original_title.lower() != title.lower() or original_artist.lower() != artist.lower():
                info(f"Metadata updated: '{original_title}' by '{original_artist}' → '{title}' by '{artist}'")
        else:
            info("No enhanced metadata found on Spotify, using original metadata")
    except ImportError:
        info("Spotify enhancement not available")
    except Exception as e:
        warning(f"Error during Spotify metadata enhancement: {e}")
    
    # Verify with user if the metadata seems incorrect
    info(f"Using final metadata: Title='{title}', Artist='{artist}'")
    
    # Create artist directory
    artist_dir = create_artist_directory(artist, music_dir)
    
    # Get album name from metadata if available
    if file_metadata:
        album_name = file_metadata.get("album", "")
    else:
        album_name = None
    
    # Create album directory inside artist directory
    album_dir = create_album_directory(artist_dir, album_name, title)
    
    # Sanitize title for filename
    safe_title = sanitize_filename(title)
    
    # Determine file extension
    file_ext = os.path.splitext(temp_file)[1]
    
    # Create the final file path (now in album directory)
    final_file = os.path.join(album_dir, f"{safe_title}{file_ext}")
    
    # If the file already exists, add a number to make it unique
    counter = 1
    while os.path.exists(final_file):
        final_file = os.path.join(album_dir, f"{safe_title}_{counter}{file_ext}")
        counter += 1
    
    # Move the file to the album directory
    system(f"Moving file to: {final_file}", "FILE")
    shutil.move(temp_file, final_file)
    
    # Use album from metadata if available, otherwise use artist name
    album = album_name if album_name else (title or f"{artist}")
    
    system(f"Fetching album cover for {title} by {artist}", "ART")
    image_url = get_album_cover_url(title, artist)
    cover_path = None
    if image_url:
        # Use the cover image for embedding but don't save it to disk
        temp_cover_path = os.path.join(os.getcwd(), f"temp_cover_{os.getpid()}.jpg")
        
        if download_cover_image(image_url, temp_cover_path):
            success("Album cover downloaded for embedding")
            cover_path = temp_cover_path
        else:
            warning("Failed to download album cover")
            cover_path = None
    
    # Fetch and add lyrics using the final enhanced metadata
    lyrics_text = None
    try:
        from meta_ops.lyrics import search_lyrics_with_fallbacks, add_lyrics_to_metadata
        system(f"Searching for lyrics: {title} by {artist}", "LYRICS")
        lyrics_text = search_lyrics_with_fallbacks(title, artist)
        
        if lyrics_text:
            success(f"Found lyrics ({len(lyrics_text)} characters)", "LYRICS")
        else:
            warning("No lyrics found for this song", "LYRICS")
    except ImportError:
        warning("Lyrics functionality not available. Install lyricsgenius: pip install lyricsgenius", "LYRICS")
    except Exception as e:
        warning(f"Error searching for lyrics: {e}", "LYRICS")
    
    # Add metadata to the file
    system(f"Adding metadata to {final_file}", "META")
    try:
        # Save the full artist string as album_artist
        full_artist = artist
        # Add metadata with both artist and album_artist
        add_metadata(final_file, title=title, artist=artist, album=album, cover_path=cover_path, album_artist=full_artist)
        
        # Add lyrics to the file if we found them
        if lyrics_text:
            try:
                system("Adding lyrics to metadata", "LYRICS")
                if add_lyrics_to_metadata(final_file, lyrics_text):
                    success("Lyrics successfully embedded in file", "LYRICS")
                else:
                    warning("Failed to embed lyrics in file", "LYRICS")
            except Exception as e:
                warning(f"Error adding lyrics to file: {e}", "LYRICS")
        
        # Check if this was a Bandcamp download and enhance with Spotify metadata if available
        from meta_ops.downloader import is_bandcamp_url
        if is_url(query) and is_bandcamp_url(query):
            try:
                from meta_ops.spotify_metadata import enhance_bandcamp_file_with_spotify
                system("Enhancing Bandcamp file with Spotify metadata...", "BANDCAMP")
                enhancement_success, enhanced_metadata = enhance_bandcamp_file_with_spotify(final_file)
                if enhancement_success:
                    success("Successfully enhanced with additional metadata", "BANDCAMP")
                else:
                    warning("No additional metadata enhancement available", "BANDCAMP")
            except ImportError:
                warning("Spotify enhancement not available for Bandcamp files")
            except Exception as e:
                warning(f"Error enhancing Bandcamp file: {e}")
        
        # Remove the temporary cover file if it exists
        if cover_path and os.path.exists(cover_path):
            os.remove(cover_path)
            
        success(f"Finished downloading and tagging: {final_file}", "COMPLETE")
        return True
    except Exception as e:
        error(f"Error adding metadata: {e}")
        # Clean up temp file even if there was an error
        if cover_path and os.path.exists(cover_path):
            os.remove(cover_path)
        return False

def run_playlist(query, output_file=None, music_dir=None, audio_format='opus'):
    """
    Download a YouTube playlist, add metadata, and embed album art for each track.
    
    Args:
        query (str): The YouTube playlist URL
        output_file (str, optional): The output file path pattern. Defaults to '%(title)s.%(ext)s'.
        music_dir (str, optional): Base directory for music organization. Defaults to 'Music' in user's home.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
    
    Returns:
        bool: True if successful, False otherwise
    """
    if output_file is None:
        output_file = '%(title)s.%(ext)s'
    
    # Download the playlist
    try:
        downloaded_files = download_audio(query, output_file, False, audio_format)
        if not downloaded_files or not isinstance(downloaded_files, list):
            error("Playlist download failed", "FAIL")
            return False
    except Exception as e:
        error(f"Error downloading playlist: {e}")
        return False
    
    if get_verbose_logging():
        print(f"\033[32m[SUCCESS]\033[0m Downloaded {len(downloaded_files)} tracks from playlist")
    
    # Process each file in the playlist
    success_count = 0
    for file_path in downloaded_files:
        try:
            # Extract metadata from the downloaded file
            if get_verbose_logging():
                print(f"\033[32m[META]\033[0m Extracting metadata from: {os.path.basename(file_path)}")
            file_metadata = extract_metadata(file_path)
            
            # Initialize title and artist variables
            title = ""
            artist = ""
            
            # If we have metadata from the file, use it
            if file_metadata['title'] and file_metadata['artist']:
                if get_verbose_logging():
                    print(f"\033[32m[META]\033[0m Found metadata in file: Title='{file_metadata['title']}', Artist='{file_metadata['artist']}'")
                title = file_metadata['title']
                artist = file_metadata['artist']
            else:
                # Fallback: Parse title and artist from filename
                if get_verbose_logging():
                    print(f"\033[33m[META]\033[0m No metadata found in file, using fallback method...")
                filename = os.path.basename(file_path)
                # Remove playlist index if present (e.g., "01 - ")
                if re.match(r'^\d+\s*-\s*', filename):
                    filename = re.sub(r'^\d+\s*-\s*', '', filename)
                
                name_without_ext = os.path.splitext(filename)[0]
                
                # Simple heuristic: try to split at " - " which is common in music filenames
                if " - " in name_without_ext:
                    artist, title = name_without_ext.split(" - ", 1)
                else:
                    # If no clear separator, use the whole name as title
                    title = name_without_ext
                    artist = "Unknown"
            
            # Try to enhance metadata with Spotify search before proceeding
            enhanced_metadata = None
            try:
                from meta_ops.spotify_metadata import search_spotify_for_metadata
                if get_verbose_logging():
                    print(f"\033[32m[SPOTIFY]\033[0m Searching Spotify for enhanced metadata: {title} by {artist}")
                enhanced_metadata = search_spotify_for_metadata(title, artist)
                
                if enhanced_metadata:
                    # Use the enhanced metadata from Spotify
                    original_title = title
                    original_artist = artist
                    title = enhanced_metadata['title']
                    artist = enhanced_metadata['artist']
                    if get_verbose_logging():
                        print(f"\033[32m[SPOTIFY]\033[0m Enhanced metadata from Spotify: '{title}' by '{artist}'")
                    
                    # If the metadata changed significantly, inform the user
                    if original_title.lower() != title.lower() or original_artist.lower() != artist.lower():
                        if get_verbose_logging():
                            print(f"\033[32m[INFO]\033[0m Metadata updated: '{original_title}' by '{original_artist}' → '{title}' by '{artist}'")
                else:
                    if get_verbose_logging():
                        print("\033[32m[INFO]\033[0m No enhanced metadata found on Spotify, using original metadata")
            except ImportError:
                if get_verbose_logging():
                    print("\033[32m[INFO]\033[0m Spotify enhancement not available")
            except Exception as e:
                if get_verbose_logging():
                    print(f"\033[33m[WARNING]\033[0m Error during Spotify metadata enhancement: {e}")
            
            # Create artist directory
            artist_dir = create_artist_directory(artist, music_dir)
            
            # Get album name from metadata if available
            if file_metadata:
                album_name = file_metadata.get("album", "")
            else:
                album_name = None
            
            # Create album directory inside artist directory
            album_dir = create_album_directory(artist_dir, album_name, title)
            
            # Sanitize title for filename
            safe_title = sanitize_filename(title)
            
            # Determine file extension
            file_ext = os.path.splitext(file_path)[1]
            
            # Create the final file path (now in album directory)
            final_file = os.path.join(album_dir, f"{safe_title}{file_ext}")
            
            # If the file already exists, add a number to make it unique
            counter = 1
            while os.path.exists(final_file):
                final_file = os.path.join(album_dir, f"{safe_title}_{counter}{file_ext}")
                counter += 1
            
            # Move the file to the album directory
            if get_verbose_logging():
                print(f"\033[32m[FILE]\033[0m Moving file to: {final_file}")
            shutil.move(file_path, final_file)
            
            # Use album from metadata if available, otherwise use artist name
            album = album_name if album_name else (title or f"{artist}")
            
            if get_verbose_logging():
                print(f"\033[32m[ART]\033[0m Fetching album cover for {title} by {artist}")
            image_url = get_album_cover_url(title, artist)
            cover_path = None
            if image_url:
                # Use the cover image for embedding but don't save it to disk
                temp_cover_path = os.path.join(os.getcwd(), f"temp_cover_{os.getpid()}.jpg")
                
                if download_cover_image(image_url, temp_cover_path):
                    print(f"\033[32m[SUCCESS]\033[0m Album cover downloaded for embedding")
                    cover_path = temp_cover_path
                else:
                    print("\033[33m[WARNING]\033[0m Failed to download album cover")
                    cover_path = None
            
            # Fetch and add lyrics using the final enhanced metadata
            lyrics_text = None
            try:
                from meta_ops.lyrics import search_lyrics_with_fallbacks, add_lyrics_to_metadata
                if get_verbose_logging():
                    print(f"\033[32m[LYRICS]\033[0m Searching for lyrics: {title} by {artist}")
                lyrics_text = search_lyrics_with_fallbacks(title, artist)
                
                if lyrics_text:
                    if get_verbose_logging():
                        print(f"\033[32m[LYRICS]\033[0m Found lyrics ({len(lyrics_text)} characters)")
                else:
                    if get_verbose_logging():
                        print("\033[33m[LYRICS]\033[0m No lyrics found for this song")
            except ImportError:
                if get_verbose_logging():
                    print("\033[33m[LYRICS]\033[0m Lyrics functionality not available. Install lyricsgenius: pip install lyricsgenius")
            except Exception as e:
                if get_verbose_logging():
                    print(f"\033[33m[LYRICS]\033[0m Error searching for lyrics: {e}")
            
            # Add metadata to the file
            print(f"\033[32m[META]\033[0m Adding metadata to {final_file}")
            try:
                # Save the full artist string as album_artist
                full_artist = artist
                # Add metadata with both artist and album_artist
                add_metadata(final_file, title=title, artist=artist, album=album, cover_path=cover_path, album_artist=full_artist)
                
                # Add lyrics to the file if we found them
                if lyrics_text:
                    try:
                        if get_verbose_logging():
                            print("\033[32m[LYRICS]\033[0m Adding lyrics to metadata")
                        if add_lyrics_to_metadata(final_file, lyrics_text):
                            if get_verbose_logging():
                                print("\033[32m[LYRICS]\033[0m Lyrics successfully embedded in file")
                        else:
                            if get_verbose_logging():
                                print("\033[33m[LYRICS]\033[0m Failed to embed lyrics in file")
                    except Exception as e:
                        if get_verbose_logging():
                            print(f"\033[33m[LYRICS]\033[0m Error adding lyrics to file: {e}")
                
                # Remove the temporary cover file if it exists
                if cover_path and os.path.exists(cover_path):
                    os.remove(cover_path)
                print(f"\033[32m[COMPLETE]\033[0m Finished processing: {final_file}")
                success_count += 1
            except Exception as e:
                print(f"\033[31m[ERROR]\033[0m Error adding metadata: {e}")
        except Exception as e:
            print(f"\033[31m[ERROR]\033[0m Error processing file {file_path}: {e}")
    
    # Clean up the temporary playlist directory
    temp_dir = os.path.join(os.getcwd(), 'temp_playlist')
    if os.path.exists(temp_dir):
        try:
            shutil.rmtree(temp_dir)
            print(f"\033[32m[CLEANUP]\033[0m Removed temporary directory: {temp_dir}")
        except Exception as e:
            print(f"\033[33m[WARNING]\033[0m Failed to remove temporary directory: {e}")
    
    print(f"\033[32m[SUMMARY]\033[0m Successfully processed {success_count} out of {len(downloaded_files)} tracks from the playlist")
    return success_count > 0

def run_with_spotify(query, output_file=None, music_dir=None, audio_format='opus'):
    """
    Download a song with Spotify metadata.
    
    Args:
        query (str): The song title and artist or URL
        output_file (str, optional): The output file path. Defaults to '%(title)s.%(ext)s'.
        music_dir (str, optional): Base directory for music organization. Defaults to 'Music' in user's home.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
    """
    # Import Spotify metadata module
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
    spotify_metadata = None
    try:
        if is_url(query):
            download_success, spotify_metadata = process_youtube_url_with_spotify(query, temp_output, audio_format)
        else:
            # If it's a search query, prepend ytsearch1:
            download_success, spotify_metadata = process_youtube_url_with_spotify(f"ytsearch1:{query}", temp_output, audio_format)
            
        if not download_success:
            error("Download failed", "FAIL")
            return False
    except Exception as e:
        error(f"Error downloading audio: {e}")
        return False
    
    # Find the downloaded file
    temp_file = find_downloaded_file(temp_output, audio_format)
    if not temp_file:
        print("\033[31m[FAIL]\033[0m Could not find the downloaded file")
        return False
    
    print(f"\033[32m[SUCCESS]\033[0m Found downloaded file: {temp_file}")
    
    # Get metadata - prefer Spotify metadata if available
    file_metadata = None
    if spotify_metadata:
        print(f"\033[32m[INFO]\033[0m Using Spotify metadata")
        title = spotify_metadata.get('title', '')
        # Store the full artist string for later use
        full_artist = spotify_metadata.get('artist', '')
        # For directory creation, use only the first artist
        artist = full_artist.split(',')[0].strip() if ',' in full_artist else full_artist
    else:
        # Extract metadata from the file as fallback
        print(f"\033[33m[WARNING]\033[0m No Spotify metadata available, extracting from file")
        file_metadata = extract_metadata(temp_file)
        title = file_metadata.get('title', '')
        artist = file_metadata.get('artist', '')
        full_artist = artist  # No full artist info available
    
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
    
    # Get album name from metadata if available
    # Get album name from metadata if available
    if spotify_metadata:
        album_name = spotify_metadata.get("album", "")
    elif file_metadata:
        album_name = file_metadata.get("album", "")
    else:
        album_name = None
    
    # Create album directory inside artist directory
    album_dir = create_album_directory(artist_dir, album_name, title)
    
    # Sanitize title for filename
    safe_title = sanitize_filename(title)
    
    # Determine file extension
    file_ext = os.path.splitext(temp_file)[1]
    
    # Create the final file path (now in album directory)
    final_file = os.path.join(album_dir, f"{safe_title}{file_ext}")
    
    # If the file already exists, add a number to make it unique
    counter = 1
    while os.path.exists(final_file):
        final_file = os.path.join(album_dir, f"{safe_title}_{counter}{file_ext}")
        counter += 1
    
    # Move the file to the album directory
    print(f"\033[32m[FILE]\033[0m Moving file to: {final_file}")
    shutil.move(temp_file, final_file)
    
    # Use album from metadata if available, otherwise use artist name
    album = album_name if album_name else (title or f"{artist}")
    
    # Fetch and add lyrics using the final enhanced metadata
    lyrics_text = None
    try:
        from meta_ops.lyrics import search_lyrics_with_fallbacks, add_lyrics_to_metadata
        if get_verbose_logging():
            print(f"\033[32m[LYRICS]\033[0m Searching for lyrics: {title} by {full_artist}")
        lyrics_text = search_lyrics_with_fallbacks(title, full_artist)
        
        if lyrics_text:
            if get_verbose_logging():
                print(f"\033[32m[LYRICS]\033[0m Found lyrics ({len(lyrics_text)} characters)")
        else:
            if get_verbose_logging():
                print("\033[33m[LYRICS]\033[0m No lyrics found for this song")
    except ImportError:
        if get_verbose_logging():
            print("\033[33m[LYRICS]\033[0m Lyrics functionality not available. Install lyricsgenius: pip install lyricsgenius")
    except Exception as e:
        if get_verbose_logging():
            print(f"\033[33m[LYRICS]\033[0m Error searching for lyrics: {e}")

    # Add metadata to the file
    print(f"\033[32m[META]\033[0m Adding metadata to {final_file}")
    try:
        # Get date from metadata if available
        date = spotify_metadata.get('release_date', '') if spotify_metadata else file_metadata.get('date', '')
        
        # Debug logging
        print(f"\033[32m[DEBUG]\033[0m Title: '{title}'")
        print(f"\033[32m[DEBUG]\033[0m Artist (for directory): '{artist}'")
        print(f"\033[32m[DEBUG]\033[0m Full Artist (for album_artist): '{full_artist}'")
        print(f"\033[32m[DEBUG]\033[0m Album: '{album}'")
        
        # Add metadata with both artist and album_artist
        add_metadata(
            file_path=final_file, 
            title=title, 
            artist=full_artist,  # Pass the full artist string to add_metadata
            album=album, 
            date=date, 
            album_artist=full_artist  # Explicitly set album_artist to the full artist string
        )
        
        # Add lyrics to the file if we found them
        if lyrics_text:
            try:
                if get_verbose_logging():
                    print("\033[32m[LYRICS]\033[0m Adding lyrics to metadata")
                if add_lyrics_to_metadata(final_file, lyrics_text):
                    if get_verbose_logging():
                        print("\033[32m[LYRICS]\033[0m Lyrics successfully embedded in file")
                else:
                    if get_verbose_logging():
                        print("\033[33m[LYRICS]\033[0m Failed to embed lyrics in file")
            except Exception as e:
                if get_verbose_logging():
                    print(f"\033[33m[LYRICS]\033[0m Error adding lyrics to file: {e}")
        
        print(f"\033[32m[COMPLETE]\033[0m Finished downloading and tagging: {final_file}")
        return True
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error adding metadata: {e}")
        return False

def run_playlist_with_spotify(query, output_file=None, music_dir=None, audio_format='opus'):
    """
    Download a YouTube playlist with Spotify metadata for each track.
    
    Args:
        query (str): The YouTube playlist URL
        output_file (str, optional): The output file path pattern. Defaults to '%(title)s.%(ext)s'.
        music_dir (str, optional): Base directory for music organization. Defaults to 'Music' in user's home.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
    
    Returns:
        bool: True if successful, False otherwise
    """
    # Import Spotify metadata module
    try:
        from meta_ops.spotify_metadata import search_metadata_with_fallback, apply_spotify_metadata_to_file
    except ImportError:
        error("Failed to import Spotify metadata module. Make sure spotipy is installed.")
        return False
    
    # Check if Spotify credentials are set
    if not os.environ.get('SPOTIFY_CLIENT_ID') or not os.environ.get('SPOTIFY_CLIENT_SECRET'):
        warning("Spotify credentials not found. Falling back to regular playlist download.")
        return run_playlist(query, output_file, music_dir, audio_format)
    
    if output_file is None:
        output_file = '%(title)s.%(ext)s'
    
    # Download the playlist
    try:
        downloaded_files = download_audio(query, output_file, False, audio_format)
        if not downloaded_files or not isinstance(downloaded_files, list):
            error("Playlist download failed", "FAIL")
            return False
    except Exception as e:
        error(f"Error downloading playlist: {e}")
        return False
    
    if get_verbose_logging():
        print(f"\033[32m[SUCCESS]\033[0m Downloaded {len(downloaded_files)} tracks from playlist")
    
    # Process each file in the playlist with Spotify metadata
    success_count = 0
    for file_path in downloaded_files:
        try:
            # Extract metadata from the downloaded file
            print(f"\033[32m[META]\033[0m Processing: {os.path.basename(file_path)}")
            file_metadata = extract_metadata(file_path)
            
            # Initialize title and artist variables
            title = ""
            artist = ""
            
            # If we have metadata from the file, use it
            if file_metadata['title'] and file_metadata['artist']:
                print(f"\033[32m[META]\033[0m Found metadata in file: Title='{file_metadata['title']}', Artist='{file_metadata['artist']}'")
                title = file_metadata['title']
                artist = file_metadata['artist']
            else:
                # Fallback: Parse title and artist from filename
                print(f"\033[33m[META]\033[0m No metadata found in file, using fallback method...")
                filename = os.path.basename(file_path)
                # Remove playlist index if present (e.g., "01 - ")
                if re.match(r'^\d+\s*-\s*', filename):
                    filename = re.sub(r'^\d+\s*-\s*', '', filename)
                
                name_without_ext = os.path.splitext(filename)[0]
                
                # Simple heuristic: try to split at " - " which is common in music filenames
                if " - " in name_without_ext:
                    artist, title = name_without_ext.split(" - ", 1)
                else:
                    # If no clear separator, use the whole name as title
                    title = name_without_ext
                    artist = "Unknown"
            
            # Search for metadata using Spotify with iTunes fallback
            print(f"\033[32m[METADATA]\033[0m Searching for metadata: {title} by {artist}")
            spotify_metadata = search_metadata_with_fallback(title, artist)
            
            if spotify_metadata:
                # Use enhanced metadata (Spotify + iTunes fallback)
                print(f"\033[32m[METADATA]\033[0m Found enhanced metadata for: {spotify_metadata['title']} by {spotify_metadata['artist']}")
                if spotify_metadata.get('genre'):
                    print(f"\033[32m[METADATA]\033[0m Genre: {spotify_metadata['genre']}")
                title = spotify_metadata['title']
                artist = spotify_metadata['artist']
                album = spotify_metadata['album']
            else:
                # Fallback to basic metadata
                print(f"\033[33m[WARNING]\033[0m No enhanced metadata found, using basic metadata")
                album = file_metadata.get("album", "") if file_metadata else (title or f"{artist}")
            
            # Create artist directory
            artist_dir = create_artist_directory(artist, music_dir)
            
            # Create album directory inside artist directory
            album_dir = create_album_directory(artist_dir, album, title)
            
            # Sanitize title for filename
            safe_title = sanitize_filename(title)
            
            # Determine file extension
            file_ext = os.path.splitext(file_path)[1]
            
            # Create the final file path (now in album directory)
            final_file = os.path.join(album_dir, f"{safe_title}{file_ext}")
            
            # If the file already exists, add a number to make it unique
            counter = 1
            while os.path.exists(final_file):
                final_file = os.path.join(album_dir, f"{safe_title}_{counter}{file_ext}")
                counter += 1
            
            # Move the file to the album directory
            if get_verbose_logging():
                print(f"\033[32m[FILE]\033[0m Moving file to: {final_file}")
            shutil.move(file_path, final_file)
            
            # Apply metadata to the file
            if spotify_metadata:
                # Apply Spotify metadata with all the enhanced fields
                metadata_success = apply_spotify_metadata_to_file(final_file, spotify_metadata)
                if metadata_success:
                    if get_verbose_logging():
                        print(f"\033[32m[COMPLETE]\033[0m Applied Spotify metadata to: {final_file}")
                    success_count += 1
                else:
                    error(f"Failed to apply Spotify metadata to: {final_file}")
            else:
                # Apply basic metadata
                try:
                    full_artist = artist
                    add_metadata(final_file, title=title, artist=artist, album=album, album_artist=full_artist)
                    print(f"\033[32m[COMPLETE]\033[0m Applied basic metadata to: {final_file}")
                    success_count += 1
                except Exception as e:
                    print(f"\033[31m[ERROR]\033[0m Error adding basic metadata: {e}")
                    
        except Exception as e:
            print(f"\033[31m[ERROR]\033[0m Error processing file {file_path}: {e}")
    
    # Clean up the temporary playlist directory
    temp_dir = os.path.join(os.getcwd(), 'temp_playlist')
    if os.path.exists(temp_dir):
        try:
            shutil.rmtree(temp_dir)
            print(f"\033[32m[CLEANUP]\033[0m Removed temporary directory: {temp_dir}")
        except Exception as e:
            print(f"\033[33m[WARNING]\033[0m Failed to remove temporary directory: {e}")
    
    print(f"\033[32m[SUMMARY]\033[0m Successfully processed {success_count} out of {len(downloaded_files)} tracks from the playlist with enhanced metadata")
    return success_count > 0

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
    
    # Check if the query is a YouTube playlist
    if is_url(query) and is_youtube_playlist(query):
        if get_verbose_logging():
            print("\033[32m[PLAYLIST]\033[0m Detected YouTube playlist URL")
            print("\033[32m[INFO]\033[0m Processing playlist using yt-dlp's playlist handling")
        # Don't use Spotify for playlists as it can cause issues
        result = run_playlist(query, args.output, music_dir, audio_format)
    elif use_spotify:
        result = run_with_spotify(query, args.output, music_dir, audio_format)
    else:
        result = run(query, args.output, music_dir, audio_format)
    
    return 0 if result else 1

if __name__ == "__main__":
    sys.exit(main())