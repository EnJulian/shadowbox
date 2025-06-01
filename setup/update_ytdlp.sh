#!/bin/bash

# Update yt-dlp to fix YouTube download issues
# This script updates yt-dlp to the latest version

echo "🔄 Shadowbox yt-dlp Updater"
echo "=========================================="
echo

# Check if we're in a virtual environment
if [[ "$VIRTUAL_ENV" != "" ]]; then
    echo "✅ Virtual environment detected: $VIRTUAL_ENV"
else
    echo "⚠️  No virtual environment detected"
    echo "   Consider activating your virtual environment first"
fi

echo

# Check current version
echo "📋 Checking current yt-dlp version..."
if command -v yt-dlp &> /dev/null; then
    current_version=$(yt-dlp --version 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo "✅ Current version: $current_version"
    else
        echo "❌ yt-dlp found but not working properly"
        current_version=""
    fi
else
    echo "❌ yt-dlp not found"
    current_version=""
fi

echo
echo "🚀 Updating yt-dlp to latest version..."

# Try different update methods
update_success=false

# Method 1: pip3
if command -v pip3 &> /dev/null; then
    echo "🔧 Trying: pip3 install --upgrade yt-dlp"
    if pip3 install --upgrade yt-dlp; then
        echo "✅ Updated successfully with pip3!"
        update_success=true
    else
        echo "❌ Failed with pip3"
    fi
fi

# Method 2: pip (if pip3 failed)
if [ "$update_success" = false ] && command -v pip &> /dev/null; then
    echo "🔧 Trying: pip install --upgrade yt-dlp"
    if pip install --upgrade yt-dlp; then
        echo "✅ Updated successfully with pip!"
        update_success=true
    else
        echo "❌ Failed with pip"
    fi
fi

# Method 3: python3 -m pip (if others failed)
if [ "$update_success" = false ] && command -v python3 &> /dev/null; then
    echo "🔧 Trying: python3 -m pip install --upgrade yt-dlp"
    if python3 -m pip install --upgrade yt-dlp; then
        echo "✅ Updated successfully with python3 -m pip!"
        update_success=true
    else
        echo "❌ Failed with python3 -m pip"
    fi
fi

echo

if [ "$update_success" = true ]; then
    # Check new version
    if command -v yt-dlp &> /dev/null; then
        new_version=$(yt-dlp --version 2>/dev/null)
        if [ $? -eq 0 ]; then
            echo "📋 New version: $new_version"
            
            if [ -n "$current_version" ] && [ "$new_version" != "$current_version" ]; then
                echo "🎉 Successfully upgraded from $current_version to $new_version"
            elif [ -n "$current_version" ] && [ "$new_version" = "$current_version" ]; then
                echo "ℹ️  You already had the latest version"
            else
                echo "✅ yt-dlp is now installed and ready to use"
            fi
        else
            echo "⚠️  Update completed but couldn't verify new version"
        fi
    fi
    
    echo
    echo "🎵 YouTube download issues should now be resolved!"
    echo "   You can now use Shadowbox to download music from YouTube."
else
    echo "❌ All update methods failed!"
    echo
    echo "🛠️  Manual update options:"
    echo "   1. Try: python3 -m pip install --upgrade yt-dlp"
    echo "   2. Try: pip3 install --upgrade yt-dlp"
    echo "   3. Check if you have permission to install packages"
    echo "   4. Consider using a virtual environment"
    echo "   5. Run: source .venv/bin/activate (if you have a .venv directory)"
fi

echo
read -p "Press Enter to exit..."