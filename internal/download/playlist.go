package download

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	applog "github.com/EnJulian/shadowbox/internal/log"
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
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("playlist download failed: %s", truncate(string(out), 200))
	}

	files := matchingFiles(dir, d.Format)
	if len(files) == 0 {
		return nil, fmt.Errorf("no %s files found after playlist download", d.Format)
	}
	sortByPlaylistIndex(files)
	applog.Successf("DOWNLOAD", "Downloaded %d tracks from playlist", len(files))
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
