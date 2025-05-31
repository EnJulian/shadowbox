import requests
import os
import urllib.parse
import time
import spotipy
from spotipy.oauth2 import SpotifyClientCredentials
import sys

# Import enhanced terminal UI
try:
    # Try to import from parent directory (core)
    sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'core'))
    from terminal_ui import api, warning, success, error
except ImportError:
    # Fallback to basic print if terminal_ui is not available
    def api(msg, tag="API"): print(f"\033[32m[{tag}]\033[0m {msg}")
    def warning(msg, tag="WARNING"): print(f"\033[33m[{tag}]\033[0m {msg}")
    def success(msg, tag="SUCCESS"): print(f"\033[32m[{tag}]\033[0m {msg}")
    def error(msg, tag="ERROR"): print(f"\033[31m[{tag}]\033[0m {msg}")

# Spotify API credentials will be loaded dynamically when needed
# This allows the credentials to be loaded after the module is imported

def get_album_cover_url(title, artist):
    """
    Get the album cover URL, trying Spotify first, then falling back to iTunes API.
    
    Args:
        title (str): Song title
        artist (str): Artist name
    
    Returns:
        str or None: URL of the album cover, or None if not found
    """
    # Get Spotify credentials dynamically
    spotify_client_id = os.environ.get('SPOTIFY_CLIENT_ID')
    spotify_client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET')
    
    # Debug information about credentials
    if spotify_client_id:
        api(f"Spotify credentials found: ID {spotify_client_id[:4]}...{spotify_client_id[-4:]}")
    
    # Try Spotify first (if credentials are available)
    if spotify_client_id and spotify_client_secret:
        api(f"Searching Spotify for album cover: '{title}' by '{artist}'")
        url = _search_spotify_api(title, artist, spotify_client_id, spotify_client_secret)
        if url:
            return url
        
        # If Spotify search fails with full artist name, try with just the first artist
        first_artist = artist.split(',')[0].strip() if ',' in artist else artist
        if first_artist != artist:
            warning(f"Spotify search failed, trying with first artist only: '{title}' by '{first_artist}'", "API")
            url = _search_spotify_api(title, first_artist, spotify_client_id, spotify_client_secret)
            if url:
                return url
        
        # If Spotify search fails, try iTunes with the same search query first
        warning(f"Spotify search failed, trying iTunes with same query: '{title}' by '{artist}'", "API")
        url = _search_itunes_api(f"{title} {artist}")
        if url:
            return url
        
        # If iTunes fails with full artist, try with just the first artist
        if first_artist != artist:
            warning(f"iTunes search failed, trying with first artist only: '{title}' by '{first_artist}'", "API")
            url = _search_itunes_api(f"{title} {first_artist}")
            if url:
                return url
            
        # If that fails, try with just the title on Spotify
        warning(f"iTunes search failed, trying Spotify with just the title: '{title}'", "API")
        url = _search_spotify_api(title, None, spotify_client_id, spotify_client_secret)
        if url:
            return url
    else:
        warning("Spotify credentials not found, skipping Spotify search", "API")
    
    # Fall back to iTunes API if not already tried
    if spotify_client_id and spotify_client_secret:
        # We already tried iTunes with title+artist above, so now try with just title
        warning(f"Trying iTunes with just the title: '{title}'", "API")
    else:
        # If Spotify was skipped, this is our first attempt with iTunes
        warning(f"Falling back to iTunes API for: '{title}' by '{artist}'", "API")
        # Try with both title and artist first
        url = _search_itunes_api(f"{title} {artist}")
        if url:
            return url
        
        # If iTunes fails with full artist, try with just the first artist
        first_artist = artist.split(',')[0].strip() if ',' in artist else artist
        if first_artist != artist:
            print(f"\033[33m[API]\033[0m iTunes search failed, trying with first artist only: '{title}' by '{first_artist}'")
            url = _search_itunes_api(f"{title} {first_artist}")
            if url:
                return url
        
        print(f"\033[33m[API]\033[0m iTunes search failed, trying with just the title: '{title}'")
    
    # Try with just the title
    url = _search_itunes_api(title)
    if url:
        return url
    
    # If that fails too, try with just the artist
    print(f"\033[33m[API]\033[0m iTunes search failed, trying with just the artist: '{artist}'")
    url = _search_itunes_api(artist)
    if url:
        return url
    
    # If all searches fail, return None
    return None

def _search_spotify_api(query, artist=None, client_id=None, client_secret=None):
    """
    Helper function to search Spotify API.
    
    Args:
        query (str): Search query (typically the song title)
        artist (str, optional): Artist name to refine search
        client_id (str): Spotify API client ID
        client_secret (str): Spotify API client secret
    
    Returns:
        str or None: URL of the album cover, or None if not found
    """
    try:
        # Initialize Spotify client
        sp = spotipy.Spotify(auth_manager=SpotifyClientCredentials(
            client_id=client_id,
            client_secret=client_secret
        ))
        
        # Build search query
        search_query = query
        if artist:
            search_query = f"track:{query} artist:{artist}"
        
        # Search for tracks
        results = sp.search(q=search_query, type='track', limit=1)
        
        # Check if we got any results
        if not results['tracks']['items']:
            return None
        
        # Get the album cover URL (highest resolution)
        track = results['tracks']['items'][0]
        images = track['album']['images']
        
        # Sort images by size (largest first)
        if images:
            images.sort(key=lambda x: x['width'] * x['height'], reverse=True)
            return images[0]['url']
        
        return None
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error searching Spotify API: {e}")
        return None

def _search_itunes_api(query):
    """
    Helper function to search iTunes API.
    
    Args:
        query (str): Search query
    
    Returns:
        str or None: URL of the album cover, or None if not found
    """
    encoded_query = urllib.parse.quote(query)
    url = f"https://itunes.apple.com/search?term={encoded_query}&entity=song&limit=1"
    
    try:
        # Add a user agent to avoid potential blocks
        headers = {
            'User-Agent': 'shadowbox/0.1.0'
        }
        response = requests.get(url, headers=headers, timeout=10)
        
        # Check if the request was successful
        if response.status_code != 200:
            print(f"\033[33m[API]\033[0m iTunes API returned status code {response.status_code}")
            return None
        
        # Parse the JSON response
        data = response.json()
        
        # Check if we got any results
        if not data.get('results') or len(data['results']) == 0:
            print(f"\033[33m[API]\033[0m No results found in iTunes API for query: '{query}'")
            return None
        
        # Get the artwork URL and replace with higher resolution
        artwork_url = data['results'][0].get('artworkUrl100')
        if artwork_url:
            # Replace with higher resolution image
            return artwork_url.replace('100x100bb', '600x600bb')
        
        return None
    except requests.exceptions.RequestException as e:
        print(f"\033[31m[ERROR]\033[0m Error fetching album cover URL: {e}")
        return None
    except ValueError as e:
        print(f"\033[31m[ERROR]\033[0m Error parsing JSON response: {e}")
        return None
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Unexpected error: {e}")
        return None

def download_cover_image(image_url, file_path='cover.jpg'):
    """
    Download an image from a URL and save it to a file.
    
    Args:
        image_url (str): URL of the image
        file_path (str, optional): Path to save the image. Defaults to 'cover.jpg'.
    
    Returns:
        str or None: Path to the saved image, or None if download failed
    """
    if not image_url:
        return None
    
    # Ensure the directory exists
    directory = os.path.dirname(file_path)
    if directory and not os.path.exists(directory):
        os.makedirs(directory)
    
    try:
        # Add a user agent to avoid potential blocks
        headers = {
            'User-Agent': 'shadowbox/0.1.0'
        }
        
        # Download the image with a timeout
        response = requests.get(image_url, headers=headers, timeout=10)
        
        # Check if the request was successful
        if response.status_code != 200:
            print(f"\033[31m[FAIL]\033[0m Image download failed with status code {response.status_code}")
            return None
        
        # Save the image to a file
        with open(file_path, 'wb') as f:
            f.write(response.content)
        
        # Verify the file was created
        if os.path.exists(file_path) and os.path.getsize(file_path) > 0:
            return file_path
        else:
            print("\033[31m[FAIL]\033[0m Image file was not created or is empty")
            return None
    except requests.exceptions.RequestException as e:
        print(f"\033[31m[ERROR]\033[0m Error downloading image: {e}")
        return None
    except IOError as e:
        print(f"\033[31m[ERROR]\033[0m Error saving image: {e}")
        return None
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Unexpected error: {e}")
        return None