@echo off
REM Shadowbox Docker Runner Script for Windows
REM This script provides easy commands to run Shadowbox with Docker

setlocal enabledelayedexpansion

REM Function to print colored output (basic version for Windows)
set "GREEN=[92m"
set "YELLOW=[93m"
set "RED=[91m"
set "BLUE=[94m"
set "NC=[0m"

REM Check if Docker is running
docker info >nul 2>&1
if errorlevel 1 (
    echo %RED%[ERROR]%NC% Docker is not running. Please start Docker and try again.
    exit /b 1
)

REM Create necessary directories
if not exist "downloads" mkdir downloads
if not exist "music" mkdir music
if not exist "config" mkdir config

REM Setup environment file
if not exist ".env" (
    echo %GREEN%[INFO]%NC% Creating .env file from template...
    copy .env.example .env >nul
    echo %YELLOW%[WARNING]%NC% Please edit .env file with your API credentials for enhanced features
)

REM Main command handling
if "%1"=="" goto :run
if "%1"=="setup" goto :setup
if "%1"=="build" goto :build
if "%1"=="rebuild" goto :rebuild
if "%1"=="run" goto :run
if "%1"=="cli" goto :cli
if "%1"=="background" goto :background
if "%1"=="logs" goto :logs
if "%1"=="stop" goto :stop
if "%1"=="update" goto :update
if "%1"=="check" goto :check
if "%1"=="cleanup" goto :cleanup
if "%1"=="shell" goto :shell
if "%1"=="help" goto :help
if "%1"=="-h" goto :help
if "%1"=="--help" goto :help

echo %RED%[ERROR]%NC% Unknown command: %1
goto :help

:setup
echo %GREEN%[INFO]%NC% Building Shadowbox Docker image...
docker-compose build
echo %GREEN%[INFO]%NC% Setup complete! Run 'docker-run.bat run' to start Shadowbox
goto :end

:build
echo %GREEN%[INFO]%NC% Building Shadowbox Docker image...
docker-compose build
goto :end

:rebuild
echo %GREEN%[INFO]%NC% Rebuilding Shadowbox Docker image (no cache)...
docker-compose build --no-cache
echo %GREEN%[INFO]%NC% Docker image rebuilt successfully
goto :end

:run
echo %BLUE%Starting Shadowbox in Interactive Mode%NC%
echo %GREEN%[INFO]%NC% Use Ctrl+C to exit
echo %GREEN%[INFO]%NC% Ensuring proper terminal interaction...
REM Use docker-compose run instead of up for better interactive support
docker-compose run --rm shadowbox python /app/core/shbox.py
goto :end

:cli
if "%2"=="" (
    echo %RED%[ERROR]%NC% Please provide a search query
    echo Usage: %0 cli "Artist - Song Title"
    exit /b 1
)
echo %BLUE%Running Shadowbox CLI Mode%NC%
echo %GREEN%[INFO]%NC% Searching for: %2
docker-compose run --rm shadowbox python /app/core/main.py -q %2
goto :end

:background
echo %BLUE%Starting Shadowbox in Background Mode%NC%
docker-compose up -d
echo %GREEN%[INFO]%NC% Shadowbox is running in background
echo %GREEN%[INFO]%NC% Use 'docker-run.bat logs' to view logs
echo %GREEN%[INFO]%NC% Use 'docker-run.bat stop' to stop the container
goto :end

:logs
echo %BLUE%Shadowbox Logs%NC%
docker-compose logs -f
goto :end

:stop
echo %GREEN%[INFO]%NC% Stopping Shadowbox containers...
docker-compose down
echo %GREEN%[INFO]%NC% Containers stopped
goto :end

:update
echo %GREEN%[INFO]%NC% Updating yt-dlp...
docker-compose exec shadowbox pip install --upgrade yt-dlp
echo %GREEN%[INFO]%NC% yt-dlp updated
goto :end

:check
echo %BLUE%Checking Dependencies in Container%NC%
docker-compose ps | findstr "shadowbox.*Up" >nul
if errorlevel 1 (
    echo %GREEN%[INFO]%NC% Starting temporary container for dependency check...
    docker-compose run --rm shadowbox ./docker-deps-check.sh
) else (
    docker-compose exec shadowbox ./docker-deps-check.sh
)
goto :end

:cleanup
echo %YELLOW%[WARNING]%NC% This will remove all containers, images, and volumes
set /p "confirm=Are you sure? (y/N): "
if /i "!confirm!"=="y" (
    echo %GREEN%[INFO]%NC% Cleaning up...
    docker-compose down -v --rmi all
    echo %GREEN%[INFO]%NC% Cleanup complete
) else (
    echo %GREEN%[INFO]%NC% Cleanup cancelled
)
goto :end

:shell
echo %BLUE%Opening Shell in Shadowbox Container%NC%
docker-compose ps | findstr "shadowbox.*Up" >nul
if errorlevel 1 (
    echo %GREEN%[INFO]%NC% Starting temporary container...
    docker-compose run --rm shadowbox bash
) else (
    docker-compose exec shadowbox bash
)
goto :end

:help
echo %BLUE%Shadowbox Docker Runner%NC%
echo.
echo Usage: %0 [command] [options]
echo.
echo Commands:
echo   setup          - Initial setup (build image)
echo   build          - Build the Docker image
echo   rebuild        - Rebuild the Docker image (no cache)
echo   run            - Run in interactive mode (default)
echo   cli "query"    - Run single download via CLI
echo   background     - Run in background mode
echo   logs           - Show container logs
echo   stop           - Stop running containers
echo   update         - Update yt-dlp in running container
echo   check          - Check all dependencies in container
echo   cleanup        - Remove all containers, images, and volumes
echo   shell          - Open shell in container
echo   help           - Show this help message
echo.
echo Examples:
echo   %0 setup                           # Initial setup
echo   %0 run                             # Interactive mode
echo   %0 cli "Imagine Dragons Believer"  # Download specific song
echo   %0 background                      # Run in background
echo.
goto :end

:end
endlocal