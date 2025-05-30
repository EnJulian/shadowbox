import os
import re
import spotipy
from spotipy.oauth2 import SpotifyClientCredentials
from meta_ops.metadata import add_metadata
from meta_ops.cover_art import download_cover_image

def extract_title_from_youtube_url(url):
    """
    Extract the title from a YouTube URL using yt-dlp.
    
    Args:
        url (str): YouTube URL
        
    Returns:
        tuple: (title, artist) if successful, (None, None) otherwise
    """
    try:
        import subprocess
        import json
        
        # Use yt-dlp to get video info in JSON format
        cmd = [
            "yt-dlp",
            "-j",
            "--skip-download",
            url
        ]
        
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        video_info = json.loads(result.stdout)
        
        # Get the title
        title = video_info.get('title', '')
        
        # Try to extract artist from title
        artist = None
        
        # Common patterns in YouTube music titles
        patterns = [
            r'^(.*?)\s*-\s*(.*?)$',  # Artist - Title
            r'^(.*?)\s*–\s*(.*?)$',  # Artist – Title (en dash)
            r'^(.*?)\s*—\s*(.*?)$',  # Artist — Title (em dash)
            r'^(.*?)\s*:\s*(.*?)$',  # Artist: Title
            r'^(.*?)\s*\|\s*(.*?)$'  # Artist | Title
        ]
        
        for pattern in patterns:
            match = re.match(pattern, title)
            if match:
                artist = match.group(1).strip()
                title = match.group(2).strip()
                break
        
        # If no pattern matched, try to use the uploader as artist
        if not artist and 'uploader' in video_info:
            artist = video_info.get('uploader', '').replace(' - Topic', '')
        
        # Clean up title and artist
        if title:
            # Remove common suffixes like "(Official Video)", "[Official Music Video]", etc.
            title = re.sub(r'\s*[\(\[].*?(official|music|video|audio|lyrics|hd|4k).*?[\)\]]', '', title, flags=re.IGNORECASE)
            title = title.strip()
        
        if artist:
            artist = artist.strip()
        
        return title, artist
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error extracting title from YouTube URL: {e}")
        return None, None

def search_spotify_for_metadata(title, artist=None):
    """
    Search Spotify for metadata using the title and artist.
    
    Args:
        title (str): Song title
        artist (str, optional): Artist name
        
    Returns:
        dict or None: Dictionary with metadata if successful, None otherwise
    """
    # Get Spotify credentials dynamically
    spotify_client_id = os.environ.get('SPOTIFY_CLIENT_ID')
    spotify_client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')
    
    if not spotify_client_id or not spotify_client_secret:
        print("\033[33m[WARNING]\033[0m Spotify credentials not found. Run setup_spotify.sh first.")
        return None
    
    try:
        # Initialize Spotify client
        sp = spotipy.Spotify(auth_manager=SpotifyClientCredentials(
            client_id=spotify_client_id,
            client_secret=spotify_client_secret
        ))
        
        # Build search query
        search_query = title
        if artist:
            search_query = f"track:{title} artist:{artist}"
        
        print(f"\033[32m[SPOTIFY]\033[0m Searching for: {search_query}")
        
        # Search for tracks
        results = sp.search(q=search_query, type='track', limit=5)
        
        # Check if we got any results
        if not results['tracks']['items']:
            print(f"\033[33m[SPOTIFY]\033[0m No results found for: {search_query}")
            
            # Try with track name and first word of artist if artist contains multiple words
            if artist and ' ' in artist:
                first_word_of_artist = artist.split(' ')[0]
                print(f"\033[33m[SPOTIFY]\033[0m Trying with track and first word of artist: {title} - {first_word_of_artist}")
                modified_results = sp.search(q=f"track:{title} artist:{first_word_of_artist}", type='track', limit=5)
                
                if modified_results['tracks']['items']:
                    # We found results with this combination
                    results = modified_results
                    # Continue with the rest of the function to process these results
                else:
                    # Try with just the title if artist was provided
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with just the title: {title}")
                    return search_spotify_for_metadata(title, None)
            elif artist:
                # Try with just the title if artist was provided
                print(f"\033[33m[SPOTIFY]\033[0m Trying with just the title: {title}")
                return search_spotify_for_metadata(title, None)
            else:
                return None
        
        # If we reach here, we have results either from the original search or from one of the fallback searches
        
        # Get the first track
        track = results['tracks']['items'][0]
        
        # Extract metadata
        metadata = {
            'title': track['name'],
            'artist': ', '.join([artist['name'] for artist in track['artists']]),
            'album': track['album']['name'],
            'release_date': track['album']['release_date'],
            'track_number': track['track_number'],
            'total_tracks': track['album']['total_tracks'],
            'duration_ms': track['duration_ms'],
            'explicit': track['explicit'],
            'spotify_url': track['external_urls']['spotify'],
            'preview_url': track['preview_url'],
            'cover_url': None
        }
        
        # Get album cover URL (highest resolution)
        images = track['album']['images']
        if images:
            images.sort(key=lambda x: x['width'] * x['height'], reverse=True)
            metadata['cover_url'] = images[0]['url']
        
        print(f"\033[32m[SPOTIFY]\033[0m Found metadata for: {metadata['title']} by {metadata['artist']}")
        return metadata
    
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error searching Spotify API: {e}")
        return None

def apply_spotify_metadata_to_file(file_path, metadata, download_cover=True):
    """
    Apply Spotify metadata to an audio file.
    
    Args:
        file_path (str): Path to the audio file
        metadata (dict): Metadata from Spotify
        download_cover (bool, optional): Whether to download and embed cover art. Defaults to True.
        
    Returns:
        bool: True if successful, False otherwise
    """
    if not os.path.exists(file_path):
        print(f"\033[31m[ERROR]\033[0m File not found: {file_path}")
        return False
    
    if not metadata:
        print(f"\033[31m[ERROR]\033[0m No metadata provided")
        return False
    
    try:
        # Download cover art if requested
        cover_path = None
        if download_cover and metadata.get('cover_url'):
            # Create a temporary file for the cover art
            cover_dir = os.path.dirname(file_path)
            cover_filename = os.path.splitext(os.path.basename(file_path))[0] + "_cover.jpg"
            cover_path = os.path.join(cover_dir, cover_filename)
            
            # Download the cover art
            cover_path = download_cover_image(metadata['cover_url'], cover_path)
            if not cover_path:
                print(f"\033[33m[WARNING]\033[0m Failed to download cover art")
        
        # Apply metadata to the file
        add_metadata(
            file_path=file_path,
            title=metadata['title'],
            artist=metadata['artist'],
            album=metadata['album'],
            cover_path=cover_path,
            date=metadata.get('release_date')
        )
        
        print(f"\033[32m[SUCCESS]\033[0m Applied Spotify metadata to: {file_path}")
        
        # Clean up cover file if it was created
        if cover_path and os.path.exists(cover_path):
            os.remove(cover_path)
        
        return True
    
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error applying metadata: {e}")
        return False

def process_youtube_url_with_spotify(youtube_url, output_file=None, audio_format='opus'):
    """
    Process a YouTube URL by downloading it and applying Spotify metadata.
    
    Args:
        youtube_url (str): YouTube URL
        output_file (str, optional): Output file path. If None, uses the default yt-dlp naming.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
        
    Returns:
        bool: True if successful, False otherwise
    """
    from meta_ops.downloader import download_from_youtube
    
    # Extract title and artist from YouTube URL
    title, artist = extract_title_from_youtube_url(youtube_url)
    
    if not title:
        print(f"\033[31m[ERROR]\033[0m Could not extract title from YouTube URL")
        return False
    
    print(f"\033[32m[INFO]\033[0m Extracted from YouTube: Title='{title}', Artist='{artist or 'Unknown'}'")
    
    # Search Spotify for metadata
    metadata = search_spotify_for_metadata(title, artist)
    
    if not metadata:
        print(f"\033[33m[WARNING]\033[0m Could not find metadata on Spotify. Proceeding with download only.")
        return download_from_youtube(youtube_url, output_file, audio_format)
    
    # Download the audio
    success = download_from_youtube(youtube_url, output_file or '%(title)s.%(ext)s', audio_format)
    
    if not success:
        print(f"\033[31m[ERROR]\033[0m Failed to download audio from YouTube")
        return False
    
    # Find the downloaded file
    if output_file and '%' not in output_file:
        # If output_file is a specific file path without yt-dlp variables
        downloaded_file = output_file
    else:
        # Try to find the downloaded file
        output_dir = "."
        if output_file:
            output_dir = os.path.dirname(output_file) or "."
        
        # Look for recently created files with the specified format
        format_files = [os.path.join(output_dir, f) for f in os.listdir(output_dir) 
                     if f.endswith(f'.{audio_format}') and os.path.isfile(os.path.join(output_dir, f))]
        
        if not format_files:
            # If no files with the specified format, look for common audio formats
            for ext in ['opus', 'm4a', 'mp3', 'flac', 'wav', 'aac']:
                # Skip the format we already checked
                if ext == audio_format:
                    continue
                    
                format_files = [os.path.join(output_dir, f) for f in os.listdir(output_dir) 
                             if f.endswith(f'.{ext}') and os.path.isfile(os.path.join(output_dir, f))]
                if format_files:
                    break
        
        if not format_files:
            print(f"\033[31m[ERROR]\033[0m Could not find downloaded file")
            return False
        
        # Sort by modification time (newest first)
        format_files.sort(key=lambda x: os.path.getmtime(x), reverse=True)
        downloaded_file = format_files[0]
    
    # Apply Spotify metadata to the downloaded file
    return apply_spotify_metadata_to_file(downloaded_file, metadata)

def enhance_existing_file_with_spotify(file_path, title=None, artist=None):
    """
    Enhance an existing audio file with Spotify metadata.
    
    Args:
        file_path (str): Path to the audio file
        title (str, optional): Song title. If None, tries to extract from filename.
        artist (str, optional): Artist name.
        
    Returns:
        bool: True if successful, False otherwise
    """
    if not os.path.exists(file_path):
        print(f"\033[31m[ERROR]\033[0m File not found: {file_path}")
        return False
    
    # If title is not provided, try to extract from filename
    if not title:
        filename = os.path.splitext(os.path.basename(file_path))[0]
        
        # Try to extract title and artist from filename
        patterns = [
            r'^(.*?)\s*-\s*(.*?)$',  # Artist - Title
            r'^(.*?)\s*–\s*(.*?)$',  # Artist – Title (en dash)
            r'^(.*?)\s*—\s*(.*?)$',  # Artist — Title (em dash)
            r'^(.*?)\s*:\s*(.*?)$',  # Artist: Title
            r'^(.*?)\s*\|\s*(.*?)$'  # Artist | Title
        ]
        
        for pattern in patterns:
            match = re.match(pattern, filename)
            if match:
                if not artist:
                    artist = match.group(1).strip()
                title = match.group(2).strip()
                break
        
        # If no pattern matched, use the whole filename as title
        if not title:
            title = filename
    
    print(f"\033[32m[INFO]\033[0m Using title='{title}', artist='{artist or 'Unknown'}' for search")
    
    # Search Spotify for metadata
    metadata = search_spotify_for_metadata(title, artist)
    
    if not metadata:
        print(f"\033[31m[ERROR]\033[0m Could not find metadata on Spotify")
        return False
    
    # Apply Spotify metadata to the file
    return apply_spotify_metadata_to_file(file_path, metadata)