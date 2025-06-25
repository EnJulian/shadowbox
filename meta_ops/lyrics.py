import os
import re
import time
from typing import Optional, Tuple

def get_lyrics_from_genius(title: str, artist: str) -> Optional[str]:
    """
    Fetch lyrics from Genius API using the title and artist.
    
    Args:
        title (str): Song title
        artist (str): Artist name
        
    Returns:
        str or None: Lyrics text if successful, None otherwise
    """
    # Get Genius API token from settings or environment
    try:
        from settings import get_genius_access_token, get_use_genius
        
        # Check if lyrics are enabled in settings
        if not get_use_genius():
            return None
            
        genius_token = get_genius_access_token()
    except ImportError:
        # Fallback to environment variable if settings module not available
        genius_token = os.environ.get('GENIUS_ACCESS_TOKEN')
    
    if not genius_token:
        print("\033[33m[WARNING]\033[0m Genius API token not found. Configure it in settings or run setup_genius.sh first to enable lyrics.")
        return None
    
    try:
        import lyricsgenius
        
        # Initialize Genius client
        genius = lyricsgenius.Genius(genius_token)
        
        # Configure client settings
        genius.verbose = False  # Turn off status messages
        genius.remove_section_headers = True  # Remove section headers like [Verse 1]
        genius.skip_non_songs = True  # Skip non-song results
        genius.excluded_terms = ["(Remix)", "(Live)", "(Acoustic)", "(Demo)"]  # Skip certain versions
        
        print(f"\033[32m[GENIUS]\033[0m Searching for lyrics: {title} by {artist}")
        
        # Search for the song
        song = genius.search_song(title, artist)
        
        if song and song.lyrics:
            # Clean up the lyrics
            lyrics = song.lyrics
            
            # Remove the title and artist info that Genius adds at the beginning
            # Pattern: "Title by Artist Lyrics" or similar
            lyrics = re.sub(r'^.*?Lyrics\s*\n', '', lyrics, flags=re.IGNORECASE)
            
            # Remove "You might also like" sections
            lyrics = re.sub(r'You might also like.*?\n', '', lyrics, flags=re.IGNORECASE)
            
            # Remove embed info at the end
            lyrics = re.sub(r'\d+Embed$', '', lyrics)
            
            # Clean up extra whitespace
            lyrics = re.sub(r'\n\s*\n\s*\n', '\n\n', lyrics)  # Replace multiple newlines with double newlines
            lyrics = lyrics.strip()
            
            if lyrics:
                print(f"\033[32m[GENIUS]\033[0m Successfully found lyrics ({len(lyrics)} characters)")
                return lyrics
            else:
                print(f"\033[33m[GENIUS]\033[0m Found song but lyrics were empty")
                return None
        else:
            print(f"\033[33m[GENIUS]\033[0m No lyrics found for: {title} by {artist}")
            return None
            
    except ImportError:
        print("\033[33m[WARNING]\033[0m lyricsgenius library not installed. Install it with: pip install lyricsgenius")
        return None
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Error fetching lyrics from Genius: {e}")
        return None

def search_lyrics_with_fallbacks(title: str, artist: str) -> Optional[str]:
    """
    Search for lyrics using multiple strategies and fallbacks.
    
    Args:
        title (str): Song title
        artist (str): Artist name
        
    Returns:
        str or None: Lyrics text if successful, None otherwise
    """
    # Try with original title and artist
    lyrics = get_lyrics_from_genius(title, artist)
    if lyrics:
        return lyrics
    
    # Try with cleaned title (remove brackets, parentheses, etc.)
    clean_title = title
    
    # Remove content within various types of brackets
    clean_title = re.sub(r'\s*\([^)]*\)', '', clean_title)  # Remove parentheses content
    clean_title = re.sub(r'\s*\[[^\]]*\]', '', clean_title)  # Remove square brackets content
    clean_title = re.sub(r'\s*\{[^}]*\}', '', clean_title)  # Remove curly braces content
    clean_title = re.sub(r'\s*(feat\.|ft\.|featuring).*', '', clean_title, flags=re.IGNORECASE)  # Remove features
    clean_title = clean_title.strip()
    
    if clean_title != title and clean_title:
        print(f"\033[33m[GENIUS]\033[0m Trying with cleaned title: {clean_title}")
        lyrics = get_lyrics_from_genius(clean_title, artist)
        if lyrics:
            return lyrics
    
    # Try with simplified artist (first artist only if multiple)
    if ',' in artist:
        first_artist = artist.split(',')[0].strip()
        print(f"\033[33m[GENIUS]\033[0m Trying with first artist only: {first_artist}")
        lyrics = get_lyrics_from_genius(title, first_artist)
        if lyrics:
            return lyrics
        
        # Try cleaned title with first artist
        if clean_title != title and clean_title:
            lyrics = get_lyrics_from_genius(clean_title, first_artist)
            if lyrics:
                return lyrics
    
    # Try with core title (before separator characters)
    core_title = title
    separator_chars = ['~', '|', ':', '-', '–', '—', '•', '►', '♪', '♫']
    
    for separator in separator_chars:
        if separator in title:
            potential_core = title.split(separator)[0].strip()
            if potential_core and len(potential_core) > 3:
                core_title = potential_core
                break
    
    if core_title != title and core_title != clean_title:
        print(f"\033[33m[GENIUS]\033[0m Trying with core title: {core_title}")
        lyrics = get_lyrics_from_genius(core_title, artist)
        if lyrics:
            return lyrics
    
    return None

def add_lyrics_to_metadata(file_path: str, lyrics: str) -> bool:
    """
    Add lyrics to the metadata of an audio file.
    
    Args:
        file_path (str): Path to the audio file
        lyrics (str): Lyrics text to add
        
    Returns:
        bool: True if successful, False otherwise
    """
    if not os.path.exists(file_path):
        print(f"\033[31m[ERROR]\033[0m Audio file not found: {file_path}")
        return False
    
    if not lyrics or not lyrics.strip():
        print(f"\033[33m[WARNING]\033[0m No lyrics provided to add to metadata")
        return False
    
    # Determine file type based on extension
    file_ext = os.path.splitext(file_path)[1].lower()
    
    try:
        if file_ext == '.mp3':
            return add_lyrics_to_mp3(file_path, lyrics)
        elif file_ext == '.opus':
            return add_lyrics_to_opus(file_path, lyrics)
        elif file_ext == '.m4a':
            return add_lyrics_to_m4a(file_path, lyrics)
        else:
            print(f"\033[33m[WARNING]\033[0m Unsupported file format for lyrics: {file_ext}")
            return False
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Failed to add lyrics to {file_ext}: {e}")
        return False

def add_lyrics_to_mp3(file_path: str, lyrics: str) -> bool:
    """Add lyrics to MP3 file using USLT (Unsynchronized Lyrics) tag."""
    try:
        from mutagen.mp3 import MP3
        from mutagen.id3 import ID3, USLT
        
        audio = MP3(file_path)
        
        # Make sure we have ID3 tags
        if not audio.tags:
            audio.add_tags()
        
        # Add unsynchronized lyrics
        audio.tags.add(
            USLT(
                encoding=3,  # UTF-8
                lang='eng',  # Language code
                desc='',     # Description
                text=lyrics
            )
        )
        
        audio.save()
        print(f"\033[32m[LYRICS]\033[0m Added lyrics to MP3 file")
        return True
        
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Failed to add lyrics to MP3: {e}")
        return False

def add_lyrics_to_opus(file_path: str, lyrics: str) -> bool:
    """Add lyrics to Opus file using LYRICS Vorbis comment."""
    try:
        from mutagen.oggopus import OggOpus
        
        audio = OggOpus(file_path)
        audio['LYRICS'] = lyrics
        audio.save()
        
        print(f"\033[32m[LYRICS]\033[0m Added lyrics to Opus file")
        return True
        
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Failed to add lyrics to Opus: {e}")
        return False

def add_lyrics_to_m4a(file_path: str, lyrics: str) -> bool:
    """Add lyrics to M4A file using ©lyr atom."""
    try:
        from mutagen.mp4 import MP4
        
        audio = MP4(file_path)
        audio['©lyr'] = [lyrics]
        audio.save()
        
        print(f"\033[32m[LYRICS]\033[0m Added lyrics to M4A file")
        return True
        
    except Exception as e:
        print(f"\033[31m[ERROR]\033[0m Failed to add lyrics to M4A: {e}")
        return False

def extract_lyrics_from_file(file_path: str) -> Optional[str]:
    """
    Extract lyrics from an audio file's metadata.
    
    Args:
        file_path (str): Path to the audio file
        
    Returns:
        str or None: Lyrics text if found, None otherwise
    """
    if not os.path.exists(file_path):
        return None
    
    file_ext = os.path.splitext(file_path)[1].lower()
    
    try:
        if file_ext == '.mp3':
            from mutagen.mp3 import MP3
            from mutagen.id3 import USLT
            
            audio = MP3(file_path)
            if audio.tags:
                for tag in audio.tags.values():
                    if isinstance(tag, USLT):
                        return tag.text
                        
        elif file_ext == '.opus':
            from mutagen.oggopus import OggOpus
            
            audio = OggOpus(file_path)
            if 'LYRICS' in audio:
                return audio['LYRICS'][0]
                
        elif file_ext == '.m4a':
            from mutagen.mp4 import MP4
            
            audio = MP4(file_path)
            if '©lyr' in audio:
                return audio['©lyr'][0]
                
    except Exception as e:
        print(f"\033[33m[WARNING]\033[0m Error extracting lyrics from {file_ext}: {e}")
    
    return None