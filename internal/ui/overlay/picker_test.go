// internal/ui/overlay/picker_test.go
package overlay

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestPickerViewRendersOptions(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	p := Picker{
		Title: "Select track",
		Options: []app.PromptOption{
			{Label: "Song A — Artist", Detail: "Album (2020) · 3:30"},
			{Label: "Song B — Artist", Detail: "Single (2021)"},
		},
	}
	view := p.View(st)
	if !strings.Contains(view, "Select track") || !strings.Contains(view, "Song A — Artist") {
		t.Fatalf("missing expected content: %q", view)
	}
}

func TestPickerMoveClampsAtBounds(t *testing.T) {
	p := Picker{Options: []app.PromptOption{{Label: "a"}, {Label: "b"}}}
	p.MoveUp()
	if p.Cursor != 0 {
		t.Fatalf("Cursor = %d, want 0 (clamped)", p.Cursor)
	}
	p.MoveDown()
	if p.Cursor != 1 {
		t.Fatalf("Cursor = %d, want 1", p.Cursor)
	}
	p.MoveDown()
	if p.Cursor != 1 {
		t.Fatalf("Cursor = %d, want 1 (clamped)", p.Cursor)
	}
}

func TestThemeMoveAndSelected(t *testing.T) {
	th := Theme{}
	if th.Selected() != style.ThemeOrder[0] {
		t.Fatalf("Selected() = %q, want %q", th.Selected(), style.ThemeOrder[0])
	}
	th.MoveDown()
	if th.Selected() != style.ThemeOrder[1] {
		t.Fatalf("Selected() = %q, want %q", th.Selected(), style.ThemeOrder[1])
	}
}
