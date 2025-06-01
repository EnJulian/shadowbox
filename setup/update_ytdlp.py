#!/usr/bin/env python3
"""
Update yt-dlp to fix YouTube download issues.

This script updates yt-dlp to the latest version to resolve common
YouTube download problems like "Sign in to confirm you're not a bot".
"""

import subprocess
import sys
import os

def main():
    print("🔄 Shadowbox yt-dlp Updater")
    print("=" * 40)
    print()
    
    # Check current version
    print("📋 Checking current yt-dlp version...")
    try:
        result = subprocess.run(['yt-dlp', '--version'], 
                              check=True, capture_output=True, text=True)
        current_version = result.stdout.strip()
        print(f"✅ Current version: {current_version}")
    except (subprocess.CalledProcessError, FileNotFoundError):
        print("❌ yt-dlp not found or not working properly")
        current_version = None
    
    print()
    print("🚀 Updating yt-dlp to latest version...")
    
    # Try different pip commands
    pip_commands = ['pip3', 'pip', f'{sys.executable} -m pip']
    
    for pip_cmd in pip_commands:
        try:
            print(f"🔧 Trying: {pip_cmd} install --upgrade yt-dlp")
            
            if pip_cmd.startswith(sys.executable):
                # Handle python -m pip case
                cmd_parts = pip_cmd.split() + ['install', '--upgrade', 'yt-dlp']
            else:
                cmd_parts = [pip_cmd, 'install', '--upgrade', 'yt-dlp']
            
            result = subprocess.run(cmd_parts, check=True, capture_output=True, text=True)
            
            print("✅ yt-dlp updated successfully!")
            print()
            
            # Check new version
            try:
                result = subprocess.run(['yt-dlp', '--version'], 
                                      check=True, capture_output=True, text=True)
                new_version = result.stdout.strip()
                print(f"📋 New version: {new_version}")
                
                if current_version and new_version != current_version:
                    print(f"🎉 Successfully upgraded from {current_version} to {new_version}")
                elif current_version and new_version == current_version:
                    print("ℹ️  You already had the latest version")
                else:
                    print("✅ yt-dlp is now installed and ready to use")
                    
            except (subprocess.CalledProcessError, FileNotFoundError):
                print("⚠️  Update completed but couldn't verify new version")
            
            print()
            print("🎵 YouTube download issues should now be resolved!")
            print("   You can now use Shadowbox to download music from YouTube.")
            return True
            
        except (subprocess.CalledProcessError, FileNotFoundError) as e:
            print(f"❌ Failed with {pip_cmd}: {e}")
            continue
    
    # If all methods failed
    print()
    print("❌ All update methods failed!")
    print()
    print("🛠️  Manual update options:")
    print("   1. Try: python3 -m pip install --upgrade yt-dlp")
    print("   2. Try: python -m pip install --upgrade yt-dlp")
    print("   3. Check if you have permission to install packages")
    print("   4. Consider using a virtual environment")
    
    return False

if __name__ == "__main__":
    try:
        success = main()
        if success:
            input("\nPress Enter to exit...")
        else:
            input("\nPress Enter to exit...")
            sys.exit(1)
    except KeyboardInterrupt:
        print("\n\n⏹️  Update cancelled by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        input("\nPress Enter to exit...")
        sys.exit(1)