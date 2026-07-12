package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/EnJulian/shadowbox/internal/player"
)

// renderBannerWithPlayback renders the banner as-is when nothing is
// playing, or with a left-to-right color sweep (muted -> accent) tracking
// playback progress, plus a now-playing title/time line below it.
func renderBannerWithPlayback(st styles, theme Theme, state player.State) string {
	if state.Title == "" {
		return st.title.Render(banner)
	}

	lines := strings.Split(banner, "\n")
	width := 0
	for _, l := range lines {
		if n := len([]rune(l)); n > width {
			width = n
		}
	}

	var fraction float64
	if state.Duration > 0 {
		fraction = float64(state.Elapsed) / float64(state.Duration)
		if fraction > 1 {
			fraction = 1
		}
		if fraction < 0 {
			fraction = 0
		}
	}
	fillTo := int(fraction * float64(width))

	mutedStyle := lipgloss.NewStyle().Foreground(theme.Muted)
	accentStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)

	var b strings.Builder
	for _, line := range lines {
		runes := []rune(line)
		for i, r := range runes {
			if i < fillTo {
				b.WriteString(accentStyle.Render(string(r)))
			} else {
				b.WriteString(mutedStyle.Render(string(r)))
			}
		}
		b.WriteString("\n")
	}

	nowPlaying := fmt.Sprintf("  %s  %s / %s", state.Title, formatDuration(state.Elapsed), formatDuration(state.Duration))
	b.WriteString(st.subtitle.Render(nowPlaying))
	return b.String()
}

// formatDuration renders a duration as m:ss, matching the mockup the design
// was approved against.
func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	if total < 0 {
		total = 0
	}
	return fmt.Sprintf("%d:%02d", total/60, total%60)
}
