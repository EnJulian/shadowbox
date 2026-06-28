// Package tag reads and writes audio metadata (tags, cover art, and lyrics)
// across the formats Shadowbox produces: Opus, MP3, M4A/MP4, and FLAC.
//
// All writers are pure Go so the resulting binary links statically without
// CGO. Opus and FLAC are handled by rewriting Vorbis comments and embedded
// pictures directly; MP3 uses ID3v2 frames; MP4 uses iTunes-style atoms.
package tag

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Metadata holds the tag fields Shadowbox can read or write. Empty fields are
// left untouched on write where the underlying format allows it.
type Metadata struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Date        string // year (YYYY)
	Genre       string
	TrackNumber string
	TotalTracks string
	DiscNumber  string
	TotalDiscs  string
	Composer    string
	Performer   string

	// Lyrics is plain-text, unsynchronised lyrics.
	Lyrics string

	// Cover is raw image bytes (JPEG or PNG). CoverMIME describes its type.
	Cover     []byte
	CoverMIME string
}

// HasCover reports whether cover image data is present.
func (m *Metadata) HasCover() bool { return len(m.Cover) > 0 }

// ErrUnsupportedFormat is returned for file extensions Shadowbox cannot tag.
type ErrUnsupportedFormat struct{ Ext string }

func (e *ErrUnsupportedFormat) Error() string {
	return fmt.Sprintf("unsupported file format for tagging: %q", e.Ext)
}

// Write applies the metadata to the file at path, dispatching by extension.
func Write(path string, m *Metadata) error {
	switch ext(path) {
	case ".mp3":
		return writeMP3(path, m)
	case ".opus", ".ogg":
		return writeOpus(path, m)
	case ".m4a", ".mp4", ".alac", ".aac":
		return writeMP4(path, m)
	case ".flac":
		return writeFLAC(path, m)
	default:
		return &ErrUnsupportedFormat{Ext: ext(path)}
	}
}

// Read extracts metadata from the file at path, dispatching by extension.
func Read(path string) (*Metadata, error) {
	switch ext(path) {
	case ".mp3":
		return readMP3(path)
	case ".opus", ".ogg":
		return readOpus(path)
	case ".m4a", ".mp4", ".alac", ".aac":
		return readMP4(path)
	case ".flac":
		return readFLAC(path)
	default:
		return nil, &ErrUnsupportedFormat{Ext: ext(path)}
	}
}

// splitSlashPair splits ID3/Vorbis "n/total" values into separate parts.
func splitSlashPair(raw string) (n, total string) {
	raw = strings.TrimSpace(raw)
	if i := strings.IndexByte(raw, '/'); i >= 0 {
		return strings.TrimSpace(raw[:i]), strings.TrimSpace(raw[i+1:])
	}
	return raw, ""
}

func ext(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

// firstArtist returns the first artist when several are comma-separated.
func firstArtist(artist string) string {
	if i := strings.Index(artist, ","); i >= 0 {
		return strings.TrimSpace(artist[:i])
	}
	return artist
}

// coverMIME normalises a MIME type, defaulting to image/jpeg.
func coverMIME(m *Metadata) string {
	if m.CoverMIME != "" {
		return m.CoverMIME
	}
	return "image/jpeg"
}
