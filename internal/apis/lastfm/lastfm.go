// Package lastfm is a minimal client for the Last.fm API, used only to source a
// genre when MusicBrainz does not provide one.
package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PublicAPIKey is the shared, read-only Last.fm key used by Shadowbox (carried
// over from the original implementation).
const PublicAPIKey = "b25b959554ed76058ac220b7b2e0a026"

// Client queries the Last.fm REST API.
type Client struct {
	APIKey  string
	HTTP    *http.Client
	BaseURL string // overridable for testing
}

// New returns a Last.fm client using the public API key.
func New() *Client {
	return &Client{
		APIKey:  PublicAPIKey,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
		BaseURL: "https://ws.audioscrobbler.com/2.0/",
	}
}

type tagList struct {
	Tag []struct {
		Name string `json:"name"`
	} `json:"tag"`
}

type trackInfoResponse struct {
	Track struct {
		TopTags tagList `json:"toptags"`
	} `json:"track"`
}

type artistTagsResponse struct {
	TopTags tagList `json:"toptags"`
}

// Genre returns a best-effort genre for the track, trying the track's top tags
// first and falling back to the artist's top tags. Returns "" when nothing is
// found.
func (c *Client) Genre(ctx context.Context, title, artist string) (string, error) {
	if artist != "" && title != "" {
		var ti trackInfoResponse
		params := url.Values{
			"method":      {"track.getinfo"},
			"api_key":     {c.APIKey},
			"artist":      {artist},
			"track":       {title},
			"autocorrect": {"1"},
			"format":      {"json"},
		}
		if err := c.get(ctx, params, &ti); err == nil {
			if g := firstTag(ti.Track.TopTags); g != "" {
				return g, nil
			}
		}
	}

	if artist != "" {
		var at artistTagsResponse
		params := url.Values{
			"method":      {"artist.gettoptags"},
			"api_key":     {c.APIKey},
			"artist":      {artist},
			"autocorrect": {"1"},
			"format":      {"json"},
		}
		if err := c.get(ctx, params, &at); err == nil {
			if g := firstTag(at.TopTags); g != "" {
				return g, nil
			}
		}
	}
	return "", nil
}

func (c *Client) get(ctx context.Context, params url.Values, out any) error {
	u := c.BaseURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lastfm request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// firstTag returns the first non-empty, title-cased tag name.
func firstTag(tags tagList) string {
	for _, t := range tags.Tag {
		if name := strings.TrimSpace(t.Name); name != "" {
			return name
		}
	}
	return ""
}
