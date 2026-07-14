package ui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// helpContent is the full keyboard-shortcut and important-things-to-remember
// reference, as lines — fixed facts about the app with no dependency on live
// state (it never shows whether mpv/aria2 are currently installed; that's
// what the Setup Wizard itself is for), so this is computed once at package
// load and never recomputed.
var helpContent = []string{
	"Global",
	"  Ctrl+C          force quit, always works",
	"  ?               open/close this help screen",
	"  space/n/p/s     playback: pause, next, previous, stop",
	"  ← / →           playback: seek back/forward 10s",
	"  (playback keys work everywhere except Library, text inputs, and Settings edit mode,",
	"   since those screens need those keys for typing instead)",
	"",
	"Menu",
	"  ↑/↓ or k/j      navigate",
	"  enter           select",
	"  q               quit",
	"",
	"Library",
	"  ↑/↓/←/→          navigate / open",
	"  (any letter)    type to filter the current list",
	"  backspace       remove the last filter character",
	"",
	"Settings",
	"  ↑/↓ or k/j      navigate",
	"  enter           edit / toggle / open",
	"  esc or q        back to menu",
	"",
	"Settings edit (text/secret fields)",
	"  enter           save",
	"  esc             cancel",
	"",
	"Theme picker",
	"  ↑/↓ or k/j      live preview",
	"  enter           apply (saves)",
	"  esc or q        cancel (reverts the preview)",
	"",
	"Download Log",
	"  ↑/↓ or k/j      scroll",
	"  pgup/pgdown     scroll a page",
	"  home / end      jump to top / bottom",
	"  r               reload",
	"  esc or q        back",
	"",
	"Search result picker",
	"  ↑/↓ or k/j      navigate",
	"  enter           select",
	"  esc             cancel",
	"",
	"URL / Playlist / Search input",
	"  enter           submit",
	"  esc             cancel",
	"",
	"Enhance Existing Files (folder picker)",
	"  ↑/↓/←/→          navigate / open",
	"  enter           open folder, or use it if \"[ Use this folder ]\" is selected",
	"  (any letter)    type to filter the current list",
	"  backspace       remove the last filter character",
	"  esc             go up a folder (or back to the menu at the top)",
	"",
	"Setup Wizard",
	"  ↑/↓ or k/j      navigate",
	"  enter           install (or show instructions)",
	"  esc             done (saves and returns)",
	"",
	"Things to remember",
	"  Config lives at ~/.config/shadowbox/config.yaml — edit it through Settings",
	"  (auto-saved) or directly.",
	"  yt-dlp/ffmpeg are required; aria2 (faster downloads) and mpv (in-app",
	"  playback) are optional — the Setup Wizard (first run, or Settings ->",
	"  \"Run setup wizard\") can install them, or shows the exact command.",
	"  Supported audio formats: opus, mp3, flac, m4a, wav, ogg, webm, aac.",
	"  Library expects Artist/Album/Track folders, but a track directly under",
	"  an Artist folder (no album) works too — selecting it plays immediately.",
}

// openHelp switches to the help screen, remembering returnTo so closing it
// (via '?', esc, or q) knows where to go back to.
func (m model) openHelp(returnTo screen) (tea.Model, tea.Cmd) {
	m.screen = screenHelp
	m.helpReturnTo = returnTo
	m.helpScroll = 0
	return m, nil
}

func (m model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	viewport := helpViewport(m.height)
	maxScroll := maxHelpScroll(viewport)

	switch msg.String() {
	case "?", "esc", "q":
		m.screen = m.helpReturnTo
		return m, nil
	case "up", "k":
		if m.helpScroll > 0 {
			m.helpScroll--
		}
	case "down", "j":
		if m.helpScroll < maxScroll {
			m.helpScroll++
		}
	case "pgup":
		m.helpScroll = max(0, m.helpScroll-viewport)
	case "pgdown":
		m.helpScroll = min(maxScroll, m.helpScroll+viewport)
	}
	return m, nil
}

// helpViewport returns how many content lines fit given the terminal
// height, matching internal/ui/log_view.go's logViewport reserved-rows math.
func helpViewport(height int) int {
	h := height - 12
	if h < 5 {
		return 5
	}
	return h
}

func maxHelpScroll(viewport int) int {
	if len(helpContent) <= viewport {
		return 0
	}
	return len(helpContent) - viewport
}

func visibleHelpLines(viewport, scroll int) []string {
	maxScroll := maxHelpScroll(viewport)
	if scroll > maxScroll {
		scroll = maxScroll
	}
	start := scroll
	end := scroll + viewport
	if end > len(helpContent) {
		end = len(helpContent)
	}
	return helpContent[start:end]
}

func (m model) viewHelp() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Help") + "\n\n")

	viewport := helpViewport(m.height)
	visible := visibleHelpLines(viewport, m.helpScroll)
	for _, line := range visible {
		b.WriteString("  " + m.st.item.Render(line) + "\n")
	}
	if len(helpContent) > viewport {
		end := m.helpScroll + viewport
		if end > len(helpContent) {
			end = len(helpContent)
		}
		b.WriteString("\n")
		b.WriteString(m.st.help.Render(
			"  lines " + strconv.Itoa(m.helpScroll+1) + "-" + strconv.Itoa(end) + " of " + strconv.Itoa(len(helpContent)),
		))
	}

	b.WriteString("\n\n")
	b.WriteString(m.st.help.Render("  up/down: scroll   ?/esc/q: close"))
	return b.String()
}
