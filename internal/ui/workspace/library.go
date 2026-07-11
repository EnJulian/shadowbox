// internal/ui/workspace/library.go
package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/suggest"
	tea "github.com/charmbracelet/bubbletea"
)

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true}

// Library is the workspace.Workspace for the Library nav section: a
// drill-down Artist -> Album -> Track browser with type-ahead filtering.
type Library struct {
	cfg *config.Config
	st  style.Styles

	level    int // 0 = artists, 1 = albums, 2 = tracks
	cursor   int
	artist   string
	album    string
	allNames []string // unfiltered entries at the current level
	filter   string
}

// NewLibrary builds the Library workspace.
func NewLibrary(cfg *config.Config, st style.Styles) *Library {
	return &Library{cfg: cfg, st: st}
}

func (l *Library) Activate() Workspace {
	l.level, l.cursor, l.artist, l.album, l.filter = 0, 0, "", "", ""
	l.allNames = listDirs(l.cfg.MusicDirectory)
	return l
}

func (l *Library) entries() []string {
	return suggest.Filter(l.allNames, l.filter)
}

func (l *Library) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return l, nil
	}
	// Navigation is dispatched on keyMsg.Type (not keyMsg.String()) so that
	// letter keys are never swallowed as vim-style h/j/k/l shortcuts — this
	// is a type-ahead filter, so every printable rune must always reach the
	// filter, including h, j, k, and l themselves.
	switch keyMsg.Type {
	case tea.KeyUp:
		if l.cursor > 0 {
			l.cursor--
		}
	case tea.KeyDown:
		if l.cursor < len(l.entries())-1 {
			l.cursor++
		}
	case tea.KeyLeft, tea.KeyEsc:
		return l.back()
	case tea.KeyRight, tea.KeyEnter:
		return l.enter()
	case tea.KeyBackspace:
		if l.filter != "" {
			l.filter = l.filter[:len(l.filter)-1]
			l.cursor = 0
		}
	case tea.KeyRunes:
		if len(keyMsg.Runes) == 1 {
			l.filter += string(keyMsg.Runes)
			l.cursor = 0
		}
	}
	return l, nil
}

func (l *Library) back() (Workspace, tea.Cmd) {
	l.filter = ""
	l.cursor = 0
	switch l.level {
	case 0:
		return l, shell.RequestNavFocus()
	case 1:
		l.level = 0
		l.artist = ""
		l.allNames = listDirs(l.cfg.MusicDirectory)
	case 2:
		l.level = 1
		l.album = ""
		l.allNames = listDirs(filepath.Join(l.cfg.MusicDirectory, l.artist))
	}
	return l, nil
}

func (l *Library) enter() (Workspace, tea.Cmd) {
	entries := l.entries()
	if len(entries) == 0 || l.level == 2 {
		return l, nil
	}
	selected := entries[l.cursor]
	l.filter = ""
	l.cursor = 0
	switch l.level {
	case 0:
		l.artist = selected
		l.level = 1
		l.allNames = listDirs(filepath.Join(l.cfg.MusicDirectory, l.artist))
	case 1:
		l.album = selected
		l.level = 2
		l.allNames = listTracks(filepath.Join(l.cfg.MusicDirectory, l.artist, l.album))
	}
	return l, nil
}

func (l *Library) View(width, height int) string {
	var b strings.Builder
	var crumb string
	switch l.level {
	case 0:
		crumb = "Library / Artists"
	case 1:
		crumb = "Library / " + l.artist
	case 2:
		crumb = "Library / " + l.artist + " / " + l.album
	}
	b.WriteString(l.st.Subtitle.Render(crumb) + "\n")
	if l.filter != "" {
		b.WriteString(l.st.Help.Render("filter: "+l.filter) + "\n")
	}
	b.WriteString("\n")

	entries := l.entries()
	if len(entries) == 0 {
		b.WriteString(l.st.Item.Render("(empty)") + "\n")
	}
	for i, e := range entries {
		cursor := "  "
		label := l.st.Item.Render(e)
		if i == l.cursor {
			cursor = l.st.Accent.Render("> ")
			label = l.st.Selected.Render(e)
		}
		b.WriteString(cursor + label + "\n")
	}
	return b.String()
}

func listDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}

func listTracks(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if audioExts[strings.ToLower(filepath.Ext(e.Name()))] {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
