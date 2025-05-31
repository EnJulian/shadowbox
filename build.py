#!/usr/bin/env python3
"""
Build script for creating Shadowbox executables using PyInstaller.
"""

import os
import sys
import subprocess
import platform
import shutil

def run_command(cmd, cwd=None):
    """Run a command and return the result."""
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
        return False
    print(result.stdout)
    return True

def verify_dependencies():
    """Verify all required dependencies can be imported."""
    print("Verifying dependencies...")
    
    required_modules = [
        'yt_dlp',
        'mutagen',
        'requests', 
        'PIL',
        'spotipy'
    ]
    
    for module in required_modules:
        try:
            __import__(module)
            print(f"✅ {module}")
        except ImportError as e:
            print(f"❌ {module}: {e}")
            return False
    
    return True

def install_dependencies():
    """Install required dependencies."""
    print("Installing dependencies...")
    
    # Install Python dependencies
    if not run_command([sys.executable, "-m", "pip", "install", "-r", "setup/requirements.txt"]):
        return False
    
    if not run_command([sys.executable, "-m", "pip", "install", "pyinstaller"]):
        return False
    
    # Verify all dependencies are properly installed
    if not verify_dependencies():
        print("Dependency verification failed!")
        return False
    
    return True

def build_executable():
    """Build the executable using PyInstaller."""
    print("Building executable...")
    
    # Clean previous builds
    if os.path.exists("dist"):
        shutil.rmtree("dist")
    if os.path.exists("build"):
        shutil.rmtree("build")
    
    # Run PyInstaller
    if not run_command([sys.executable, "-m", "PyInstaller", "shadowbox.spec"]):
        return False
    
    return True

def create_archive():
    """Create an archive of the built executable."""
    print("Creating archive...")
    
    system = platform.system().lower()
    machine = platform.machine().lower()
    
    # Map machine types
    if machine in ['x86_64', 'amd64']:
        arch = 'x64'
    elif machine in ['arm64', 'aarch64']:
        arch = 'arm64'
    else:
        arch = machine
    
    archive_name = f"shadowbox-{system}-{arch}"
    
    if system == "windows":
        # Create ZIP for Windows
        shutil.make_archive(f"dist/{archive_name}", 'zip', 'dist', 'shadowbox.exe')
        print(f"Created: dist/{archive_name}.zip")
    else:
        # Create tar.gz for Unix-like systems
        shutil.make_archive(f"dist/{archive_name}", 'gztar', 'dist', 'shadowbox')
        print(f"Created: dist/{archive_name}.tar.gz")
    
    return True

def main():
    """Main build function."""
    print("Shadowbox Build Script")
    print("=" * 50)
    
    # Check if we're in the right directory
    if not os.path.exists("core/main.py"):
        print("Error: Please run this script from the shadowbox root directory")
        sys.exit(1)
    
    # Install dependencies
    if not install_dependencies():
        print("Failed to install dependencies")
        sys.exit(1)
    
    # Build executable
    if not build_executable():
        print("Failed to build executable")
        sys.exit(1)
    
    # Create archive
    if not create_archive():
        print("Failed to create archive")
        sys.exit(1)
    
    print("\nBuild completed successfully!")
    print("Check the 'dist' directory for your executable and archive.")

if __name__ == "__main__":
    main()