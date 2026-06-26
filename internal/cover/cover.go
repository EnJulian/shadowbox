// Package cover resolves and downloads album cover art, trying Spotify first and
// falling back to the iTunes Search API, mirroring the original cascade.
package cover

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/spotify"
)

// Resolver finds cover art URLs across providers.
type Resolver struct {
	Spotify *spotify.Client
	ITunes  *itunes.Client
	HTTP    *http.Client
}

// New builds a Resolver from the given clients.
func New(sp *spotify.Client, it *itunes.Client) *Resolver {
	return &Resolver{
		Spotify: sp,
		ITunes:  it,
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

// URL returns the best cover-art URL for the track, or "" if none is found.
func (r *Resolver) URL(ctx context.Context, title, artist string) string {
	first := firstArtist(artist)

	if r.Spotify != nil && r.Spotify.Configured() {
		if u := r.spotifyCover(ctx, title, artist); u != "" {
			return u
		}
		if first != artist {
			if u := r.spotifyCover(ctx, title, first); u != "" {
				return u
			}
		}
		if u := r.itunesCover(ctx, title+" "+artist); u != "" {
			return u
		}
		if first != artist {
			if u := r.itunesCover(ctx, title+" "+first); u != "" {
				return u
			}
		}
		if u := r.spotifyCover(ctx, title, ""); u != "" {
			return u
		}
	} else {
		if u := r.itunesCover(ctx, title+" "+artist); u != "" {
			return u
		}
		if first != artist {
			if u := r.itunesCover(ctx, title+" "+first); u != "" {
				return u
			}
		}
	}

	if u := r.itunesCover(ctx, title); u != "" {
		return u
	}
	return r.itunesCover(ctx, artist)
}

func (r *Resolver) spotifyCover(ctx context.Context, title, artist string) string {
	m, err := r.Spotify.Search(ctx, title, artist)
	if err != nil || m == nil {
		return ""
	}
	return m.CoverURL
}

func (r *Resolver) itunesCover(ctx context.Context, query string) string {
	if r.ITunes == nil || strings.TrimSpace(query) == "" {
		return ""
	}
	u, err := r.ITunes.CoverURL(ctx, query)
	if err != nil {
		return ""
	}
	return u
}

// Download fetches the image at url, returning its bytes and MIME type.
func (r *Resolver) Download(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "shadowbox/1.0")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" || !strings.HasPrefix(mime, "image/") {
		mime = guessMIME(url)
	}
	return data, mime, nil
}

func guessMIME(url string) string {
	if strings.Contains(strings.ToLower(url), ".png") {
		return "image/png"
	}
	return "image/jpeg"
}

func firstArtist(artist string) string {
	if i := strings.Index(artist, ","); i >= 0 {
		return strings.TrimSpace(artist[:i])
	}
	return artist
}
