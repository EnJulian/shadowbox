// internal/ui/workspace/settings.go
package workspace

import (
	"fmt"
	"strings"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type settingKind int

const (
	settingToggle settingKind = iota
	settingText
	settingSecret
	settingTheme
)

type settingItem struct {
	key   string
	label string
	kind  settingKind
}

var settingItems = []settingItem{
	{"audio_format", "Audio format", settingText},
	{"music_directory", "Music directory", settingText},
	{"use_genius", "Use Genius lyrics", settingToggle},
	{"verbose", "Verbose logging", settingToggle},
	{"theme", "Theme", settingTheme},
	{"genius.access_token", "Genius access token", settingSecret},
}

// SettingsThemeRequestMsg asks root to open the theme picker overlay.
type SettingsThemeRequestMsg struct{}

// SettingsChangedMsg asks root to reload theme/app/workspaces from cfg after a save.
type SettingsChangedMsg struct{}

// Settings is the workspace.Workspace for the Settings nav section.
type Settings struct {
	cfg *config.Config
	st  style.Styles

	cursor  int
	editing bool
	editKey string
	input   textinput.Model
}

// NewSettings builds the Settings workspace.
func NewSettings(cfg *config.Config, st style.Styles) *Settings {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Width = 60
	return &Settings{cfg: cfg, st: st, input: ti}
}

func (s *Settings) Activate() Workspace {
	s.editing = false
	return s
}

func (s *Settings) value(key string) string {
	switch key {
	case "audio_format":
		return s.cfg.AudioFormat
	case "music_directory":
		return s.cfg.MusicDirectory
	case "use_genius":
		return boolLabel(s.cfg.UseGenius)
	case "verbose":
		return boolLabel(s.cfg.Verbose)
	case "theme":
		return s.cfg.Theme
	case "genius.access_token":
		return maskSecret(s.cfg.Genius.AccessToken)
	}
	return ""
}

func (s *Settings) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}
	if s.editing {
		return s.updateEdit(keyMsg)
	}

	switch keyMsg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(settingItems)-1 {
			s.cursor++
		}
	case "enter":
		return s.activate()
	}
	return s, nil
}

func (s *Settings) activate() (Workspace, tea.Cmd) {
	item := settingItems[s.cursor]
	switch item.kind {
	case settingToggle:
		s.toggle(item.key)
		_ = config.Save(s.cfg)
		return s, func() tea.Msg { return SettingsChangedMsg{} }
	case settingTheme:
		return s, func() tea.Msg { return SettingsThemeRequestMsg{} }
	case settingText, settingSecret:
		s.editing = true
		s.editKey = item.key
		s.input.SetValue("")
		s.input.Placeholder = "type a new value, enter to save"
		s.input.Focus()
	}
	return s, nil
}

func (s *Settings) toggle(key string) {
	switch key {
	case "use_genius":
		s.cfg.UseGenius = !s.cfg.UseGenius
	case "verbose":
		s.cfg.Verbose = !s.cfg.Verbose
	}
}

func (s *Settings) updateEdit(msg tea.KeyMsg) (Workspace, tea.Cmd) {
	switch msg.String() {
	case "esc":
		s.editing = false
		return s, nil
	case "enter":
		value := strings.TrimSpace(s.input.Value())
		s.editing = false
		if value == "" {
			return s, nil
		}
		s.apply(s.editKey, value)
		_ = config.Save(s.cfg)
		return s, func() tea.Msg { return SettingsChangedMsg{} }
	}
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *Settings) apply(key, value string) {
	switch key {
	case "audio_format":
		s.cfg.AudioFormat = value
	case "music_directory":
		s.cfg.MusicDirectory = value
	case "genius.access_token":
		s.cfg.Genius.AccessToken = value
	}
}

func (s *Settings) View(width, height int) string {
	if s.editing {
		var b strings.Builder
		b.WriteString(s.st.Subtitle.Render("Edit "+s.editKey) + "\n\n")
		b.WriteString(s.input.View() + "\n\n")
		b.WriteString(s.st.Help.Render("enter: save   esc: cancel"))
		return b.String()
	}

	var b strings.Builder
	for i, item := range settingItems {
		cursor := "  "
		paddedLabel := fmt.Sprintf("%-26s", item.label)
		label := s.st.Item.Render(paddedLabel)
		if i == s.cursor {
			cursor = s.st.Accent.Render("> ")
			label = s.st.Selected.Render(paddedLabel)
		}
		value := s.st.Accent.Render(s.value(item.key))
		fmt.Fprintf(&b, "%s%s %s\n", cursor, label, value)
	}
	return b.String()
}

func boolLabel(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func maskSecret(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 4 {
		return strings.Repeat("*", len(v))
	}
	return strings.Repeat("*", len(v)-4) + v[len(v)-4:]
}
