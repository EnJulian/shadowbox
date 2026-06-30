package app

import (
	"path/filepath"
	"regexp"
	"strings"
)

// titleSuffixRe strips common YouTube-style suffixes like "(Official Video)".
var titleSuffixRe = regexp.MustCompile(`(?i)\s*[\(\[][^\)\]]*?(official|music|video|audio|lyrics|hd|4k)[^\)\]]*?[\)\]]`)

// parenAnnotationRe strips trailing parenthetical/bracket annotations such as
// "(Lo-Fi Synthwave)" or "[Remastered]".
var parenAnnotationRe = regexp.MustCompile(`(?i)\s*[\(\[][^\)\]]*[\)\]]`)

// bySeparatorRe splits "Title by Artist" search queries (case-insensitive).
var bySeparatorRe = regexp.MustCompile(`(?i)\s+by\s+`)

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
// Preferred format: "<title> by <artist>". Also supports "Artist - Title" and
// legacy "Artist Title" (artist-first) heuristics.
func parseFromQuery(query string) (title, artist string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", "Unknown"
	}

	if loc := bySeparatorRe.FindStringIndex(query); loc != nil {
		return cleanTitle(strings.TrimSpace(query[:loc[0]])), strings.TrimSpace(query[loc[1]:])
	}

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

// cleanTitle removes promotional suffixes and trailing annotations from a title.
func cleanTitle(title string) string {
	title = strings.TrimSpace(titleSuffixRe.ReplaceAllString(title, ""))
	for {
		trimmed := strings.TrimSpace(parenAnnotationRe.ReplaceAllString(title, ""))
		if trimmed == title {
			break
		}
		title = trimmed
	}
	return title
}

func queryUsesByFormat(query string) bool {
	return bySeparatorRe.MatchString(strings.TrimSpace(query))
}

// parseYouTubeTitle splits a typical YouTube upload title "Artist - Title".
func parseYouTubeTitle(display string) (title, artist string, ok bool) {
	display = strings.TrimSpace(display)
	for _, sep := range separators {
		if i := strings.Index(display, sep); i >= 0 {
			return cleanTitle(strings.TrimSpace(display[i+len(sep):])), strings.TrimSpace(display[:i]), true
		}
	}
	return "", "", false
}

// applySelectedTrackMeta refines search metadata from the chosen YouTube result.
// When the user searched with "title by artist", both fields are kept as parsed
// from the query and the YouTube display title is not applied.
func applySelectedTrackMeta(opts *Options, videoTitle string) {
	if !opts.searchMeta || opts.searchByFormat {
		return
	}
	ytTitle, ytArtist, ok := parseYouTubeTitle(videoTitle)
	if !ok {
		return
	}
	opts.searchTitle = ytTitle
	if (opts.searchArtist == "" || opts.searchArtist == "Unknown") && ytArtist != "" {
		opts.searchArtist = ytArtist
	}
}
