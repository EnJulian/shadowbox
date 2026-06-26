// Package itunes is a minimal client for the public iTunes Search API, used as
// a fallback source for album cover art.
package itunes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client queries the iTunes Search API.
type Client struct {
	HTTP    *http.Client
	BaseURL string // overridable for testing
}

// New returns an iTunes client with the default endpoint.
func New() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 10 * time.Second},
		BaseURL: "https://itunes.apple.com/search",
	}
}

type searchResponse struct {
	Results []struct {
		ArtworkURL100   string `json:"artworkUrl100"`
		PrimaryGenre    string `json:"primaryGenreName"`
		TrackName       string `json:"trackName"`
		ArtistName      string `json:"artistName"`
		CollectionName  string `json:"collectionName"`
		ReleaseDate     string `json:"releaseDate"`
		TrackNumber     int    `json:"trackNumber"`
		TrackCount      int    `json:"trackCount"`
	} `json:"results"`
}

// CoverURL returns a high-resolution cover art URL for the query, or "" if none.
func (c *Client) CoverURL(ctx context.Context, query string) (string, error) {
	res, err := c.search(ctx, query)
	if err != nil {
		return "", err
	}
	if len(res.Results) == 0 {
		return "", nil
	}
	art := res.Results[0].ArtworkURL100
	if art == "" {
		return "", nil
	}
	// Upgrade the thumbnail to a larger render.
	return strings.Replace(art, "100x100bb", "600x600bb", 1), nil
}

func (c *Client) search(ctx context.Context, query string) (*searchResponse, error) {
	u := fmt.Sprintf("%s?term=%s&entity=song&limit=1", c.BaseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "shadowbox/1.0")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("itunes request failed: %s", resp.Status)
	}
	var res searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
