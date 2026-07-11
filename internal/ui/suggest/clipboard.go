package suggest

import (
	"strings"

	"github.com/atotto/clipboard"
)

// LooksLikeURL reports whether s (after trimming whitespace) starts with
// http:// or https://.
func LooksLikeURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// ClipboardURL returns the current clipboard contents if they look like a
// URL, or "" if the clipboard is empty, unreadable, or not URL-shaped.
func ClipboardURL() string {
	text, err := clipboard.ReadAll()
	if err != nil {
		return ""
	}
	text = strings.TrimSpace(text)
	if LooksLikeURL(text) {
		return text
	}
	return ""
}
