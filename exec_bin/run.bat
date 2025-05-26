@echo off
REM Shadowbox Music Downloader Run Script for Windows

REM Get the directory where the script is located
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%"

REM Check if virtual environment exists
if not exist "%PROJECT_ROOT%\.venv" (
    echo X Virtual environment not found. Please run setup\install.bat first.
    exit /b 1
)

REM Activate virtual environment
call "%PROJECT_ROOT%\.venv\Scripts\activate.bat"

REM Check if aria2c is installed
where aria2c >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ! aria2c is not installed. For better download performance, install aria2c:
    echo Download from: https://github.com/aria2/aria2/releases
    echo Continuing without aria2c...
)

REM Run the application
echo Starting lsnen's shadowbox...
echo ------------------------------------
echo Usage examples:
echo 1. Download by search: run.bat -q "Imagine Dragons Believer"
echo 2. Download by URL: run.bat -q "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
echo 3. Specify music directory: run.bat -q "Imagine Dragons Believer" -d "C:\Users\YourName\Music\MyCollection"
echo 4. Interactive mode: run.bat
echo ------------------------------------
echo Options:
echo -q, --query: Song title and artist or URL
echo -d, --directory: Base music directory (default: User's Music folder)
echo -o, --output: Output file name (optional, rarely needed)
echo ------------------------------------

REM Check if arguments were provided
if "%~1"=="" (
    REM No arguments, run in interactive mode
    python "%PROJECT_ROOT%\core\main.py"
) else (
    REM Pass all arguments to the script
    python "%PROJECT_ROOT%\core\main.py" %*
)

REM Keep the window open
pause