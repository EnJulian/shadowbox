package download

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/progress"
)

// DownloadPlaylist downloads every track of a YouTube playlist into dir and
// returns the downloaded file paths, ordered by playlist index.
func (d *Downloader) DownloadPlaylist(ctx context.Context, url, dir string) ([]string, error) {
	if err := ValidateInput(url); err != nil {
		return nil, err
	}
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	if IsKHInsider(url) {
		return d.downloadKHInsiderAlbum(ctx, url, dir)
	}

	output := filepath.Join(dir, "%(playlist_index)s - %(title)s.%(ext)s")
	args := []string{
		"-f", "bestaudio",
		"--extract-audio",
		"--audio-format", d.Format,
		"--audio-quality", "0",
		"--embed-metadata",
		"--embed-thumbnail",
		"--add-metadata",
		"--metadata-from-title", "%(artist)s - %(title)s",
		"--parse-metadata", "%(playlist)s:%(album)s",
		"-o", output,
	}
	args = appendTarget(args, url)
	applog.Systemf("GET", "yt-dlp %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	var logBuf bytes.Buffer
	var logMu sync.Mutex
	safeLog := &lockedWriter{mu: &logMu, w: &logBuf}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	done := make(chan struct{}, 2)
	tee := func(r io.Reader) io.Reader {
		return io.TeeReader(r, safeLog)
	}
	go func() {
		d.scanYTDLPProgress(tee(stdout))
		done <- struct{}{}
	}()
	go func() {
		d.scanYTDLPProgress(tee(stderr))
		done <- struct{}{}
	}()
	cmdErr := cmd.Wait()
	<-done
	<-done
	if cmdErr != nil {
		return nil, fmt.Errorf("playlist download failed: %s", truncate(logBuf.String(), 200))
	}

	files := matchingFiles(dir, d.Format)
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s files found after playlist download", d.Format)
	}
	sortByPlaylistIndex(files)
	applog.Successf("DOWNLOAD", "Downloaded %d tracks from playlist", len(files))
	return files, nil
}

func (d *Downloader) downloadKHInsiderAlbum(ctx context.Context, albumURL, dir string) ([]string, error) {
	applog.Infof("PLAYLIST", "Detected KHInsider album")
	tracks, err := scrapeKHInsiderAlbum(ctx, albumURL)
	if err != nil {
		return nil, err
	}
	applog.Infof("PLAYLIST", "Found %d tracks", len(tracks))

	for i, track := range tracks {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(khScrapeDelay):
			}
		}

		directURL, title, err := resolveKHInsiderTrack(ctx, track.PageURL, d.Format)
		if err != nil {
			return nil, fmt.Errorf("track %d (%s): %w", track.Index, track.Title, err)
		}
		if title == "" {
			title = track.Title
		}
		safeTitle := sanitizeKHOutputTitle(title)
		output := filepath.Join(dir, fmt.Sprintf("%02d - %s.%%(ext)s", track.Index, safeTitle))
		applog.Infof("PLAYLIST", "Downloading track %d/%d: %s", track.Index, len(tracks), safeTitle)
		d.reportProgress(track.Index, len(tracks))
		if _, err := d.runDirectStrategies(ctx, directURL, output, dir); err != nil {
			return nil, fmt.Errorf("track %d (%s): %w", track.Index, safeTitle, err)
		}
	}

	files := matchingFiles(dir, d.Format)
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s files found after KHInsider album download", d.Format)
	}
	sortByPlaylistIndex(files)
	applog.Successf("DOWNLOAD", "Downloaded %d tracks from KHInsider album", len(files))
	return files, nil
}

// DownloadFromBandcamp downloads a Bandcamp track/album into dir and returns the
// path to the (first) downloaded file.
func (d *Downloader) DownloadFromBandcamp(ctx context.Context, url, dir string) (string, error) {
	if err := ValidateInput(url); err != nil {
		return "", err
	}
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	output := filepath.Join(dir, "%(title)s.%(ext)s")
	args := []string{
		"--audio-quality", "0",
		"--audio-format", d.Format,
		"-x",
		"-f", "ba",
		"--embed-metadata",
		"--embed-thumbnail",
		"-o", output,
	}
	if d.UseAria2 {
		args = append([]string{"--downloader", "aria2c"}, args...)
	}
	args = appendTarget(args, url)
	applog.Systemf("GET", "yt-dlp %s", strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("bandcamp download failed: %s", truncate(string(out), 200))
	}

	if file := newestAudioFile(dir, d.Format); file != "" {
		return file, nil
	}
	return "", fmt.Errorf("no %s files found after bandcamp download", d.Format)
}

// matchingFiles returns all files in dir matching the target format.
func matchingFiles(dir, format string) []string {
	wantExt := "." + strings.ToLower(format)
	if strings.ToLower(format) == "alac" {
		wantExt = ".m4a"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(e.Name())) == wantExt {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files
}

// sortByPlaylistIndex orders files by the leading "NN - " playlist index.
func sortByPlaylistIndex(files []string) {
	indexOf := func(path string) int {
		base := filepath.Base(path)
		if i := strings.Index(base, " - "); i > 0 {
			if n, err := strconv.Atoi(strings.TrimSpace(base[:i])); err == nil {
				return n
			}
		}
		return 999
	}
	sort.SliceStable(files, func(i, j int) bool {
		return indexOf(files[i]) < indexOf(files[j])
	})
}

var ytDLPItemProgress = regexp.MustCompile(`(?i)(?:item|video)\s+(\d+)\s+of\s+(\d+)`)

type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

func (d *Downloader) reportProgress(current, total int) {
	if d.Progress != nil {
		heading := "Downloading playlist"
		if total > 0 && current > 0 {
			heading = fmt.Sprintf("Downloading playlist track %d/%d", current, total)
		}
		d.Progress(progress.Update{Heading: heading, Current: current, Total: total})
	}
}

func (d *Downloader) scanYTDLPProgress(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if m := ytDLPItemProgress.FindStringSubmatch(line); len(m) == 3 {
			current, _ := strconv.Atoi(m[1])
			total, _ := strconv.Atoi(m[2])
			if current > 0 && total > 0 {
				d.reportProgress(current, total)
			}
		}
	}
}
