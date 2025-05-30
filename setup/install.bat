@echo off
REM Shadowbox Music Downloader Installation Script for Windows

echo Installing Shadowbox Music Downloader...

REM Check if Python 3 is installed
python --version 2>nul | findstr /r "^Python 3" >nul
if %ERRORLEVEL% NEQ 0 (
    echo X Python 3 is not installed. Please install Python 3 and try again.
    echo Download from: https://www.python.org/downloads/
    pause
    exit /b 1
)

REM Check if FFmpeg is installed
where ffmpeg >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo X FFmpeg is not installed. Please install FFmpeg and try again.
    echo Download from: https://ffmpeg.org/download.html
    echo Make sure to add FFmpeg to your PATH environment variable.
    pause
    exit /b 1
)

REM Create virtual environment if it doesn't exist
if not exist ..\.venv (
    echo Creating virtual environment...
    python -m venv ..\.venv
)

REM Activate virtual environment
echo Activating virtual environment...
call ..\.venv\Scripts\activate.bat

REM Get the directory where the script is located
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."

REM Install dependencies
echo Installing dependencies...
pip install -r "%SCRIPT_DIR%requirements.txt"

REM Explicitly check and install spotipy
echo Ensuring Spotify API module is installed...
echo Installing spotipy module...
pip install --force-reinstall spotipy>=2.23.0

REM Verify the installation
python -c "import spotipy" 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo First installation attempt failed. Trying alternative method...
    pip install --upgrade pip
    pip install --force-reinstall spotipy>=2.23.0
    
    REM Final verification
    python -c "import spotipy" 2>nul
    if %ERRORLEVEL% NEQ 0 (
        echo Warning: Could not verify spotipy installation. Spotify features may not work.
    ) else (
        echo Spotipy successfully installed!
    )
) else (
    echo Spotipy successfully installed!
)

REM Ensure meta_ops is a proper Python package
echo Ensuring meta_ops is a proper Python package...
if not exist "%PROJECT_ROOT%\meta_ops\__init__.py" (
    echo # meta_ops package initialization > "%PROJECT_ROOT%\meta_ops\__init__.py"
    echo Created meta_ops/__init__.py
)

echo Installation complete!
echo To run the application, double-click shbox.bat in the project root or use:
echo %PROJECT_ROOT%\.venv\Scripts\activate.bat ^&^& python %PROJECT_ROOT%\core\main.py

pause