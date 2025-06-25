@echo off
setlocal enabledelayedexpansion

echo === Genius API Setup ===
echo.
echo This script will help you set up Genius API credentials for lyrics functionality.
echo.

REM Check if .env file exists
if not exist ".env" (
    echo Creating .env file...
    type nul > .env
)

echo Step 1: Get your Genius API Access Token
echo.
echo 1. Go to https://genius.com/api-clients
echo 2. Sign in or create a Genius account
echo 3. Click 'New API Client'
echo 4. Fill in the form:
echo    - App Name: ShadowBox (or any name you prefer)
echo    - App Website URL: http://localhost (or your website)
echo    - Redirect URI: http://localhost (can be anything for this use case)
echo 5. Click 'Save'
echo 6. Copy the 'Client Access Token' (NOT the Client ID or Client Secret)
echo.
echo Note: The Client Access Token is all you need for lyrics fetching.
echo.

REM Prompt for access token
set /p genius_token="Enter your Genius Client Access Token: "

if "%genius_token%"=="" (
    echo Error: No access token provided. Exiting.
    pause
    exit /b 1
)

REM Check if GENIUS_ACCESS_TOKEN already exists in .env
findstr /c:"GENIUS_ACCESS_TOKEN=" .env >nul 2>&1
if !errorlevel! equ 0 (
    REM Update existing entry
    powershell -Command "(Get-Content .env) -replace 'GENIUS_ACCESS_TOKEN=.*', 'GENIUS_ACCESS_TOKEN=%genius_token%' | Set-Content .env"
    echo Updated existing GENIUS_ACCESS_TOKEN in .env
) else (
    REM Add new entry
    echo GENIUS_ACCESS_TOKEN=%genius_token% >> .env
    echo Added GENIUS_ACCESS_TOKEN to .env
)

echo.
echo Step 2: Testing the connection...

REM Set the token for this session
set GENIUS_ACCESS_TOKEN=%genius_token%

REM Test the connection using Python
python -c "
import os
import sys

# Add the current directory to Python path
sys.path.insert(0, '.')

try:
    from meta_ops.lyrics import get_lyrics_from_genius
    
    print('Testing Genius API connection...')
    
    # Test with a well-known song
    test_lyrics = get_lyrics_from_genius('Bohemian Rhapsody', 'Queen')
    
    if test_lyrics:
        print('✅ SUCCESS: Genius API connection working!')
        print(f'   Retrieved lyrics ({len(test_lyrics)} characters)')
    else:
        print('⚠️  WARNING: Could not retrieve test lyrics, but API token seems valid')
        print('   This might be due to the specific test song not being found')
        
except ImportError as e:
    print('❌ ERROR: Missing required library')
    print('   Please run: pip install lyricsgenius')
    sys.exit(1)
except Exception as e:
    print(f'❌ ERROR: {e}')
    print('   Please check your access token and try again')
    sys.exit(1)
"

if !errorlevel! equ 0 (
    echo.
    echo === Setup Complete! ===
    echo.
    echo Your Genius API credentials have been saved to .env
    echo Lyrics will now be automatically fetched and embedded when downloading music.
    echo.
    echo Usage:
    echo   The lyrics functionality is now integrated into the main download process.
    echo   When you download music, lyrics will be automatically:
    echo   • Searched for using the song title and artist
    echo   • Embedded into the audio file metadata
    echo   • Available in music players that support lyrics display
    echo.
    echo Note: You may need to restart your command prompt or application
    echo to ensure the new environment variables are loaded.
) else (
    echo.
    echo === Setup Failed ===
    echo.
    echo Please check your access token and try running this script again.
    echo Make sure you copied the 'Client Access Token' and not the Client ID or Secret.
)

pause