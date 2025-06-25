"""
Settings module for Shadowbox Music Downloader.
Handles saving and loading user preferences.
"""
import os
import json
import logging

# Default settings
DEFAULT_SETTINGS = {
    'audio_format': 'opus',
    'music_directory': os.path.join(os.path.expanduser('~'), 'Music'),
    'use_spotify': False,
    'use_genius': True,  # Enable lyrics by default if API key is available
    'default_output_pattern': '%(title)s.%(ext)s',
    'verbose_logging': False,
    'theme': 'hacker',
    'genius_access_token': '',  # Store Genius API token
    'spotify_client_id': '',    # Store Spotify credentials in settings too
    'spotify_client_secret': ''
}

# Settings file path
SETTINGS_FILE = os.path.join(os.path.expanduser('~'), '.shadowbox_settings.json')

def load_settings():
    """
    Load settings from the settings file.
    
    Returns:
        dict: The loaded settings, or default settings if the file doesn't exist
    """
    try:
        if os.path.exists(SETTINGS_FILE):
            with open(SETTINGS_FILE, 'r') as f:
                settings = json.load(f)
                
            # Ensure all default settings exist
            for key, value in DEFAULT_SETTINGS.items():
                if key not in settings:
                    settings[key] = value
                    
            return settings
        else:
            return DEFAULT_SETTINGS.copy()
    except Exception as e:
        logging.error(f"Error loading settings: {e}")
        return DEFAULT_SETTINGS.copy()

def save_settings(settings):
    """
    Save settings to the settings file.
    
    Args:
        settings (dict): The settings to save
        
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        with open(SETTINGS_FILE, 'w') as f:
            json.dump(settings, f, indent=4)
        return True
    except Exception as e:
        logging.error(f"Error saving settings: {e}")
        return False

def get_setting(key, default=None):
    """
    Get a specific setting.
    
    Args:
        key (str): The setting key
        default: The default value to return if the key doesn't exist
        
    Returns:
        The setting value, or the default if the key doesn't exist
    """
    settings = load_settings()
    return settings.get(key, default)

def set_setting(key, value):
    """
    Set a specific setting.
    
    Args:
        key (str): The setting key
        value: The setting value
        
    Returns:
        bool: True if successful, False otherwise
    """
    settings = load_settings()
    settings[key] = value
    return save_settings(settings)

def get_audio_format():
    """
    Get the preferred audio format.
    
    Returns:
        str: The preferred audio format
    """
    return get_setting('audio_format', DEFAULT_SETTINGS['audio_format'])

def set_audio_format(format_name):
    """
    Set the preferred audio format.
    
    Args:
        format_name (str): The audio format name
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('audio_format', format_name)

def get_music_directory():
    """
    Get the preferred music directory.
    
    Returns:
        str: The preferred music directory
    """
    return get_setting('music_directory', DEFAULT_SETTINGS['music_directory'])

def set_music_directory(directory):
    """
    Set the preferred music directory.
    
    Args:
        directory (str): The music directory path
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('music_directory', directory)

def get_use_spotify():
    """
    Get whether to use Spotify for metadata.
    
    Returns:
        bool: True if Spotify should be used, False otherwise
    """
    return get_setting('use_spotify', DEFAULT_SETTINGS['use_spotify'])

def set_use_spotify(use_spotify):
    """
    Set whether to use Spotify for metadata.
    
    Args:
        use_spotify (bool): Whether to use Spotify
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('use_spotify', use_spotify)

def get_verbose_logging():
    """
    Get whether verbose logging is enabled.
    
    Returns:
        bool: True if verbose logging is enabled, False otherwise
    """
    return get_setting('verbose_logging', DEFAULT_SETTINGS['verbose_logging'])

def set_verbose_logging(verbose):
    """
    Set whether to enable verbose logging.
    
    Args:
        verbose (bool): Whether to enable verbose logging
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('verbose_logging', verbose)

def get_theme():
    """
    Get the preferred UI theme.
    
    Returns:
        str: The preferred theme name
    """
    return get_setting('theme', DEFAULT_SETTINGS['theme'])

def set_theme(theme_name):
    """
    Set the preferred UI theme.
    
    Args:
        theme_name (str): The theme name
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('theme', theme_name)

def get_use_genius():
    """
    Get whether to use Genius for lyrics.
    
    Returns:
        bool: True if Genius should be used, False otherwise
    """
    return get_setting('use_genius', DEFAULT_SETTINGS['use_genius'])

def set_use_genius(use_genius):
    """
    Set whether to use Genius for lyrics.
    
    Args:
        use_genius (bool): Whether to use Genius
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('use_genius', use_genius)

def get_genius_access_token():
    """
    Get the Genius API access token.
    
    Returns:
        str: The Genius access token, or empty string if not set
    """
    # First check environment variable, then settings
    env_token = os.environ.get('GENIUS_ACCESS_TOKEN', '')
    if env_token:
        return env_token
    return get_setting('genius_access_token', DEFAULT_SETTINGS['genius_access_token'])

def set_genius_access_token(token):
    """
    Set the Genius API access token.
    
    Args:
        token (str): The Genius access token
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('genius_access_token', token)

def get_spotify_client_id():
    """
    Get the Spotify client ID.
    
    Returns:
        str: The Spotify client ID, or empty string if not set
    """
    # First check environment variable, then settings
    env_id = os.environ.get('SPOTIFY_CLIENT_ID', '')
    if env_id:
        return env_id
    return get_setting('spotify_client_id', DEFAULT_SETTINGS['spotify_client_id'])

def set_spotify_client_id(client_id):
    """
    Set the Spotify client ID.
    
    Args:
        client_id (str): The Spotify client ID
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('spotify_client_id', client_id)

def get_spotify_client_secret():
    """
    Get the Spotify client secret.
    
    Returns:
        str: The Spotify client secret, or empty string if not set
    """
    # First check environment variable, then settings
    env_secret = os.environ.get('SPOTIFY_CLIENT_SECRET', '')
    if env_secret:
        return env_secret
    return get_setting('spotify_client_secret', DEFAULT_SETTINGS['spotify_client_secret'])

def set_spotify_client_secret(client_secret):
    """
    Set the Spotify client secret.
    
    Args:
        client_secret (str): The Spotify client secret
        
    Returns:
        bool: True if successful, False otherwise
    """
    return set_setting('spotify_client_secret', client_secret)

def update_ytdlp():
    """
    Update yt-dlp to the latest version to fix YouTube download issues.
    
    Returns:
        bool: True if successful, False otherwise
    """
    import subprocess
    import sys
    
    try:
        print("\033[33m[UPDATE]\033[0m Updating yt-dlp to latest version...")
        
        # Try pip3 first, then pip
        for pip_cmd in ['pip3', 'pip']:
            try:
                result = subprocess.run([pip_cmd, 'install', '--upgrade', 'yt-dlp'], 
                                      check=True, capture_output=True, text=True)
                print(f"\033[32m[SUCCESS]\033[0m yt-dlp updated successfully using {pip_cmd}")
                print(f"Output: {result.stdout}")
                return True
            except (subprocess.CalledProcessError, FileNotFoundError):
                continue
        
        # If both pip commands failed, try with python -m pip
        try:
            result = subprocess.run([sys.executable, '-m', 'pip', 'install', '--upgrade', 'yt-dlp'], 
                                  check=True, capture_output=True, text=True)
            print(f"\033[32m[SUCCESS]\033[0m yt-dlp updated successfully using python -m pip")
            print(f"Output: {result.stdout}")
            return True
        except subprocess.CalledProcessError as e:
            print(f"\033[31m[ERROR]\033[0m Failed to update yt-dlp: {e.stderr}")
            return False
            
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Unexpected error updating yt-dlp: {e}")
        return False

def check_ytdlp_version():
    """
    Check the current version of yt-dlp.
    
    Returns:
        str: The version string, or None if yt-dlp is not installed
    """
    import subprocess
    
    try:
        result = subprocess.run(['yt-dlp', '--version'], 
                              check=True, capture_output=True, text=True)
        version = result.stdout.strip()
        print(f"\033[32m[INFO]\033[0m Current yt-dlp version: {version}")
        return version
    except (subprocess.CalledProcessError, FileNotFoundError):
        print(f"\033[31m[ERROR]\033[0m yt-dlp is not installed or not found in PATH")
        return None