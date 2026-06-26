package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
)

// openInput switches to the text-input screen for the given context.
func (m model) openInput(ctxName, title string) (tea.Model, tea.Cmd) {
	m.screen = screenInput
	m.inputContext = ctxName
	m.inputTitle = title
	m.input.SetValue("")
	m.input.Placeholder = ""
	m.input.Focus()
	return m, nil
}

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenMenu
		return m, nil
	case "enter":
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			m.screen = screenMenu
			return m, nil
		}
		return m.submitInput(value)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) submitInput(value string) (tea.Model, tea.Cmd) {
	a := m.app
	opts := app.Options{UseSpotify: m.cfg.UseSpotify}

	switch m.inputContext {
	case "search":
		cmd := m.startTask("Download", func(ctx context.Context) error {
			return a.Run(ctx, value, opts)
		})
		return m, cmd
	case "url":
		cmd := m.startTask("Download", func(ctx context.Context) error {
			return a.Run(ctx, value, opts)
		})
		return m, cmd
	case "playlist":
		cmd := m.startTask("Playlist download", func(ctx context.Context) error {
			return a.RunPlaylist(ctx, value, opts)
		})
		return m, cmd
	case "enhance":
		cmd := m.startTask("Enhance", func(ctx context.Context) error {
			return a.EnhanceDir(ctx, value, true, []string{"opus", "mp3", "m4a", "flac"}, false)
		})
		return m, cmd
	}
	m.screen = screenMenu
	return m, nil
}

func (m model) viewInput() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render(m.inputTitle) + "\n\n")
	b.WriteString("  " + m.input.View() + "\n\n")
	b.WriteString(m.st.help.Render("  enter: submit   esc: back"))
	return b.String()
}

func (m model) viewRunning() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	b.WriteString("  " + m.spinner.View() + " " + m.st.accent.Render(m.result+"...") + "\n\n")
	b.WriteString(m.st.help.Render("  working - this can take a while for downloads"))
	return b.String()
}

func (m model) viewResult() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")
	if m.resultErr != nil {
		b.WriteString("  " + m.st.danger.Render("x "+m.result) + "\n")
		b.WriteString("  " + m.st.item.Render(m.resultErr.Error()) + "\n\n")
	} else {
		b.WriteString("  " + m.st.selected.Render("+ "+m.result) + "\n\n")
	}
	b.WriteString(m.st.help.Render("  press any key to return to the menu"))
	return b.String()
}
