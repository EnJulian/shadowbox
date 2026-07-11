// Package suggest provides local-only autocomplete data sources: recent
// search history, library fuzzy matching, and clipboard URL detection.
package suggest

import (
	"os"
	"strings"
)

const maxHistoryEntries = 20

// History stores recent search queries, most-recent-first.
type History struct {
	entries []string
}

// LoadHistory reads a newline-delimited history file. A missing file is not
// an error — it just means no history yet.
func LoadHistory(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &History{}, nil
		}
		return nil, err
	}
	h := &History{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
	return h, nil
}

// Save writes the history to path, most-recent-first, one query per line.
func (h *History) Save(path string) error {
	return os.WriteFile(path, []byte(strings.Join(h.entries, "\n")), 0o600)
}

// Add records a query as most recent, de-duplicating and capping the list.
func (h *History) Add(query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	filtered := h.entries[:0]
	for _, e := range h.entries {
		if !strings.EqualFold(e, query) {
			filtered = append(filtered, e)
		}
	}
	h.entries = append([]string{query}, filtered...)
	if len(h.entries) > maxHistoryEntries {
		h.entries = h.entries[:maxHistoryEntries]
	}
}

// Matches returns up to limit history entries whose prefix case-insensitively
// matches prefix, most-recent-first. An empty prefix matches everything.
func (h *History) Matches(prefix string, limit int) []string {
	var out []string
	lowerPrefix := strings.ToLower(prefix)
	for _, e := range h.entries {
		if strings.HasPrefix(strings.ToLower(e), lowerPrefix) {
			out = append(out, e)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}
