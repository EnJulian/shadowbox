package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/player"
)

// libState tracks navigation through the Artist/Album/Track hierarchy, plus
// a live type-ahead filter over the current level's entries.
type libState struct {
	level   int // 0 = artists, 1 = albums, 2 = tracks
	cursor  int
	artist  string
	album   string
	entries []string // full, unfiltered list for the current level
	filter  string
}

// visible returns entries narrowed by filter (case-insensitive substring
// match), or entries unchanged when filter is empty.
func (l libState) visible() []string {
	if l.filter == "" {
		return l.entries
	}
	lower := strings.ToLower(l.filter)
	var out []string
	for _, e := range l.entries {
		if strings.Contains(strings.ToLower(e), lower) {
			out = append(out, e)
		}
	}
	return out
}

// startPlaybackMsg asks the root model to start playing tracks[index].
type startPlaybackMsg struct {
	tracks []player.Track
	index  int
}

var audioExts = map[string]bool{".opus": true, ".mp3": true, ".m4a": true, ".flac": true, ".wav": true, ".ogg": true, ".webm": true, ".aac": true}

func (m model) openLibrary() (tea.Model, tea.Cmd) {
	m.screen = screenLibrary
	m.lib = libState{level: 0, entries: listDirs(m.cfg.MusicDirectory)}
	return m, nil
}

func (m model) updateLibrary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyLeft:
		return m.libraryBack()
	case tea.KeyUp:
		if m.lib.cursor > 0 {
			m.lib.cursor--
		}
	case tea.KeyDown:
		visible := m.lib.visible()
		if m.lib.cursor < len(visible)-1 {
			m.lib.cursor++
		}
	case tea.KeyRight, tea.KeyEnter:
		return m.libraryEnter()
	case tea.KeyBackspace:
		if len(m.lib.filter) > 0 {
			r := []rune(m.lib.filter)
			m.lib.filter = string(r[:len(r)-1])
			m.lib.cursor = 0
		}
	case tea.KeyRunes:
		// Every printable rune reaches the filter unconditionally — no
		// vim-key (h/j/k/l) shortcuts here, since any of those letters can
		// legitimately start a search query (e.g. "kanye", "long"), and a
		// guard like "only when filter is empty" is always true for the
		// very first keystroke, which would silently eat it as navigation
		// instead. Real navigation already has dedicated keys above
		// (arrows, enter, esc, backspace).
		m.lib.filter += msg.String()
		m.lib.cursor = 0
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
		m.lib = libState{level: 1, artist: m.lib.artist, entries: listAlbumsAndBareTracks(artistDir)}
	}
	return m, nil
}

func (m model) libraryEnter() (tea.Model, tea.Cmd) {
	visible := m.lib.visible()
	if len(visible) == 0 {
		return m, nil
	}
	if m.lib.cursor >= len(visible) {
		m.lib.cursor = len(visible) - 1
	}
	selected := visible[m.lib.cursor]

	switch m.lib.level {
	case 0:
		artistDir := filepath.Join(m.cfg.MusicDirectory, selected)
		m.lib = libState{level: 1, artist: selected, entries: listAlbumsAndBareTracks(artistDir)}
	case 1:
		entryPath := filepath.Join(m.cfg.MusicDirectory, m.lib.artist, selected)
		if info, err := os.Stat(entryPath); err == nil && !info.IsDir() {
			// A track sitting directly under the artist folder, with no
			// album level to descend into — play it immediately as a
			// single-track queue instead of trying (and failing) to list
			// tracks inside it as if it were an album directory.
			title := strings.TrimSuffix(selected, filepath.Ext(selected))
			return m, func() tea.Msg {
				return startPlaybackMsg{tracks: []player.Track{{Path: entryPath, Title: title}}, index: 0}
			}
		}
		m.lib = libState{level: 2, artist: m.lib.artist, album: selected, entries: listTracks(entryPath)}
	case 2:
		albumDir := filepath.Join(m.cfg.MusicDirectory, m.lib.artist, m.lib.album)
		index := 0
		for i, name := range m.lib.entries {
			if name == selected {
				index = i
			}
		}
		tracks := make([]player.Track, len(m.lib.entries))
		for i, name := range m.lib.entries {
			base := filepath.Base(name)
			tracks[i] = player.Track{
				Path:  filepath.Join(albumDir, filepath.FromSlash(name)),
				Title: strings.TrimSuffix(base, filepath.Ext(base)),
			}
		}
		return m, func() tea.Msg { return startPlaybackMsg{tracks: tracks, index: index} }
	}
	return m, nil
}

func (m model) viewLibrary() string {
	var b strings.Builder
	b.WriteString(renderBannerWithPlayback(m.st, m.theme, m.playback))
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

	visible := m.lib.visible()
	if m.lib.filter != "" {
		b.WriteString("  " + m.st.help.Render("filter: "+m.lib.filter) + "\n\n")
	}
	if len(visible) == 0 {
		b.WriteString("  " + m.st.item.Render("(no matches)") + "\n")
	}
	for i, e := range visible {
		cursor := "  "
		label := m.st.item.Render(e)
		if i == m.lib.cursor {
			cursor = m.st.accent.Render("> ")
			label = m.st.selected.Render(e)
		}
		b.WriteString("  " + cursor + label + "\n")
	}

	if m.playerErr != "" {
		b.WriteString("\n  " + m.st.danger.Render(m.playerErr) + "\n")
	}
	if m.playback.LastError != "" {
		b.WriteString("\n  " + m.st.danger.Render(m.playback.LastError) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(m.st.help.Render("  type to filter   enter: open   esc: back"))
	return b.String()
}

// listAlbumsAndBareTracks returns the sorted names of album subdirectories
// plus any audio files sitting directly in dir (a track with no album
// level) — used for the Artist-level Library view, since some tracks (not
// downloaded through Shadowbox's own Artist/Album/Track pipeline) may sit
// directly under an artist folder with no album directory at all.
func listAlbumsAndBareTracks(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
			continue
		}
		if audioExts[strings.ToLower(filepath.Ext(e.Name()))] {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
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

// listTracks returns the sorted, dir-relative paths (forward-slash separated)
// of every audio file under dir, including ones nested arbitrarily deep in
// subdirectories — e.g. multi-disc albums organized into CD1/CD2 folders, or
// libraries imported from elsewhere with extra nesting.
func listTracks(dir string) []string {
	var names []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries, keep walking the rest
		}
		if d.IsDir() {
			return nil
		}
		if !audioExts[strings.ToLower(filepath.Ext(d.Name()))] {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		names = append(names, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(names)
	return names
}
