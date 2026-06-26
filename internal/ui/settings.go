package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/config"
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
	{"use_spotify", "Use Spotify metadata", settingToggle},
	{"use_genius", "Use Genius lyrics", settingToggle},
	{"verbose", "Verbose logging", settingToggle},
	{"theme", "Theme", settingTheme},
	{"spotify.client_id", "Spotify client ID", settingSecret},
	{"spotify.client_secret", "Spotify client secret", settingSecret},
	{"genius.access_token", "Genius access token", settingSecret},
}

func (m model) settingValue(key string) string {
	switch key {
	case "audio_format":
		return m.cfg.AudioFormat
	case "music_directory":
		return m.cfg.MusicDirectory
	case "use_spotify":
		return boolLabel(m.cfg.UseSpotify)
	case "use_genius":
		return boolLabel(m.cfg.UseGenius)
	case "verbose":
		return boolLabel(m.cfg.Verbose)
	case "theme":
		return m.cfg.Theme
	case "spotify.client_id":
		return maskSecret(m.cfg.Spotify.ClientID)
	case "spotify.client_secret":
		return maskSecret(m.cfg.Spotify.ClientSecret)
	case "genius.access_token":
		return maskSecret(m.cfg.Genius.AccessToken)
	}
	return ""
}

func (m model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case "down", "j":
		if m.settingsCursor < len(settingItems)-1 {
			m.settingsCursor++
		}
	case "enter":
		return m.activateSetting()
	}
	return m, nil
}

func (m model) activateSetting() (tea.Model, tea.Cmd) {
	item := settingItems[m.settingsCursor]
	switch item.kind {
	case settingToggle:
		m.toggleSetting(item.key)
		_ = config.Save(m.cfg)
		return m, nil
	case settingTheme:
		m.screen = screenThemePicker
		m.themeCursor = themeIndex(m.cfg.Theme)
		return m, nil
	case settingText, settingSecret:
		m.screen = screenSettingEdit
		m.editKey = item.key
		m.input.SetValue("")
		m.input.Placeholder = "type a new value, enter to save"
		m.input.Focus()
		return m, nil
	}
	return m, nil
}

func (m *model) toggleSetting(key string) {
	switch key {
	case "use_spotify":
		m.cfg.UseSpotify = !m.cfg.UseSpotify
	case "use_genius":
		m.cfg.UseGenius = !m.cfg.UseGenius
	case "verbose":
		m.cfg.Verbose = !m.cfg.Verbose
	}
}

func (m model) updateSettingEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenSettings
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.input.Value())
		if value != "" {
			m.applySetting(m.editKey, value)
			_ = config.Save(m.cfg)
			m.app = rebuildApp(m.cfg)
		}
		m.screen = screenSettings
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) applySetting(key, value string) {
	switch key {
	case "audio_format":
		m.cfg.AudioFormat = value
	case "music_directory":
		m.cfg.MusicDirectory = value
	case "spotify.client_id":
		m.cfg.Spotify.ClientID = value
	case "spotify.client_secret":
		m.cfg.Spotify.ClientSecret = value
	case "genius.access_token":
		m.cfg.Genius.AccessToken = value
	}
}

func (m model) viewSettings() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Settings") + "\n\n")

	for i, item := range settingItems {
		cursor := "  "
		label := m.st.item.Render(item.label)
		if i == m.settingsCursor {
			cursor = m.st.accent.Render("> ")
			label = m.st.selected.Render(item.label)
		}
		value := m.st.accent.Render(m.settingValue(item.key))
		fmt.Fprintf(&b, "  %s%-26s %s\n", cursor, label, value)
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: navigate   enter: edit/toggle   esc: back"))
	return b.String()
}

func (m model) viewSettingEdit() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Edit "+m.editKey) + "\n\n")
	b.WriteString("  " + m.input.View() + "\n\n")
	b.WriteString(m.st.help.Render("  enter: save   esc: cancel"))
	return b.String()
}

func boolLabel(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func maskSecret(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
}
