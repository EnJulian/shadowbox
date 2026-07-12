// internal/ui/workspace/settings_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestSettingsViewListsItems(t *testing.T) {
	cfg := &config.Config{AudioFormat: "opus", Theme: "hacker"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()
	view := s.View(80, 20)
	if !strings.Contains(view, "Audio format") {
		t.Fatalf("expected settings list, got %q", view)
	}
}

func TestSettingsTextFocused(t *testing.T) {
	cfg := &config.Config{AudioFormat: "opus", Theme: "hacker"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()

	if s.(*Settings).TextFocused() {
		t.Fatal("TextFocused() = true, want false during normal list navigation")
	}

	// audio_format (cursor 0) is a settingText item; enter opens inline edit.
	ws, _ := s.Update(key("enter"))
	if !ws.(*Settings).editing {
		t.Fatal("expected enter on a settingText item to enter edit mode")
	}
	if !ws.(*Settings).TextFocused() {
		t.Fatal("TextFocused() = false, want true while inline-editing a text setting")
	}

	ws, _ = ws.Update(key("esc"))
	if ws.(*Settings).TextFocused() {
		t.Fatal("TextFocused() = true, want false after esc cancels editing")
	}
}

func TestSettingsToggleFlipsBoolAndEmitsChanged(t *testing.T) {
	// The toggle path calls config.Save, which resolves its target file via
	// os.UserConfigDir(). Redirect that to a temp dir so the test never
	// touches the real machine's config file.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())

	cfg := &config.Config{UseGenius: true}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()

	// use_genius is index 2 in the fixed settingItems ordering.
	for i := 0; i < 2; i++ {
		s, _ = s.Update(key("down"))
	}
	_, cmd := s.Update(key("enter"))
	if cfg.UseGenius {
		t.Fatal("expected UseGenius to flip to false")
	}
	if cmd == nil {
		t.Fatal("expected a SettingsChangedMsg cmd after toggling")
	}
	if _, ok := cmd().(SettingsChangedMsg); !ok {
		t.Fatal("expected SettingsChangedMsg")
	}
}

func TestSettingsThemeItemRequestsThemePicker(t *testing.T) {
	cfg := &config.Config{Theme: "hacker"}
	st := style.NewStyles(style.ThemeByName("hacker"))
	s := NewSettings(cfg, st).Activate()

	// theme is index 4.
	for i := 0; i < 4; i++ {
		s, _ = s.Update(key("down"))
	}
	_, cmd := s.Update(key("enter"))
	if cmd == nil {
		t.Fatal("expected a SettingsThemeRequestMsg cmd")
	}
	if _, ok := cmd().(SettingsThemeRequestMsg); !ok {
		t.Fatal("expected SettingsThemeRequestMsg")
	}
}
