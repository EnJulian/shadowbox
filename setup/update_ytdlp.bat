@echo off
REM Update yt-dlp to fix YouTube download issues
REM This script updates yt-dlp to the latest version

echo 🔄 Shadowbox yt-dlp Updater
echo ==========================================
echo.

REM Check if we're in a virtual environment
if defined VIRTUAL_ENV (
    echo ✅ Virtual environment detected: %VIRTUAL_ENV%
) else (
    echo ⚠️  No virtual environment detected
    echo    Consider activating your virtual environment first
)

echo.

REM Check current version
echo 📋 Checking current yt-dlp version...
yt-dlp --version >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=*" %%i in ('yt-dlp --version 2^>nul') do set current_version=%%i
    echo ✅ Current version: !current_version!
) else (
    echo ❌ yt-dlp not found or not working properly
    set current_version=
)

echo.
echo 🚀 Updating yt-dlp to latest version...

set update_success=false

REM Method 1: pip3
echo 🔧 Trying: pip3 install --upgrade yt-dlp
pip3 install --upgrade yt-dlp >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Updated successfully with pip3!
    set update_success=true
    goto check_version
) else (
    echo ❌ Failed with pip3
)

REM Method 2: pip
echo 🔧 Trying: pip install --upgrade yt-dlp
pip install --upgrade yt-dlp >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Updated successfully with pip!
    set update_success=true
    goto check_version
) else (
    echo ❌ Failed with pip
)

REM Method 3: python -m pip
echo 🔧 Trying: python -m pip install --upgrade yt-dlp
python -m pip install --upgrade yt-dlp >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Updated successfully with python -m pip!
    set update_success=true
    goto check_version
) else (
    echo ❌ Failed with python -m pip
)

REM Method 4: python3 -m pip
echo 🔧 Trying: python3 -m pip install --upgrade yt-dlp
python3 -m pip install --upgrade yt-dlp >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Updated successfully with python3 -m pip!
    set update_success=true
    goto check_version
) else (
    echo ❌ Failed with python3 -m pip
)

:check_version
echo.

if "%update_success%"=="true" (
    REM Check new version
    yt-dlp --version >nul 2>&1
    if %errorlevel% equ 0 (
        for /f "tokens=*" %%i in ('yt-dlp --version 2^>nul') do set new_version=%%i
        echo 📋 New version: !new_version!
        
        if defined current_version (
            if "!new_version!" neq "!current_version!" (
                echo 🎉 Successfully upgraded from !current_version! to !new_version!
            ) else (
                echo ℹ️  You already had the latest version
            )
        ) else (
            echo ✅ yt-dlp is now installed and ready to use
        )
    ) else (
        echo ⚠️  Update completed but couldn't verify new version
    )
    
    echo.
    echo 🎵 YouTube download issues should now be resolved!
    echo    You can now use Shadowbox to download music from YouTube.
) else (
    echo ❌ All update methods failed!
    echo.
    echo 🛠️  Manual update options:
    echo    1. Try: python -m pip install --upgrade yt-dlp
    echo    2. Try: pip install --upgrade yt-dlp
    echo    3. Check if you have permission to install packages
    echo    4. Consider using a virtual environment
    echo    5. Run: .venv\Scripts\activate.bat (if you have a .venv directory)
)

echo.
pause