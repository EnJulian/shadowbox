package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the colour palette and derived styles for the interface.
type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Muted     lipgloss.Color
	Danger    lipgloss.Color
}

// themes are the available colour schemes, mirroring the original UI options.
var themes = map[string]Theme{
	"hacker": {
		Name: "hacker", Primary: "10", Secondary: "2", Accent: "14", Muted: "245", Danger: "9",
	},
	"matrix": {
		Name: "matrix", Primary: "46", Secondary: "34", Accent: "118", Muted: "238", Danger: "196",
	},
	"cyberpunk": {
		Name: "cyberpunk", Primary: "201", Secondary: "93", Accent: "51", Muted: "240", Danger: "197",
	},
	"ocean": {
		Name: "ocean", Primary: "39", Secondary: "33", Accent: "87", Muted: "244", Danger: "203",
	},
	"fire": {
		Name: "fire", Primary: "208", Secondary: "166", Accent: "226", Muted: "242", Danger: "160",
	},
}

// themeOrder is the stable ordering used when cycling themes.
var themeOrder = []string{"hacker", "matrix", "cyberpunk", "ocean", "fire"}

func themeByName(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["hacker"]
}

// styles bundles the lipgloss styles derived from a theme.
type styles struct {
	title    lipgloss.Style
	subtitle lipgloss.Style
	selected lipgloss.Style
	item     lipgloss.Style
	help     lipgloss.Style
	accent   lipgloss.Style
	danger   lipgloss.Style
	box      lipgloss.Style
}

func newStyles(t Theme) styles {
	return styles{
		title:    lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		subtitle: lipgloss.NewStyle().Foreground(t.Secondary),
		selected: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		item:     lipgloss.NewStyle().Foreground(t.Muted),
		help:     lipgloss.NewStyle().Foreground(t.Muted).Italic(true),
		accent:   lipgloss.NewStyle().Foreground(t.Accent),
		danger:   lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		box:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Primary).Padding(0, 1),
	}
}

// banner is the ASCII logo shown at the top of the interface.
const banner = `
 ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
 ██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝
 ███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝
 ╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗
 ███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗
 ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝`
