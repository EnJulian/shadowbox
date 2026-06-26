package app

import (
	"path/filepath"
	"regexp"
	"strings"
)

// titleSuffixRe strips common YouTube-style suffixes like "(Official Video)".
var titleSuffixRe = regexp.MustCompile(`(?i)\s*[\(\[][^\)\]]*?(official|music|video|audio|lyrics|hd|4k)[^\)\]]*?[\)\]]`)

// separators used when splitting "Artist - Title" style strings.
var separators = []string{" - ", " – ", " — ", " : ", " | "}

// twoWordArtists are well-known artists whose names are two words; used to keep
// the heuristic from splitting them, mirroring the original implementation.
var twoWordArtists = map[string]bool{
	"imagine dragons": true,
	"pink floyd":      true,
	"led zeppelin":    true,
	"rolling stones":  true,
}

// parseFromQuery derives a best-effort (title, artist) from a search query.
func parseFromQuery(query string) (title, artist string) {
	for _, sep := range separators {
		if i := strings.Index(query, sep); i >= 0 {
			return cleanTitle(strings.TrimSpace(query[i+len(sep):])), strings.TrimSpace(query[:i])
		}
	}

	parts := strings.Fields(query)
	if len(parts) >= 3 && twoWordArtists[strings.ToLower(parts[0]+" "+parts[1])] {
		return strings.Join(parts[2:], " "), parts[0] + " " + parts[1]
	}
	switch {
	case len(parts) >= 3:
		return strings.Join(parts[1:], " "), parts[0]
	case len(parts) == 2:
		return parts[1], parts[0]
	default:
		return query, "Unknown"
	}
}

// parseFromFilename derives (title, artist) from a downloaded filename.
func parseFromFilename(path string) (title, artist string) {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	// Drop a leading playlist index like "01 - ".
	name = regexp.MustCompile(`^\d+\s*-\s*`).ReplaceAllString(name, "")
	if i := strings.Index(name, " - "); i >= 0 {
		return cleanTitle(strings.TrimSpace(name[i+3:])), strings.TrimSpace(name[:i])
	}
	return cleanTitle(name), "Unknown"
}

// cleanTitle removes common promotional suffixes from a title.
func cleanTitle(title string) string {
	return strings.TrimSpace(titleSuffixRe.ReplaceAllString(title, ""))
}
