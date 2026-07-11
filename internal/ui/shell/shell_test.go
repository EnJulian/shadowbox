package shell

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestComputeSplitsWidthBetweenNavAndContent(t *testing.T) {
	l := Compute(120, 40)
	if l.NavWidth != NavWidth {
		t.Fatalf("NavWidth = %d, want %d", l.NavWidth, NavWidth)
	}
	if l.ContentWidth != 120-NavWidth {
		t.Fatalf("ContentWidth = %d, want %d", l.ContentWidth, 120-NavWidth)
	}
	if l.PlaybarHeight != 0 {
		t.Fatalf("PlaybarHeight = %d, want 0 (reserved but inactive this phase)", l.PlaybarHeight)
	}
}

func TestComputeClampsContentWidthOnNarrowTerminals(t *testing.T) {
	l := Compute(50, 40)
	if l.ContentWidth < 20 {
		t.Fatalf("ContentWidth = %d, want at least 20 even on a narrow terminal", l.ContentWidth)
	}
}

func TestComputeClampsContentHeightOnShortTerminals(t *testing.T) {
	l := Compute(120, 10)
	if l.ContentHeight < 5 {
		t.Fatalf("ContentHeight = %d, want at least 5 even on a short terminal", l.ContentHeight)
	}
}

func TestRenderIncludesBannerNavAndStatus(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	out := Render(st, style.ThemeByName("hacker"), Compute(120, 40), true, "NAV BODY", "CONTENT BODY", "", "STATUS LINE")
	for _, want := range []string{"NAV BODY", "CONTENT BODY", "STATUS LINE"} {
		if !strings.Contains(out, want) {
			t.Errorf("Render output missing %q", want)
		}
	}
}

func TestRenderDoesNotExceedTerminalHeight(t *testing.T) {
	// This test verifies the invariant: for a given height passed to Compute,
	// the total number of lines in Render(...)'s output must not exceed that height.
	// This was broken when reservedRows was undercounting the actual banner height.
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"minimum terminal size", MinWidth, MinHeight},
		{"larger terminal", 120, 40},
	}

	st := style.NewStyles(style.ThemeByName("hacker"))
	theme := style.ThemeByName("hacker")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := Compute(tt.width, tt.height)
			out := Render(st, theme, layout, true, "NAV BODY", "CONTENT BODY", "", "STATUS LINE")
			lineCount := len(strings.Split(out, "\n"))

			if lineCount > tt.height {
				t.Errorf("Render output has %d lines but terminal height is only %d (overflow: %d lines)",
					lineCount, tt.height, lineCount-tt.height)
			}
		})
	}
}
