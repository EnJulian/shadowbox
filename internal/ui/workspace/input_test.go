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
	_, cmd := ws.Update(key("enter"))
	if cmd != nil {
		t.Fatal("expected no cmd when submitting an empty field")
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
