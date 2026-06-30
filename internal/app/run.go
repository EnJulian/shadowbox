package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EnJulian/shadowbox/internal/apis"
	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/musicbrainz"
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

	downloadTarget := query
	if !download.IsURL(query) {
		opts.searchMeta = true
		opts.searchTitle, opts.searchArtist = parseFromQuery(query)
		opts.searchByFormat = queryUsesByFormat(query)

		opts.heading("Searching YouTube for matches")
		results, err := dl.SearchYouTube(ctx, query, 10)
		if err != nil {
			return err
		}
		idx, err := choose(ctx, opts, youtubePrompt(results))
		if err != nil {
			return err
		}
		applySelectedTrackMeta(&opts, results[idx].Title)
		downloadTarget = results[idx].URL
		applog.Infof("SCAN", "Selected: %s", results[idx].Title)
		applog.Infof("SCAN", "Metadata search: %s by %s", opts.searchTitle, opts.searchArtist)
	}

	var file string
	opts.heading("Downloading audio")
	if download.IsURL(downloadTarget) && download.IsBandcamp(downloadTarget) {
		applog.Step("DOWNLOAD", "Fetching from Bandcamp")
		file, err = dl.DownloadFromBandcamp(ctx, downloadTarget, tmpDir)
	} else {
		applog.Step("DOWNLOAD", "Fetching: %s", downloadTarget)
		file, err = dl.Download(ctx, downloadTarget, tmpDir)
	}
	if err != nil {
		return err
	}

	var meta *tag.Metadata
	if download.IsURL(downloadTarget) && download.IsKHInsider(downloadTarget) {
		meta, err = a.buildKHInsiderMetadataFromFile(ctx, file, download.KHInsiderAlbumURL(downloadTarget), opts, nil)
		if err != nil {
			return err
		}
	} else {
		meta, err = a.buildMetadata(ctx, file, opts)
		if err != nil {
			return err
		}
	}
	useExternal := !download.IsURL(downloadTarget) || !download.IsKHInsider(downloadTarget)
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

	opts.heading("Downloading playlist")
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
		opts.headingProgress("Tagging playlist track", i+1, len(files))
		var meta *tag.Metadata
		var err error
		if khInsider {
			meta, err = a.buildKHInsiderMetadataFromFile(ctx, f, url, trackOpts, &khCover)
		} else {
			meta, err = a.buildMetadata(ctx, f, trackOpts)
		}
		if err != nil {
			applog.Error("Failed to build metadata for %s: %v", filepath.Base(f), err)
			continue
		}
		if meta == nil {
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
	opts.heading("Reading existing tags")
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

	meta, err := a.enrich(ctx, &tag.Metadata{
		Title:  title,
		Artist: artist,
		Album:  existing.Album,
		Genre:  existing.Genre,
		Date:   existing.Date,
	}, opts)
	if err != nil {
		return err
	}

	a.attachCoverAndLyrics(ctx, meta, opts, true)
	opts.heading("Writing tags to file")
	if err := tag.Write(path, meta); err != nil {
		return err
	}
	applog.Step("COMPLETE", "Tagged: %s", filepath.Base(path))
	return nil
}

// buildMetadata reads any existing tags, derives title/artist, and enriches with
// online sources.
func (a *App) buildMetadata(ctx context.Context, file string, opts Options) (*tag.Metadata, error) {
	opts.heading("Reading embedded tags")
	existing, _ := tag.Read(file)
	if existing == nil {
		existing = &tag.Metadata{}
	}

	var title, artist string
	switch {
	case opts.searchMeta:
		title, artist = cleanTitle(opts.searchTitle), opts.searchArtist
		if artist == "" {
			artist = "Unknown"
		}
	case existing.Title != "" && existing.Artist != "":
		title, artist = existing.Title, existing.Artist
	default:
		title, artist = existing.Title, existing.Artist
		if title == "" || artist == "" {
			title, artist = parseFromFilename(file)
		}
	}

	return a.enrich(ctx, &tag.Metadata{
		Title:  title,
		Artist: artist,
		Album:  existing.Album,
		Genre:  existing.Genre,
		Date:   existing.Date,
	}, opts)
}

// enrich augments metadata with iTunes (primary), MusicBrainz (fallback), and Last.fm genre.
func (a *App) enrich(ctx context.Context, m *tag.Metadata, opts Options) (*tag.Metadata, error) {
	searchTitle := cleanTitle(m.Title)
	searchArtist := strings.TrimSpace(m.Artist)

	if matched, err := a.matchFromITunes(ctx, m, opts, searchTitle, searchArtist); err != nil {
		return m, err
	} else if matched {
		applog.Successf("ITUNES", "Matched: %s by %s", m.Title, m.Artist)
	} else if matched, err = a.matchFromMusicBrainz(ctx, m, opts, searchTitle, searchArtist); err != nil {
		return m, err
	} else if matched {
		applog.Successf("MUSICBRAINZ", "Matched: %s by %s", m.Title, m.Artist)
	}

	if m.Genre == "" {
		opts.heading("Looking up genre")
		if g, err := a.lastfm.Genre(ctx, m.Title, m.Artist); err == nil && g != "" {
			m.Genre = g
			applog.Systemf("LASTFM", "Genre: %s", g)
		}
	}
	return m, nil
}

func (a *App) matchFromITunes(ctx context.Context, m *tag.Metadata, opts Options, title, artist string) (bool, error) {
	opts.heading("Matching metadata on iTunes")
	applog.Systemf("ITUNES", "Searching metadata: %s by %s", title, artist)

	candidates, err := a.itunes.SearchCandidates(ctx, title, artist)
	if err != nil {
		applog.Warningf("ITUNES", "Search failed: %v", err)
		return false, nil
	}
	candidates = filterITunesCandidates(candidates, artist)
	if len(candidates) == 0 {
		applog.Warningf("ITUNES", "No candidates matched artist %q", artist)
		return false, nil
	}

	idx, err := choose(ctx, opts, itunesPrompt(candidates))
	if err != nil {
		return false, err
	}
	if meta := a.itunes.MetaFromCandidate(candidates[idx]); meta != nil {
		applyTrackMeta(m, meta)
		return true, nil
	}
	return false, nil
}

func (a *App) matchFromMusicBrainz(ctx context.Context, m *tag.Metadata, opts Options, title, artist string) (bool, error) {
	opts.heading("Matching metadata on MusicBrainz")
	applog.Systemf("MUSICBRAINZ", "Searching metadata: %s by %s", title, artist)

	candidates, err := a.musicbrainz.SearchCandidates(ctx, title, artist)
	if err != nil {
		applog.Warningf("MUSICBRAINZ", "Search failed: %v", err)
		return false, nil
	}
	candidates = filterMBCandidates(candidates, artist)
	if len(candidates) == 0 {
		applog.Warningf("MUSICBRAINZ", "No candidates matched artist %q", artist)
		return false, nil
	}

	idx, err := choose(ctx, opts, musicbrainzPrompt(candidates))
	if err != nil {
		return false, err
	}
	if meta, err := a.musicbrainz.MetaFromCandidate(ctx, candidates[idx], title); err == nil && meta != nil {
		applyTrackMeta(m, meta)
		return true, nil
	}
	return false, nil
}

// finalize organises the file into Artist/Album, embeds cover and lyrics, writes
// tags, and returns the final path. When useExternalMeta is false, online cover
// and Genius lyrics are skipped (KHInsider supplies its own metadata).
func (a *App) finalize(ctx context.Context, srcFile string, meta *tag.Metadata, musicDir string, opts Options, useExternalMeta bool) (string, error) {
	a.attachCoverAndLyrics(ctx, meta, opts, useExternalMeta)

	opts.heading("Organizing into library folders")
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

	opts.heading("Writing tags to file")
	if err := tag.Write(finalPath, meta); err != nil {
		return "", fmt.Errorf("writing tags: %w", err)
	}
	return finalPath, nil
}

// attachCoverAndLyrics fills in cover image bytes and lyrics on the metadata.
func (a *App) attachCoverAndLyrics(ctx context.Context, meta *tag.Metadata, opts Options, useExternalSources bool) {
	if useExternalSources && !meta.HasCover() {
		opts.heading("Fetching album cover art")
		if url := a.cover.URL(ctx, meta.Title, meta.Artist); url != "" {
			if data, mime, err := a.cover.Download(ctx, url); err == nil && len(data) > 0 {
				meta.Cover = data
				meta.CoverMIME = mime
				applog.Success("Cover art embedded")
			}
		}
	}

	if useExternalSources && a.cfg.UseGenius && a.genius.Configured() {
		opts.heading("Fetching lyrics from Genius")
		applog.Systemf("LYRICS", "Searching: %s by %s", meta.Title, meta.Artist)
		if lyrics, err := a.genius.Lyrics(ctx, meta.Title, meta.Artist); err == nil && lyrics != "" {
			meta.Lyrics = lyrics
			applog.Successf("LYRICS", "Embedded %d characters", len(lyrics))
		}
	}
}

func applyTrackMeta(m *tag.Metadata, src *apis.TrackMeta) {
	m.Title = src.Title
	m.Artist = src.Artist
	m.Album = src.Album
	m.Date = src.Year()
	m.TrackNumber = src.TrackNumber
	m.TotalTracks = src.TotalTracks
	m.DiscNumber = src.DiscNumber
	m.TotalDiscs = src.TotalDiscs
	m.Performer = src.Performer
	if src.Genre != "" {
		m.Genre = src.Genre
	}
}

func youtubePrompt(results []download.SearchResult) PromptRequest {
	opts := make([]PromptOption, len(results))
	for i, r := range results {
		label := r.Title
		if r.Uploader != "" {
			label += " — " + r.Uploader
		}
		detail := r.Duration
		if r.URL != "" && detail != "" {
			detail += " · " + shortenURL(r.URL)
		} else if r.URL != "" {
			detail = shortenURL(r.URL)
		}
		opts[i] = PromptOption{Label: label, Detail: detail}
	}
	return PromptRequest{Title: "Select track", Options: opts}
}

func itunesPrompt(candidates []itunes.TrackCandidate) PromptRequest {
	opts := make([]PromptOption, len(candidates))
	for i, c := range candidates {
		label := c.Title + " — " + c.Artist
		detail := c.Album
		if c.Date != "" {
			if detail != "" {
				detail += " (" + c.Date + ")"
			} else {
				detail = c.Date
			}
		}
		if c.Genre != "" {
			if detail != "" {
				detail += " · " + c.Genre
			} else {
				detail = c.Genre
			}
		}
		opts[i] = PromptOption{Label: label, Detail: detail}
	}
	return PromptRequest{Title: "Select release", Options: opts}
}

func filterITunesCandidates(candidates []itunes.TrackCandidate, artist string) []itunes.TrackCandidate {
	artist = strings.TrimSpace(artist)
	if artist == "" || strings.EqualFold(artist, "unknown") {
		return candidates
	}
	key := strings.ToLower(artist)
	var kept []itunes.TrackCandidate
	for _, c := range candidates {
		if artistNamesMatch(key, c.Artist) {
			kept = append(kept, c)
		}
	}
	return kept
}

func filterMBCandidates(candidates []musicbrainz.RecordingCandidate, artist string) []musicbrainz.RecordingCandidate {
	artist = strings.TrimSpace(artist)
	if artist == "" || strings.EqualFold(artist, "unknown") {
		return candidates
	}
	key := strings.ToLower(artist)
	var kept []musicbrainz.RecordingCandidate
	for _, c := range candidates {
		if artistNamesMatch(key, c.Artist) {
			kept = append(kept, c)
		}
	}
	return kept
}

func artistNamesMatch(queryKey, candidateArtist string) bool {
	candidateArtist = strings.ToLower(strings.TrimSpace(candidateArtist))
	if candidateArtist == "" {
		return false
	}
	if strings.Contains(candidateArtist, queryKey) || strings.Contains(queryKey, candidateArtist) {
		return true
	}
	for _, part := range strings.Split(candidateArtist, ",") {
		part = strings.TrimSpace(part)
		if part == queryKey || strings.Contains(part, queryKey) || strings.Contains(queryKey, part) {
			return true
		}
	}
	qWords := strings.Fields(queryKey)
	cWords := strings.Fields(candidateArtist)
	if len(qWords) >= 2 && len(cWords) >= 2 {
		if qWords[0] == cWords[0] && qWords[len(qWords)-1] == cWords[len(cWords)-1] {
			return true
		}
	}
	return false
}

func musicbrainzPrompt(candidates []musicbrainz.RecordingCandidate) PromptRequest {
	opts := make([]PromptOption, len(candidates))
	for i, c := range candidates {
		label := c.Title + " — " + c.Artist
		detail := c.Album
		if c.Date != "" {
			if detail != "" {
				detail += " (" + c.Date + ")"
			} else {
				detail = c.Date
			}
		}
		if c.Length != "" {
			if detail != "" {
				detail += " · " + c.Length
			} else {
				detail = c.Length
			}
		}
		opts[i] = PromptOption{Label: label, Detail: detail}
	}
	return PromptRequest{Title: "Select release", Options: opts}
}

func shortenURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= 48 {
		return raw
	}
	return raw[:45] + "..."
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
