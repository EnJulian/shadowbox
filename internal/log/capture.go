package log

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const defaultMaxDownloadLogLines = 2000

// DownloadLog captures download pipeline output for the interactive console.
type DownloadLog struct {
	mu       sync.Mutex
	lines    []string
	maxLines int
	filePath string
}

var downloadLog = &DownloadLog{maxLines: defaultMaxDownloadLogLines}

// DownloadLogWriter returns a writer that records log output for the download log
// viewer and on-disk history.
func DownloadLogWriter() *DownloadLog { return downloadLog }

// Write implements io.Writer.
func (d *DownloadLog) Write(p []byte) (int, error) {
	text := string(p)
	if text == "" {
		return 0, nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, line := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		if line == "" {
			continue
		}
		d.appendLocked(line)
	}
	return len(p), nil
}

func (d *DownloadLog) appendLocked(line string) {
	d.lines = append(d.lines, line)
	if len(d.lines) > d.maxLines {
		d.lines = d.lines[len(d.lines)-d.maxLines:]
	}
	d.appendFileLocked(line)
}

func (d *DownloadLog) appendFileLocked(line string) {
	path, err := d.resolveFilePathLocked()
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintln(f, line)
	_ = f.Close()
}

func (d *DownloadLog) resolveFilePathLocked() (string, error) {
	if d.filePath != "" {
		return d.filePath, nil
	}
	path, err := DownloadLogPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	d.filePath = path
	return path, nil
}

// BeginDownloadSession marks the start of a new acquisition run in the log.
func BeginDownloadSession(label string) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	downloadLog.mu.Lock()
	defer downloadLog.mu.Unlock()
	downloadLog.appendLocked(fmt.Sprintf("=== %s — %s ===", ts, label))
}

// DownloadLogText returns the captured log lines joined for display.
func DownloadLogText() string {
	downloadLog.mu.Lock()
	defer downloadLog.mu.Unlock()
	if len(downloadLog.lines) == 0 {
		return ""
	}
	return strings.Join(downloadLog.lines, "\n")
}

// DownloadLogLines returns a copy of captured log lines.
func DownloadLogLines() []string {
	downloadLog.mu.Lock()
	defer downloadLog.mu.Unlock()
	out := make([]string, len(downloadLog.lines))
	copy(out, downloadLog.lines)
	return out
}

// LoadDownloadLog reads the tail of the on-disk download log into memory when the
// in-memory buffer is empty.
func LoadDownloadLog() error {
	downloadLog.mu.Lock()
	defer downloadLog.mu.Unlock()
	if len(downloadLog.lines) > 0 {
		return nil
	}
	path, err := downloadLog.resolveFilePathLocked()
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = f.Close() }()

	var tail []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		tail = append(tail, sc.Text())
		if len(tail) > defaultMaxDownloadLogLines {
			tail = tail[1:]
		}
	}
	downloadLog.lines = tail
	return sc.Err()
}

// DownloadLogPath returns ~/.config/shadowbox/download.log (or platform equivalent).
func DownloadLogPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "shadowbox", "download.log"), nil
}
