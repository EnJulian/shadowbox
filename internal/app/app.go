// Package app orchestrates the Shadowbox download-and-tag pipeline, tying
// together the downloader, metadata APIs, cover-art resolver, and tag writers.
package app

import (
	"github.com/EnJulian/shadowbox/internal/apis/genius"
	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/lastfm"
	"github.com/EnJulian/shadowbox/internal/apis/spotify"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/cover"
	"github.com/EnJulian/shadowbox/internal/download"
)

// App holds the configured clients used across the pipeline.
type App struct {
	cfg     *config.Config
	spotify *spotify.Client
	itunes  *itunes.Client
	lastfm  *lastfm.Client
	genius  *genius.Client
	cover   *cover.Resolver
}

// New builds an App from configuration.
func New(cfg *config.Config) *App {
	sp := spotify.New(cfg.Spotify.ClientID, cfg.Spotify.ClientSecret)
	it := itunes.New()
	return &App{
		cfg:     cfg,
		spotify: sp,
		itunes:  it,
		lastfm:  lastfm.New(),
		genius:  genius.New(cfg.Genius.AccessToken),
		cover:   cover.New(sp, it),
	}
}

// Options controls a single download or enhancement operation.
type Options struct {
	MusicDir   string // base music directory; defaults to config
	Output     string // optional output base filename (no extension)
	Format     string // audio format; defaults to config
	UseSpotify bool   // force Spotify metadata enrichment

	// Progress, when set, receives short human-readable descriptions of each
	// pipeline stage as it begins (e.g. "ripping audio", "writing tags"). It is
	// called from the worker goroutine, so callers must keep the handler
	// non-blocking and concurrency-safe. Optional.
	Progress func(stage string)
}

// step reports a pipeline stage to the Progress handler when one is configured.
func (o Options) step(stage string) {
	if o.Progress != nil {
		o.Progress(stage)
	}
}

func (a *App) format(opts Options) string {
	if opts.Format != "" {
		return opts.Format
	}
	return a.cfg.AudioFormat
}

func (a *App) musicDir(opts Options) string {
	if opts.MusicDir != "" {
		return opts.MusicDir
	}
	return a.cfg.MusicDirectory
}

// newDownloader builds a downloader for the effective format.
func (a *App) newDownloader(opts Options) *download.Downloader {
	return download.New(a.format(opts))
}
