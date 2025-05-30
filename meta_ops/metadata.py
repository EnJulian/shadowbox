from mutagen.easyid3 import EasyID3
from mutagen.mp3 import MP3
from mutagen.id3 import APIC, ID3
from mutagen.oggopus import OggOpus
from mutagen.flac import Picture
from mutagen.mp4 import MP4, MP4Cover
import os
import base64

def extract_year_from_date(date_str):
    """
    Extract the year from a date string in various formats.
    
    Args:
        date_str (str): The date string to parse
        
    Returns:
        str: The extracted year, or the original string if no year could be extracted
    """
    # Check if it's in the format DATE:YYYYMMDD
    if date_str.startswith('DATE:') and len(date_str) >= 9 and date_str[5:9].isdigit():
        return date_str[5:9]  # Extract year from DATE:YYYYMMDD
    # Check if it's in the format YYYYMMDD
    elif len(date_str) >= 4 and date_str[:4].isdigit():
        return date_str[:4]  # Just take the year
    # Return the original string if no pattern matches
    return date_str

def extract_metadata(file_path):
    """
    Extract metadata from an audio file (MP3 or Opus).
    
    Args:
        file_path (str): Path to the audio file
    
    Returns:
        dict: Dictionary containing title, artist, and album if available
              If metadata is not available, returns empty strings for each field
    
    Raises:
        FileNotFoundError: If the audio file doesn't exist
    """
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"Audio file not found: {file_path}")
    
    # Default empty metadata
    metadata = {
        'title': '',
        'artist': '',
        'album': '',
        'album_artist': '',
        'date': '',
        'genre': '',
        'track_number': ''
    }
    
    # Determine file type based on extension
    file_ext = os.path.splitext(file_path)[1].lower()
    
    try:
        if file_ext == '.mp3':
            try:
                audio = MP3(file_path, ID3=EasyID3)
                if 'title' in audio and audio['title']:
                    metadata['title'] = audio['title'][0]
                if 'artist' in audio and audio['artist']:
                    metadata['artist'] = audio['artist'][0]
                if 'album' in audio and audio['album']:
                    metadata['album'] = audio['album'][0]
                if 'albumartist' in audio and audio['albumartist']:
                    metadata['album_artist'] = audio['albumartist'][0]
                if 'date' in audio and audio['date']:
                    # Extract just the year from any date format
                    metadata['date'] = extract_year_from_date(audio['date'][0])
                if 'genre' in audio and audio['genre']:
                    metadata['genre'] = audio['genre'][0]
                if 'tracknumber' in audio and audio['tracknumber']:
                    metadata['track_number'] = audio['tracknumber'][0]
            except:
                print("\033[33m[WARNING]\033[0m Could not extract metadata from MP3 file")
        elif file_ext == '.opus':
            try:
                audio = OggOpus(file_path)
                if 'title' in audio and audio['title']:
                    metadata['title'] = audio['title'][0]
                if 'artist' in audio and audio['artist']:
                    metadata['artist'] = audio['artist'][0]
                if 'album' in audio and audio['album']:
                    metadata['album'] = audio['album'][0]
                if 'albumartist' in audio and audio['albumartist']:
                    metadata['album_artist'] = audio['albumartist'][0]
                if 'date' in audio and audio['date']:
                    # Extract just the year from any date format
                    metadata['date'] = extract_year_from_date(audio['date'][0])
                if 'genre' in audio and audio['genre']:
                    metadata['genre'] = audio['genre'][0]
                if 'tracknumber' in audio and audio['tracknumber']:
                    metadata['track_number'] = audio['tracknumber'][0]
            except:
                print("\033[33m[WARNING]\033[0m Could not extract metadata from Opus file")
        elif file_ext == '.m4a':
            try:
                audio = MP4(file_path)
                # MP4 tags use different keys
                if '©nam' in audio and audio['©nam']:
                    metadata['title'] = audio['©nam'][0]
                if '©ART' in audio and audio['©ART']:
                    metadata['artist'] = audio['©ART'][0]
                if '©alb' in audio and audio['©alb']:
                    metadata['album'] = audio['©alb'][0]
                if 'aART' in audio and audio['aART']:
                    metadata['album_artist'] = audio['aART'][0]
                if '©day' in audio and audio['©day']:
                    # Extract just the year from any date format
                    metadata['date'] = extract_year_from_date(audio['©day'][0])
                if '©gen' in audio and audio['©gen']:
                    metadata['genre'] = audio['©gen'][0]
                if 'trkn' in audio and audio['trkn']:
                    metadata['track_number'] = str(audio['trkn'][0][0])
            except Exception as e:
                print(f"\033[33m[WARNING]\033[0m Could not extract metadata from M4A file: {e}")
        else:
            print(f"\033[33m[WARNING]\033[0m Unsupported file format for metadata extraction: {file_ext}")
    except Exception as e:
        print(f"\033[33m[WARNING]\033[0m Error extracting metadata: {e}")
    
    return metadata

def add_metadata(file_path, title, artist, album, cover_path=None, album_artist=None, date=None, genre=None, track_number=None):
    """
    Add metadata to an audio file (MP3 or Opus).
    
    Args:
        file_path (str): Path to the audio file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
        album_artist (str, optional): Album artist name. Defaults to None.
        date (str, optional): Release date or year. Defaults to None.
        genre (str, optional): Genre of the song. Defaults to None.
        track_number (str, optional): Track number. Defaults to None.
    
    Raises:
        FileNotFoundError: If the audio file doesn't exist
        Exception: If adding metadata fails
    """
    if not os.path.exists(file_path):
        raise FileNotFoundError(f"Audio file not found: {file_path}")
    
    # Determine file type based on extension
    file_ext = os.path.splitext(file_path)[1].lower()
    
    try:
        if file_ext == '.mp3':
            add_metadata_mp3(file_path, title, artist, album, cover_path, album_artist, date, genre, track_number)
        elif file_ext == '.opus':
            add_metadata_opus(file_path, title, artist, album, cover_path, album_artist, date, genre, track_number)
        elif file_ext == '.m4a':
            add_metadata_m4a(file_path, title, artist, album, cover_path, album_artist, date, genre, track_number)
        else:
            raise Exception(f"Unsupported file format: {file_ext}")
    except Exception as e:
        raise Exception(f"Failed to add metadata: {e}")

def add_metadata_mp3(file_path, title, artist, album, cover_path=None, album_artist=None, date=None, genre=None, track_number=None):
    """
    Add metadata to an MP3 file.
    
    Args:
        file_path (str): Path to the MP3 file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
        album_artist (str, optional): Album artist name. Defaults to None.
        date (str, optional): Release date or year. Defaults to None.
        genre (str, optional): Genre of the song. Defaults to None.
        track_number (str, optional): Track number. Defaults to None.
    """
    try:
        # Try to load existing ID3 tags or create new ones
        try:
            audio = MP3(file_path, ID3=EasyID3)
        except:
            # If the file doesn't have an ID3 tag, add one
            audio = MP3(file_path)
            audio.add_tags(ID3=EasyID3)
            audio.save()
            audio = MP3(file_path, ID3=EasyID3)
        
        # Add basic metadata
        audio['title'] = title
        audio['artist'] = artist
        audio['album'] = album
        
        # Add additional metadata if provided
        if album_artist:
            audio['albumartist'] = album_artist
        elif artist:  # Use artist as album artist if not provided
            audio['albumartist'] = artist
            
        if date:
            # Extract just the year from any date format
            audio['date'] = extract_year_from_date(date)
                
        if genre:
            audio['genre'] = genre
            
        if track_number:
            audio['tracknumber'] = track_number
            
        audio.save()
        
        # Add cover art if provided
        if cover_path and os.path.exists(cover_path):
            try:
                with open(cover_path, 'rb') as cover_file:
                    cover_data = cover_file.read()
                
                audio = MP3(file_path)
                
                # Make sure we have ID3 tags
                if not audio.tags:
                    audio.add_tags()
                
                # Determine image MIME type based on file extension
                mime_type = 'image/jpeg'  # Default
                if cover_path.lower().endswith('.png'):
                    mime_type = 'image/png'
                
                # Add cover art
                audio.tags.add(
                    APIC(
                        encoding=3,  # UTF-8
                        mime=mime_type,
                        type=3,  # Cover (front)
                        desc='Cover',
                        data=cover_data
                    )
                )
                audio.save()
            except Exception as e:
                print(f"Warning: Failed to add cover art to MP3: {e}")
    except Exception as e:
        raise Exception(f"Failed to add metadata to MP3: {e}")

def add_metadata_opus(file_path, title, artist, album, cover_path=None, album_artist=None, date=None, genre=None, track_number=None):
    """
    Add metadata to an Opus file.
    
    Args:
        file_path (str): Path to the Opus file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
        album_artist (str, optional): Album artist name. Defaults to None.
        date (str, optional): Release date or year. Defaults to None.
        genre (str, optional): Genre of the song. Defaults to None.
        track_number (str, optional): Track number. Defaults to None.
    """
    try:
        # Load the Opus file
        audio = OggOpus(file_path)
        
        # Add basic metadata
        audio['title'] = title
        audio['artist'] = artist
        audio['album'] = album
        
        # Add additional metadata if provided
        if album_artist:
            audio['albumartist'] = album_artist
        elif artist:  # Use artist as album artist if not provided
            audio['albumartist'] = artist
            
        if date:
            # Extract just the year from any date format
            audio['date'] = extract_year_from_date(date)
                
        if genre:
            audio['genre'] = genre
            
        if track_number:
            audio['tracknumber'] = track_number
        
        # Add cover art if provided
        if cover_path and os.path.exists(cover_path):
            try:
                # Create a Picture object
                picture = Picture()
                
                # Determine image MIME type based on file extension
                mime_type = 'image/jpeg'  # Default
                if cover_path.lower().endswith('.png'):
                    mime_type = 'image/png'
                
                picture.mime = mime_type
                picture.type = 3  # Cover (front)
                picture.desc = 'Cover'
                
                # Read the image data
                with open(cover_path, 'rb') as cover_file:
                    picture.data = cover_file.read()
                
                # Convert the picture to base64 string
                picture_data = picture.write()
                encoded_data = base64.b64encode(picture_data).decode('ascii')
                
                # Add the metadata
                audio['metadata_block_picture'] = [encoded_data]
            except Exception as e:
                print(f"Warning: Failed to add cover art to Opus: {e}")
        
        # Save the file
        audio.save()
    except Exception as e:
        raise Exception(f"Failed to add metadata to Opus: {e}")

def add_metadata_m4a(file_path, title, artist, album, cover_path=None, album_artist=None, date=None, genre=None, track_number=None):
    """
    Add metadata to an M4A file.
    
    Args:
        file_path (str): Path to the M4A file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
        album_artist (str, optional): Album artist name. Defaults to None.
        date (str, optional): Release date or year. Defaults to None.
        genre (str, optional): Genre of the song. Defaults to None.
        track_number (str, optional): Track number. Defaults to None.
    """
    try:
        # Load the M4A file
        audio = MP4(file_path)
        
        # Add basic metadata - MP4 uses different tag keys
        audio['©nam'] = [title]
        audio['©ART'] = [artist]
        audio['©alb'] = [album]
        
        # Add additional metadata if provided
        if album_artist:
            audio['aART'] = [album_artist]
        elif artist:  # Use artist as album artist if not provided
            audio['aART'] = [artist]
            
        if date:
            # Extract just the year from any date format
            audio['©day'] = [extract_year_from_date(date)]
                
        if genre:
            audio['©gen'] = [genre]
            
        if track_number:
            # Track number in M4A is stored as a tuple of (track_number, total_tracks)
            try:
                track_num = int(track_number)
                audio['trkn'] = [(track_num, 0)]  # Second value is total tracks, use 0 if unknown
            except ValueError:
                # If track number can't be converted to int, try to parse it
                if '/' in track_number:
                    try:
                        track_num, total_tracks = map(int, track_number.split('/', 1))
                        audio['trkn'] = [(track_num, total_tracks)]
                    except ValueError:
                        pass
        
        # Add cover art if provided
        if cover_path and os.path.exists(cover_path):
            try:
                with open(cover_path, 'rb') as cover_file:
                    cover_data = cover_file.read()
                
                # Determine image format based on file extension
                cover_format = MP4Cover.FORMAT_JPEG  # Default
                if cover_path.lower().endswith('.png'):
                    cover_format = MP4Cover.FORMAT_PNG
                
                # Add the cover art
                audio['covr'] = [MP4Cover(cover_data, cover_format)]
            except Exception as e:
                print(f"Warning: Failed to add cover art to M4A: {e}")
        
        # Save the file
        audio.save()
    except Exception as e:
        raise Exception(f"Failed to add metadata to M4A: {e}")