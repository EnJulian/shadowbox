// Package app orchestrates the Shadowbox download-and-tag pipeline, tying
// together the downloader, metadata APIs, cover-art resolver, and tag writers.
package app

import (
	"fmt"

	"github.com/EnJulian/shadowbox/internal/apis/genius"
	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/lastfm"
	"github.com/EnJulian/shadowbox/internal/apis/musicbrainz"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/cover"
	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/progress"
)

// App holds the configured clients used across the pipeline.
type App struct {
	cfg         *config.Config
	musicbrainz *musicbrainz.Client
	itunes      *itunes.Client
	lastfm      *lastfm.Client
	genius      *genius.Client
	cover       *cover.Resolver
}

// New builds an App from configuration.
func New(cfg *config.Config) *App {
	mb := musicbrainz.New()
	it := itunes.New()
	return &App{
		cfg:         cfg,
		musicbrainz: mb,
		itunes:      it,
		lastfm:      lastfm.New(),
		genius:      genius.New(cfg.Genius.AccessToken),
		cover:       cover.New(mb, it),
	}
}

// Options controls a single download or enhancement operation.
type Options struct {
	MusicDir string // base music directory; defaults to config
	Output   string // optional output base filename (no extension)
	Format   string // audio format; defaults to config

	// Progress, when set, receives pipeline stage updates for the UI. It is
	// called from the worker goroutine, so callers must keep the handler
	// non-blocking and concurrency-safe. Optional.
	Progress func(progress.Update)

	// Select, when set, drives interactive prompts from the TUI. When nil,
	// choose falls back to numbered stdin prompts for CLI use.
	Select SelectFunc

	// AutoAcceptTopMatch, when set, makes choose return the top-ranked
	// candidate without prompting (skips both Select and CLI stdin). Set by
	// EnhanceDir so batch enhancement runs unattended.
	AutoAcceptTopMatch bool

	// Search metadata captured from a text query (not URLs). Set internally by Run.
	searchMeta     bool
	searchTitle    string
	searchArtist   string
	searchByFormat bool
}

// heading updates the running-screen label shown during a pipeline step.
func (o Options) heading(text string) {
	o.report(progress.Update{Heading: text})
}

// headingProgress updates the running-screen label with an optional item counter.
func (o Options) headingProgress(text string, current, total int) {
	heading := text
	if total > 0 && current > 0 {
		heading = fmt.Sprintf("%s %d/%d", text, current, total)
	}
	o.report(progress.Update{Heading: heading, Current: current, Total: total})
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
