package ui

import (
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/setup"
)

// wizardItem is one row in the setup wizard: either a tool to install, or
// the trailing "Done" entry.
type wizardItem struct {
	tool      setup.Tool
	isDone    bool
	showHint  bool
	resultMsg string
}

// wizardInstallDoneMsg reports the outcome of running an install command
// via tea.ExecProcess for the item at index.
type wizardInstallDoneMsg struct {
	index int
	tool  string
	err   error
}

// openWizard switches to the setup wizard screen, remembering returnTo so
// "Done" (or Esc) knows where to go back to.
func (m model) openWizard(returnTo screen) (tea.Model, tea.Cmd) {
	m.screen = screenSetupWizard
	m.wizardReturnTo = returnTo
	m.wizardCursor = 0
	items := make([]wizardItem, 0, len(setup.Tools)+1)
	for _, t := range setup.Tools {
		items = append(items, wizardItem{tool: t})
	}
	items = append(items, wizardItem{isDone: true})
	m.wizardItems = items
	return m, nil
}

func (m model) updateWizard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.wizardDone()
	case "up", "k":
		if m.wizardCursor > 0 {
			m.wizardCursor--
		}
	case "down", "j":
		if m.wizardCursor < len(m.wizardItems)-1 {
			m.wizardCursor++
		}
	case "enter":
		return m.wizardActivate()
	}
	return m, nil
}

func (m model) wizardActivate() (tea.Model, tea.Cmd) {
	item := m.wizardItems[m.wizardCursor]
	if item.isDone {
		return m.wizardDone()
	}
	if setup.Available(item.tool.Name) {
		return m, nil
	}
	result := setup.InstallCommand(item.tool)
	if !result.Automatable {
		items := make([]wizardItem, len(m.wizardItems))
		copy(items, m.wizardItems)
		items[m.wizardCursor].showHint = !items[m.wizardCursor].showHint
		m.wizardItems = items
		return m, nil
	}
	cmd := exec.Command(result.Command[0], result.Command[1:]...)
	index := m.wizardCursor
	toolName := item.tool.Name
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return wizardInstallDoneMsg{index: index, tool: toolName, err: err}
	})
}

func (m model) wizardDone() (tea.Model, tea.Cmd) {
	_ = config.Save(m.cfg)
	m.screen = m.wizardReturnTo
	return m, nil
}

// handleWizardInstallDone applies the result of a finished install attempt.
func (m model) handleWizardInstallDone(msg wizardInstallDoneMsg) model {
	items := make([]wizardItem, len(m.wizardItems))
	copy(items, m.wizardItems)
	switch {
	case msg.err != nil:
		items[msg.index].resultMsg = "install failed — see output above"
	case setup.Available(msg.tool):
		items[msg.index].resultMsg = "installed"
	default:
		items[msg.index].resultMsg = "install finished, but still not found on PATH"
	}
	m.wizardItems = items
	return m
}

func (m model) viewWizard() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Setup — optional tools") + "\n\n")

	for i, item := range m.wizardItems {
		label := "Done"
		if !item.isDone {
			status := "not installed"
			if setup.Available(item.tool.Name) {
				status = "installed"
			} else if item.resultMsg != "" {
				status = item.resultMsg
			}
			label = item.tool.Name + " — " + item.tool.Purpose + " (" + status + ")"
		}

		cursor := "  "
		line := m.st.item.Render(label)
		if i == m.wizardCursor {
			cursor = m.st.accent.Render("> ")
			line = m.st.selected.Render(label)
		}
		b.WriteString("  " + cursor + line + "\n")

		if !item.isDone && item.showHint {
			result := setup.InstallCommand(item.tool)
			b.WriteString("      " + m.st.help.Render(result.Instruction) + "\n")
		}
	}

	help := "  up/down: navigate   enter: install/toggle   ?: help   esc: done"
	if current := m.wizardItems[m.wizardCursor]; !current.isDone && !setup.Available(current.tool.Name) {
		if setup.InstallCommand(current.tool).Automatable {
			help = "  up/down: navigate   enter: install (may ask for your password)   ?: help   esc: done"
		}
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render(help))
	return b.String()
}
