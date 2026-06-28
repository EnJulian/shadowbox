package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/EnJulian/shadowbox/internal/progress"
)

const progressBarWidth = 24

// renderProgressBar returns a minimal text progress bar. When total is zero the
// bar shows a light indeterminate fill without a numeric label.
func renderProgressBar(u progress.Update, accent, muted lipgloss.Color) string {
	filled := progressBarWidth / 4
	if u.Total > 0 && u.Current > 0 {
		current := u.Current
		if current > u.Total {
			current = u.Total
		}
		filled = (current * progressBarWidth) / u.Total
	}

	filledStr := lipgloss.NewStyle().Foreground(accent).Render(strings.Repeat("─", filled))
	emptyStr := lipgloss.NewStyle().Foreground(muted).Render(strings.Repeat("·", progressBarWidth-filled))
	bar := "[" + filledStr + emptyStr + "]"

	if u.Total > 0 && u.Current > 0 {
		label := lipgloss.NewStyle().Foreground(muted).Render(
			fmt.Sprintf(" %d/%d", u.Current, u.Total),
		)
		return bar + label
	}
	return bar
}
