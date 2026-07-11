package shell

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

const (
	// NavWidth is the fixed outer width (including border) of the Nav pane.
	NavWidth = 20
	// MinWidth and MinHeight are the terminal dimensions the shell is
	// designed for; smaller terminals still render but content is clamped.
	MinWidth  = 100
	MinHeight = 24

	minContentWidth  = 20
	minContentHeight = 5
	// reservedRows accounts for the banner (2 lines), a blank line, and the
	// unboxed status bar (1 line) that always surround the pane row.
	reservedRows = 4
)

// Layout describes the computed pixel dimensions of each shell region for a
// given terminal size.
type Layout struct {
	NavWidth      int
	ContentWidth  int
	ContentHeight int
	// PlaybarHeight is 0 while playback is inactive (always, this phase).
	// The field exists so a later phase can grow it without touching the
	// rest of the layout math.
	PlaybarHeight int
}

// Compute derives a Layout from the terminal size reported by tea.WindowSizeMsg.
func Compute(width, height int) Layout {
	contentWidth := width - NavWidth
	if contentWidth < minContentWidth {
		contentWidth = minContentWidth
	}
	contentHeight := height - reservedRows
	if contentHeight < minContentHeight {
		contentHeight = minContentHeight
	}
	return Layout{
		NavWidth:      NavWidth,
		ContentWidth:  contentWidth,
		ContentHeight: contentHeight,
		PlaybarHeight: 0,
	}
}

func borderColor(theme style.Theme, focused bool) lipgloss.Color {
	if focused {
		return theme.Accent
	}
	return theme.Muted
}

// Render composes the full frame: banner, Nav | Content row (each in a
// rounded-border box, the focused pane's border in the theme's accent
// colour), the reserved playback bar row (only when playbarBody is
// non-empty and layout.PlaybarHeight > 0), and the unboxed status line.
func Render(st style.Styles, theme style.Theme, layout Layout, navFocused bool, navBody, contentBody, playbarBody, statusBody string) string {
	navStyle := st.Box.Width(layout.NavWidth - 2).Height(layout.ContentHeight - 2).
		BorderForeground(borderColor(theme, navFocused))
	contentStyle := st.Box.Width(layout.ContentWidth - 2).Height(layout.ContentHeight - 2).
		BorderForeground(borderColor(theme, !navFocused))

	row := lipgloss.JoinHorizontal(lipgloss.Top, navStyle.Render(navBody), contentStyle.Render(contentBody))

	parts := []string{style.RenderBanner(st), row}
	if layout.PlaybarHeight > 0 && playbarBody != "" {
		parts = append(parts, playbarBody)
	}
	parts = append(parts, statusBody)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
