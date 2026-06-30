// Package itunes is a minimal client for the public iTunes Search API, used as
// Shadowbox's primary metadata and cover-art source.
package itunes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/EnJulian/shadowbox/internal/apis"
)

const searchLimit = 10

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
	Results []trackResult `json:"results"`
}

type trackResult struct {
	ArtworkURL100  string `json:"artworkUrl100"`
	PrimaryGenre   string `json:"primaryGenreName"`
	TrackName      string `json:"trackName"`
	ArtistName     string `json:"artistName"`
	CollectionName string `json:"collectionName"`
	ReleaseDate    string `json:"releaseDate"`
	TrackNumber    int    `json:"trackNumber"`
	TrackCount     int    `json:"trackCount"`
}

// TrackCandidate is one iTunes search hit for user selection.
type TrackCandidate struct {
	Title       string
	Artist      string
	Album       string
	Genre       string
	Date        string
	TrackNumber string
	TotalTracks string
	CoverURL    string
}

// SearchCandidates looks up songs by title and optional artist.
func (c *Client) SearchCandidates(ctx context.Context, title, artist string) ([]TrackCandidate, error) {
	for _, term := range buildSearchTerms(title, artist) {
		res, err := c.searchSongs(ctx, term, searchLimit)
		if err != nil {
			return nil, err
		}
		if len(res.Results) > 0 {
			return mapCandidates(res.Results), nil
		}
	}
	return nil, nil
}

// MetaFromCandidate converts a chosen candidate into normalised track metadata.
func (c *Client) MetaFromCandidate(cand TrackCandidate) *apis.TrackMeta {
	return &apis.TrackMeta{
		Title:       cand.Title,
		Artist:      cand.Artist,
		Album:       cand.Album,
		ReleaseDate: cand.Date,
		Genre:       cand.Genre,
		TrackNumber: cand.TrackNumber,
		TotalTracks: cand.TotalTracks,
		CoverURL:    cand.CoverURL,
		Performer:   cand.Artist,
	}
}

// CoverURL returns a high-resolution cover art URL for the query, or "" if none.
func (c *Client) CoverURL(ctx context.Context, query string) (string, error) {
	res, err := c.searchSongs(ctx, query, 1)
	if err != nil {
		return "", err
	}
	if len(res.Results) == 0 {
		return "", nil
	}
	return upgradeArtworkURL(res.Results[0].ArtworkURL100), nil
}

func (c *Client) searchSongs(ctx context.Context, query string, limit int) (*searchResponse, error) {
	if limit <= 0 {
		limit = 1
	}
	u := fmt.Sprintf("%s?term=%s&entity=song&limit=%d", c.BaseURL, url.QueryEscape(strings.TrimSpace(query)), limit)
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

func buildSearchTerms(title, artist string) []string {
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)
	var terms []string
	if title != "" && artist != "" && !strings.EqualFold(artist, "unknown") {
		terms = append(terms, title+" "+artist, artist+" "+title)
	}
	if title != "" {
		terms = append(terms, title)
	}
	return terms
}

func mapCandidates(results []trackResult) []TrackCandidate {
	out := make([]TrackCandidate, 0, len(results))
	for _, r := range results {
		out = append(out, mapCandidate(r))
	}
	return out
}

func mapCandidate(r trackResult) TrackCandidate {
	c := TrackCandidate{
		Title:    strings.TrimSpace(r.TrackName),
		Artist:   strings.TrimSpace(r.ArtistName),
		Album:    strings.TrimSpace(r.CollectionName),
		Genre:    strings.TrimSpace(r.PrimaryGenre),
		Date:     yearFromRelease(r.ReleaseDate),
		CoverURL: upgradeArtworkURL(r.ArtworkURL100),
	}
	if r.TrackNumber > 0 {
		c.TrackNumber = strconv.Itoa(r.TrackNumber)
	}
	if r.TrackCount > 0 {
		c.TotalTracks = strconv.Itoa(r.TrackCount)
	}
	return c
}

func upgradeArtworkURL(art string) string {
	if art == "" {
		return ""
	}
	return strings.Replace(art, "100x100bb", "600x600bb", 1)
}

func yearFromRelease(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) >= 4 {
		return raw[:4]
	}
	return raw
}
