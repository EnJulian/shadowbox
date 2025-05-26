@echo off
REM lsnen's Foobar2000 Music Downloader Interactive Launcher for Windows

REM Check if virtual environment exists
if not exist .venv (
    echo X Virtual environment not found. Please run install.bat first.
    pause
    exit /b 1
)

REM Activate virtual environment
call .venv\Scripts\activate.bat

REM Check if aria2c is installed
where aria2c >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ! aria2c is not installed. For better download performance, install aria2c:
    echo Download from: https://github.com/aria2/aria2/releases
    echo Continuing without aria2c...
)

REM Run the interactive application
python interactive.py

REM If we get here, the application has exited
pause