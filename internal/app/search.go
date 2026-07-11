// internal/app/search.go
package app

import (
	"context"

	"github.com/EnJulian/shadowbox/internal/download"
)

// SearchTracks runs a YouTube search and returns the raw results for the
// caller to browse, without downloading or prompting. Used by the Search
// workspace to show results before the user picks one.
func (a *App) SearchTracks(ctx context.Context, query string, limit int) ([]download.SearchResult, error) {
	dl := a.newDownloader(Options{})
	return dl.SearchYouTube(ctx, query, limit)
}
