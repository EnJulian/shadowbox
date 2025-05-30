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