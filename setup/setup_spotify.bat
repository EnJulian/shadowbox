@echo off
REM Spotify API Setup Script for Windows

REM Get the directory where the script is located
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%~dp0.."

REM Check if virtual environment exists and activate it
if exist "%PROJECT_ROOT%\.venv" (
    call "%PROJECT_ROOT%\.venv\Scripts\activate.bat"
    
    REM Check if spotipy is installed
    python -c "import spotipy" 2>nul
    if errorlevel 1 (
        echo Installing spotipy module...
        python -m pip install spotipy
    )
)

echo Setting up Spotify API integration for shadowbox...
echo This script will help you configure your Spotify API credentials.
echo.
echo You need to create a Spotify Developer account and register an application to get your credentials.
echo Visit: https://developer.spotify.com/dashboard/ to create your app.
echo.

REM Check if credentials file exists
set "CREDENTIALS_FILE=%USERPROFILE%\.shadowbox_spotify"

REM Ask for client ID and secret
set /p CLIENT_ID="Enter your Spotify Client ID: "
set /p CLIENT_SECRET="Enter your Spotify Client Secret: "

REM Validate input
if "%CLIENT_ID%"=="" (
    echo X Error: Client ID is required.
    exit /b 1
)

if "%CLIENT_SECRET%"=="" (
    echo X Error: Client Secret is required.
    exit /b 1
)

REM Save credentials to file
echo @echo off > "%CREDENTIALS_FILE%"
echo set SPOTIFY_CLIENT_ID=%CLIENT_ID% >> "%CREDENTIALS_FILE%"
echo set SPOTIFY_CLIENT_SECRET=%CLIENT_SECRET% >> "%CREDENTIALS_FILE%"

REM Set environment variables for current session
set SPOTIFY_CLIENT_ID=%CLIENT_ID%
set SPOTIFY_CLIENT_SECRET=%CLIENT_SECRET%

REM Verify that the variables are set
if "%SPOTIFY_CLIENT_ID%"=="" (
    echo X Warning: Failed to set environment variables. Please try setting them manually:
    echo set SPOTIFY_CLIENT_ID=your_client_id
    echo set SPOTIFY_CLIENT_SECRET=your_client_secret
) else (
    echo ✓ Spotify credentials successfully loaded into current session.
)

echo.
echo Credentials saved to %CREDENTIALS_FILE%
echo.
echo For future sessions, you may need to run this file before using Spotify features:
echo call "%CREDENTIALS_FILE%"
echo.
echo Setup complete!

REM Keep the window open
pause