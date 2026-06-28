package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EnJulian/shadowbox/internal/download"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/organize"
	"github.com/EnJulian/shadowbox/internal/tag"
)

// Run downloads a single track or playlist for the given query/URL and tags it.
func (a *App) Run(ctx context.Context, query string, opts Options) error {
	if missing := download.MissingRequired(); len(missing) > 0 {
		return fmt.Errorf("missing required tools: %s (run 'shadowbox doctor')", strings.Join(missing, ", "))
	}
	if download.IsURL(query) && (download.IsYouTubePlaylist(query) || download.IsKHInsiderPlaylist(query)) {
		return a.RunPlaylist(ctx, query, opts)
	}

	dl := a.newDownloader(opts)
	musicDir := a.musicDir(opts)

	tmpDir, err := os.MkdirTemp("", "shadowbox-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var file string
	opts.step("ripping audio")
	if download.IsURL(query) && download.IsBandcamp(query) {
		applog.Step("DOWNLOAD", "Fetching from Bandcamp")
		file, err = dl.DownloadFromBandcamp(ctx, query, tmpDir)
	} else {
		applog.Step("DOWNLOAD", "Fetching: %s", query)
		file, err = dl.Download(ctx, query, tmpDir)
	}
	if err != nil {
		return err
	}

	var meta *tag.Metadata
	if download.IsURL(query) && download.IsKHInsider(query) {
		meta, err = a.buildKHInsiderMetadataFromFile(ctx, file, download.KHInsiderAlbumURL(query), opts, nil)
		if err != nil {
			return err
		}
	} else {
		meta = a.buildMetadata(ctx, file, query, opts)
	}
	useExternal := !(download.IsURL(query) && download.IsKHInsider(query))
	final, err := a.finalize(ctx, file, meta, musicDir, opts, useExternal)
	if err != nil {
		return err
	}
	applog.Step("COMPLETE", "Saved: %s", final)
	return nil
}

// RunPlaylist downloads and tags every track of a YouTube playlist.
func (a *App) RunPlaylist(ctx context.Context, url string, opts Options) error {
	dl := a.newDownloader(opts)
	musicDir := a.musicDir(opts)

	tmpDir, err := os.MkdirTemp("", "shadowbox-playlist-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	opts.step("ripping playlist audio")
	applog.Step("PLAYLIST", "Downloading playlist (this may take a while)")

	khInsider := download.IsKHInsider(url)

	files, err := dl.DownloadPlaylist(ctx, url, tmpDir)
	if err != nil {
		return err
	}

	trackOpts := opts
	trackOpts.Output = ""

	var ok int
	var khCover khInsiderCover
	for i, f := range files {
		opts.stepN("processing track", i+1, len(files))
		var meta *tag.Metadata
		var err error
		if khInsider {
			meta, err = a.buildKHInsiderMetadataFromFile(ctx, f, url, trackOpts, &khCover)
		} else {
			meta = a.buildMetadata(ctx, f, "", trackOpts)
		}
		if err != nil {
			applog.Error("Failed to build metadata for %s: %v", filepath.Base(f), err)
			continue
		}
		if _, err := a.finalize(ctx, f, meta, musicDir, trackOpts, !khInsider); err != nil {
			applog.Error("Failed to process %s: %v", filepath.Base(f), err)
			continue
		}
		ok++
	}
	applog.Step("SUMMARY", "Processed %d of %d tracks", ok, len(files))
	if ok == 0 {
		return fmt.Errorf("no tracks processed successfully")
	}
	return nil
}

// Enhance tags an existing file in place, optionally with title/artist overrides.
func (a *App) Enhance(ctx context.Context, path, title, artist string) error {
	return a.enhanceFile(ctx, path, title, artist, Options{})
}

// enhanceFile is the progress-aware implementation behind Enhance.
func (a *App) enhanceFile(ctx context.Context, path, title, artist string, opts Options) error {
	opts.step("reading existing tags")
	existing, _ := tag.Read(path)
	if existing == nil {
		existing = &tag.Metadata{}
	}
	if title == "" {
		title = existing.Title
	}
	if artist == "" {
		artist = existing.Artist
	}
	if title == "" {
		title, artist = parseFromFilename(path)
	}

	meta := a.enrich(ctx, &tag.Metadata{
		Title:  title,
		Artist: artist,
		Album:  existing.Album,
		Genre:  existing.Genre,
		Date:   existing.Date,
	}, true, opts)

	a.attachCoverAndLyrics(ctx, meta, opts, true)
	opts.step("writing tags")
	if err := tag.Write(path, meta); err != nil {
		return err
	}
	applog.Step("COMPLETE", "Tagged: %s", filepath.Base(path))
	return nil
}

// buildMetadata reads any existing tags, derives title/artist, and enriches with
// online sources.
func (a *App) buildMetadata(ctx context.Context, file, query string, opts Options) *tag.Metadata {
	opts.step("extracting metadata")
	existing, _ := tag.Read(file)
	if existing == nil {
		existing = &tag.Metadata{}
	}

	title, artist := existing.Title, existing.Artist
	if title == "" || artist == "" {
		if download.IsURL(query) || query == "" {
			title, artist = parseFromFilename(file)
		} else {
			title, artist = parseFromQuery(query)
		}
	}

	return a.enrich(ctx, &tag.Metadata{
		Title:  title,
		Artist: artist,
		Album:  existing.Album,
		Genre:  existing.Genre,
		Date:   existing.Date,
	}, opts.UseSpotify, opts)
}

// enrich augments metadata with Spotify (when configured) and a Last.fm genre
// fallback.
func (a *App) enrich(ctx context.Context, m *tag.Metadata, forceSpotify bool, opts Options) *tag.Metadata {
	if a.spotify.Configured() {
		opts.step("matching on Spotify")
		applog.Systemf("SPOTIFY", "Searching metadata: %s by %s", m.Title, m.Artist)
		if sp, err := a.spotify.Search(ctx, m.Title, m.Artist); err == nil && sp != nil {
			m.Title = sp.Title
			m.Artist = sp.Artist
			m.Album = sp.Album
			m.Date = sp.Year()
			m.TrackNumber = sp.TrackNumber
			m.TotalTracks = sp.TotalTracks
			m.DiscNumber = sp.DiscNumber
			m.TotalDiscs = sp.TotalDiscs
			m.Performer = sp.Performer
			if sp.Genre != "" {
				m.Genre = sp.Genre
			}
			applog.Successf("SPOTIFY", "Matched: %s by %s", m.Title, m.Artist)
		}
	} else if forceSpotify {
		applog.Warning("Spotify credentials not configured; skipping enrichment")
	}

	if m.Genre == "" {
		opts.step("finding genre")
		if g, err := a.lastfm.Genre(ctx, m.Title, m.Artist); err == nil && g != "" {
			m.Genre = g
			applog.Systemf("LASTFM", "Genre: %s", g)
		}
	}
	return m
}

// finalize organises the file into Artist/Album, embeds cover and lyrics, writes
// tags, and returns the final path. When useExternalMeta is false, Spotify/Last.fm
// cover and Genius lyrics are skipped (KHInsider supplies its own metadata).
func (a *App) finalize(ctx context.Context, srcFile string, meta *tag.Metadata, musicDir string, opts Options, useExternalMeta bool) (string, error) {
	a.attachCoverAndLyrics(ctx, meta, opts, useExternalMeta)

	opts.step("organizing files")
	artistDir, err := organize.ArtistDir(musicDir, meta.Artist)
	if err != nil {
		return "", err
	}
	albumDir, err := organize.AlbumDir(artistDir, meta.Album, meta.Title)
	if err != nil {
		return "", err
	}

	base := opts.Output
	if base == "" {
		base = organize.SanitizeFilename(meta.Title)
	}
	ext := filepath.Ext(srcFile)
	finalPath := organize.UniquePath(albumDir, base, ext)

	if err := moveFile(srcFile, finalPath); err != nil {
		return "", err
	}

	opts.step("writing tags")
	if err := tag.Write(finalPath, meta); err != nil {
		return "", fmt.Errorf("writing tags: %w", err)
	}
	return finalPath, nil
}

// attachCoverAndLyrics fills in cover image bytes and lyrics on the metadata.
func (a *App) attachCoverAndLyrics(ctx context.Context, meta *tag.Metadata, opts Options, useExternalSources bool) {
	if useExternalSources && !meta.HasCover() {
		opts.step("fetching cover art")
		if url := a.cover.URL(ctx, meta.Title, meta.Artist); url != "" {
			if data, mime, err := a.cover.Download(ctx, url); err == nil && len(data) > 0 {
				meta.Cover = data
				meta.CoverMIME = mime
				applog.Success("Cover art embedded")
			}
		}
	}

	if useExternalSources && a.cfg.UseGenius && a.genius.Configured() {
		opts.step("fetching lyrics")
		applog.Systemf("LYRICS", "Searching: %s by %s", meta.Title, meta.Artist)
		if lyrics, err := a.genius.Lyrics(ctx, meta.Title, meta.Artist); err == nil && lyrics != "" {
			meta.Lyrics = lyrics
			applog.Successf("LYRICS", "Embedded %d characters", len(lyrics))
		}
	}
}

// moveFile moves a file, falling back to copy+remove across filesystems.
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}
