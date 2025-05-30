@echo off
REM Utility script to enhance multiple music files with Spotify metadata

REM Get the directory where the script is located
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%~dp0.."

REM Check if virtual environment exists and activate it
if exist "%PROJECT_ROOT%\.venv" (
    call "%PROJECT_ROOT%\.venv\Scripts\activate.bat"
) else (
    echo X Python virtual environment not found. Please run setup\install.bat first.
    exit /b 1
)

REM Check if arguments were provided
if "%~1"=="" (
    echo Usage: enhance_files.bat [OPTIONS] ^<path^>
    echo.
    echo Options:
    echo   -r, --recursive            Recursively scan directories
    echo   -e, --extensions EXT       Comma-separated list of file extensions to process (default: .mp3,.opus)
    echo   -d, --dry-run              Show what would be done without making changes
    echo.
    echo Examples:
    echo   enhance_files.bat C:\Music
    echo   enhance_files.bat -r C:\Music
    echo   enhance_files.bat -e .mp3,.flac,.opus C:\Music
    echo   enhance_files.bat -d C:\Music\Artist
    echo.
    exit /b 1
)

REM Run the Python script with all arguments
python "%SCRIPT_DIR%\enhance_files.py" %*

REM Keep the window open
pause