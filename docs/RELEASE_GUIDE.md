# Shadowbox Release Guide

This document explains how to create and manage releases for shadowbox, including automated cross-platform builds and GitHub releases.

## Table of Contents

1. [Overview](#overview)
2. [Initial Setup](#initial-setup)
3. [Release Process](#release-process)
4. [Understanding the Build System](#understanding-the-build-system)
5. [Troubleshooting](#troubleshooting)
6. [Advanced Usage](#advanced-usage)

## Overview

Shadowbox uses an automated release system that:

- **Builds standalone executables** for multiple platforms (Linux, Windows, macOS Intel, macOS Apple Silicon)
- **Creates GitHub releases** automatically when you push a git tag
- **Packages binaries** in platform-appropriate formats (tar.gz for Unix, zip for Windows)
- **Handles dependencies** automatically (FFmpeg, aria2, Python libraries)

### What Users Get

When you create a release, users can download:
- `shadowbox-linux-x64.tar.gz` - Linux 64-bit executable
- `shadowbox-windows-x64.zip` - Windows 64-bit executable  
- `shadowbox-macos-x64.tar.gz` - macOS Intel executable
- `shadowbox-macos-arm64.tar.gz` - macOS Apple Silicon executable

Users don't need Python, pip, or any dependencies installed - everything is bundled into the executable.

## Initial Setup

### Prerequisites

1. **Git repository** with GitHub remote
2. **GitHub account** with repository access
3. **Local development environment** with Python 3.8+

### First-Time Setup

1. **Ensure all files are committed:**
   ```bash
   git add .
   git commit -m "Add release automation system"
   git push origin main
   ```

2. **Verify GitHub Actions is enabled:**
   - Go to your GitHub repository
   - Click on the "Actions" tab
   - Ensure Actions are enabled for your repository

3. **Set up Python virtual environment:**
   
   **For macOS/Linux users (recommended):**
   ```bash
   # Navigate to project directory
   cd /path/to/shadowbox
   
   # Create virtual environment (if not already exists)
   python3 -m venv .venv
   
   # Activate virtual environment
   source .venv/bin/activate
   
   # Install dependencies
   pip install -r setup/requirements.txt
   pip install pyinstaller
   ```
   
   **For Windows users:**
   ```cmd
   # Navigate to project directory
   cd C:\path\to\shadowbox
   
   # Create virtual environment (if not already exists)
   python -m venv .venv
   
   # Activate virtual environment
   .venv\Scripts\activate
   
   # Install dependencies
   pip install -r setup/requirements.txt
   pip install pyinstaller
   ```
   
   **Note for macOS users:** If you encounter an "externally-managed-environment" error when using `pip3` directly, this is expected behavior with Homebrew-managed Python. Always use a virtual environment as shown above.

4. **Test local build (optional but recommended):**
   ```bash
   # Ensure virtual environment is activated
   source .venv/bin/activate  # macOS/Linux
   # OR
   .venv\Scripts\activate     # Windows
   
   # Test local build
   python build.py
   ```

## Release Process

### Method 1: Automated Release (Recommended)

Use the `release.py` script to handle version bumping and tagging automatically:

```bash
# Patch release (0.1.0 → 0.1.1) - for bug fixes
python3 release.py patch

# Minor release (0.1.0 → 0.2.0) - for new features
python3 release.py minor

# Major release (0.1.0 → 1.0.0) - for breaking changes
python3 release.py major

# Specific version
python3 release.py 1.2.3
```

**What this does:**
1. Updates version in `pyproject.toml`
2. Commits the version change
3. Creates and pushes a git tag (e.g., `v1.2.3`)
4. Triggers GitHub Actions to build and release

### Method 2: Manual Release

If you prefer manual control:

```bash
# 1. Update version in pyproject.toml manually
# Edit the line: version = "0.1.0" → version = "0.2.0"

# 2. Commit the version change
git add pyproject.toml
git commit -m "Bump version to 0.2.0"
git push origin main

# 3. Create and push tag
git tag v0.2.0
git push origin v0.2.0
```

### Method 3: Manual Trigger (Emergency Releases)

You can manually trigger a release without creating a tag:

1. Go to your GitHub repository
2. Click "Actions" tab
3. Click "Build and Release" workflow
4. Click "Run workflow"
5. Enter the version (e.g., `v0.1.1`)
6. Click "Run workflow"

## Understanding the Build System

### File Structure

```
shadowbox/
├── .github/workflows/release.yml    # GitHub Actions workflow
├── shadowbox.spec                   # PyInstaller configuration
├── pyproject.toml                   # Project metadata and dependencies
├── build.py                         # Local build script
├── release.py                       # Release automation script
├── Makefile                         # Common development tasks
└── RELEASE_GUIDE.md                 # This file
```

### Key Components

#### 1. GitHub Actions Workflow (`.github/workflows/release.yml`)

**Triggers:**
- When you push a tag starting with `v` (e.g., `v1.0.0`)
- Manual workflow dispatch

**Build Matrix:**
- Ubuntu (Linux x64)
- Windows (Windows x64)
- macOS (Intel x64)
- macOS (Apple Silicon arm64)

**Build Process:**
1. Checks out code
2. Sets up Python 3.11
3. Installs system dependencies (FFmpeg, aria2)
4. Installs Python dependencies
5. Runs PyInstaller to create executable
6. Creates platform-appropriate archive
7. Uploads to GitHub Release

#### 2. PyInstaller Configuration (`shadowbox.spec`)

**Purpose:** Defines how to build the standalone executable

**Key Features:**
- Bundles all Python dependencies
- Includes hidden imports for yt-dlp and mutagen
- Creates single-file executable
- Optimizes for size with UPX compression

#### 3. Release Script (`release.py`)

**Purpose:** Automates version management and release creation

**Features:**
- Semantic version bumping (major.minor.patch)
- Updates `pyproject.toml` automatically
- Creates git commits and tags
- Pushes to GitHub to trigger builds

#### 4. Build Script (`build.py`)

**Purpose:** Local testing and development builds

**Use Cases:**
- Test builds before releasing
- Debug build issues
- Create local executables for testing

## Monitoring Releases

### Checking Build Status

1. **GitHub Actions:**
   - Go to repository → Actions tab
   - Click on the latest workflow run
   - Monitor build progress for each platform

2. **Build Logs:**
   - Click on individual jobs (e.g., "build (ubuntu-latest)")
   - Expand steps to see detailed logs
   - Look for errors in red-highlighted sections

### Release Artifacts

Once builds complete successfully:

1. **GitHub Release:**
   - Go to repository → Releases
   - Find your new release
   - Download and test binaries

2. **Artifact Structure:**
   ```
   shadowbox-linux-x64.tar.gz
   ├── shadowbox                    # Executable
   
   shadowbox-windows-x64.zip
   ├── shadowbox.exe               # Executable
   
   shadowbox-macos-x64.tar.gz
   ├── shadowbox                   # Executable
   
   shadowbox-macos-arm64.tar.gz
   ├── shadowbox                   # Executable
   ```

## Troubleshooting

### Common Issues

#### 1. Externally-Managed-Environment Error (macOS)

**Symptoms:** 
```
error: externally-managed-environment
× This environment is externally managed
```

**Solution:** Use a virtual environment instead of installing packages globally:
```bash
# Navigate to project directory
cd /path/to/shadowbox

# Activate existing virtual environment
source .venv/bin/activate

# Or create new one if it doesn't exist
python3 -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -r setup/requirements.txt
```

**IDE Setup (PyCharm):**
1. Go to **PyCharm → Settings** (or **Preferences** on macOS)
2. Navigate to **Project: shadowbox → Python Interpreter**
3. Click the gear icon and select **Add...**
4. Choose **Existing Environment**
5. Browse to `/path/to/shadowbox/.venv/bin/python`

#### 2. Build Fails - Missing Dependencies

**Symptoms:** Build fails with import errors
**Solution:** Update `shadowbox.spec` to include missing modules:

```python
hiddenimports += [
    'missing_module_name',
    # Add other missing imports
]
```

#### 3. Build Fails - System Dependencies

**Symptoms:** FFmpeg or aria2 not found
**Solution:** Check GitHub Actions workflow system dependency installation:

```yaml
# For Ubuntu
- name: Install system dependencies (Ubuntu)
  if: matrix.os == 'ubuntu-latest'
  run: |
    sudo apt-get update
    sudo apt-get install -y ffmpeg aria2
```

#### 4. Release Not Created

**Symptoms:** Build succeeds but no GitHub release appears
**Solution:** 
- Ensure tag starts with `v` (e.g., `v1.0.0`, not `1.0.0`)
- Check repository permissions for GitHub Actions
- Verify `GITHUB_TOKEN` has release permissions

#### 5. Large Executable Size

**Symptoms:** Executable is unexpectedly large (>100MB)
**Solution:**
- Review `excludes` in `shadowbox.spec`
- Consider excluding unused modules
- Verify UPX compression is working

### Debug Local Builds

```bash
# Activate virtual environment first
source .venv/bin/activate  # macOS/Linux
# OR
.venv\Scripts\activate     # Windows

# Clean previous builds
make clean

# Install dependencies
make install

# Build with verbose output
python build.py

# Test the executable
./dist/shadowbox --help
```

### Debug GitHub Actions

1. **Enable debug logging:**
   - Go to repository Settings → Secrets and variables → Actions
   - Add secret: `ACTIONS_STEP_DEBUG` = `true`

2. **Check workflow syntax:**
   ```bash
   # Install GitHub CLI
   gh workflow view release.yml
   ```

## Advanced Usage

### Custom Build Configurations

#### Building for Specific Platforms

```bash
# Local builds for testing
make build-linux    # Linux only
make build-windows  # Windows only (if on Windows)
make build-macos    # macOS only (if on macOS)
```

#### Modifying Build Settings

Edit `shadowbox.spec` for custom configurations:

```python
# Example: Exclude specific modules to reduce size
excludes=[
    'tkinter',      # GUI toolkit (not needed)
    'matplotlib',   # Plotting library (not needed)
],

# Example: Add custom data files
datas += [
    ('config/', 'config/'),  # Include config directory
],
```

### Release Channels

#### Pre-release Versions

For beta/alpha releases:

```bash
# Create pre-release tag
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1
```

Modify `.github/workflows/release.yml` to mark as pre-release:

```yaml
- name: Create Release
  with:
    prerelease: true  # Add this line
```

#### Release Branches

For maintaining multiple versions:

```bash
# Create release branch
git checkout -b release/v1.0
git push origin release/v1.0

# Tag from release branch
git checkout release/v1.0
git tag v1.0.1
git push origin v1.0.1
```

### Customizing Release Notes

Edit the release body in `.github/workflows/release.yml`:

```yaml
body: |
  ## Shadowbox ${{ steps.version.outputs.version }}
  
  ### 🎵 New Features
  - Feature 1
  - Feature 2
  
  ### 🐛 Bug Fixes
  - Fix 1
  - Fix 2
  
  ### 📦 Downloads
  Choose the appropriate file for your platform:
  - **Linux (x64)**: `shadowbox-linux-x64.tar.gz`
  - **Windows (x64)**: `shadowbox-windows-x64.zip`
  - **macOS (Intel)**: `shadowbox-macos-x64.tar.gz`
  - **macOS (Apple Silicon)**: `shadowbox-macos-arm64.tar.gz`
```

## Best Practices

### Version Management

1. **Follow Semantic Versioning:**
   - `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
   - Major: Breaking changes
   - Minor: New features (backward compatible)
   - Patch: Bug fixes

2. **Update CHANGELOG.md:**
   ```markdown
   ## [1.2.3] - 2024-01-15
   
   ### Added
   - New feature description
   
   ### Fixed
   - Bug fix description
   ```

3. **Test Before Release:**
   ```bash
   # Always test locally first
   python build.py
   ./dist/shadowbox --help
   ```

### Release Workflow

1. **Development → Testing → Release:**
   ```bash
   # 1. Develop features
   git checkout -b feature/new-feature
   # ... make changes ...
   git commit -m "Add new feature"
   
   # 2. Merge to main
   git checkout main
   git merge feature/new-feature
   git push origin main
   
   # 3. Test build locally
   python build.py
   
   # 4. Create release
   python release.py minor
   ```

2. **Monitor and Verify:**
   - Watch GitHub Actions build
   - Test downloaded binaries
   - Update documentation if needed

### Security Considerations

1. **Dependency Updates:**
   ```bash
   # Regularly update dependencies
   pip list --outdated
   # Update setup/requirements.txt as needed
   ```

2. **Vulnerability Scanning:**
   - GitHub automatically scans for vulnerabilities
   - Review Dependabot alerts
   - Update dependencies promptly

## Quick Reference

### Common Commands

```bash
# Activate virtual environment (always do this first)
source .venv/bin/activate  # macOS/Linux
# OR
.venv\Scripts\activate     # Windows

# Create patch release (bug fixes)
python release.py patch

# Create minor release (new features)
python release.py minor

# Create major release (breaking changes)
python release.py major

# Build locally for testing
python build.py

# Clean build artifacts
make clean

# View current version
grep 'version = ' pyproject.toml

# Deactivate virtual environment when done
deactivate
```

### Important URLs

- **Repository Actions:** `https://github.com/USERNAME/shadowbox/actions`
- **Releases Page:** `https://github.com/USERNAME/shadowbox/releases`
- **Workflow File:** `.github/workflows/release.yml`

### File Checklist

Before creating a release, ensure these files are up to date:

- [ ] `pyproject.toml` - Dependencies and metadata
- [ ] `setup/requirements.txt` - Python dependencies
- [ ] `shadowbox.spec` - PyInstaller configuration
- [ ] `CHANGELOG.md` - Release notes
- [ ] `README.md` - Installation instructions
- [ ] `DEPENDENCY_AUDIT.md` - Dependency synchronization check

### Dependency Verification

Run this command to verify all dependencies are properly synchronized:

```bash
# Activate virtual environment first
source .venv/bin/activate  # macOS/Linux
# OR
.venv\Scripts\activate     # Windows

# Verify local dependencies
python build.py  # Includes automatic dependency verification

# Manual verification
python -c "
import yt_dlp, mutagen, requests, PIL, spotipy
print('✅ All dependencies verified')
"
```

---

## Summary

The Shadowbox release system provides:

✅ **Automated cross-platform builds**  
✅ **One-command releases**  
✅ **Standalone executables**  
✅ **Professional GitHub releases**  
✅ **Easy maintenance and updates**

For most releases, you'll simply run:
```bash
python release.py patch  # or minor/major
```

Then monitor the GitHub Actions build and verify the release artifacts. The system handles everything else automatically!