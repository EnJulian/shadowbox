package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuCursor > 0 {
			m.menuCursor--
		}
	case "down", "j":
		if m.menuCursor < len(mainMenu)-1 {
			m.menuCursor++
		}
	case "enter":
		return m.selectMenu()
	}
	return m, nil
}

func (m model) selectMenu() (tea.Model, tea.Cmd) {
	switch m.menuCursor {
	case 0:
		return m.openInput("search", "Enter song title and artist (e.g. Imagine Dragons Believer)")
	case 1:
		return m.openInput("url", "Enter a YouTube or Bandcamp URL")
	case 2:
		return m.openInput("playlist", "Enter a YouTube playlist URL")
	case 3:
		return m.openInput("enhance", "Enter a directory of audio files to enhance")
	case 4:
		return m.openLibrary()
	case 5:
		m.screen = screenSettings
		m.settingsCursor = 0
		return m, nil
	case 6:
		return m, tea.Quit
	}
	return m, nil
}

func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString(m.st.subtitle.Render("  Music acquisition console"))
	b.WriteString("\n\n")

	for i, item := range mainMenu {
		cursor := "  "
		line := m.st.item.Render(item)
		if i == m.menuCursor {
			cursor = m.st.accent.Render("> ")
			line = m.st.selected.Render(item)
		}
		b.WriteString("  " + cursor + line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: navigate   enter: select   q: quit"))
	return b.String()
}
