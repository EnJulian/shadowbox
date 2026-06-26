package download

import (
	"regexp"
	"strings"
)

var urlRe = regexp.MustCompile(`(?i)^(?:http|https)://`)

// IsURL reports whether text looks like an http(s) URL.
func IsURL(text string) bool {
	return urlRe.MatchString(strings.TrimSpace(text))
}

// IsBandcamp reports whether the URL points to Bandcamp.
func IsBandcamp(url string) bool {
	return strings.Contains(strings.ToLower(url), "bandcamp.com")
}

// IsYouTube reports whether the URL points to YouTube.
func IsYouTube(url string) bool {
	u := strings.ToLower(url)
	return strings.Contains(u, "youtube.com") || strings.Contains(u, "youtu.be")
}

// IsYouTubePlaylist reports whether the URL is a YouTube playlist.
func IsYouTubePlaylist(url string) bool {
	u := strings.ToLower(url)
	return IsYouTube(url) && (strings.Contains(u, "playlist") || strings.Contains(u, "list="))
}
