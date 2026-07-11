// Package shell renders Shadowbox's persistent two-pane frame (Nav |
// Content) and owns which pane currently has keyboard focus.
package shell

import tea "github.com/charmbracelet/bubbletea"

// Pane identifies which region of the shell currently has keyboard focus.
type Pane int

const (
	PaneNav Pane = iota
	PaneContent
)

// Toggle returns the other pane. With exactly two panes, Tab and Shift+Tab
// both cycle to the same place, matching the spec's "always works, no dead
// ends" focus model.
func (p Pane) Toggle() Pane {
	if p == PaneNav {
		return PaneContent
	}
	return PaneNav
}

// FocusNavMsg is emitted by a workspace to ask the root model to return
// focus to Nav (e.g. pressing "left" from the leftmost Library column).
type FocusNavMsg struct{}

// RequestNavFocus returns a tea.Cmd that emits FocusNavMsg.
func RequestNavFocus() tea.Cmd {
	return func() tea.Msg { return FocusNavMsg{} }
}
