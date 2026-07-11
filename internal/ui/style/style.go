// Package style holds Shadowbox's terminal color themes, derived lipgloss
// styles, and the ASCII banner — the one leaf package every other ui
// sub-package (shell, workspace, overlay) depends on for rendering.
package style

import "github.com/charmbracelet/lipgloss"

// Theme defines the colour palette for the interface.
type Theme struct {
	Name      string
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Muted     lipgloss.Color
	Danger    lipgloss.Color
}

// themes is the curated v2 palette, retuned so border colours and active-nav
// highlights stay legible against the new rounded-border, box-heavy layout.
var themes = map[string]Theme{
	"hacker": {
		Name: "hacker", Primary: "10", Secondary: "2", Accent: "14", Muted: "245", Danger: "9",
	},
	"cyberpunk": {
		Name: "cyberpunk", Primary: "201", Secondary: "135", Accent: "51", Muted: "240", Danger: "197",
	},
	"ocean": {
		Name: "ocean", Primary: "39", Secondary: "33", Accent: "87", Muted: "244", Danger: "203",
	},
	"ember": {
		Name: "ember", Primary: "208", Secondary: "166", Accent: "220", Muted: "242", Danger: "160",
	},
	"matrix": {
		Name: "matrix", Primary: "46", Secondary: "34", Accent: "118", Muted: "238", Danger: "196",
	},
}

// ThemeOrder is the stable ordering used when cycling themes in the picker.
var ThemeOrder = []string{"hacker", "cyberpunk", "ocean", "ember", "matrix"}

// ThemeByName returns the named theme, or the "hacker" default if unknown.
func ThemeByName(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["hacker"]
}

// Styles bundles the lipgloss styles derived from a Theme.
type Styles struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Selected lipgloss.Style
	Item     lipgloss.Style
	Help     lipgloss.Style
	Accent   lipgloss.Style
	Danger   lipgloss.Style
	Box      lipgloss.Style
}

// NewStyles derives the full style set from a Theme.
func NewStyles(t Theme) Styles {
	return Styles{
		Title:    lipgloss.NewStyle().Foreground(t.Primary).Bold(true),
		Subtitle: lipgloss.NewStyle().Foreground(t.Secondary),
		Selected: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		Item:     lipgloss.NewStyle().Foreground(t.Muted),
		Help:     lipgloss.NewStyle().Foreground(t.Muted).Italic(true),
		Accent:   lipgloss.NewStyle().Foreground(t.Accent),
		Danger:   lipgloss.NewStyle().Foreground(t.Danger).Bold(true),
		Box:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
	}
}

// Banner is the ASCII logo shown at the top of the shell on every screen.
const Banner = `
 ███████╗██╗  ██╗ █████╗ ██████╗  ██████╗ ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
 ██╔════╝██║  ██║██╔══██╗██╔══██╗██╔═══██╗██║    ██║██╔══██╗██╔═══██╗╚██╗██╔╝
 ███████╗███████║███████║██║  ██║██║   ██║██║ █╗ ██║██████╔╝██║   ██║ ╚███╔╝
 ╚════██║██╔══██║██╔══██║██║  ██║██║   ██║██║███╗██║██╔══██╗██║   ██║ ██╔██╗
 ███████║██║  ██║██║  ██║██████╔╝╚██████╔╝╚███╔███╔╝██████╔╝╚██████╔╝██╔╝ ██╗
 ╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚══╝╚══╝ ╚═════╝  ╚═════╝ ╚═╝  ╚═╝`

// RenderBanner renders the centered banner plus subtitle line, styled.
func RenderBanner(st Styles) string {
	return st.Title.Render(Banner) + "\n" + st.Subtitle.Render("                    Music Acquisition Console")
}
