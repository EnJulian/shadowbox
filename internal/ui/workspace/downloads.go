// internal/ui/workspace/downloads.go
package workspace

import (
	"fmt"
	"strings"

	"github.com/EnJulian/shadowbox/internal/progress"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

const progressBarWidth = 24

// Downloads is the workspace.Workspace for the Downloads nav section: the
// active job's heading, progress bar, and completion result.
type Downloads struct {
	st style.Styles

	running       bool
	heading       string
	progress      progress.Update
	spinnerFrame  string
	resultSummary string
	resultErr     error
}

// NewDownloads builds the Downloads workspace.
func NewDownloads(st style.Styles) *Downloads {
	return &Downloads{st: st}
}

func (d *Downloads) Activate() Workspace { return d }

// SetSpinnerFrame is called by root each spinner tick so the running view
// can show an animated frame without Downloads owning its own spinner.Model.
func (d *Downloads) SetSpinnerFrame(frame string) { d.spinnerFrame = frame }

// SetRunning transitions to the running state with the given initial heading.
func (d *Downloads) SetRunning(heading string) {
	d.running = true
	d.heading = heading
	d.progress = progress.Update{}
	d.resultSummary = ""
	d.resultErr = nil
}

// Finish transitions out of the running state with a result.
func (d *Downloads) Finish(summary string, err error) {
	d.running = false
	d.resultSummary = summary
	d.resultErr = err
}

func (d *Downloads) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	switch msg := msg.(type) {
	case progress.Update:
		if msg.Heading != "" {
			d.heading = msg.Heading
		}
		d.progress = msg
		return d, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "left", "h":
			return d, shell.RequestNavFocus()
		case "c":
			if d.running {
				return d, CancelTask()
			}
		}
	}
	return d, nil
}

func (d *Downloads) View(width, height int) string {
	var b strings.Builder
	switch {
	case d.running:
		heading := d.heading
		if heading == "" {
			heading = "Initializing"
		}
		b.WriteString(d.spinnerFrame + " " + d.st.Accent.Render(heading) + "\n\n")
		b.WriteString(d.ProgressBar() + "\n\n")
		b.WriteString(d.st.Help.Render("c: cancel"))
	case d.resultSummary != "":
		if d.resultErr != nil {
			b.WriteString(d.st.Danger.Render("x "+d.resultSummary) + "\n")
			b.WriteString(d.st.Item.Render(d.resultErr.Error()))
		} else {
			b.WriteString(d.st.Selected.Render("+ " + d.resultSummary))
		}
	default:
		b.WriteString(d.st.Item.Render("No active download — start one from Search, URL, or Playlist."))
	}
	return b.String()
}

// ProgressBar returns a minimal text progress bar, reusing the same visual
// language as the pre-redesign UI. Exported so root can render it standalone
// (e.g. in a status line) without duplicating the bar logic.
func (d *Downloads) ProgressBar() string {
	return renderProgressBar(d.progress, d.st)
}

// renderProgressBar returns a minimal text progress bar, reusing the same
// visual language as the pre-redesign UI.
func renderProgressBar(u progress.Update, st style.Styles) string {
	filled := progressBarWidth / 4
	if u.Total > 0 && u.Current > 0 {
		current := u.Current
		if current > u.Total {
			current = u.Total
		}
		filled = (current * progressBarWidth) / u.Total
	}
	bar := "[" + st.Accent.Render(strings.Repeat("─", filled)) + st.Help.Render(strings.Repeat("·", progressBarWidth-filled)) + "]"
	if u.Total > 0 && u.Current > 0 {
		return bar + st.Help.Render(fmt.Sprintf(" %d/%d", u.Current, u.Total))
	}
	return bar
}
