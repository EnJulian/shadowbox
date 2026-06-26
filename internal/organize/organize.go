// Package organize handles filename sanitisation and the Artist/Album directory
// layout Shadowbox uses to file downloaded tracks.
package organize

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var invalidChars = regexp.MustCompile(`[\\/*?:"<>|]`)

// SanitizeFilename makes a string safe to use as a file or directory name.
func SanitizeFilename(name string) string {
	s := invalidChars.ReplaceAllString(name, "_")
	s = strings.Trim(s, ". ")
	if s == "" {
		return "Unknown"
	}
	return s
}

// DefaultMusicDir returns ~/Music, falling back to the current directory.
func DefaultMusicDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "Music"
	}
	return filepath.Join(home, "Music")
}

// ArtistDir returns (creating if needed) the directory for an artist beneath
// base. Only the first of several comma-separated artists is used.
func ArtistDir(base, artist string) (string, error) {
	if base == "" {
		base = DefaultMusicDir()
	}
	first := artist
	if i := strings.Index(artist, ","); i >= 0 {
		first = strings.TrimSpace(artist[:i])
	}
	dir := filepath.Join(base, SanitizeFilename(first))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// AlbumDir returns (creating if needed) the album directory inside artistDir.
// It falls back to the track title, then "Unknown Album", and strips a leading
// "Album - " prefix, matching the original behaviour.
func AlbumDir(artistDir, album, title string) (string, error) {
	if album == "" {
		if title != "" {
			album = title
		} else {
			album = "Unknown Album"
		}
	}
	album = strings.TrimPrefix(album, "Album - ")
	dir := filepath.Join(artistDir, SanitizeFilename(album))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// UniquePath returns a path in dir for the given base name and extension,
// appending _1, _2, ... if a file already exists.
func UniquePath(dir, base, ext string) string {
	candidate := filepath.Join(dir, base+ext)
	for i := 1; fileExists(candidate); i++ {
		candidate = filepath.Join(dir, base+"_"+itoa(i)+ext)
	}
	return candidate
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}
