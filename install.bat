@echo off
REM lsnen's Foobar2000 Music Downloader Installation Script for Windows

echo Installing lsnen's Foobar2000 Music Downloader...

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
if not exist .venv (
    echo Creating virtual environment...
    python -m venv .venv
)

REM Activate virtual environment
echo Activating virtual environment...
call .venv\Scripts\activate.bat

REM Install dependencies
echo Installing dependencies...
pip install -r requirements.txt

echo Installation complete!
echo To run the application, double-click run.bat or use:
echo .venv\Scripts\activate.bat ^&^& python main.py

pause