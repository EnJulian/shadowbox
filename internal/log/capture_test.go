package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownloadLogCaptureAndSession(t *testing.T) {
	orig := downloadLog
	t.Cleanup(func() { downloadLog = orig })

	dir := t.TempDir()
	downloadLog = &DownloadLog{maxLines: 10, filePath: filepath.Join(dir, "download.log")}

	if _, err := downloadLog.Write([]byte("[STEP] starting\n")); err != nil {
		t.Fatal(err)
	}
	BeginDownloadSession("Downloading")
	if _, err := downloadLog.Write([]byte("[SYSTEM] yt-dlp test\n")); err != nil {
		t.Fatal(err)
	}

	text := DownloadLogText()
	if !strings.Contains(text, "Downloading") {
		t.Errorf("missing session header: %q", text)
	}
	if !strings.Contains(text, "yt-dlp test") {
		t.Errorf("missing log line: %q", text)
	}

	data, err := os.ReadFile(downloadLog.filePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "yt-dlp test") {
		t.Errorf("file missing log line: %q", data)
	}
}

func TestDownloadLogMaxLines(t *testing.T) {
	d := &DownloadLog{maxLines: 3, filePath: filepath.Join(t.TempDir(), "download.log")}
	for range 5 {
		_, _ = d.Write([]byte("line\n"))
	}
	d.mu.Lock()
	n := len(d.lines)
	d.mu.Unlock()
	if n != 3 {
		t.Fatalf("got %d lines, want 3", n)
	}
}

func TestLoadDownloadLogFromFile(t *testing.T) {
	orig := downloadLog
	t.Cleanup(func() { downloadLog = orig })

	dir := t.TempDir()
	path := filepath.Join(dir, "download.log")
	if err := os.WriteFile(path, []byte("older\nnewer\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	downloadLog = &DownloadLog{maxLines: 100, filePath: path}

	if err := LoadDownloadLog(); err != nil {
		t.Fatal(err)
	}
	lines := DownloadLogLines()
	if len(lines) != 2 || lines[1] != "newer" {
		t.Fatalf("got %v", lines)
	}
}
