@echo off
REM Script to download YouTube audio and tag it with Spotify metadata

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
    echo Usage: spotify_tag.bat [OPTIONS] ^<youtube_url or search_query^>
    echo.
    echo Options:
    echo   -h, --help                 Show this help message
    echo   -o, --output FILE          Specify output file (default: auto-generated)
    echo   -f, --file FILE            Tag an existing file instead of downloading
    echo   -t, --title TITLE          Specify title for search (with -f)
    echo   -a, --artist ARTIST        Specify artist for search (with -f)
    echo.
    echo Examples:
    echo   spotify_tag.bat https://www.youtube.com/watch?v=dQw4w9WgXcQ
    echo   spotify_tag.bat -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ
    echo   spotify_tag.bat -f existing_song.opus -t "Never Gonna Give You Up" -a "Rick Astley"
    echo   spotify_tag.bat "Rick Astley Never Gonna Give You Up"
    echo.
    exit /b 1
)

REM Create a temporary Python script
set "TEMP_SCRIPT=%TEMP%\spotify_tag_temp.py"

echo import sys > "%TEMP_SCRIPT%"
echo import os >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo # Add project root to Python path >> "%TEMP_SCRIPT%"
echo project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__))) >> "%TEMP_SCRIPT%"
echo sys.path.insert(0, project_root) >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo from meta_ops.downloader import download_audio >> "%TEMP_SCRIPT%"
echo from meta_ops.spotify_metadata import enhance_existing_file_with_spotify >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo def main(): >> "%TEMP_SCRIPT%"
echo     args = sys.argv[1:] >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo     # Parse arguments >> "%TEMP_SCRIPT%"
echo     mode = args[0] >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo     if mode == "download": >> "%TEMP_SCRIPT%"
echo         query = args[1] >> "%TEMP_SCRIPT%"
echo         output_file = args[2] if len(args) ^> 2 and args[2] else None >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo         # Download and tag >> "%TEMP_SCRIPT%"
echo         success = download_audio(query, output_file or '%%^(title^)s.%%^(ext^)s', use_spotify_metadata=True) >> "%TEMP_SCRIPT%"
echo         sys.exit(0 if success else 1) >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo     elif mode == "tag": >> "%TEMP_SCRIPT%"
echo         file_path = args[1] >> "%TEMP_SCRIPT%"
echo         title = args[2] if len(args) ^> 2 else None >> "%TEMP_SCRIPT%"
echo         artist = args[3] if len(args) ^> 3 else None >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo         # Tag existing file >> "%TEMP_SCRIPT%"
echo         success = enhance_existing_file_with_spotify(file_path, title, artist) >> "%TEMP_SCRIPT%"
echo         sys.exit(0 if success else 1) >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo     else: >> "%TEMP_SCRIPT%"
echo         print(f"Unknown mode: {mode}") >> "%TEMP_SCRIPT%"
echo         sys.exit(1) >> "%TEMP_SCRIPT%"
echo. >> "%TEMP_SCRIPT%"
echo if __name__ == "__main__": >> "%TEMP_SCRIPT%"
echo     main() >> "%TEMP_SCRIPT%"

REM Parse command line arguments
set "OUTPUT_FILE="
set "EXISTING_FILE="
set "TITLE="
set "ARTIST="
set "QUERY="

:parse_args
if "%~1"=="" goto run_script

if /i "%~1"=="-h" (
    echo Usage: spotify_tag.bat [OPTIONS] ^<youtube_url or search_query^>
    echo.
    echo Options:
    echo   -h, --help                 Show this help message
    echo   -o, --output FILE          Specify output file (default: auto-generated)
    echo   -f, --file FILE            Tag an existing file instead of downloading
    echo   -t, --title TITLE          Specify title for search (with -f)
    echo   -a, --artist ARTIST        Specify artist for search (with -f)
    echo.
    echo Examples:
    echo   spotify_tag.bat https://www.youtube.com/watch?v=dQw4w9WgXcQ
    echo   spotify_tag.bat -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ
    echo   spotify_tag.bat -f existing_song.opus -t "Never Gonna Give You Up" -a "Rick Astley"
    echo   spotify_tag.bat "Rick Astley Never Gonna Give You Up"
    echo.
    exit /b 0
) else if /i "%~1"=="--help" (
    echo Usage: spotify_tag.bat [OPTIONS] ^<youtube_url or search_query^>
    echo.
    echo Options:
    echo   -h, --help                 Show this help message
    echo   -o, --output FILE          Specify output file (default: auto-generated)
    echo   -f, --file FILE            Tag an existing file instead of downloading
    echo   -t, --title TITLE          Specify title for search (with -f)
    echo   -a, --artist ARTIST        Specify artist for search (with -f)
    echo.
    echo Examples:
    echo   spotify_tag.bat https://www.youtube.com/watch?v=dQw4w9WgXcQ
    echo   spotify_tag.bat -o my_song.opus https://www.youtube.com/watch?v=dQw4w9WgXcQ
    echo   spotify_tag.bat -f existing_song.opus -t "Never Gonna Give You Up" -a "Rick Astley"
    echo   spotify_tag.bat "Rick Astley Never Gonna Give You Up"
    echo.
    exit /b 0
) else if /i "%~1"=="-o" (
    set "OUTPUT_FILE=%~2"
    shift
) else if /i "%~1"=="--output" (
    set "OUTPUT_FILE=%~2"
    shift
) else if /i "%~1"=="-f" (
    set "EXISTING_FILE=%~2"
    shift
) else if /i "%~1"=="--file" (
    set "EXISTING_FILE=%~2"
    shift
) else if /i "%~1"=="-t" (
    set "TITLE=%~2"
    shift
) else if /i "%~1"=="--title" (
    set "TITLE=%~2"
    shift
) else if /i "%~1"=="-a" (
    set "ARTIST=%~2"
    shift
) else if /i "%~1"=="--artist" (
    set "ARTIST=%~2"
    shift
) else (
    set "QUERY=%~1"
)

shift
goto parse_args

:run_script
REM Check if we have a query or existing file
if "%QUERY%"=="" (
    if "%EXISTING_FILE%"=="" (
        echo X Error: No YouTube URL, search query, or existing file specified.
        echo Run spotify_tag.bat -h for help.
        exit /b 1
    )
)

REM Run the Python script with appropriate arguments
if not "%EXISTING_FILE%"=="" (
    REM Tag existing file
    python "%TEMP_SCRIPT%" "tag" "%EXISTING_FILE%" "%TITLE%" "%ARTIST%"
) else (
    REM Download and tag
    python "%TEMP_SCRIPT%" "download" "%QUERY%" "%OUTPUT_FILE%"
)

REM Clean up
del "%TEMP_SCRIPT%"

REM Keep the window open
pause