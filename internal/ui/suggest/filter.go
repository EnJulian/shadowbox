package suggest

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Filter returns the entries whose text case-insensitively contains query.
// An empty query returns entries unchanged.
func Filter(entries []string, query string) []string {
	if query == "" {
		return entries
	}
	lowerQuery := strings.ToLower(query)
	var out []string
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e), lowerQuery) {
			out = append(out, e)
		}
	}
	return out
}

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true}

// LibraryMatches scans musicDir (Artist/Album/Track, one level of each) and
// returns up to limit artist, album, and track names whose text
// case-insensitively contains query.
func LibraryMatches(musicDir, query string, limit int) []string {
	var out []string
	appendMatch := func(name string) bool {
		if strings.Contains(strings.ToLower(name), strings.ToLower(query)) {
			out = append(out, name)
		}
		return len(out) >= limit
	}

	artists, err := os.ReadDir(musicDir)
	if err != nil {
		return nil
	}
	sort.Slice(artists, func(i, j int) bool { return artists[i].Name() < artists[j].Name() })

	for _, artist := range artists {
		if !artist.IsDir() {
			continue
		}
		if appendMatch(artist.Name()) {
			return out
		}
		artistDir := filepath.Join(musicDir, artist.Name())
		albums, err := os.ReadDir(artistDir)
		if err != nil {
			continue
		}
		for _, album := range albums {
			if album.IsDir() {
				if appendMatch(album.Name()) {
					return out
				}
				albumDir := filepath.Join(artistDir, album.Name())
				tracks, err := os.ReadDir(albumDir)
				if err != nil {
					continue
				}
				for _, track := range tracks {
					if !track.IsDir() && audioExts[strings.ToLower(filepath.Ext(track.Name()))] {
						if appendMatch(track.Name()) {
							return out
						}
					}
				}
			}
		}
	}
	return out
}
