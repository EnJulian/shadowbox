// internal/ui/overlay/picker.go

// Package overlay renders modal content shown on top of the shell: the
// metadata/URL disambiguation picker, the keybinding help screen, the
// task-result toast, and the theme picker.
package overlay

import (
	"strings"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/ui/style"
)

// Picker is the disambiguation overlay shown when a pipeline step needs the
// user to choose among several candidates (metadata match, URL variant, …).
type Picker struct {
	Title   string
	Options []app.PromptOption
	Cursor  int
}

func (p *Picker) MoveUp() {
	if p.Cursor > 0 {
		p.Cursor--
	}
}

func (p *Picker) MoveDown() {
	if p.Cursor < len(p.Options)-1 {
		p.Cursor++
	}
}

func (p Picker) View(st style.Styles) string {
	var b strings.Builder
	b.WriteString(st.Subtitle.Render(p.Title) + "\n\n")
	if len(p.Options) == 0 {
		b.WriteString(st.Item.Render("(no options)") + "\n")
	}
	for i, o := range p.Options {
		cursor := "  "
		label := st.Item.Render(o.Label)
		if i == p.Cursor {
			cursor = st.Accent.Render("> ")
			label = st.Selected.Render(o.Label)
		}
		b.WriteString(cursor + label + "\n")
		if o.Detail != "" {
			b.WriteString("    " + st.Help.Render(o.Detail) + "\n")
		}
	}
	b.WriteString("\n" + st.Help.Render("up/down: navigate   enter: select   esc: cancel"))
	return b.String()
}
