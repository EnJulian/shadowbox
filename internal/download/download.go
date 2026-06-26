package download

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	applog "github.com/EnJulian/shadowbox/internal/log"
)

// audioExtensions are the output extensions yt-dlp may produce, in preference order.
var audioExtensions = []string{".opus", ".m4a", ".mp3", ".flac", ".wav", ".webm", ".ogg", ".aac"}

// Downloader runs yt-dlp to fetch and convert audio.
type Downloader struct {
	// Format is the target audio format (opus, m4a, mp3, ...).
	Format string
	// UseAria2 enables the aria2c-accelerated strategy when aria2c is present.
	UseAria2 bool
}

// New returns a Downloader for the given format, auto-detecting aria2c.
func New(format string) *Downloader {
	if format == "" {
		format = "opus"
	}
	return &Downloader{Format: format, UseAria2: HasAria2()}
}

// strategy is a single yt-dlp invocation attempt.
type strategy struct {
	name string
	args func(format, output, target string) []string
}

// strategies returns the ordered download strategies, mirroring the Python
// implementation: aria2c acceleration, a standard download, then a
// browser-simulating attempt for anti-bot situations.
func (d *Downloader) strategies() []strategy {
	var out []strategy
	if d.UseAria2 {
		out = append(out, strategy{
			name: "aria2c Accelerated",
			args: func(format, output, target string) []string {
				return []string{
					"--downloader", "aria2c",
					"--downloader-args", "aria2c:-x 16 -s 16 -j 16 --max-connection-per-server=16",
					"-x",
					"-f", "bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio/best",
					"--embed-metadata",
					"--audio-format", format,
					"--retry-sleep", "1",
					"--retries", "3",
					"-o", output,
					target,
				}
			},
		})
	}
	out = append(out,
		strategy{
			name: "Standard Download",
			args: func(format, output, target string) []string {
				return []string{
					"-x",
					"-f", "bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio/best",
					"--embed-metadata",
					"--audio-format", format,
					"--no-check-certificates",
					"--retry-sleep", "2",
					"--retries", "5",
					"--socket-timeout", "30",
					"-o", output,
					target,
				}
			},
		},
		strategy{
			name: "Browser Simulation",
			args: func(format, output, target string) []string {
				return []string{
					"--user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
					"--referer", "https://www.youtube.com/",
					"-x",
					"-f", "bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio/best",
					"--embed-metadata",
					"--audio-format", format,
					"--no-check-certificates",
					"--no-warnings",
					"--retry-sleep", "3",
					"--retries", "3",
					"--socket-timeout", "45",
					"-o", output,
					target,
				}
			},
		},
	)
	return out
}

// fatalErrorPhrases indicate a video that no strategy can recover.
var fatalErrorPhrases = []string{
	"Video unavailable", "Private video", "This video is not available",
	"Video has been removed", "This video is private",
}

// Download fetches a single track for the given query or URL, writing the result
// into dir, and returns the path to the downloaded file.
func (d *Downloader) Download(ctx context.Context, query, dir string) (string, error) {
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	target := query
	switch {
	case IsURL(query):
		if IsBandcamp(query) {
			applog.Infof("AUDIO", "Detected Bandcamp URL")
		} else {
			applog.Infof("AUDIO", "Detected YouTube URL")
		}
	default:
		applog.Infof("SCAN", "Searching for: %s", query)
		target = "ytsearch1:" + query
	}

	output := filepath.Join(dir, "shadowbox_download.%(ext)s")
	return d.runStrategies(ctx, target, output, dir)
}

// runStrategies attempts each strategy in order until one yields a valid file.
func (d *Downloader) runStrategies(ctx context.Context, target, output, dir string) (string, error) {
	strategies := d.strategies()
	for i, s := range strategies {
		if i > 0 {
			delay := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
			applog.Warning("Waiting %.1fs before next strategy...", delay.Seconds())
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
		}
		applog.Warningf("STRATEGY", "Attempt %d/%d: %s", i+1, len(strategies), s.name)

		args := s.args(d.Format, output, target)
		applog.Systemf("GET", "yt-dlp %s", strings.Join(args, " "))

		cmd := exec.CommandContext(ctx, "yt-dlp", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			msg := string(out)
			if containsAny(msg, fatalErrorPhrases) {
				return "", fmt.Errorf("video is unavailable, private, or removed")
			}
			applog.Warningf("STRATEGY", "%q failed: %s", s.name, truncate(msg, 150))
			continue
		}

		if file := newestAudioFile(dir, d.Format); file != "" {
			if info, err := os.Stat(file); err == nil && info.Size() > 1000 {
				applog.Successf("DOWNLOAD", "Complete: %s (%d bytes)", filepath.Base(file), info.Size())
				return file, nil
			}
		}
		applog.Warningf("STRATEGY", "%q completed but produced no valid file", s.name)
	}
	return "", fmt.Errorf("all download strategies failed")
}

// newestAudioFile returns the most recently modified audio file in dir,
// preferring the target format.
func newestAudioFile(dir, format string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	wantExt := "." + strings.ToLower(format)
	if strings.ToLower(format) == "alac" {
		wantExt = ".m4a" // yt-dlp converts ALAC into an M4A container
	}

	type cand struct {
		path string
		mod  time.Time
	}
	var preferred, fallback []cand
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(dir, e.Name())
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == wantExt {
			preferred = append(preferred, cand{path, info.ModTime()})
		} else if isAudioExt(ext) {
			fallback = append(fallback, cand{path, info.ModTime()})
		}
	}

	pick := func(cs []cand) string {
		if len(cs) == 0 {
			return ""
		}
		sort.Slice(cs, func(i, j int) bool { return cs[i].mod.After(cs[j].mod) })
		return cs[0].path
	}
	if p := pick(preferred); p != "" {
		return p
	}
	return pick(fallback)
}

func isAudioExt(ext string) bool {
	for _, a := range audioExtensions {
		if ext == a {
			return true
		}
	}
	return false
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
