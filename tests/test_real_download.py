#!/usr/bin/env python3
"""
Real test script to verify download strategies actually work
"""

import sys
import os
import subprocess
import tempfile
import shutil

# Add the project root to the path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

def test_individual_strategies():
    """Test each strategy individually to see which ones actually work"""
    
    # Use a simple, reliable YouTube video for testing
    test_url = "ytsearch1:test audio"  # Simple search that should work
    
    strategies = [
        {
            "name": "Original with aria2c",
            "cmd": [
                "yt-dlp",
                "--downloader", "aria2c",
                "--downloader-args", "aria2c:-x 16 -s 16",
                "-x",
                "-f", "bestaudio/best",
                "--embed-metadata",
                "--audio-format", "opus",
                "-o", "test_aria2c.%(ext)s",
                test_url
            ]
        },
        {
            "name": "Basic Method",
            "cmd": [
                "yt-dlp",
                "-x",
                "-f", "bestaudio/best",
                "--embed-metadata",
                "--audio-format", "opus",
                "--no-check-certificates",
                "-o", "test_basic.%(ext)s",
                test_url
            ]
        },
        {
            "name": "Alternative Extractor",
            "cmd": [
                "yt-dlp",
                "--extractor-args", "youtube:player_client=android",
                "--user-agent", "Mozilla/5.0 (Linux; Android 11; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
                "-x",
                "-f", "bestaudio/best",
                "--embed-metadata",
                "--audio-format", "opus",
                "--no-check-certificates",
                "--no-warnings",
                "-o", "test_android.%(ext)s",
                test_url
            ]
        },
        {
            "name": "iOS Client",
            "cmd": [
                "yt-dlp",
                "--extractor-args", "youtube:player_client=ios",
                "--user-agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1",
                "-x",
                "-f", "bestaudio/best",
                "--embed-metadata",
                "--audio-format", "opus",
                "--no-check-certificates",
                "--no-warnings",
                "-o", "test_ios.%(ext)s",
                test_url
            ]
        },
        {
            "name": "Simple Headers",
            "cmd": [
                "yt-dlp",
                "--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
                "-x",
                "-f", "bestaudio/best",
                "--embed-metadata",
                "--audio-format", "opus",
                "--no-check-certificates",
                "--no-warnings",
                "-o", "test_headers.%(ext)s",
                test_url
            ]
        }
    ]
    
    print("Testing individual download strategies...")
    print("=" * 60)
    
    # Create a temporary directory for testing
    with tempfile.TemporaryDirectory() as temp_dir:
        os.chdir(temp_dir)
        
        results = {}
        
        for i, strategy in enumerate(strategies):
            print(f"\n🧪 Testing Strategy {i+1}: {strategy['name']}")
            print("-" * 40)
            
            try:
                # Run the command with a timeout
                print(f"Command: {' '.join(strategy['cmd'][:4])}...")
                result = subprocess.run(
                    strategy['cmd'], 
                    check=True, 
                    capture_output=True, 
                    text=True,
                    timeout=60  # 60 second timeout
                )
                
                # Check if any files were created
                files = [f for f in os.listdir('.') if f.startswith('test_')]
                if files:
                    print(f"✅ SUCCESS: Downloaded {files[0]}")
                    file_size = os.path.getsize(files[0])
                    print(f"   File size: {file_size} bytes")
                    results[strategy['name']] = "SUCCESS"
                    # Clean up the file
                    for f in files:
                        os.remove(f)
                else:
                    print(f"⚠️  PARTIAL: Command succeeded but no file found")
                    results[strategy['name']] = "PARTIAL"
                    
            except subprocess.TimeoutExpired:
                print(f"⏰ TIMEOUT: Strategy took too long (>60s)")
                results[strategy['name']] = "TIMEOUT"
                
            except subprocess.CalledProcessError as e:
                error_msg = e.stderr if e.stderr else str(e)
                print(f"❌ FAILED: {error_msg[:200]}...")
                results[strategy['name']] = "FAILED"
                
            except Exception as e:
                print(f"💥 ERROR: {str(e)}")
                results[strategy['name']] = "ERROR"
    
    print("\n" + "=" * 60)
    print("📊 RESULTS SUMMARY:")
    print("=" * 60)
    
    for strategy_name, result in results.items():
        status_emoji = {
            "SUCCESS": "✅",
            "PARTIAL": "⚠️ ",
            "TIMEOUT": "⏰",
            "FAILED": "❌",
            "ERROR": "💥"
        }.get(result, "❓")
        
        print(f"{status_emoji} {strategy_name}: {result}")
    
    # Count successful strategies
    successful = sum(1 for r in results.values() if r == "SUCCESS")
    print(f"\n📈 {successful}/{len(strategies)} strategies working successfully")
    
    return results

def test_aria2c_availability():
    """Test if aria2c is available and working"""
    print("🔍 Checking aria2c availability...")
    try:
        result = subprocess.run(['aria2c', '--version'], capture_output=True, text=True, timeout=10)
        if result.returncode == 0:
            version_line = result.stdout.split('\n')[0]
            print(f"✅ aria2c found: {version_line}")
            return True
        else:
            print("❌ aria2c not working properly")
            return False
    except FileNotFoundError:
        print("❌ aria2c not installed")
        return False
    except Exception as e:
        print(f"❌ aria2c check failed: {e}")
        return False

def test_ytdlp_availability():
    """Test if yt-dlp is available and working"""
    print("🔍 Checking yt-dlp availability...")
    try:
        result = subprocess.run(['yt-dlp', '--version'], capture_output=True, text=True, timeout=10)
        if result.returncode == 0:
            version = result.stdout.strip()
            print(f"✅ yt-dlp found: {version}")
            return True
        else:
            print("❌ yt-dlp not working properly")
            return False
    except FileNotFoundError:
        print("❌ yt-dlp not installed")
        return False
    except Exception as e:
        print(f"❌ yt-dlp check failed: {e}")
        return False

if __name__ == "__main__":
    print("🧪 SHADOWBOX DOWNLOAD STRATEGY TESTING")
    print("=" * 60)
    
    # Check prerequisites
    aria2c_ok = test_aria2c_availability()
    ytdlp_ok = test_ytdlp_availability()
    
    if not ytdlp_ok:
        print("\n❌ Cannot proceed without yt-dlp. Please install it first.")
        sys.exit(1)
    
    if not aria2c_ok:
        print("\n⚠️  aria2c not available. The first strategy will fail.")
    
    print("\n" + "=" * 60)
    
    # Test the strategies
    results = test_individual_strategies()
    
    print("\n🎯 RECOMMENDATIONS:")
    print("=" * 60)
    
    if results.get("Original with aria2c") == "SUCCESS":
        print("✅ aria2c strategy is working - good choice for first strategy!")
    elif not aria2c_ok:
        print("⚠️  Install aria2c to enable the first strategy:")
        print("   macOS: brew install aria2")
        print("   Linux: sudo apt install aria2")
    else:
        print("❌ aria2c strategy failed - may need different configuration")
    
    successful_strategies = [name for name, result in results.items() if result == "SUCCESS"]
    if successful_strategies:
        print(f"\n✅ Working strategies: {', '.join(successful_strategies)}")
    else:
        print("\n❌ No strategies are working! This needs investigation.")