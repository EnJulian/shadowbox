package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/config"
)

func themeIndex(name string) int {
	for i, n := range themeOrder {
		if n == name {
			return i
		}
	}
	return 0
}

func (m model) updateThemePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.screen = screenSettings
		return m, nil
	case "up", "k":
		if m.themeCursor > 0 {
			m.themeCursor--
		}
		m.previewTheme()
	case "down", "j":
		if m.themeCursor < len(themeOrder)-1 {
			m.themeCursor++
		}
		m.previewTheme()
	case "enter":
		m.cfg.Theme = themeOrder[m.themeCursor]
		m.applyTheme(m.cfg.Theme)
		_ = config.Save(m.cfg)
		m.screen = screenSettings
		return m, nil
	}
	return m, nil
}

// previewTheme applies the highlighted theme live for instant feedback.
func (m *model) previewTheme() {
	m.applyTheme(themeOrder[m.themeCursor])
}

func (m *model) applyTheme(name string) {
	m.theme = themeByName(name)
	m.st = newStyles(m.theme)
}

func (m model) viewThemePicker() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Choose a theme") + "\n\n")

	for i, name := range themeOrder {
		cursor := "  "
		label := m.st.item.Render(name)
		if i == m.themeCursor {
			cursor = m.st.accent.Render("> ")
			label = m.st.selected.Render(name)
		}
		b.WriteString("  " + cursor + label + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: preview   enter: apply   esc: cancel"))
	return b.String()
}
