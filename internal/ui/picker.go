package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
)

type promptOutgoing struct {
	req  app.PromptRequest
	resp chan promptResult
}

type promptResult struct {
	idx int
	err error
}

type promptRequestMsg struct {
	out promptOutgoing
}

type pickerState struct {
	title   string
	options []app.PromptOption
	cursor  int
	pending *promptOutgoing
}

func (m model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.picker.pending == nil {
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.picker.pending.resp <- promptResult{idx: -1, err: app.ErrSelectionCancelled}
		m.picker.pending = nil
		if m.taskCancel != nil {
			m.taskCancel()
		}
		m.screen = screenRunning
		return m, nil
	case "up", "k":
		if m.picker.cursor > 0 {
			m.picker.cursor--
		}
	case "down", "j":
		if m.picker.cursor < len(m.picker.options)-1 {
			m.picker.cursor++
		}
	case "enter":
		idx := m.picker.cursor
		pending := m.picker.pending
		m.picker.pending = nil
		m.screen = screenRunning
		pending.resp <- promptResult{idx: idx, err: nil}
		return m, m.taskListenCmd()
	}
	return m, nil
}

func (m model) viewPicker() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render(m.picker.title) + "\n\n")

	if len(m.picker.options) == 0 {
		b.WriteString("  " + m.st.item.Render("(no options)") + "\n")
	} else {
		for i, o := range m.picker.options {
			cursor := "  "
			label := m.st.item.Render(o.Label)
			if i == m.picker.cursor {
				cursor = m.st.accent.Render("> ")
				label = m.st.selected.Render(o.Label)
			}
			b.WriteString("  " + cursor + label + "\n")
			if o.Detail != "" {
				b.WriteString("      " + m.st.help.Render(o.Detail) + "\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: navigate   enter: select   ?: help   esc: cancel"))
	return b.String()
}

func waitForPrompt(reqCh chan promptOutgoing) tea.Cmd {
	return func() tea.Msg {
		out, ok := <-reqCh
		if !ok {
			return nil
		}
		return promptRequestMsg{out: out}
	}
}

func makeSelectFunc(reqCh chan promptOutgoing) app.SelectFunc {
	return func(ctx context.Context, req app.PromptRequest) (int, error) {
		respCh := make(chan promptResult, 1)
		out := promptOutgoing{req: req, resp: respCh}
		select {
		case reqCh <- out:
		case <-ctx.Done():
			return -1, ctx.Err()
		}
		select {
		case r := <-respCh:
			return r.idx, r.err
		case <-ctx.Done():
			return -1, ctx.Err()
		}
	}
}
