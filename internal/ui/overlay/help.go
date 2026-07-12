// internal/ui/overlay/help.go
package overlay

import "github.com/EnJulian/shadowbox/internal/ui/style"

// Help is the static keybinding reference overlay, toggled by "?".
type Help struct{}

var helpLines = []string{
	"1-8       jump to a nav section",
	"Tab       cycle focus: Nav <-> Content",
	"<- -> h l switch panes (Library: columns first)",
	"up down   move within the focused pane",
	"/         focus the Search query input",
	"enter     activate / submit / select",
	"esc       back / cancel / close",
	"?         toggle this help",
	"q         quit (confirms if a download is running)",
	"ctrl+c    force quit",
}

func (Help) View(st style.Styles) string {
	out := st.Subtitle.Render("Keybindings") + "\n\n"
	for _, line := range helpLines {
		out += st.Item.Render(line) + "\n"
	}
	return out + "\n" + st.Help.Render("? or esc: close")
}
