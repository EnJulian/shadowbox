package download

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	applog "github.com/EnJulian/shadowbox/internal/log"
)

var urlRe = regexp.MustCompile(`(?i)^(?:http|https)://`)

// ErrDisallowedHost is returned when a URL points to a non-allowlisted host.
var ErrDisallowedHost = errors.New("URL host is not an allowed download source")

// IsURL reports whether text looks like an http(s) URL.
func IsURL(text string) bool {
	return urlRe.MatchString(strings.TrimSpace(text))
}

// IsBandcamp reports whether the URL points to Bandcamp.
func IsBandcamp(rawURL string) bool {
	return strings.Contains(strings.ToLower(rawURL), "bandcamp.com")
}

// IsYouTube reports whether the URL points to YouTube.
func IsYouTube(rawURL string) bool {
	u := strings.ToLower(rawURL)
	return strings.Contains(u, "youtube.com") || strings.Contains(u, "youtu.be")
}

// IsYouTubePlaylist reports whether the URL is a YouTube playlist.
func IsYouTubePlaylist(rawURL string) bool {
	u := strings.ToLower(rawURL)
	return IsYouTube(rawURL) && (strings.Contains(u, "playlist") || strings.Contains(u, "list="))
}

// ValidateInput checks query/URL inputs before passing them to yt-dlp.
func ValidateInput(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return errors.New("empty input")
	}
	for _, r := range input {
		if unicode.IsControl(r) {
			return fmt.Errorf("input contains control character")
		}
	}
	if !IsURL(input) {
		if strings.Contains(input, "://") {
			return fmt.Errorf("unsupported URL scheme")
		}
		return nil
	}
	return validateURL(input)
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "https":
	case "http":
		applog.Warning("HTTP URL used; prefer HTTPS for downloads")
	default:
		return fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	if !isAllowedHost(host) {
		return fmt.Errorf("%w: %s", ErrDisallowedHost, host)
	}
	return nil
}

func isAllowedHost(host string) bool {
	if host == "" {
		return false
	}
	if host == "youtu.be" ||
		host == "youtube.com" ||
		host == "www.youtube.com" ||
		host == "music.youtube.com" ||
		strings.HasSuffix(host, ".youtube.com") {
		return true
	}
	if host == "bandcamp.com" || strings.HasSuffix(host, ".bandcamp.com") {
		return true
	}
	return false
}

// appendTarget appends the option terminator and target so user input cannot be
// parsed as yt-dlp flags.
func appendTarget(args []string, target string) []string {
	return append(args, "--", target)
}
