@echo off
REM Test Spotify API credentials for Windows

REM Get the directory where the script is located
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%~dp0.."

REM Check if credentials file exists
set "CREDENTIALS_FILE=%USERPROFILE%\.shadowbox_spotify"
if exist "%CREDENTIALS_FILE%" (
    call "%CREDENTIALS_FILE%"
) else (
    echo X Spotify credentials file not found.
    echo Please run setup_spotify.bat first.
    pause
    exit /b 1
)

REM Check if credentials are set
if "%SPOTIFY_CLIENT_ID%"=="" (
    echo X Spotify Client ID not found in environment variables.
    echo Please run setup_spotify.bat again.
    pause
    exit /b 1
)

if "%SPOTIFY_CLIENT_SECRET%"=="" (
    echo X Spotify Client Secret not found in environment variables.
    echo Please run setup_spotify.bat again.
    pause
    exit /b 1
)

REM Check if virtual environment exists and activate it
if exist "%PROJECT_ROOT%\.venv" (
    call "%PROJECT_ROOT%\.venv\Scripts\activate.bat"
) else (
    echo X Virtual environment not found. Please run install.bat first.
    pause
    exit /b 1
)

REM Create a temporary Python script to test the credentials
set "TEMP_SCRIPT=%TEMP%\test_spotify.py"

echo import os > "%TEMP_SCRIPT%"
echo import sys >> "%TEMP_SCRIPT%"
echo try: >> "%TEMP_SCRIPT%"
echo     import spotipy >> "%TEMP_SCRIPT%"
echo     from spotipy.oauth2 import SpotifyClientCredentials >> "%TEMP_SCRIPT%"
echo except ImportError: >> "%TEMP_SCRIPT%"
echo     print("X spotipy module not found. Installing...") >> "%TEMP_SCRIPT%"
echo     import subprocess >> "%TEMP_SCRIPT%"
echo     subprocess.check_call([sys.executable, "-m", "pip", "install", "spotipy"]) >> "%TEMP_SCRIPT%"
echo     import spotipy >> "%TEMP_SCRIPT%"
echo     from spotipy.oauth2 import SpotifyClientCredentials >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo client_id = os.environ.get('SPOTIFY_CLIENT_ID') >> "%TEMP_SCRIPT%"
echo client_secret = os.environ.get('SPOTIFY_CLIENT_SECRET') >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo if not client_id or not client_secret: >> "%TEMP_SCRIPT%"
echo     print("X Spotify credentials not found in environment variables.") >> "%TEMP_SCRIPT%"
echo     sys.exit(1) >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo print(f"Testing Spotify credentials: ID {client_id[:4]}...{client_id[-4:]}") >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo try: >> "%TEMP_SCRIPT%"
echo     sp = spotipy.Spotify(auth_manager=SpotifyClientCredentials( >> "%TEMP_SCRIPT%"
echo         client_id=client_id, >> "%TEMP_SCRIPT%"
echo         client_secret=client_secret >> "%TEMP_SCRIPT%"
echo     )) >> "%TEMP_SCRIPT%"
echo     results = sp.search(q='artist:Queen track:Bohemian Rhapsody', limit=1) >> "%TEMP_SCRIPT%"
echo     if results and results['tracks']['items']: >> "%TEMP_SCRIPT%"
echo         track = results['tracks']['items'][0] >> "%TEMP_SCRIPT%"
echo         print(f"✓ Success! Found: {track['name']} by {track['artists'][0]['name']}") >> "%TEMP_SCRIPT%"
echo         print("Your Spotify credentials are working correctly.") >> "%TEMP_SCRIPT%"
echo     else: >> "%TEMP_SCRIPT%"
echo         print("X No results found. This might be an API issue.") >> "%TEMP_SCRIPT%"
echo except Exception as e: >> "%TEMP_SCRIPT%"
echo     print(f"X Error testing Spotify API: {e}") >> "%TEMP_SCRIPT%"
echo     sys.exit(1) >> "%TEMP_SCRIPT%"

REM Run the test script
python "%TEMP_SCRIPT%"

REM Clean up
del "%TEMP_SCRIPT%"

REM Keep the window open
pause