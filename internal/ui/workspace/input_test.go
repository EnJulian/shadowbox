// internal/ui/workspace/input_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestInputActivateFocusesField(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()
	if !ws.(*Input).input.Focused() {
		t.Fatal("Activate() did not focus the input")
	}
}

func TestInputSubmitStartsTaskAndSwitchesToDownloads(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()

	for _, r := range "https://youtu.be/abc" {
		ws, _ = ws.Update(key(string(r)))
	}
	_, cmd := ws.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a cmd after submitting a URL")
	}
	msgs := flattenBatch(cmd())
	var sawStart, sawSwitch bool
	for _, m := range msgs {
		if _, ok := m.(StartTaskMsg); ok {
			sawStart = true
		}
		if ss, ok := m.(SwitchSectionMsg); ok && ss.Section == SectionDownloads {
			sawSwitch = true
		}
	}
	if !sawStart || !sawSwitch {
		t.Fatalf("expected StartTaskMsg and SwitchSectionMsg{Downloads}, got %#v", msgs)
	}
}

func TestInputEmptySubmitDoesNothing(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()
	// No clipboard hint present (Activate() only sets clipHint from a real
	// clipboard read, which is not URL-shaped in the test environment), so
	// this exercises the plain "empty field, no hint" no-op path.
	if ws.(*Input).clipHint != "" {
		t.Fatalf("test assumes no clipboard hint is present, got %q", ws.(*Input).clipHint)
	}
	_, cmd := ws.Update(key("enter"))
	if cmd != nil {
		t.Fatal("expected no cmd when submitting an empty field")
	}
}

func TestInputEnterFillsFieldFromClipboardHintWhenEmpty(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()
	in := ws.(*Input)
	in.clipHint = "https://youtu.be/abc123"

	ws, cmd := in.Update(key("enter"))
	if cmd != nil {
		t.Fatal("expected no cmd (no submit) when accepting the clipboard hint")
	}
	if got := ws.(*Input).input.Value(); got != "https://youtu.be/abc123" {
		t.Fatalf("input value = %q, want the clipboard hint to be filled in", got)
	}
}

func TestInputEnterSubmitsNormallyWhenFieldNonEmptyEvenWithHint(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewURL(a, st).Activate()
	in := ws.(*Input)
	// A hint being present must not interfere with submitting a value the
	// user has already typed.
	in.clipHint = "https://youtu.be/other"

	for _, r := range "https://youtu.be/abc" {
		ws, _ = ws.Update(key(string(r)))
	}
	_, cmd := ws.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a cmd after submitting a non-empty URL, even with a hint present")
	}
	msgs := flattenBatch(cmd())
	var sawStart, sawSwitch bool
	for _, m := range msgs {
		if _, ok := m.(StartTaskMsg); ok {
			sawStart = true
		}
		if ss, ok := m.(SwitchSectionMsg); ok && ss.Section == SectionDownloads {
			sawSwitch = true
		}
	}
	if !sawStart || !sawSwitch {
		t.Fatalf("expected StartTaskMsg and SwitchSectionMsg{Downloads}, got %#v", msgs)
	}
}

func TestEnhanceRunsEnhanceDir(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewEnhance(a, st).Activate()
	for _, r := range "." {
		ws, _ = ws.Update(key(string(r)))
	}
	_, cmd := ws.Update(key("enter"))
	msgs := flattenBatch(cmd())
	for _, m := range msgs {
		if st, ok := m.(StartTaskMsg); ok {
			if !strings.Contains(st.Summary, "Enhance") {
				t.Fatalf("Summary = %q, want it to mention Enhance", st.Summary)
			}
			return
		}
	}
	t.Fatal("expected a StartTaskMsg")
}

func TestInputPlaylistSubmitUsesPlaylistSummary(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	a := app.New(&config.Config{})
	ws := NewPlaylist(a, st).Activate()
	for _, r := range "https://youtube.com/playlist?list=abc" {
		ws, _ = ws.Update(key(string(r)))
	}
	_, cmd := ws.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a cmd after submitting a playlist URL")
	}
	msgs := flattenBatch(cmd())
	for _, m := range msgs {
		if st, ok := m.(StartTaskMsg); ok {
			if !strings.Contains(st.Summary, "Playlist") {
				t.Fatalf("Summary = %q, want it to mention Playlist (proves NewPlaylist wired to a.RunPlaylist, not a.Run)", st.Summary)
			}
			return
		}
	}
	t.Fatal("expected a StartTaskMsg")
}
