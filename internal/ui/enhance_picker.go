package ui

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
)

// enhancePickerUseThisFolder is the pinned first row of every directory
// listing in the picker — selecting it confirms cwd as the Enhance target.
const enhancePickerUseThisFolder = "[ Use this folder ]"

// enhancePickerState tracks folder-browser navigation for "Enhance Existing
// Files". Unlike libState (fixed Artist/Album/Track depth), it can descend
// or ascend to any directory on disk, starting from the configured music
// directory.
type enhancePickerState struct {
	cwd     string   // absolute path currently being browsed
	cursor  int
	entries []string // subdirectory names in cwd, sorted, dotfiles excluded
	filter  string
}

// visible returns the pinned "use this folder" row followed by entries
// narrowed by filter (case-insensitive substring match).
func (e enhancePickerState) visible() []string {
	out := make([]string, 0, len(e.entries)+1)
	out = append(out, enhancePickerUseThisFolder)
	if e.filter == "" {
		return append(out, e.entries...)
	}
	lower := strings.ToLower(e.filter)
	for _, name := range e.entries {
		if strings.Contains(strings.ToLower(name), lower) {
			out = append(out, name)
		}
	}
	return out
}

func (m model) openEnhancePicker() (tea.Model, tea.Cmd) {
	m.screen = screenEnhancePicker
	m.enhancePicker = enhancePickerState{cwd: m.cfg.MusicDirectory, entries: listSubdirs(m.cfg.MusicDirectory)}
	return m, nil
}

func (m model) updateEnhancePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyLeft:
		return m.enhancePickerUp()
	case tea.KeyUp:
		if m.enhancePicker.cursor > 0 {
			m.enhancePicker.cursor--
		}
	case tea.KeyDown:
		visible := m.enhancePicker.visible()
		if m.enhancePicker.cursor < len(visible)-1 {
			m.enhancePicker.cursor++
		}
	case tea.KeyRight, tea.KeyEnter:
		return m.enhancePickerEnter()
	case tea.KeyBackspace:
		if len(m.enhancePicker.filter) > 0 {
			r := []rune(m.enhancePicker.filter)
			m.enhancePicker.filter = string(r[:len(r)-1])
			m.enhancePicker.cursor = 0
		}
	case tea.KeyRunes:
		m.enhancePicker.filter += msg.String()
		m.enhancePicker.cursor = 0
	}
	return m, nil
}

// enhancePickerUp navigates to the parent of cwd, or returns to the main
// menu when cwd has no parent (the filesystem root).
func (m model) enhancePickerUp() (tea.Model, tea.Cmd) {
	parent := filepath.Dir(m.enhancePicker.cwd)
	if parent == m.enhancePicker.cwd {
		m.screen = screenMenu
		return m, nil
	}
	m.enhancePicker = enhancePickerState{cwd: parent, entries: listSubdirs(parent)}
	return m, nil
}

func (m model) enhancePickerEnter() (tea.Model, tea.Cmd) {
	visible := m.enhancePicker.visible()
	if m.enhancePicker.cursor >= len(visible) {
		m.enhancePicker.cursor = len(visible) - 1
	}

	if m.enhancePicker.cursor == 0 {
		a := m.app
		dir := m.enhancePicker.cwd
		cmd := m.startTask("Enhancement", func(ctx context.Context, opts app.Options) error {
			return a.EnhanceDir(ctx, dir, true, []string{"opus", "mp3", "m4a", "flac"}, false, opts)
		})
		return m, cmd
	}

	selected := visible[m.enhancePicker.cursor]
	child := filepath.Join(m.enhancePicker.cwd, selected)
	m.enhancePicker = enhancePickerState{cwd: child, entries: listSubdirs(child)}
	return m, nil
}

func (m model) viewEnhancePicker() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
	b.WriteString("\n\n")
	b.WriteString("  " + m.st.subtitle.Render("Enhance Existing Files / "+m.enhancePicker.cwd) + "\n\n")

	if m.enhancePicker.filter != "" {
		b.WriteString("  " + m.st.help.Render("filter: "+m.enhancePicker.filter) + "\n\n")
	}
	for i, e := range m.enhancePicker.visible() {
		cursor := "  "
		label := m.st.item.Render(e)
		if i == m.enhancePicker.cursor {
			cursor = m.st.accent.Render("> ")
			label = m.st.selected.Render(e)
		}
		b.WriteString("  " + cursor + label + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  type to filter   enter: open / use folder   esc: up"))
	return b.String()
}

// listSubdirs returns the sorted names of subdirectories in dir, excluding
// dotfiles.
func listSubdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}
