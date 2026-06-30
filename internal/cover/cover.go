// Package cover resolves and downloads album cover art, trying the iTunes Search
// API first and falling back to MusicBrainz/Cover Art Archive.
package cover

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/musicbrainz"
)

const maxCoverBytes = 10 << 20 // 10 MiB

// Resolver finds cover art URLs across providers.
type Resolver struct {
	MusicBrainz *musicbrainz.Client
	ITunes      *itunes.Client
	HTTP        *http.Client
}

// New builds a Resolver from the given clients.
func New(mb *musicbrainz.Client, it *itunes.Client) *Resolver {
	return &Resolver{
		MusicBrainz: mb,
		ITunes:      it,
		HTTP:        &http.Client{Timeout: 15 * time.Second},
	}
}

// URL returns the best cover-art URL for the track, or "" if none is found.
func (r *Resolver) URL(ctx context.Context, title, artist string) string {
	first := firstArtist(artist)

	if u := r.itunesCover(ctx, title+" "+artist); u != "" {
		return u
	}
	if first != artist {
		if u := r.itunesCover(ctx, title+" "+first); u != "" {
			return u
		}
	}
	if u := r.itunesCover(ctx, title); u != "" {
		return u
	}
	if u := r.itunesCover(ctx, artist); u != "" {
		return u
	}

	if r.MusicBrainz != nil {
		if u := r.musicBrainzCover(ctx, title, artist); u != "" {
			return u
		}
		if first != artist {
			if u := r.musicBrainzCover(ctx, title, first); u != "" {
				return u
			}
		}
	}
	return ""
}

func (r *Resolver) musicBrainzCover(ctx context.Context, title, artist string) string {
	m, err := r.MusicBrainz.Search(ctx, title, artist)
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
func (r *Resolver) Download(ctx context.Context, rawURL string) ([]byte, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, "", fmt.Errorf("invalid cover URL: %w", err)
	}
	if strings.ToLower(u.Scheme) != "https" {
		return nil, "", fmt.Errorf("cover URL must use HTTPS")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "shadowbox/1.0")
	resp, err := r.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("cover download failed: %s", resp.Status)
	}

	mime := resp.Header.Get("Content-Type")
	if mime == "" || !strings.HasPrefix(strings.ToLower(mime), "image/") {
		return nil, "", fmt.Errorf("cover response is not an image (content-type: %q)", mime)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxCoverBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxCoverBytes {
		return nil, "", fmt.Errorf("cover image exceeds %d byte limit", maxCoverBytes)
	}
	return data, mime, nil
}

func firstArtist(artist string) string {
	if i := strings.Index(artist, ","); i >= 0 {
		return strings.TrimSpace(artist[:i])
	}
	return artist
}
