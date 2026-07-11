// Package workspace implements the content pane's per-section behavior
// (Search, URL, Playlist, Library, Downloads, Enhance, Log, Settings). Each
// Workspace reports side effects (starting a task, switching section,
// cancelling) to the root model via tea.Cmd-returned messages rather than
// calling back directly, so this package never imports the root ui package.
package workspace

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
)

// Section identifies one nav sidebar entry / content pane.
type Section int

const (
	SectionSearch Section = iota
	SectionURL
	SectionPlaylist
	SectionLibrary
	SectionDownloads
	SectionEnhance
	SectionLog
	SectionSettings
)

// Meta describes one nav sidebar entry.
type Meta struct {
	Section Section
	Label   string
}

// Order is the fixed nav sidebar ordering, matching the spec's Nav Sections table.
var Order = []Meta{
	{SectionSearch, "Search"},
	{SectionURL, "URL"},
	{SectionPlaylist, "Playlist"},
	{SectionLibrary, "Library"},
	{SectionDownloads, "Downloads"},
	{SectionEnhance, "Enhance"},
	{SectionLog, "Log"},
	{SectionSettings, "Settings"},
}

// Workspace is a single nav section's content-pane behavior.
type Workspace interface {
	// Activate resets internal focus/cursor state (e.g. focuses the query
	// input). Called whenever this section becomes active, and whenever the
	// Content pane regains focus while this section is already active.
	Activate() Workspace
	Update(msg tea.Msg) (Workspace, tea.Cmd)
	View(width, height int) string
}

// StartTaskMsg asks the root model to run a background pipeline operation,
// identical in shape to the pre-redesign model.startTask contract.
type StartTaskMsg struct {
	Summary string
	Run     func(ctx context.Context, opts app.Options) error
}

// SwitchSectionMsg asks the root model to change the active nav section and
// focus Content on it.
type SwitchSectionMsg struct {
	Section Section
}

// CancelTaskMsg asks the root model to cancel the running task.
type CancelTaskMsg struct{}

// StartTask returns a tea.Cmd that emits StartTaskMsg.
func StartTask(summary string, run func(ctx context.Context, opts app.Options) error) tea.Cmd {
	return func() tea.Msg { return StartTaskMsg{Summary: summary, Run: run} }
}

// SwitchSection returns a tea.Cmd that emits SwitchSectionMsg.
func SwitchSection(s Section) tea.Cmd {
	return func() tea.Msg { return SwitchSectionMsg{Section: s} }
}

// CancelTask returns a tea.Cmd that emits CancelTaskMsg.
func CancelTask() tea.Cmd {
	return func() tea.Msg { return CancelTaskMsg{} }
}
