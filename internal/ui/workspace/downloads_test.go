package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/progress"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestDownloadsActivateShowsIdleState(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st).Activate()
	view := d.View(80, 20)
	if !strings.Contains(view, "No active download") {
		t.Fatalf("expected idle message, got %q", view)
	}
}

func TestDownloadsSetRunningThenProgressUpdate(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st)
	d.SetRunning("Downloading audio")
	ws, _ := d.Update(progress.Update{Heading: "Tagging track 2/5", Current: 2, Total: 5})
	view := ws.View(80, 20)
	if !strings.Contains(view, "Tagging track 2/5") {
		t.Fatalf("expected heading in view, got %q", view)
	}
	if !strings.Contains(view, "2/5") {
		t.Fatalf("expected numbered progress in view, got %q", view)
	}
}

func TestDownloadsFinishShowsResult(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st)
	d.SetRunning("Downloading audio")
	d.Finish("Download complete", nil)
	view := d.View(80, 20)
	if !strings.Contains(view, "Download complete") {
		t.Fatalf("expected completion summary, got %q", view)
	}
}

func TestDownloadsCancelKeyEmitsCancelTaskMsg(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	d := NewDownloads(st)
	d.SetRunning("Downloading audio")
	_, cmd := d.Update(key("c"))
	if cmd == nil {
		t.Fatal("expected a cmd for the cancel key")
	}
	if _, ok := cmd().(CancelTaskMsg); !ok {
		t.Fatal("expected CancelTaskMsg")
	}
}
