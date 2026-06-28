package app

import (
	"context"
	"fmt"

	"github.com/EnJulian/shadowbox/internal/download"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/tag"
)

// khInsiderCover caches album cover bytes across a playlist run.
type khInsiderCover struct {
	data    []byte
	mime    string
	fetched bool
}

func (a *App) buildKHInsiderMetadataFromFile(ctx context.Context, file, albumURL string, opts Options, cover *khInsiderCover) (*tag.Metadata, error) {
	opts.step("reading embedded KHInsider tags")
	m, err := tag.Read(file)
	if err != nil {
		applog.Warningf("KHINSIDER", "Could not read embedded tags: %v", err)
		m = &tag.Metadata{}
	}
	if m.Title == "" {
		title, _ := parseFromFilename(file)
		if title != "" {
			m.Title = title
		}
	}
	if m.Title == "" {
		return nil, fmt.Errorf("no title in embedded tags or filename for %s", file)
	}

	if cover != nil {
		cover.attach(ctx, albumURL, m, opts)
	} else {
		attachKHInsiderCover(ctx, albumURL, m, opts)
	}
	return m, nil
}

func attachKHInsiderCover(ctx context.Context, albumURL string, m *tag.Metadata, opts Options) {
	if m.HasCover() || albumURL == "" {
		return
	}
	opts.step("fetching KHInsider cover art")
	album, err := download.FetchKHInsiderAlbumInfo(ctx, albumURL)
	if err != nil {
		applog.Warningf("KHINSIDER", "Album page fetch failed: %v", err)
		return
	}
	if album.CoverURL == "" {
		return
	}
	applog.Systemf("KHINSIDER", "Cover: %s", album.CoverURL)
	data, mime, err := download.FetchKHCover(ctx, album.CoverURL)
	if err != nil {
		applog.Warningf("KHINSIDER", "Cover download failed: %v", err)
		return
	}
	if len(data) > 0 {
		m.Cover = data
		m.CoverMIME = mime
		applog.Success("KHInsider cover art embedded")
	}
}

func (c *khInsiderCover) attach(ctx context.Context, albumURL string, m *tag.Metadata, opts Options) {
	if m.HasCover() {
		return
	}
	if c.fetched {
		if len(c.data) > 0 {
			m.Cover = c.data
			m.CoverMIME = c.mime
		}
		return
	}
	c.fetched = true
	attachKHInsiderCover(ctx, albumURL, m, opts)
	c.data = m.Cover
	c.mime = m.CoverMIME
}
