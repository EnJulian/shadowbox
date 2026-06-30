package ui

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
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
	m.runningHeading = "Initializing"

	view := m.viewRunning()
	if !strings.Contains(view, "please hold") {
		t.Fatalf("expected please hold, got %q", view)
	}
	if !strings.Contains(view, "Initializing") {
		t.Fatalf("expected Initializing heading, got %q", view)
	}
	if strings.Contains(view, "»") {
		t.Fatalf("should not show stage line, got %q", view)
	}
	if strings.Contains(strings.ToLower(view), "wait") {
		t.Fatalf("should not contain wait text, got %q", view)
	}
}

func TestViewRunningHeadingFromProgress(t *testing.T) {
	m := newTestModel()
	m.screen = screenRunning
	m.runningHeading = "Initializing"
	m.progress = progress.Update{Heading: "Matching metadata on iTunes"}

	view := m.viewRunning()
	if !strings.Contains(view, "Matching metadata on iTunes") {
		t.Fatalf("expected iTunes heading, got %q", view)
	}
}

func TestViewRunningProgressBar(t *testing.T) {
	m := newTestModel()
	m.screen = screenRunning
	m.runningHeading = "Downloading playlist"
	m.progress = progress.Update{Heading: "Tagging playlist track 2/5", Current: 2, Total: 5}

	view := m.viewRunning()
	if !strings.Contains(view, "2/5") {
		t.Fatalf("expected numbered progress, got %q", view)
	}
	if !strings.Contains(view, "Tagging playlist track 2/5") {
		t.Fatalf("expected heading in view, got %q", view)
	}
	if strings.Contains(view, "»") {
		t.Fatalf("should not show stage line, got %q", view)
	}
}

func TestViewPickerRendersOptions(t *testing.T) {
	m := newTestModel()
	m.screen = screenPicker
	m.picker = pickerState{
		title: "Select track",
		options: []app.PromptOption{
			{Label: "Song A — Artist", Detail: "Album (2020) · 3:30"},
			{Label: "Song B — Artist", Detail: "Single (2021)"},
		},
		cursor: 0,
	}
	view := m.viewPicker()
	if !strings.Contains(view, "Select track") {
		t.Fatalf("missing title: %q", view)
	}
	if !strings.Contains(view, "Song A — Artist") {
		t.Fatalf("missing option: %q", view)
	}
}

func TestPickerNavigation(t *testing.T) {
	m := newTestModel()
	resp := make(chan promptResult, 1)
	m.picker = pickerState{
		title:   "Select artist",
		options: []app.PromptOption{{Label: "a"}, {Label: "b"}},
		cursor:  0,
		pending: &promptOutgoing{resp: resp},
	}
	m.screen = screenPicker

	next, _ := m.updatePicker(key("down"))
	m = next.(model)
	if m.picker.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.picker.cursor)
	}

	next, _ = m.updatePicker(key("enter"))
	m = next.(model)
	select {
	case r := <-resp:
		if r.idx != 1 || r.err != nil {
			t.Fatalf("result = %+v", r)
		}
	default:
		t.Fatal("expected response on channel")
	}
	if m.screen != screenRunning {
		t.Fatalf("screen = %v, want running", m.screen)
	}
}
