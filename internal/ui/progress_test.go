package ui

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/progress"
)

func TestRenderProgressBarNumbered(t *testing.T) {
	bar := renderProgressBar(progress.Update{Stage: "downloading track", Current: 3, Total: 10}, "14", "245")
	if !strings.Contains(bar, "3/10") {
		t.Fatalf("expected numbered label, got %q", bar)
	}
	if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
		t.Fatalf("expected bar brackets, got %q", bar)
	}
}

func TestRenderProgressBarIndeterminate(t *testing.T) {
	bar := renderProgressBar(progress.Update{Stage: "ripping audio"}, "14", "245")
	if strings.Contains(bar, "/") {
		t.Fatalf("indeterminate bar should not show fraction, got %q", bar)
	}
}

func TestViewRunningPleaseHold(t *testing.T) {
	m := newTestModel()
	m.screen = screenRunning
	m.result = "Downloading"
	m.progress = progress.Update{Stage: "ripping audio"}

	view := m.viewRunning()
	if !strings.Contains(view, "please hold") {
		t.Fatalf("expected please hold, got %q", view)
	}
	if strings.Contains(strings.ToLower(view), "wait") {
		t.Fatalf("should not contain wait text, got %q", view)
	}
}

func TestViewRunningProgressBar(t *testing.T) {
	m := newTestModel()
	m.screen = screenRunning
	m.result = "Downloading playlist"
	m.progress = progress.Update{Stage: "downloading track", Current: 2, Total: 5}

	view := m.viewRunning()
	if !strings.Contains(view, "2/5") {
		t.Fatalf("expected numbered progress, got %q", view)
	}
	if !strings.Contains(view, "downloading track") {
		t.Fatalf("expected stage label, got %q", view)
	}
}
