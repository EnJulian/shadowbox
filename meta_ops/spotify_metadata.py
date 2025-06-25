import os
import re
import requests
import urllib.parse
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
            
            # Initialize clean_title variable
            clean_title = title
            
            # First, try to extract core title before separator characters
            core_title = title
            separator_chars = ['~', '|', ':', '-', '–', '—', '•', '►', '♪', '♫']
            
            for separator in separator_chars:
                if separator in title:
                    # Split by separator and take the first part
                    potential_core = title.split(separator)[0].strip()
                    if potential_core and len(potential_core) > 3:  # Ensure it's not too short
                        core_title = potential_core
                        break
            
            # Try with core title and artist first if it's different from original
            if core_title != title and artist:
                print(f"\033[33m[SPOTIFY]\033[0m Trying with core title: {core_title} by {artist}")
                modified_results = sp.search(q=f"track:{core_title} artist:{artist}", type='track', limit=5)
                
                if modified_results['tracks']['items']:
                    # We found results with this combination
                    results = modified_results
                    # Continue with the rest of the function to process these results
                else:
                    # Try with just the core title (no artist)
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with just core title: {core_title}")
                    modified_results = sp.search(q=f"track:{core_title}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with title without brackets, parentheses, or common features text
            # Check if title contains any type of brackets or common feature indicators
            has_brackets = any(x in title for x in ['(', ')', '[', ']', '{', '}', '<', '>', 'feat.', 'ft.', 'featuring'])
            if has_brackets and not results['tracks']['items']:
                # Remove content within various types of brackets and common feature indicators
                # Remove parentheses content
                clean_title = re.sub(r'\s*\([^)]*\)', '', clean_title)
                # Remove square brackets content
                clean_title = re.sub(r'\s*\[[^\]]*\]', '', clean_title)
                # Remove curly braces content
                clean_title = re.sub(r'\s*\{[^}]*\}', '', clean_title)
                # Remove angle brackets content
                clean_title = re.sub(r'\s*<[^>]*>', '', clean_title)
                # Remove "feat." or "ft." and everything after it
                clean_title = re.sub(r'\s*(feat\.|ft\.|featuring).*', '', clean_title, flags=re.IGNORECASE)
                
                clean_title = clean_title.strip()
                
                if clean_title and clean_title != title:
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with clean title: {clean_title} by {artist}")
                    modified_results = sp.search(q=f"track:{clean_title} artist:{artist}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Try with just the clean title (no artist)
                        print(f"\033[33m[SPOTIFY]\033[0m Trying with just clean title: {clean_title}")
                        modified_results = sp.search(q=f"track:{clean_title}", type='track', limit=5)
                        
                        if modified_results['tracks']['items']:
                            # We found results with this combination
                            results = modified_results
                            # Continue with the rest of the function to process these results
                        else:
                            # Continue with other fallback strategies
                            pass
            
            # Try with track name and first word of artist if artist contains multiple words
            if artist and ' ' in artist and not results['tracks']['items']:
                first_word_of_artist = artist.split(' ')[0]
                print(f"\033[33m[SPOTIFY]\033[0m Trying with track and first word of artist: {title} - {first_word_of_artist}")
                modified_results = sp.search(q=f"track:{title} artist:{first_word_of_artist}", type='track', limit=5)
                
                if modified_results['tracks']['items']:
                    # We found results with this combination
                    results = modified_results
                    # Continue with the rest of the function to process these results
                else:
                    # Continue with other fallback strategies
                    pass
            
            # Initialize simplified_artist variable
            simplified_artist = artist if artist else ""
            
            # Try with simplified artist name (remove special characters) if artist contains special characters
            if artist and any(c in artist for c in "'-_&+.") and not results['tracks']['items']:
                simplified_artist = re.sub(r'[\'"\-_&+.]', '', artist)
                if simplified_artist and simplified_artist != artist:
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with simplified artist name: {title} by {simplified_artist}")
                    modified_results = sp.search(q=f"track:{title} artist:{simplified_artist}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with first 2 words of title and artist if title contains multiple words
            if artist and ' ' in title and not results['tracks']['items']:
                title_words = title.split(' ')
                if len(title_words) >= 2:
                    first_two_words_of_title = ' '.join(title_words[:2])
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with first 2 words of title and artist: {first_two_words_of_title} by {artist}")
                    modified_results = sp.search(q=f"track:{first_two_words_of_title} artist:{artist}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with first 3 words of title and artist if title contains multiple words
            if artist and ' ' in title and not results['tracks']['items']:
                title_words = title.split(' ')
                if len(title_words) >= 3:
                    first_three_words_of_title = ' '.join(title_words[:3])
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with first 3 words of title and artist: {first_three_words_of_title} by {artist}")
                    modified_results = sp.search(q=f"track:{first_three_words_of_title} artist:{artist}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with first word of title and artist if title contains multiple words
            if artist and ' ' in title and not results['tracks']['items']:
                first_word_of_title = title.split(' ')[0]
                print(f"\033[33m[SPOTIFY]\033[0m Trying with first word of title and artist: {first_word_of_title} by {artist}")
                modified_results = sp.search(q=f"track:{first_word_of_title} artist:{artist}", type='track', limit=5)
                
                if modified_results['tracks']['items']:
                    # We found results with this combination
                    results = modified_results
                    # Continue with the rest of the function to process these results
                else:
                    # Continue with other fallback strategies
                    pass
            
            # Try with both clean title and simplified artist if both are different from the originals
            if artist and not results['tracks']['items']:
                if clean_title != title and simplified_artist != artist:
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with clean title and simplified artist: {clean_title} by {simplified_artist}")
                    modified_results = sp.search(q=f"track:{clean_title} artist:{simplified_artist}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with just the first 2 words of the title (no artist)
            if ' ' in title and not results['tracks']['items']:
                title_words = title.split(' ')
                if len(title_words) >= 2:
                    first_two_words_of_title = ' '.join(title_words[:2])
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with first 2 words of title: {first_two_words_of_title}")
                    modified_results = sp.search(q=f"track:{first_two_words_of_title}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with just the first 3 words of the title (no artist)
            if ' ' in title and not results['tracks']['items']:
                title_words = title.split(' ')
                if len(title_words) >= 3:
                    first_three_words_of_title = ' '.join(title_words[:3])
                    print(f"\033[33m[SPOTIFY]\033[0m Trying with first 3 words of title: {first_three_words_of_title}")
                    modified_results = sp.search(q=f"track:{first_three_words_of_title}", type='track', limit=5)
                    
                    if modified_results['tracks']['items']:
                        # We found results with this combination
                        results = modified_results
                        # Continue with the rest of the function to process these results
                    else:
                        # Continue with other fallback strategies
                        pass
            
            # Try with just the first word of the title (no artist)
            if ' ' in title and not results['tracks']['items']:
                first_word_of_title = title.split(' ')[0]
                print(f"\033[33m[SPOTIFY]\033[0m Trying with just the first word of title: {first_word_of_title}")
                modified_results = sp.search(q=f"track:{first_word_of_title}", type='track', limit=5)
                
                if modified_results['tracks']['items']:
                    # We found results with this combination
                    results = modified_results
                    # Continue with the rest of the function to process these results
                else:
                    # Continue with other fallback strategies
                    pass
            
            # If all previous attempts failed and we have an artist, try with just the title
            if artist and not results['tracks']['items']:
                print(f"\033[33m[SPOTIFY]\033[0m Trying with just the title: {title}")
                return search_spotify_for_metadata(title, None)
            
            # If we have no artist and no results, return None
            if not results['tracks']['items']:
                return None
        
        # If we reach here, we have results either from the original search or from one of the fallback searches
        
        # Get the first track
        track = results['tracks']['items'][0]
        
        # Get additional album information for genre and total_discs
        album_id = track['album']['id']
        album_info = sp.album(album_id)
        
        # Calculate total discs by finding the highest disc number in the album
        total_discs = 1
        if 'tracks' in album_info and 'items' in album_info['tracks']:
            disc_numbers = [t.get('disc_number', 1) for t in album_info['tracks']['items']]
            total_discs = max(disc_numbers) if disc_numbers else 1
        
        # Extract genre from album
        genres = album_info.get('genres', [])
        genre = genres[0] if genres else None
        
        # Extract metadata
        metadata = {
            'title': track['name'],
            'artist': ', '.join([artist['name'] for artist in track['artists']]),
            'album': track['album']['name'],
            'release_date': track['album']['release_date'],
            'track_number': track['track_number'],
            'total_tracks': track['album']['total_tracks'],
            'disc_number': track.get('disc_number', 1),
            'total_discs': total_discs,
            'duration_ms': track['duration_ms'],
            'explicit': track['explicit'],
            'spotify_url': track['external_urls']['spotify'],
            'preview_url': track['preview_url'],
            'cover_url': None,
            'genre': genre,
            'composer': None,  # Spotify doesn't provide composer information
            'performer': ', '.join([artist['name'] for artist in track['artists']])
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

def search_itunes_for_metadata(title, artist=None):
    """
    Search iTunes API for metadata using the title and artist.
    
    Args:
        title (str): Song title
        artist (str, optional): Artist name
        
    Returns:
        dict or None: Dictionary with metadata if successful, None otherwise
    """
    try:
        # Build search query - use first word of artist if artist has multiple words
        if artist and ' ' in artist:
            search_artist = artist.split(' ')[0]
            search_query = f"{title} {search_artist}"
        elif artist:
            search_artist = artist
            search_query = f"{title} {artist}"
        else:
            search_artist = None
            search_query = title
        
        encoded_query = urllib.parse.quote(search_query)
        url = f"https://itunes.apple.com/search?term={encoded_query}&entity=song&limit=5"
        
        print(f"\033[32m[ITUNES]\033[0m Searching for: {search_query}")
        
        # Add a user agent to avoid potential blocks
        headers = {
            'User-Agent': 'shadowbox/0.1.0'
        }
        response = requests.get(url, headers=headers, timeout=10)
        
        # Check if the request was successful
        if response.status_code != 200:
            print(f"\033[33m[ITUNES]\033[0m iTunes API returned status code {response.status_code}")
            return None
        
        # Parse the JSON response
        data = response.json()
        
        # Check if we got any results
        if not data.get('results') or len(data['results']) == 0:
            print(f"\033[33m[ITUNES]\033[0m No results found for: {search_query}")
            
            # Initialize variables for fallback strategies
            clean_title = title
            found_results = False
            
            # Try with title without brackets, parentheses, or common features text
            has_brackets = any(x in title for x in ['(', ')', '[', ']', '{', '}', '<', '>', 'feat.', 'ft.', 'featuring'])
            if has_brackets and not found_results:
                # Remove content within various types of brackets and common feature indicators
                clean_title = re.sub(r'\s*\([^)]*\)', '', clean_title)
                clean_title = re.sub(r'\s*\[[^\]]*\]', '', clean_title)
                clean_title = re.sub(r'\s*\{[^}]*\}', '', clean_title)
                clean_title = re.sub(r'\s*<[^>]*>', '', clean_title)
                clean_title = re.sub(r'\s*(feat\.|ft\.|featuring).*', '', clean_title, flags=re.IGNORECASE)
                clean_title = clean_title.strip()
                
                if clean_title and clean_title != title:
                    print(f"\033[33m[ITUNES]\033[0m Trying with clean title: {clean_title}")
                    if artist:
                        clean_search_query = f"{clean_title} {artist}"
                    else:
                        clean_search_query = clean_title
                    
                    encoded_clean_query = urllib.parse.quote(clean_search_query)
                    clean_url = f"https://itunes.apple.com/search?term={encoded_clean_query}&entity=song&limit=5"
                    
                    clean_response = requests.get(clean_url, headers=headers, timeout=10)
                    if clean_response.status_code == 200:
                        clean_data = clean_response.json()
                        if clean_data.get('results') and len(clean_data['results']) > 0:
                            data = clean_data
                            found_results = True
            
            # Try with full artist name if we used first word in primary search
            if artist and ' ' in artist and not found_results:
                print(f"\033[33m[ITUNES]\033[0m Trying with full artist name: {title} - {artist}")
                fallback_search_query = f"{title} {artist}"
                
                encoded_fallback_query = urllib.parse.quote(fallback_search_query)
                fallback_url = f"https://itunes.apple.com/search?term={encoded_fallback_query}&entity=song&limit=5"
                
                fallback_response = requests.get(fallback_url, headers=headers, timeout=10)
                if fallback_response.status_code == 200:
                    fallback_data = fallback_response.json()
                    if fallback_data.get('results') and len(fallback_data['results']) > 0:
                        data = fallback_data
                        found_results = True
            
            # Initialize simplified_artist variable
            simplified_artist = artist if artist else ""
            
            # Try with simplified artist name (remove special characters) if artist contains special characters
            if artist and any(c in artist for c in "'-_&+.") and not found_results:
                simplified_artist = re.sub(r'[\'"\-_&+.]', '', artist)
                if simplified_artist and simplified_artist != artist:
                    print(f"\033[33m[ITUNES]\033[0m Trying with simplified artist name: {title} by {simplified_artist}")
                    simplified_search_query = f"{title} {simplified_artist}"
                    
                    encoded_simplified_query = urllib.parse.quote(simplified_search_query)
                    simplified_url = f"https://itunes.apple.com/search?term={encoded_simplified_query}&entity=song&limit=5"
                    
                    simplified_response = requests.get(simplified_url, headers=headers, timeout=10)
                    if simplified_response.status_code == 200:
                        simplified_data = simplified_response.json()
                        if simplified_data.get('results') and len(simplified_data['results']) > 0:
                            data = simplified_data
                            found_results = True
            
            # Try with both clean title and simplified artist if both are different from the originals
            if artist and not found_results:
                if clean_title != title and simplified_artist != artist:
                    print(f"\033[33m[ITUNES]\033[0m Trying with clean title and simplified artist: {clean_title} by {simplified_artist}")
                    combined_search_query = f"{clean_title} {simplified_artist}"
                    
                    encoded_combined_query = urllib.parse.quote(combined_search_query)
                    combined_url = f"https://itunes.apple.com/search?term={encoded_combined_query}&entity=song&limit=5"
                    
                    combined_response = requests.get(combined_url, headers=headers, timeout=10)
                    if combined_response.status_code == 200:
                        combined_data = combined_response.json()
                        if combined_data.get('results') and len(combined_data['results']) > 0:
                            data = combined_data
                            found_results = True
            
            # If we still don't have results after all fallback attempts, return None
            if not found_results:
                return None
        
        # Validate and get the best matching track
        track = None
        for result in data['results']:
            track_name = result.get('trackName', '').lower()
            artist_name = result.get('artistName', '').lower()
            
            # Check if this track matches our search criteria
            title_match = title.lower() in track_name or track_name in title.lower()
            
            # For artist matching, check against both full artist and first word
            artist_match = False
            if artist:
                artist_lower = artist.lower()
                first_word_artist = artist.split(' ')[0].lower() if ' ' in artist else artist_lower
                
                artist_match = (
                    artist_lower in artist_name or 
                    artist_name in artist_lower or
                    first_word_artist in artist_name or
                    artist_name in first_word_artist
                )
            else:
                artist_match = True  # If no artist specified, don't filter by artist
            
            if title_match and artist_match:
                track = result
                break
        
        # If no good match found, return None
        if not track:
            print(f"\033[33m[ITUNES]\033[0m No matching track found for: {title} by {search_artist or 'Unknown'}")
            return None
        
        # Extract metadata from iTunes response
        metadata = {
            'title': track.get('trackName', title),
            'artist': track.get('artistName', artist or ''),
            'album': track.get('collectionName', ''),
            'release_date': track.get('releaseDate', '').split('T')[0] if track.get('releaseDate') else None,
            'track_number': track.get('trackNumber'),
            'total_tracks': track.get('trackCount'),
            'disc_number': track.get('discNumber', 1),
            'total_discs': track.get('discCount', 1),
            'duration_ms': track.get('trackTimeMillis'),
            'explicit': track.get('trackExplicitness') == 'explicit',
            'itunes_url': track.get('trackViewUrl'),
            'preview_url': track.get('previewUrl'),
            'cover_url': None,
            'genre': track.get('primaryGenreName'),
            'composer': None,  # iTunes doesn't typically provide composer in search results
            'performer': track.get('artistName', artist or '')
        }
        
        # Get album cover URL (replace with higher resolution)
        artwork_url = track.get('artworkUrl100')
        if artwork_url:
            metadata['cover_url'] = artwork_url.replace('100x100bb', '600x600bb')
        
        print(f"\033[32m[ITUNES]\033[0m Found metadata for: {metadata['title']} by {metadata['artist']}")
        if metadata['genre']:
            print(f"\033[32m[ITUNES]\033[0m Genre found: {metadata['genre']}")
        
        return metadata
    
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error searching iTunes API: {e}")
        return None

def search_lastfm_for_metadata(title, artist=None):
    """
    Search Last.fm API for metadata using the title and artist.
    Last.fm has excellent genre data and is free to use.
    
    Args:
        title (str): Song title
        artist (str, optional): Artist name
        
    Returns:
        dict or None: Dictionary with metadata if successful, None otherwise
    """
    try:
        # Last.fm API endpoint for track.getInfo
        api_key = "b25b959554ed76058ac220b7b2e0a026"  # Public API key for Last.fm
        
        # Clean up artist name for better matching
        search_artist = artist
        
        if search_artist:
            # Try direct track.getinfo first with autocorrect enabled
            url = f"http://ws.audioscrobbler.com/2.0/?method=track.getinfo&api_key={api_key}&artist={urllib.parse.quote(search_artist)}&track={urllib.parse.quote(title)}&autocorrect=1&format=json"
            print(f"\033[32m[LASTFM]\033[0m Searching for: {title} by {search_artist}")
        else:
            # If no artist, try track search
            url = f"http://ws.audioscrobbler.com/2.0/?method=track.search&api_key={api_key}&track={urllib.parse.quote(title)}&format=json&limit=10"
            print(f"\033[32m[LASTFM]\033[0m Searching for: {title}")
        
        headers = {
            'User-Agent': 'shadowbox/0.1.0'
        }
        response = requests.get(url, headers=headers, timeout=10)
        
        if response.status_code != 200:
            print(f"\033[33m[LASTFM]\033[0m Last.fm API returned status code {response.status_code}")
            return None
        
        data = response.json()
        
        # Handle track.getinfo response
        if 'track' in data and data['track']:
            track = data['track']
            
            # Validate the track matches our search
            track_name = track.get('name', '').lower()
            artist_name = track.get('artist', {}).get('name', '').lower() if isinstance(track.get('artist'), dict) else str(track.get('artist', '')).lower()
            
            title_match = title.lower() in track_name or track_name in title.lower()
            artist_match = True
            
            if search_artist:
                search_artist_lower = search_artist.lower()
                artist_match = (
                    search_artist_lower in artist_name or 
                    artist_name in search_artist_lower
                )
            
            if title_match and artist_match:
                # Extract genre from tags - check multiple possible tag sources
                genre = None
                
                # Check toptags first
                if 'toptags' in track and 'tag' in track['toptags']:
                    tags = track['toptags']['tag']
                    if tags and len(tags) > 0:
                        # Get the top 2 tags as genre
                        if isinstance(tags, list):
                            genre_list = []
                            for i, tag in enumerate(tags[:2]):  # Take first 2 tags
                                tag_name = tag.get('name', '').title()
                                if tag_name:
                                    genre_list.append(tag_name)
                            genre = ', '.join(genre_list) if genre_list else None
                        else:
                            genre = tags.get('name', '').title()
                
                # If no genre from toptags, try to get artist tags
                if not genre and isinstance(track.get('artist'), dict) and 'name' in track['artist']:
                    artist_name_for_tags = track['artist']['name']
                    artist_url = f"http://ws.audioscrobbler.com/2.0/?method=artist.gettoptags&api_key={api_key}&artist={urllib.parse.quote(artist_name_for_tags)}&autocorrect=1&format=json"
                    try:
                        artist_response = requests.get(artist_url, headers=headers, timeout=10)
                        if artist_response.status_code == 200:
                            artist_data = artist_response.json()
                            if 'toptags' in artist_data and 'tag' in artist_data['toptags']:
                                artist_tags = artist_data['toptags']['tag']
                                if artist_tags and len(artist_tags) > 0:
                                    if isinstance(artist_tags, list):
                                        genre_list = []
                                        for i, tag in enumerate(artist_tags[:2]):  # Take first 2 tags
                                            tag_name = tag.get('name', '').title()
                                            if tag_name:
                                                genre_list.append(tag_name)
                                        genre = ', '.join(genre_list) if genre_list else None
                                    else:
                                        genre = artist_tags.get('name', '').title()
                    except Exception as e:
                        print(f"\033[33m[LASTFM]\033[0m Error fetching artist tags: {e}")
                
                metadata = {
                    'title': track.get('name', title),
                    'artist': artist_name.title() if artist_name else (artist or ''),
                    'album': track.get('album', {}).get('title', '') if isinstance(track.get('album'), dict) else '',
                    'genre': genre,
                    'lastfm_url': track.get('url'),
                    'duration_ms': int(track.get('duration', 0)) * 1000 if track.get('duration') else None
                }
                
                print(f"\033[32m[LASTFM]\033[0m Found metadata for: {metadata['title']} by {metadata['artist']}")
                if metadata['genre']:
                    print(f"\033[32m[LASTFM]\033[0m Genre found: {metadata['genre']}")
                else:
                    print(f"\033[33m[LASTFM]\033[0m No genre found in track or artist tags")
                
                return metadata
            else:
                print(f"\033[33m[LASTFM]\033[0m No matching track found for: {title} by {search_artist or 'Unknown'}")
                # Try fallback search if direct search failed
                if search_artist:
                    print(f"\033[33m[LASTFM]\033[0m Trying fallback search...")
                    return search_lastfm_fallback(title, search_artist, api_key, headers)
                return None
        
        # Handle track.search response
        elif 'results' in data and 'trackmatches' in data['results']:
            tracks = data['results']['trackmatches'].get('track', [])
            if not tracks:
                print(f"\033[33m[LASTFM]\033[0m No results found for: {title}")
                return None
            
            # Find best match
            for track in tracks if isinstance(tracks, list) else [tracks]:
                track_name = track.get('name', '').lower()
                artist_name = track.get('artist', '').lower()
                
                title_match = title.lower() in track_name or track_name in title.lower()
                
                if title_match:
                    # Get detailed info for this track with autocorrect
                    detail_url = f"http://ws.audioscrobbler.com/2.0/?method=track.getinfo&api_key={api_key}&artist={urllib.parse.quote(track.get('artist', ''))}&track={urllib.parse.quote(track.get('name', ''))}&autocorrect=1&format=json"
                    detail_response = requests.get(detail_url, headers=headers, timeout=10)
                    
                    if detail_response.status_code == 200:
                        detail_data = detail_response.json()
                        if 'track' in detail_data and detail_data['track']:
                            detail_track = detail_data['track']
                            
                            # Extract genre from tags
                            genre = None
                            if 'toptags' in detail_track and 'tag' in detail_track['toptags']:
                                tags = detail_track['toptags']['tag']
                                if tags and len(tags) > 0:
                                    # Get the top 2 tags as genre
                                    if isinstance(tags, list):
                                        genre_list = []
                                        for i, tag in enumerate(tags[:2]):  # Take first 2 tags
                                            tag_name = tag.get('name', '').title()
                                            if tag_name:
                                                genre_list.append(tag_name)
                                        genre = ', '.join(genre_list) if genre_list else None
                                    else:
                                        genre = tags.get('name', '').title()
                            
                            # If no genre from track tags, try artist tags
                            if not genre and isinstance(detail_track.get('artist'), dict) and 'name' in detail_track['artist']:
                                artist_name_for_tags = detail_track['artist']['name']
                                artist_url = f"http://ws.audioscrobbler.com/2.0/?method=artist.gettoptags&api_key={api_key}&artist={urllib.parse.quote(artist_name_for_tags)}&autocorrect=1&format=json"
                                try:
                                    artist_response = requests.get(artist_url, headers=headers, timeout=10)
                                    if artist_response.status_code == 200:
                                        artist_data = artist_response.json()
                                        if 'toptags' in artist_data and 'tag' in artist_data['toptags']:
                                            artist_tags = artist_data['toptags']['tag']
                                            if artist_tags and len(artist_tags) > 0:
                                                if isinstance(artist_tags, list):
                                                    genre_list = []
                                                    for i, tag in enumerate(artist_tags[:2]):  # Take first 2 tags
                                                        tag_name = tag.get('name', '').title()
                                                        if tag_name:
                                                            genre_list.append(tag_name)
                                                    genre = ', '.join(genre_list) if genre_list else None
                                                else:
                                                    genre = artist_tags.get('name', '').title()
                                except Exception as e:
                                    print(f"\033[33m[LASTFM]\033[0m Error fetching artist tags: {e}")
                            
                            metadata = {
                                'title': detail_track.get('name', title),
                                'artist': detail_track.get('artist', {}).get('name', track.get('artist', '')) if isinstance(detail_track.get('artist'), dict) else str(detail_track.get('artist', track.get('artist', ''))),
                                'album': detail_track.get('album', {}).get('title', '') if isinstance(detail_track.get('album'), dict) else '',
                                'genre': genre,
                                'lastfm_url': detail_track.get('url'),
                                'duration_ms': int(detail_track.get('duration', 0)) * 1000 if detail_track.get('duration') else None
                            }
                            
                            print(f"\033[32m[LASTFM]\033[0m Found metadata for: {metadata['title']} by {metadata['artist']}")
                            if metadata['genre']:
                                print(f"\033[32m[LASTFM]\033[0m Genre found: {metadata['genre']}")
                            
                            return metadata
            
            print(f"\033[33m[LASTFM]\033[0m No matching track found for: {title}")
            return None
        
        else:
            print(f"\033[33m[LASTFM]\033[0m No results found for: {title} by {search_artist or 'Unknown'}")
            # Try fallback search if direct search failed
            if search_artist:
                print(f"\033[33m[LASTFM]\033[0m Trying fallback search...")
                return search_lastfm_fallback(title, search_artist, api_key, headers)
            return None
    
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error searching Last.fm API: {e}")
        return None


def search_lastfm_fallback(title, artist, api_key, headers):
    """
    Fallback search method for Last.fm when direct search fails.
    Tries different search strategies.
    """
    try:
        # Try search with just the first word of artist name
        if ' ' in artist:
            first_word_artist = artist.split(' ')[0]
            url = f"http://ws.audioscrobbler.com/2.0/?method=track.getinfo&api_key={api_key}&artist={urllib.parse.quote(first_word_artist)}&track={urllib.parse.quote(title)}&autocorrect=1&format=json"
            print(f"\033[33m[LASTFM]\033[0m Trying with first word of artist: {first_word_artist}")
            
            response = requests.get(url, headers=headers, timeout=10)
            if response.status_code == 200:
                data = response.json()
                if 'track' in data and data['track']:
                    track = data['track']
                    
                    # Extract genre from tags
                    genre = None
                    if 'toptags' in track and 'tag' in track['toptags']:
                        tags = track['toptags']['tag']
                        if tags and len(tags) > 0:
                            if isinstance(tags, list):
                                genre_list = []
                                for i, tag in enumerate(tags[:2]):  # Take first 2 tags
                                    tag_name = tag.get('name', '').title()
                                    if tag_name:
                                        genre_list.append(tag_name)
                                genre = ', '.join(genre_list) if genre_list else None
                            else:
                                genre = tags.get('name', '').title()
                    
                    # If no genre from track tags, try artist tags
                    if not genre and isinstance(track.get('artist'), dict) and 'name' in track['artist']:
                        artist_name_for_tags = track['artist']['name']
                        artist_url = f"http://ws.audioscrobbler.com/2.0/?method=artist.gettoptags&api_key={api_key}&artist={urllib.parse.quote(artist_name_for_tags)}&autocorrect=1&format=json"
                        try:
                            artist_response = requests.get(artist_url, headers=headers, timeout=10)
                            if artist_response.status_code == 200:
                                artist_data = artist_response.json()
                                if 'toptags' in artist_data and 'tag' in artist_data['toptags']:
                                    artist_tags = artist_data['toptags']['tag']
                                    if artist_tags and len(artist_tags) > 0:
                                        if isinstance(artist_tags, list):
                                            genre_list = []
                                            for i, tag in enumerate(artist_tags[:2]):  # Take first 2 tags
                                                tag_name = tag.get('name', '').title()
                                                if tag_name:
                                                    genre_list.append(tag_name)
                                            genre = ', '.join(genre_list) if genre_list else None
                                        else:
                                            genre = artist_tags.get('name', '').title()
                        except Exception as e:
                            print(f"\033[33m[LASTFM]\033[0m Error fetching artist tags in fallback: {e}")
                    
                    artist_name = track.get('artist', {}).get('name', '') if isinstance(track.get('artist'), dict) else str(track.get('artist', ''))
                    
                    metadata = {
                        'title': track.get('name', title),
                        'artist': artist_name,
                        'album': track.get('album', {}).get('title', '') if isinstance(track.get('album'), dict) else '',
                        'genre': genre,
                        'lastfm_url': track.get('url'),
                        'duration_ms': int(track.get('duration', 0)) * 1000 if track.get('duration') else None
                    }
                    
                    print(f"\033[32m[LASTFM]\033[0m Fallback found metadata for: {metadata['title']} by {metadata['artist']}")
                    if metadata['genre']:
                        print(f"\033[32m[LASTFM]\033[0m Genre found: {metadata['genre']}")
                    
                    return metadata
        
        return None
    
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error in Last.fm fallback search: {e}")
        return None

def search_metadata_with_fallback(title, artist=None):
    """
    Search for metadata using Spotify with Last.fm for genre. iTunes completely removed.
    
    Args:
        title (str): Song title
        artist (str, optional): Artist name
        
    Returns:
        dict or None: Dictionary with metadata if successful, None otherwise
    """
    # Try Spotify for all metadata
    spotify_metadata = search_spotify_for_metadata(title, artist)
    
    # If Spotify fails completely, return None (no iTunes fallback)
    if not spotify_metadata:
        print(f"\033[33m[METADATA]\033[0m Spotify search failed, no metadata available")
        return None
    
    # Spotify succeeded - get Last.fm genre if Spotify has no genre
    if not spotify_metadata.get('genre'):
        print(f"\033[33m[METADATA]\033[0m Spotify metadata found but no genre, trying Last.fm for genre")
        lastfm_metadata = search_lastfm_for_metadata(title, artist)
        
        if lastfm_metadata and lastfm_metadata.get('genre'):
            print(f"\033[32m[METADATA]\033[0m Using Last.fm genre: {lastfm_metadata['genre']}")
            spotify_metadata['genre'] = lastfm_metadata['genre']
        else:
            print(f"\033[33m[METADATA]\033[0m Last.fm search failed, no genre available")
    
    return spotify_metadata

def apply_spotify_metadata_to_file(file_path, metadata, download_cover=True):
    """
    Apply enhanced metadata (Spotify + Last.fm for genre) to an audio file.
    
    Args:
        file_path (str): Path to the audio file
        metadata (dict): Metadata from Spotify with Last.fm genre
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
            temp_cover_path = os.path.join(os.getcwd(), f"temp_cover_{os.getpid()}.jpg")
            
            # Download the cover art
            if download_cover_image(metadata['cover_url'], temp_cover_path):
                cover_path = temp_cover_path
                print(f"\033[32m[SUCCESS]\033[0m Album cover downloaded for embedding")
            else:
                print(f"\033[33m[WARNING]\033[0m Failed to download cover art")
        
        # Apply metadata to the file
        # Save the full artist string as album_artist
        full_artist = metadata['artist']
        
        add_metadata(
            file_path=file_path,
            title=metadata['title'],
            artist=metadata['artist'],
            album=metadata['album'],
            cover_path=cover_path,
            date=metadata.get('release_date'),
            album_artist=full_artist,
            genre=metadata.get('genre'),
            track_number=metadata.get('track_number'),
            total_tracks=metadata.get('total_tracks'),
            disc_number=metadata.get('disc_number'),
            total_discs=metadata.get('total_discs'),
            composer=metadata.get('composer'),
            performer=metadata.get('performer')
        )
        
        print(f"\033[32m[SUCCESS]\033[0m Applied enhanced metadata to: {file_path}")
        
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
        tuple: (success, metadata) where success is a boolean indicating if the operation was successful,
               and metadata is a dictionary containing the Spotify metadata (or None if not found)
    """
    from meta_ops.downloader import download_from_youtube
    
    # Extract title and artist from YouTube URL
    title, artist = extract_title_from_youtube_url(youtube_url)
    
    if not title:
        print(f"\033[31m[ERROR]\033[0m Could not extract title from YouTube URL")
        return False
    
    print(f"\033[32m[INFO]\033[0m Extracted from YouTube: Title='{title}', Artist='{artist or 'Unknown'}'")
    
    # Search for metadata using Spotify with Last.fm for genre
    metadata = search_metadata_with_fallback(title, artist)
    
    if not metadata:
        print(f"\033[33m[WARNING]\033[0m Could not find metadata on Spotify. Proceeding with download only.")
        success = download_from_youtube(youtube_url, output_file, audio_format)
        return success, None
    
    # Download the audio
    success = download_from_youtube(youtube_url, output_file or '%(title)s.%(ext)s', audio_format)
    
    if not success:
        print(f"\033[31m[ERROR]\033[0m Failed to download audio from YouTube")
        return False, None
    
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
            return False, None
        
        # Sort by modification time (newest first)
        format_files.sort(key=lambda x: os.path.getmtime(x), reverse=True)
        downloaded_file = format_files[0]
    
    # Apply Spotify metadata to the downloaded file
    success = apply_spotify_metadata_to_file(downloaded_file, metadata)
    return success, metadata

def enhance_bandcamp_file_with_spotify(file_path):
    """
    Enhance a Bandcamp-downloaded file with additional Spotify metadata if available.
    
    Args:
        file_path (str): Path to the Bandcamp audio file
        
    Returns:
        tuple: (success, metadata) where success is a boolean indicating if the operation was successful,
               and metadata is a dictionary containing the Spotify metadata (or None if not found)
    """
    if not os.path.exists(file_path):
        print(f"\033[31m[ERROR]\033[0m File not found: {file_path}")
        return False, None
    
    # Extract existing metadata from the Bandcamp file
    from meta_ops.metadata import extract_metadata
    existing_metadata = extract_metadata(file_path)
    
    if not existing_metadata or not existing_metadata.get('title'):
        print(f"\033[33m[WARNING]\033[0m No existing metadata found in Bandcamp file")
        return False, None
    
    title = existing_metadata['title']
    artist = existing_metadata.get('artist', '')
    
    print(f"\033[32m[INFO]\033[0m Enhancing Bandcamp file: '{title}' by '{artist}'")
    
    # Search for additional metadata using Spotify with iTunes fallback
    spotify_metadata = search_metadata_with_fallback(title, artist)
    
    if not spotify_metadata:
        print(f"\033[33m[WARNING]\033[0m No enhanced metadata found for enhancement")
        return False, None
    
    # Check if the Bandcamp file already has the enhanced metadata fields
    needs_enhancement = False
    enhancement_fields = ['genre', 'disc_number', 'total_discs', 'performer']
    
    for field in enhancement_fields:
        if not existing_metadata.get(field) and spotify_metadata.get(field):
            needs_enhancement = True
            break
    
    if not needs_enhancement:
        print(f"\033[32m[INFO]\033[0m Bandcamp file already has comprehensive metadata")
        return True, existing_metadata
    
    # Merge existing Bandcamp metadata with Spotify enhancements
    enhanced_metadata = existing_metadata.copy()
    
    # Only add Spotify fields that are missing from Bandcamp metadata
    spotify_enhancements = {
        'genre': spotify_metadata.get('genre'),
        'disc_number': spotify_metadata.get('disc_number'),
        'total_discs': spotify_metadata.get('total_discs'),
        'performer': spotify_metadata.get('performer'),
        'track_number': spotify_metadata.get('track_number'),
        'total_tracks': spotify_metadata.get('total_tracks')
    }
    
    for field, value in spotify_enhancements.items():
        if value and not enhanced_metadata.get(field):
            enhanced_metadata[field] = value
    
    # Apply the enhanced metadata
    try:
        from meta_ops.metadata import add_metadata
        
        add_metadata(
            file_path=file_path,
            title=enhanced_metadata['title'],
            artist=enhanced_metadata['artist'],
            album=enhanced_metadata.get('album', ''),
            date=enhanced_metadata.get('date'),
            album_artist=enhanced_metadata.get('album_artist'),
            genre=enhanced_metadata.get('genre'),
            track_number=enhanced_metadata.get('track_number'),
            total_tracks=enhanced_metadata.get('total_tracks'),
            disc_number=enhanced_metadata.get('disc_number'),
            total_discs=enhanced_metadata.get('total_discs'),
            composer=enhanced_metadata.get('composer'),
            performer=enhanced_metadata.get('performer')
        )
        
        print(f"\033[32m[SUCCESS]\033[0m Enhanced Bandcamp file with Spotify metadata")
        return True, enhanced_metadata
        
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error enhancing Bandcamp metadata: {e}")
        return False, None

def enhance_existing_file_with_spotify(file_path, title=None, artist=None):
    """
    Enhance an existing audio file with Spotify metadata.
    
    Args:
        file_path (str): Path to the audio file
        title (str, optional): Song title. If None, tries to extract from filename.
        artist (str, optional): Artist name.
        
    Returns:
        tuple: (success, metadata) where success is a boolean indicating if the operation was successful,
               and metadata is a dictionary containing the Spotify metadata (or None if not found)
    """
    if not os.path.exists(file_path):
        print(f"\033[31m[ERROR]\033[0m File not found: {file_path}")
        return False, None
    
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
    
    # Search for metadata using Spotify with Last.fm for genre
    metadata = search_metadata_with_fallback(title, artist)
    
    if not metadata:
        print(f"\033[31m[ERROR]\033[0m Could not find metadata on Spotify")
        return False, None
    
    # Apply enhanced metadata to the file
    success = apply_spotify_metadata_to_file(file_path, metadata)
    return success, metadata