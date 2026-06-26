package ui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// libState tracks navigation through the Artist/Album/Track hierarchy.
type libState struct {
	level   int // 0 = artists, 1 = albums, 2 = tracks
	cursor  int
	artist  string
	album   string
	entries []string
}

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true}

func (m model) openLibrary() (tea.Model, tea.Cmd) {
	m.screen = screenLibrary
	m.lib = libState{level: 0, entries: listDirs(m.cfg.MusicDirectory)}
	return m, nil
}

func (m model) updateLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		m.screen = screenMenu
		return m, nil
	case "esc", "left", "h":
		return m.libraryBack()
	case "up", "k":
		if m.lib.cursor > 0 {
			m.lib.cursor--
		}
	case "down", "j":
		if m.lib.cursor < len(m.lib.entries)-1 {
			m.lib.cursor++
		}
	case "enter", "right", "l":
		return m.libraryEnter()
	}
	return m, nil
}

func (m model) libraryBack() (tea.Model, tea.Cmd) {
	switch m.lib.level {
	case 0:
		m.screen = screenMenu
	case 1:
		m.lib = libState{level: 0, entries: listDirs(m.cfg.MusicDirectory)}
	case 2:
		artistDir := filepath.Join(m.cfg.MusicDirectory, m.lib.artist)
		m.lib = libState{level: 1, artist: m.lib.artist, entries: listDirs(artistDir)}
	}
	return m, nil
}

func (m model) libraryEnter() (tea.Model, tea.Cmd) {
	if len(m.lib.entries) == 0 {
		return m, nil
	}
	selected := m.lib.entries[m.lib.cursor]
	switch m.lib.level {
	case 0:
		artistDir := filepath.Join(m.cfg.MusicDirectory, selected)
		m.lib = libState{level: 1, artist: selected, entries: listDirs(artistDir)}
	case 1:
		albumDir := filepath.Join(m.cfg.MusicDirectory, m.lib.artist, selected)
		m.lib = libState{level: 2, artist: m.lib.artist, album: selected, entries: listTracks(albumDir)}
	}
	return m, nil
}

func (m model) viewLibrary() string {
	var b strings.Builder
	b.WriteString(m.st.title.Render(banner))
	b.WriteString("\n\n")

	var crumb string
	switch m.lib.level {
	case 0:
		crumb = "Library / Artists"
	case 1:
		crumb = "Library / " + m.lib.artist
	case 2:
		crumb = "Library / " + m.lib.artist + " / " + m.lib.album
	}
	b.WriteString("  " + m.st.subtitle.Render(crumb) + "\n\n")

	if len(m.lib.entries) == 0 {
		b.WriteString("  " + m.st.item.Render("(empty)") + "\n")
	}
	for i, e := range m.lib.entries {
		cursor := "  "
		label := m.st.item.Render(e)
		if i == m.lib.cursor {
			cursor = m.st.accent.Render("> ")
			label = m.st.selected.Render(e)
		}
		b.WriteString("  " + cursor + label + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  up/down: navigate   enter: open   esc: back   q: menu"))
	return b.String()
}

// listDirs returns the sorted names of subdirectories in dir.
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

// listTracks returns the sorted names of audio files in dir.
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
