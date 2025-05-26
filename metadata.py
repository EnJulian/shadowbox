from mutagen.easyid3 import EasyID3
from mutagen.mp3 import MP3
from mutagen.id3 import APIC, ID3
from mutagen.oggopus import OggOpus
from mutagen.flac import Picture
import os
import base64

def add_metadata(file_path, title, artist, album, cover_path=None):
    """
    Add metadata to an audio file (MP3 or Opus).
    
    Args:
        file_path (str): Path to the audio file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
    
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
            add_metadata_mp3(file_path, title, artist, album, cover_path)
        elif file_ext == '.opus':
            add_metadata_opus(file_path, title, artist, album, cover_path)
        else:
            raise Exception(f"Unsupported file format: {file_ext}")
    except Exception as e:
        raise Exception(f"Failed to add metadata: {e}")

def add_metadata_mp3(file_path, title, artist, album, cover_path=None):
    """
    Add metadata to an MP3 file.
    
    Args:
        file_path (str): Path to the MP3 file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
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

def add_metadata_opus(file_path, title, artist, album, cover_path=None):
    """
    Add metadata to an Opus file.
    
    Args:
        file_path (str): Path to the Opus file
        title (str): Song title
        artist (str): Artist name
        album (str): Album name
        cover_path (str, optional): Path to the cover image. Defaults to None.
    """
    try:
        # Load the Opus file
        audio = OggOpus(file_path)
        
        # Add basic metadata
        audio['title'] = title
        audio['artist'] = artist
        audio['album'] = album
        
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