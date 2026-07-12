// internal/ui/overlay/theme.go
package overlay

import (
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

// Theme is the theme picker overlay, applied live for instant preview as
// the cursor moves.
type Theme struct {
	Cursor int
}

func (t *Theme) MoveUp() {
	if t.Cursor > 0 {
		t.Cursor--
	}
}

func (t *Theme) MoveDown() {
	if t.Cursor < len(style.ThemeOrder)-1 {
		t.Cursor++
	}
}

// Selected returns the name of the currently highlighted theme.
func (t Theme) Selected() string {
	return style.ThemeOrder[t.Cursor]
}

func (t Theme) View(st style.Styles) string {
	out := st.Subtitle.Render("Choose a theme") + "\n\n"
	for i, name := range style.ThemeOrder {
		cursor := "  "
		label := st.Item.Render(name)
		if i == t.Cursor {
			cursor = st.Accent.Render("> ")
			label = st.Selected.Render(name)
		}
		out += cursor + label + "\n"
	}
	return out + "\n" + st.Help.Render("up/down: preview   enter: apply   esc: cancel")
}
