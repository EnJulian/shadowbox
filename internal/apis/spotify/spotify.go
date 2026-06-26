// Package spotify is a minimal client for the Spotify Web API using the
// Client Credentials flow. It supports the track search and album lookup
// Shadowbox needs for metadata enrichment.
package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/EnJulian/shadowbox/internal/apis"
)

// ErrNoCredentials is returned when client credentials are not configured.
var ErrNoCredentials = errors.New("spotify credentials not configured")

// Client talks to the Spotify Web API.
type Client struct {
	ID     string
	Secret string
	HTTP   *http.Client

	// TokenURL and APIURL are overridable for testing.
	TokenURL string
	APIURL   string

	token    string
	tokenExp time.Time
}

// New returns a Spotify client with the default endpoints.
func New(id, secret string) *Client {
	return &Client{
		ID:       id,
		Secret:   secret,
		HTTP:     &http.Client{Timeout: 15 * time.Second},
		TokenURL: "https://accounts.spotify.com/api/token",
		APIURL:   "https://api.spotify.com/v1",
	}
}

// Configured reports whether credentials are present.
func (c *Client) Configured() bool {
	return c.ID != "" && c.Secret != ""
}

func (c *Client) ensureToken(ctx context.Context) error {
	if !c.Configured() {
		return ErrNoCredentials
	}
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}

	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.ID, c.Secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("spotify token request failed: %s", resp.Status)
	}

	var tr struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return err
	}
	c.token = tr.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(tr.ExpiresIn-30) * time.Second)
	return nil
}

// Spotify API response shapes (only the fields we use).
type searchResponse struct {
	Tracks struct {
		Items []track `json:"items"`
	} `json:"tracks"`
}

type track struct {
	Name        string   `json:"name"`
	TrackNumber int      `json:"track_number"`
	DiscNumber  int      `json:"disc_number"`
	Artists     []artist `json:"artists"`
	Album       album    `json:"album"`
}

type artist struct {
	Name string `json:"name"`
}

type album struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ReleaseDate string  `json:"release_date"`
	TotalTracks int     `json:"total_tracks"`
	Images      []image `json:"images"`
}

type image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type albumResponse struct {
	Genres []string `json:"genres"`
	Tracks struct {
		Items []struct {
			DiscNumber int `json:"disc_number"`
		} `json:"items"`
	} `json:"tracks"`
}

// Search looks up a track by title and optional artist, returning normalised
// metadata. It returns (nil, nil) when nothing is found.
func (c *Client) Search(ctx context.Context, title, artist string) (*apis.TrackMeta, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	query := title
	if artist != "" {
		query = fmt.Sprintf("track:%s artist:%s", title, artist)
	}

	res, err := c.search(ctx, query)
	if err != nil {
		return nil, err
	}
	// Fall back to a title-only search when an artist-qualified query is empty.
	if len(res.Tracks.Items) == 0 && artist != "" {
		if res, err = c.search(ctx, title); err != nil {
			return nil, err
		}
	}
	if len(res.Tracks.Items) == 0 {
		return nil, nil
	}

	t := res.Tracks.Items[0]
	artists := make([]string, len(t.Artists))
	for i, a := range t.Artists {
		artists[i] = a.Name
	}
	joined := strings.Join(artists, ", ")

	meta := &apis.TrackMeta{
		Title:       t.Name,
		Artist:      joined,
		Album:       t.Album.Name,
		ReleaseDate: t.Album.ReleaseDate,
		TrackNumber: strconv.Itoa(t.TrackNumber),
		TotalTracks: strconv.Itoa(t.Album.TotalTracks),
		DiscNumber:  strconv.Itoa(t.DiscNumber),
		Performer:   joined,
		CoverURL:    largestImage(t.Album.Images),
	}

	// Enrich with album genre and total disc count.
	if t.Album.ID != "" {
		if ai, err := c.album(ctx, t.Album.ID); err == nil {
			if len(ai.Genres) > 0 {
				meta.Genre = ai.Genres[0]
			}
			total := 1
			for _, it := range ai.Tracks.Items {
				if it.DiscNumber > total {
					total = it.DiscNumber
				}
			}
			meta.TotalDiscs = strconv.Itoa(total)
		}
	}
	return meta, nil
}

func (c *Client) search(ctx context.Context, query string) (*searchResponse, error) {
	u := fmt.Sprintf("%s/search?q=%s&type=track&limit=5", c.APIURL, url.QueryEscape(query))
	var res searchResponse
	if err := c.get(ctx, u, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) album(ctx context.Context, id string) (*albumResponse, error) {
	u := fmt.Sprintf("%s/albums/%s", c.APIURL, url.PathEscape(id))
	var res albumResponse
	if err := c.get(ctx, u, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) get(ctx context.Context, u string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("spotify request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func largestImage(images []image) string {
	best := ""
	bestArea := -1
	for _, img := range images {
		if a := img.Width * img.Height; a > bestArea {
			bestArea = a
			best = img.URL
		}
	}
	return best
}
