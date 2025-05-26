from yt_dlp import YoutubeDL
import os
import logging
import re
import subprocess

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

def download_audio(query, output_file='%(title)s.%(ext)s'):
    """
    Download audio from YouTube or Bandcamp based on a search query or URL.
    
    Args:
        query (str): The search query (song title and artist) or URL
        output_file (str, optional): The output file path. Defaults to '%(title)s.%(ext)s'.
    
    Returns:
        bool: True if successful, False otherwise
    
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
            print("🎵 Detected Bandcamp URL, using Bandcamp-specific settings...")
            return download_from_bandcamp(query, output_file)
        else:
            print("🎵 Detected YouTube URL, using YouTube-specific settings...")
            return download_from_youtube(query, output_file)
    else:
        # If it's not a URL, search on YouTube
        print(f"🔍 Searching for: {query}")
        return download_from_youtube(f"ytsearch1:{query}", output_file)

def download_from_youtube(url_or_query, output_file):
    """
    Download audio from YouTube using specified settings.
    
    Args:
        url_or_query (str): The YouTube URL or search query
        output_file (str): The output file path
        
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
            "--audio-format", "opus",
            "-o", output_file,
            url_or_query
        ]
        
        print(f"📥 Downloading audio using command: {' '.join(cmd)}")
        subprocess.run(cmd, check=True)
        
        # Check if file exists after download
        # Note: yt-dlp will replace %(title)s and %(ext)s with actual values
        # So we need to check if any opus file was created in the directory
        if output_dir := os.path.dirname(output_file):
            if not output_dir:
                output_dir = "."
            
            # Check if any opus files were created
            opus_files = [f for f in os.listdir(output_dir) if f.endswith('.opus')]
            if opus_files:
                print(f"✅ Download complete: {opus_files[0]}")
                return True
        else:
            opus_files = [f for f in os.listdir('.') if f.endswith('.opus')]
            if opus_files:
                print(f"✅ Download complete: {opus_files[0]}")
                return True
                
        print("❌ No opus files found after download")
        return False
    except Exception as e:
        print(f"❌ Error downloading from YouTube: {e}")
        return False

def download_from_bandcamp(url, output_file):
    """
    Download audio from Bandcamp using specified settings.
    
    Args:
        url (str): The Bandcamp URL
        output_file (str): The output file path
        
    Returns:
        bool: True if successful, False otherwise
    """
    try:
        cmd = [
            "yt-dlp",
            "--downloader", "aria2c",
            "--audio-quality", "0",
            "--audio-format", "opus",
            "-x",
            "-f", "ba",
            "--embed-metadata",
            "--embed-thumbnail",
            "-o", output_file,
            url
        ]
        
        print(f"📥 Downloading audio using command: {' '.join(cmd)}")
        subprocess.run(cmd, check=True)
        
        # Check if file exists after download
        # Note: yt-dlp will replace %(title)s and %(ext)s with actual values
        # So we need to check if any opus file was created in the directory
        if output_dir := os.path.dirname(output_file):
            if not output_dir:
                output_dir = "."
            
            # Check if any opus files were created
            opus_files = [f for f in os.listdir(output_dir) if f.endswith('.opus')]
            if opus_files:
                print(f"✅ Download complete: {opus_files[0]}")
                return True
        else:
            opus_files = [f for f in os.listdir('.') if f.endswith('.opus')]
            if opus_files:
                print(f"✅ Download complete: {opus_files[0]}")
                return True
                
        print("❌ No opus files found after download")
        return False
    except Exception as e:
        print(f"❌ Error downloading from Bandcamp: {e}")
        return False