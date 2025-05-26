import requests
import os
import urllib.parse
import time

def get_album_cover_url(title, artist):
    """
    Get the album cover URL from iTunes API.
    
    Args:
        title (str): Song title
        artist (str): Artist name
    
    Returns:
        str or None: URL of the album cover, or None if not found
    """
    # Try with both title and artist first
    url = _search_itunes_api(f"{title} {artist}")
    if url:
        return url
    
    # If that fails, try with just the title
    print(f"\033[33m[API]\033[0m First search failed, trying with just the title: '{title}'")
    url = _search_itunes_api(title)
    if url:
        return url
    
    # If that fails too, try with just the artist
    print(f"\033[33m[API]\033[0m Second search failed, trying with just the artist: '{artist}'")
    url = _search_itunes_api(artist)
    if url:
        return url
    
    # If all searches fail, return None
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