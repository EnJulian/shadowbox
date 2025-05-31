from yt_dlp import YoutubeDL
import os
import logging
import re
import subprocess
import sys

# Import enhanced terminal UI
try:
    # Try to import from parent directory (core)
    sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'core'))
    from terminal_ui import audio, scan, system, error, warning, success
except ImportError:
    # Fallback to basic print if terminal_ui is not available
    def audio(msg, tag="AUDIO"): print(f"\033[32m[{tag}]\033[0m {msg}")
    def scan(msg, tag="SCAN"): print(f"\033[37m[{tag}]\033[0m {msg}")
    def system(msg, tag="SYSTEM"): print(f"\033[32m[{tag}]\033[0m {msg}")
    def error(msg, tag="ERROR"): print(f"\033[31m[{tag}]\033[0m {msg}")
    def warning(msg, tag="WARNING"): print(f"\033[33m[{tag}]\033[0m {msg}")
    def success(msg, tag="SUCCESS"): print(f"\033[32m[{tag}]\033[0m {msg}")

def is_url(text):
    """
    Check if the text is a URL.
    
    Args:
        text (str): The text to check
        
    Returns:
        bool: True if the text is a URL, False otherwise
    """
    url_pattern = re.compile(
        r'^(?:http|https)://'  # http:// or https://
        r'(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|'  # domain
        r'localhost|'  # localhost
        r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})'  # or IP
        r'(?::\d+)?'  # optional port
        r'(?:/?|[/?]\S+)$', re.IGNORECASE)
    return bool(url_pattern.match(text))

def is_bandcamp_url(url):
    """
    Check if the URL is from Bandcamp.
    
    Args:
        url (str): The URL to check
        
    Returns:
        bool: True if the URL is from Bandcamp, False otherwise
    """
    return 'bandcamp.com' in url.lower()

def is_youtube_url(url):
    """
    Check if the URL is from YouTube.
    
    Args:
        url (str): The URL to check
        
    Returns:
        bool: True if the URL is from YouTube, False otherwise
    """
    return ('youtube.com' in url.lower() or 'youtu.be' in url.lower())

def is_youtube_playlist(url):
    """
    Check if the URL is a YouTube playlist.
    
    Args:
        url (str): The URL to check
        
    Returns:
        bool: True if the URL is a YouTube playlist, False otherwise
    """
    return is_youtube_url(url) and ('playlist' in url.lower() or 'list=' in url.lower())

def download_audio(query, output_file='%(title)s.%(ext)s', use_spotify_metadata=False, audio_format='opus'):
    """
    Download audio from YouTube or Bandcamp based on a search query or URL.
    
    Args:
        query (str): The search query (song title and artist) or URL
        output_file (str, optional): The output file path. Defaults to '%(title)s.%(ext)s'.
        use_spotify_metadata (bool, optional): Whether to use Spotify for metadata. Defaults to False.
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
    
    Returns:
        bool: True if successful, False otherwise
        If a playlist is downloaded, returns a list of downloaded file paths
    
    Raises:
        Exception: If download fails
    """
    # Ensure output directory exists
    output_dir = os.path.dirname(output_file)
    if output_dir and not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    # Check if query is a URL
    if is_url(query):
        # Use different command based on URL type
        if is_bandcamp_url(query):
            audio("Detected Bandcamp URL, using Bandcamp-specific settings...")
            return download_from_bandcamp(query, output_file, audio_format)
        elif is_youtube_playlist(query):
            audio("Detected YouTube playlist URL, downloading playlist...")
            return download_youtube_playlist(query, output_file, audio_format)
        elif is_youtube_url(query) and use_spotify_metadata:
            audio("Detected YouTube URL, using Spotify for metadata...")
            from meta_ops.spotify_metadata import process_youtube_url_with_spotify
            return process_youtube_url_with_spotify(query, output_file, audio_format)
        else:
            audio("Detected YouTube URL, using YouTube-specific settings...")
            return download_from_youtube(query, output_file, audio_format)
    else:
        # If it's not a URL, search on YouTube
        scan(f"Searching for: {query}")
        if use_spotify_metadata:
            from meta_ops.spotify_metadata import process_youtube_url_with_spotify
            return process_youtube_url_with_spotify(f"ytsearch1:{query}", output_file, audio_format)
        else:
            return download_from_youtube(f"ytsearch1:{query}", output_file, audio_format)

def download_from_youtube(url_or_query, output_file, audio_format='opus'):
    """
    Download audio from YouTube using specified settings.
    
    Args:
        url_or_query (str): The YouTube URL or search query
        output_file (str): The output file path
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
        
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        cmd = [
            "yt-dlp",
            "--downloader", "aria2c",
            "-x",
            "-f", "ba",
            "--embed-metadata",
            "--audio-format", audio_format,
            "-o", output_file,
            url_or_query
        ]
        
        print(f"\033[32m[GET]\033[0m Downloading audio using command: {' '.join(cmd)}")
        subprocess.run(cmd, check=True)
        
        # Check if file exists after download
        # Note: yt-dlp will replace %(title)s and %(ext)s with actual values
        # So we need to check if any files with the specified format were created
        output_dir = os.path.dirname(output_file)
        if not output_dir:
            output_dir = "."
        
        # Special case for ALAC format which gets converted to M4A by yt-dlp
        if audio_format.lower() == 'alac':
            m4a_files = [f for f in os.listdir(output_dir) if f.endswith('.m4a')]
            if m4a_files:
                print(f"\033[32m[SUCCESS]\033[0m Download complete (converted from ALAC to M4A): {m4a_files[0]}")
                return True
        
        # Check for files with the specified format
        format_files = [f for f in os.listdir(output_dir) if f.endswith(f'.{audio_format}')]
        if format_files:
            print(f"\033[32m[SUCCESS]\033[0m Download complete: {format_files[0]}")
            return True
        
        # If we're looking for ALAC but didn't find M4A files, show a more helpful error
        if audio_format.lower() == 'alac':
            print(f"\033[31m[FAIL]\033[0m No m4a files found after download (ALAC is converted to M4A)")
        else:
            print(f"\033[31m[FAIL]\033[0m No {audio_format} files found after download")
        return False
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error downloading from YouTube: {e}")
        return False

def download_youtube_playlist(url, output_file, audio_format='opus'):
    """
    Download a YouTube playlist using specified settings.
    
    Args:
        url (str): The YouTube playlist URL
        output_file (str): The output file path pattern
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
        
    Returns:
        list: List of downloaded file paths if successful, empty list if failed
    """
    try:
        # Create a temporary directory for playlist downloads
        temp_dir = os.path.join(os.getcwd(), 'temp_playlist')
        os.makedirs(temp_dir, exist_ok=True)
        
        # Modify output pattern to include playlist index
        playlist_output = os.path.join(temp_dir, "%(playlist_index)s - %(title)s.%(ext)s")
        
        # Use the exact command format that's known to work well with playlists
        cmd = [
            "yt-dlp",
            "-f", "bestaudio",
            "--extract-audio",
            "--audio-format", audio_format,
            "--audio-quality", "0",
            "--embed-metadata",
            "--embed-thumbnail",
            "--add-metadata",
            "--metadata-from-title", "%(artist)s - %(title)s",
            "--parse-metadata", "%(playlist)s:%(album)s",  # Set playlist name as album
            "-o", playlist_output,
            url
        ]
        
        print(f"\033[32m[GET]\033[0m Downloading playlist using command: {' '.join(cmd)}")
        subprocess.run(cmd, check=True)
        
        # Check if files exist after download
        downloaded_files = []
        
        # Special case for ALAC format which gets converted to M4A by yt-dlp
        if audio_format.lower() == 'alac':
            downloaded_files = [os.path.join(temp_dir, f) for f in os.listdir(temp_dir) if f.endswith('.m4a')]
        else:
            downloaded_files = [os.path.join(temp_dir, f) for f in os.listdir(temp_dir) if f.endswith(f'.{audio_format}')]
        
        if downloaded_files:
            print(f"\033[32m[SUCCESS]\033[0m Downloaded {len(downloaded_files)} files from playlist")
            # Sort files by playlist index (which should be at the start of the filename)
            downloaded_files.sort(key=lambda x: int(os.path.basename(x).split(' - ')[0]) if os.path.basename(x).split(' - ')[0].isdigit() else 999)
            return downloaded_files
        
        # If no files were found, show an error
        if audio_format.lower() == 'alac':
            print(f"\033[31m[FAIL]\033[0m No m4a files found after download (ALAC is converted to M4A)")
        else:
            print(f"\033[31m[FAIL]\033[0m No {audio_format} files found after download")
        return []
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error downloading YouTube playlist: {e}")
        return []

def download_from_bandcamp(url, output_file, audio_format='opus'):
    """
    Download audio from Bandcamp using specified settings.
    
    Args:
        url (str): The Bandcamp URL
        output_file (str): The output file path
        audio_format (str, optional): The audio format to download. Defaults to 'opus'.
                                     Supported formats: opus, m4a, mp3, flac, wav, etc.
        
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        cmd = [
            "yt-dlp",
            "--downloader", "aria2c",
            "--audio-quality", "0",
            "--audio-format", audio_format,
            "-x",
            "-f", "ba",
            "--embed-metadata",
            "--embed-thumbnail",
            "-o", output_file,
            url
        ]
        
        print(f"\033[32m[GET]\033[0m Downloading audio using command: {' '.join(cmd)}")
        subprocess.run(cmd, check=True)
        
        # Check if file exists after download
        # Note: yt-dlp will replace %(title)s and %(ext)s with actual values
        # So we need to check if any files with the specified format were created
        output_dir = os.path.dirname(output_file)
        if not output_dir:
            output_dir = "."
        
        # Special case for ALAC format which gets converted to M4A by yt-dlp
        if audio_format.lower() == 'alac':
            m4a_files = [f for f in os.listdir(output_dir) if f.endswith('.m4a')]
            if m4a_files:
                print(f"\033[32m[SUCCESS]\033[0m Download complete (converted from ALAC to M4A): {m4a_files[0]}")
                return True
        
        # Check for files with the specified format
        format_files = [f for f in os.listdir(output_dir) if f.endswith(f'.{audio_format}')]
        if format_files:
            print(f"\033[32m[SUCCESS]\033[0m Download complete: {format_files[0]}")
            return True
        
        # If we're looking for ALAC but didn't find M4A files, show a more helpful error
        if audio_format.lower() == 'alac':
            print(f"\033[31m[FAIL]\033[0m No m4a files found after download (ALAC is converted to M4A)")
        else:
            print(f"\033[31m[FAIL]\033[0m No {audio_format} files found after download")
        return False
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error downloading from Bandcamp: {e}")
        return False