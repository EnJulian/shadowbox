
# Making Shadowbox Available via Package Managers

To turn your project into a command line package that can be installed via package managers like `winget` (Windows) or `brew` (macOS), you'll need to create and submit package definitions to these repositories. Here's a comprehensive guide on how to do this:

## 1. Homebrew (macOS)

Homebrew is the most popular package manager for macOS. To make your application available via `brew install shadowbox`:

### Create a Homebrew Formula

1. **Create a Ruby formula file** named `shadowbox.rb`:

```ruby
class Shadowbox < Formula
  desc "Music acquisition tool that rips audio from YouTube/Bandcamp, converts to Opus, and injects metadata + album art"
  homepage "https://github.com/lsnen/shadowbox"
  url "https://github.com/lsnen/shadowbox/releases/download/v1.1.1/shadowbox-macos-x64.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256_HASH"
  version "1.1.1"
  
  depends_on "ffmpeg"
  depends_on "aria2" => :optional
  
  def install
    bin.install "shadowbox"
  end
  
  test do
    system "#{bin}/shadowbox", "--help"
  end
end
```

2. **Calculate the SHA256 hash** of your release archive:
   ```bash
   shasum -a 256 shadowbox-macos-x64.tar.gz
   ```
   Replace `REPLACE_WITH_ACTUAL_SHA256_HASH` with the output.

3. **Test your formula locally**:
   ```bash
   brew install --build-from-source ./shadowbox.rb
   ```

### Submit to Homebrew

You have two options:

#### Option A: Submit to Homebrew Core (Official Repository)
1. Fork the [Homebrew Core repository](https://github.com/Homebrew/homebrew-core)
2. Add your formula to the appropriate directory
3. Submit a pull request

#### Option B: Create Your Own Tap (Recommended for Starters)
1. Create a new GitHub repository named `homebrew-shadowbox`
2. Add your `shadowbox.rb` formula to this repository
3. Users can then install it with:
   ```bash
   brew tap lsnen/shadowbox
   brew install shadowbox
   ```

### Automate Formula Updates

Add a step to your GitHub Actions workflow to automatically update the formula when you release a new version:

```yaml
- name: Update Homebrew formula
  if: startsWith(github.ref, 'refs/tags/')
  run: |
    # Calculate SHA256 of the new release
    SHA256=$(shasum -a 256 release_files/shadowbox-macos-x64.tar.gz | cut -d ' ' -f 1)
    VERSION=${GITHUB_REF#refs/tags/v}
    
    # Update formula in homebrew-shadowbox repository
    # This requires setting up a personal access token with repo permissions
    git clone https://github.com/lsnen/homebrew-shadowbox.git
    cd homebrew-shadowbox
    sed -i "s|url \".*\"|url \"https://github.com/lsnen/shadowbox/releases/download/v${VERSION}/shadowbox-macos-x64.tar.gz\"|g" shadowbox.rb
    sed -i "s|sha256 \".*\"|sha256 \"${SHA256}\"|g" shadowbox.rb
    sed -i "s|version \".*\"|version \"${VERSION}\"|g" shadowbox.rb
    
    git config user.name "GitHub Actions"
    git config user.email "actions@github.com"
    git add shadowbox.rb
    git commit -m "Update to version ${VERSION}"
    git push https://${{ secrets.GH_TOKEN }}@github.com/lsnen/homebrew-shadowbox.git main
```

## 2. WinGet (Windows)

WinGet is Microsoft's official package manager for Windows. To make your application available via `winget install shadowbox`:

### Create a WinGet Manifest

1. **Create a manifest directory structure**:
   ```
   manifests/l/lsnen/shadowbox/1.1.1/
   ```

2. **Create the manifest files**:

   `lsnen.shadowbox.installer.yaml`:
   ```yaml
   PackageIdentifier: lsnen.shadowbox
   PackageVersion: 1.1.1
   MinimumOSVersion: 10.0.0.0
   InstallerType: zip
   NestedInstallerType: portable
   NestedInstallerFiles:
     - RelativeFilePath: shadowbox.exe
       PortableCommandAlias: shadowbox
   Installers:
     - Architecture: x64
       InstallerUrl: https://github.com/lsnen/shadowbox/releases/download/v1.1.1/shadowbox-windows-x64.zip
       InstallerSha256: REPLACE_WITH_ACTUAL_SHA256_HASH
   ManifestType: installer
   ManifestVersion: 1.4.0
   ```

   `lsnen.shadowbox.locale.en-US.yaml`:
   ```yaml
   PackageIdentifier: lsnen.shadowbox
   PackageVersion: 1.1.1
   PackageLocale: en-US
   Publisher: lsnen
   PackageName: shadowbox
   License: MIT
   ShortDescription: Music acquisition tool that rips audio from YouTube/Bandcamp, converts to Opus, and injects metadata + album art
   Description: A Python application that downloads music from YouTube or Bandcamp, adds metadata, and embeds album art
   Tags:
     - music
     - youtube
     - bandcamp
     - audio
     - download
     - cli
   ManifestType: defaultLocale
   ManifestVersion: 1.4.0
   ```

   `lsnen.shadowbox.yaml`:
   ```yaml
   PackageIdentifier: lsnen.shadowbox
   PackageVersion: 1.1.1
   DefaultLocale: en-US
   ManifestType: version
   ManifestVersion: 1.4.0
   ```

3. **Calculate the SHA256 hash** of your Windows release archive:
   ```bash
   certutil -hashfile shadowbox-windows-x64.zip SHA256
   ```
   Replace `REPLACE_WITH_ACTUAL_SHA256_HASH` with the output.

### Submit to WinGet Repository

1. Fork the [WinGet-pkgs repository](https://github.com/microsoft/winget-pkgs)
2. Add your manifest files to the appropriate directory structure
3. Submit a pull request

### Automate WinGet Manifest Updates

Add a step to your GitHub Actions workflow to automatically create and submit WinGet manifests:

```yaml
- name: Update WinGet manifest
  if: startsWith(github.ref, 'refs/tags/')
  run: |
    # Calculate SHA256 of the new release
    SHA256=$(certutil -hashfile release_files/shadowbox-windows-x64.zip SHA256 | grep -v "hash" | grep -v "CertUtil" | tr -d " ")
    VERSION=${GITHUB_REF#refs/tags/v}
    
    # Create manifest files
    # (Implementation details would go here)
    
    # Submit PR to WinGet repository using GitHub CLI
    # This requires setting up a personal access token with repo permissions
    # and installing the GitHub CLI
```

## 3. Modify Your Build Process

To ensure your package works well with package managers:

1. **Update your PyInstaller spec** to create a truly standalone executable:
   - Ensure all dependencies are bundled
   - Make sure the executable works without any additional files

2. **Add a version command** to your CLI:
   ```python
   if args.version:
       print(f"shadowbox version {__version__}")
       sys.exit(0)
   ```

3. **Ensure your executable is in PATH** when installed:
   - For Homebrew, this happens automatically when installed to `bin/`
   - For WinGet, you need to specify the `PortableCommandAlias`

## 4. Update Your GitHub Release Workflow

Modify your `.github/workflows/release.yml` to:

1. **Calculate and publish SHA256 hashes** for all release artifacts
2. **Generate package manager definitions** automatically
3. **Include installation instructions** for package managers in the release notes

```yaml
- name: Add package manager instructions to release notes
  run: |
    echo "### Package Manager Installation" >> release_notes.md
    echo "" >> release_notes.md
    echo "**macOS (Homebrew):**" >> release_notes.md
    echo '```bash' >> release_notes.md
    echo "brew tap lsnen/shadowbox" >> release_notes.md
    echo "brew install shadowbox" >> release_notes.md
    echo '```' >> release_notes.md
    echo "" >> release_notes.md
    echo "**Windows (WinGet):**" >> release_notes.md
    echo '```bash' >> release_notes.md
    echo "winget install lsnen.shadowbox" >> release_notes.md
    echo '```' >> release_notes.md
```

## 5. Update Documentation

Update your README.md to include package manager installation instructions:

```markdown
## Installation

### Using Package Managers (Recommended)

**macOS (Homebrew):**
```bash
brew tap lsnen/shadowbox
brew install shadowbox
```

**Windows (WinGet):**
```bash
winget install lsnen.shadowbox
```

### Manual Installation

Download the latest release for your platform from the [Releases page](https://github.com/lsnen/shadowbox/releases):
...
```

## Conclusion

By following these steps, you'll make your shadowbox application available through popular package managers, allowing users to easily install, update, and use it from anywhere in their terminal. The key is to:

1. Create the appropriate package definitions
2. Submit them to the respective repositories
3. Automate updates to keep the packages in sync with your releases
4. Ensure your application works well as a standalone command-line tool

This approach significantly improves the user experience and makes your tool more accessible to a wider audience.