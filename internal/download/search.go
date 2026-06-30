package download

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SearchResult is one YouTube search hit from yt-dlp.
type SearchResult struct {
	Title    string
	Uploader string
	Duration string
	URL      string
}

type ytSearchPlaylist struct {
	Entries []ytSearchEntry `json:"entries"`
}

type ytSearchEntry struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	URL      string  `json:"url"`
	Webpage  string  `json:"webpage_url"`
	Uploader string  `json:"uploader"`
	Duration float64 `json:"duration"`
}

// SearchYouTube queries YouTube via yt-dlp and returns up to limit results.
func (d *Downloader) SearchYouTube(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty search query")
	}

	target := fmt.Sprintf("ytsearch%d:%s", limit, query)
	args := []string{
		"--flat-playlist",
		"--dump-single-json",
		"--no-warnings",
		"--no-download",
		target,
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("youtube search failed: %w", err)
	}

	results, err := parseSearchJSON(out)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no youtube results for %q", query)
	}
	return results, nil
}

func parseSearchJSON(data []byte) ([]SearchResult, error) {
	var pl ytSearchPlaylist
	if err := json.Unmarshal(data, &pl); err == nil && len(pl.Entries) > 0 {
		return mapSearchEntries(pl.Entries), nil
	}

	var results []SearchResult
	dec := json.NewDecoder(strings.NewReader(string(data)))
	for dec.More() {
		var e ytSearchEntry
		if err := dec.Decode(&e); err != nil {
			continue
		}
		if r := entryToResult(e); r.URL != "" {
			results = append(results, r)
		}
	}
	if len(results) > 0 {
		return results, nil
	}
	return nil, fmt.Errorf("no entries in search response")
}

func mapSearchEntries(entries []ytSearchEntry) []SearchResult {
	var out []SearchResult
	for _, e := range entries {
		if r := entryToResult(e); r.URL != "" {
			out = append(out, r)
		}
	}
	return out
}

func entryToResult(e ytSearchEntry) SearchResult {
	url := e.URL
	if url == "" {
		url = e.Webpage
	}
	if url == "" && e.ID != "" {
		url = "https://www.youtube.com/watch?v=" + e.ID
	}
	return SearchResult{
		Title:    strings.TrimSpace(e.Title),
		Uploader: strings.TrimSpace(e.Uploader),
		Duration: formatDuration(e.Duration),
		URL:      url,
	}
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return ""
	}
	total := int(seconds + 0.5)
	m, s := total/60, total%60
	return strconv.Itoa(m) + ":" + fmt.Sprintf("%02d", s)
}
