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
    'default_output_pattern': '%(title)s.%(ext)s'
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