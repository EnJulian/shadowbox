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
	"github.com/EnJulian/shadowbox/internal/progress"
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

	// Progress, when set, receives pipeline stage updates for the UI. It is
	// called from the worker goroutine, so callers must keep the handler
	// non-blocking and concurrency-safe. Optional.
	Progress func(progress.Update)
}

// step reports a pipeline stage without numeric position.
func (o Options) step(stage string) {
	o.report(progress.Update{Stage: stage})
}

// stepN reports a numbered pipeline stage (1-based current).
func (o Options) stepN(stage string, current, total int) {
	o.report(progress.Update{Stage: stage, Current: current, Total: total})
}

func (o Options) report(u progress.Update) {
	if o.Progress != nil {
		o.Progress(u)
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
	dl := download.New(a.format(opts))
	if opts.Progress != nil {
		dl.Progress = opts.Progress
	}
	return dl
}
