// internal/ui/workspace/log.go
package workspace

import (
	"strconv"
	"strings"

	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

// Log is the workspace.Workspace for the Log nav section: the scrollable
// download log viewer.
type Log struct {
	st     style.Styles
	lines  []string
	scroll int
	height int // last known viewport height, for scroll clamping
}

// NewLog builds the Log workspace.
func NewLog(st style.Styles) *Log {
	return &Log{st: st, height: 10}
}

func (l *Log) Activate() Workspace {
	l.reload()
	return l
}

func (l *Log) reload() {
	_ = applog.LoadDownloadLog()
	l.lines = applog.DownloadLogLines()
	l.scroll = l.maxScroll()
}

func (l *Log) maxScroll() int {
	if len(l.lines) <= l.height {
		return 0
	}
	return len(l.lines) - l.height
}

func (l *Log) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return l, nil
	}
	switch keyMsg.String() {
	case "esc", "left", "h":
		return l, shell.RequestNavFocus()
	case "up", "k":
		if l.scroll > 0 {
			l.scroll--
		}
	case "down", "j":
		if l.scroll < l.maxScroll() {
			l.scroll++
		}
	case "pgup":
		l.scroll = max(0, l.scroll-l.height)
	case "pgdown":
		l.scroll = min(l.maxScroll(), l.scroll+l.height)
	case "home":
		l.scroll = 0
	case "end", "G":
		l.scroll = l.maxScroll()
	case "r":
		l.reload()
	}
	return l, nil
}

func (l *Log) View(width, height int) string {
	if height > 4 {
		l.height = height - 4
	}
	if l.scroll > l.maxScroll() {
		l.scroll = l.maxScroll()
	}

	var b strings.Builder
	if len(l.lines) == 0 {
		b.WriteString(l.st.Item.Render("(no download logs yet — run a download first)"))
		return b.String()
	}

	end := l.scroll + l.height
	if end > len(l.lines) {
		end = len(l.lines)
	}
	for _, line := range l.lines[l.scroll:end] {
		b.WriteString(l.st.Item.Render(line) + "\n")
	}
	if len(l.lines) > l.height {
		b.WriteString("\n" + l.st.Help.Render("lines "+strconv.Itoa(l.scroll+1)+"-"+strconv.Itoa(end)+" of "+strconv.Itoa(len(l.lines))))
	}
	return b.String()
}
