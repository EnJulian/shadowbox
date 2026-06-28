package ui

import (
	"strings"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	applog "github.com/EnJulian/shadowbox/internal/log"
)

func (m model) openDownloadLog() (tea.Model, tea.Cmd) {
	_ = applog.LoadDownloadLog()
	m.screen = screenDownloadLog
	m.logLines = applog.DownloadLogLines()
	m.logScroll = 0
	if len(m.logLines) == 0 {
		m.logScroll = 0
	} else {
		m.logScroll = maxLogScroll(m.logLines, m.logViewport())
	}
	return m, nil
}

func (m model) updateDownloadLog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	viewport := m.logViewport()
	maxScroll := maxLogScroll(m.logLines, viewport)

	switch msg.String() {
	case "q", "esc":
		m.screen = screenMenu
		return m, nil
	case "up", "k":
		if m.logScroll > 0 {
			m.logScroll--
		}
	case "down", "j":
		if m.logScroll < maxScroll {
			m.logScroll++
		}
	case "pgup":
		m.logScroll = max(0, m.logScroll-viewport)
	case "pgdown":
		m.logScroll = min(maxScroll, m.logScroll+viewport)
	case "home":
		m.logScroll = 0
	case "end", "G":
		m.logScroll = maxScroll
	case "r":
		_ = applog.LoadDownloadLog()
		m.logLines = applog.DownloadLogLines()
		maxScroll = maxLogScroll(m.logLines, viewport)
		m.logScroll = maxScroll
	}
	return m, nil
}

func (m model) logViewport() int {
	h := m.height - 12
	if h < 5 {
		return 5
	}
	return h
}

func maxLogScroll(lines []string, viewport int) int {
	if len(lines) <= viewport {
		return 0
	}
	return len(lines) - viewport
}

func visibleLogLines(lines []string, viewport, scroll int) []string {
	if len(lines) == 0 {
		return nil
	}
	maxScroll := maxLogScroll(lines, viewport)
	if scroll > maxScroll {
		scroll = maxScroll
	}
	start := scroll
	end := scroll + viewport
	if end > len(lines) {
		end = len(lines)
	}
	return lines[start:end]
}

func (m model) viewDownloadLog() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Download log") + "\n\n")

	viewport := m.logViewport()
	visible := visibleLogLines(m.logLines, viewport, m.logScroll)

	if len(m.logLines) == 0 {
		b.WriteString("  " + m.st.item.Render("(no download logs yet — run a download first)") + "\n")
	} else {
		for _, line := range visible {
			b.WriteString("  " + m.st.item.Render(line) + "\n")
		}
		if len(m.logLines) > viewport {
			b.WriteString("\n")
			b.WriteString(m.st.help.Render(
				"  lines " + strconv.Itoa(m.logScroll+1) + "-" + strconv.Itoa(min(len(m.logLines), m.logScroll+viewport)) +
					" of " + strconv.Itoa(len(m.logLines)),
			))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(m.st.help.Render("  up/down: scroll   home/end: top/bottom   r: refresh   esc: menu"))
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
