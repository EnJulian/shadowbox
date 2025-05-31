#!/usr/bin/env python3
"""
Release script for Shadowbox - handles version bumping and tagging.
"""

import os
import sys
import subprocess
import re
from pathlib import Path

def run_command(cmd, cwd=None):
    """Run a command and return the result."""
    print(f"Running: {' '.join(cmd)}")
    result = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
        return False, result.stderr
    return True, result.stdout.strip()

def get_current_version():
    """Get the current version from pyproject.toml."""
    try:
        with open("pyproject.toml", "r") as f:
            content = f.read()
        
        match = re.search(r'version = "([^"]+)"', content)
        if match:
            return match.group(1)
    except FileNotFoundError:
        pass
    
    return None

def update_version(new_version):
    """Update version in pyproject.toml."""
    try:
        with open("pyproject.toml", "r") as f:
            content = f.read()
        
        # Update version
        content = re.sub(r'version = "[^"]+"', f'version = "{new_version}"', content)
        
        with open("pyproject.toml", "w") as f:
            f.write(content)
        
        return True
    except Exception as e:
        print(f"Error updating version: {e}")
        return False

def bump_version(current_version, bump_type):
    """Bump version based on type (major, minor, patch)."""
    try:
        parts = current_version.split('.')
        major, minor, patch = int(parts[0]), int(parts[1]), int(parts[2])
        
        if bump_type == "major":
            major += 1
            minor = 0
            patch = 0
        elif bump_type == "minor":
            minor += 1
            patch = 0
        elif bump_type == "patch":
            patch += 1
        else:
            raise ValueError(f"Invalid bump type: {bump_type}")
        
        return f"{major}.{minor}.{patch}"
    except Exception as e:
        print(f"Error bumping version: {e}")
        return None

def create_git_tag(version):
    """Create and push a git tag."""
    tag = f"v{version}"
    
    # Check if tag already exists
    success, output = run_command(["git", "tag", "-l", tag])
    if not success:
        return False
    
    if tag in output:
        print(f"Tag {tag} already exists!")
        return False
    
    # Create tag
    success, _ = run_command(["git", "tag", "-a", tag, "-m", f"Release {tag}"])
    if not success:
        return False
    
    # Push tag
    success, _ = run_command(["git", "push", "origin", tag])
    if not success:
        return False
    
    return True

def main():
    """Main release function."""
    if len(sys.argv) < 2:
        print("Usage: python release.py <version|major|minor|patch>")
        print("Examples:")
        print("  python release.py 1.2.3     # Set specific version")
        print("  python release.py major     # Bump major version")
        print("  python release.py minor     # Bump minor version")
        print("  python release.py patch     # Bump patch version")
        sys.exit(1)
    
    version_arg = sys.argv[1]
    
    # Check if we're in a git repository
    if not os.path.exists(".git"):
        print("Error: Not in a git repository")
        sys.exit(1)
    
    # Get current version
    current_version = get_current_version()
    if not current_version:
        print("Error: Could not find current version in pyproject.toml")
        sys.exit(1)
    
    print(f"Current version: {current_version}")
    
    # Determine new version
    if version_arg in ["major", "minor", "patch"]:
        new_version = bump_version(current_version, version_arg)
        if not new_version:
            sys.exit(1)
    else:
        # Assume it's a specific version
        new_version = version_arg
    
    print(f"New version: {new_version}")
    
    # Confirm with user
    response = input(f"Create release {new_version}? (y/N): ")
    if response.lower() != 'y':
        print("Release cancelled")
        sys.exit(0)
    
    # Update version in files
    if not update_version(new_version):
        print("Failed to update version")
        sys.exit(1)
    
    # Commit version change
    success, _ = run_command(["git", "add", "pyproject.toml"])
    if not success:
        print("Failed to stage version changes")
        sys.exit(1)
    
    success, _ = run_command(["git", "commit", "-m", f"Bump version to {new_version}"])
    if not success:
        print("Failed to commit version changes")
        sys.exit(1)
    
    # Push changes
    success, _ = run_command(["git", "push"])
    if not success:
        print("Failed to push version changes")
        sys.exit(1)
    
    # Create and push tag
    if not create_git_tag(new_version):
        print("Failed to create git tag")
        sys.exit(1)
    
    print(f"\nRelease {new_version} created successfully!")
    print("GitHub Actions will automatically build and create the release.")
    print(f"Check: https://github.com/lsnen/shadowbox/actions")

if __name__ == "__main__":
    main()