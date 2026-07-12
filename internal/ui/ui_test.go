package ui

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/player"
)

func newTestModel() model {
	cfg := &config.Config{AudioFormat: "opus", MusicDirectory: ".", Theme: "hacker", UseGenius: true}
	theme := themeByName(cfg.Theme)
	ti := textinput.New()
	return model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      newStyles(theme),
		input:   ti,
		spinner: spinner.New(),
		screen:  screenMenu,
	}
}

func key(s string) tea.KeyMsg {
	switch s {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestMenuRendersAllScreens(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // isolate config writes
	m := newTestModel()

	if !strings.Contains(m.viewMenu(), "Search & Download") {
		t.Error("menu missing first item")
	}

	// Navigate to Settings (index 6) and open it.
	for i := 0; i < 6; i++ {
		next, _ := m.updateMenu(key("down"))
		m = next.(model)
	}
	next, _ := m.handleKey(key("enter"))
	m = next.(model)
	if m.screen != screenSettings {
		t.Fatalf("expected settings screen, got %v", m.screen)
	}
	if !strings.Contains(m.viewSettings(), "Audio format") {
		t.Error("settings view missing items")
	}

	// Toggle a boolean setting (use_genius at index 2).
	m.settingsCursor = 2
	before := m.cfg.UseGenius
	next, _ = m.activateSetting()
	m = next.(model)
	if m.cfg.UseGenius == before {
		t.Error("toggle did not flip use_genius")
	}

	// Open theme picker and apply a different theme.
	m.settingsCursor = 4
	next, _ = m.activateSetting()
	m = next.(model)
	if m.screen != screenThemePicker {
		t.Fatalf("expected theme picker, got %v", m.screen)
	}
	next, _ = m.updateThemePicker(key("down"))
	m = next.(model)
	if !strings.Contains(m.viewThemePicker(), "matrix") {
		t.Error("theme picker missing themes")
	}
}

func TestThemePickerEscRevertsPreview(t *testing.T) {
	m := newTestModel() // Theme: "hacker"
	m.screen = screenThemePicker
	m.themeCursor = 0

	next, _ := m.updateThemePicker(key("down")) // preview "matrix"
	m = next.(model)
	if m.theme.Name != "matrix" {
		t.Fatalf("expected live preview to switch to matrix, got %q", m.theme.Name)
	}

	next, _ = m.updateThemePicker(key("esc"))
	m = next.(model)
	if m.screen != screenSettings {
		t.Fatalf("expected esc to return to settings, got %v", m.screen)
	}
	if m.theme.Name != "hacker" {
		t.Errorf("esc did not revert the previewed theme: got %q, want %q (m.cfg.Theme)", m.theme.Name, "hacker")
	}
	if m.cfg.Theme != "hacker" {
		t.Errorf("esc must not persist the previewed theme to cfg: got %q, want %q", m.cfg.Theme, "hacker")
	}
}

func TestLibraryTypeAheadFiltersEntries(t *testing.T) {
	m := newTestModel()
	m.lib = libState{level: 0, entries: []string{"Nujabes", "Kanye West", "Aphex Twin"}}

	next, _ := m.updateLibrary(key("k")) // 'k' must filter, not be treated as "up"
	m = next.(model)
	next, _ = m.updateLibrary(key("a"))
	m = next.(model)
	next, _ = m.updateLibrary(key("n"))
	m = next.(model)

	if m.lib.filter != "kan" {
		t.Fatalf("filter = %q, want %q", m.lib.filter, "kan")
	}
	visible := m.lib.visible()
	if len(visible) != 1 || visible[0] != "Kanye West" {
		t.Fatalf("visible() = %v, want [Kanye West]", visible)
	}
}

func TestLibraryFilterBackspaceShortens(t *testing.T) {
	m := newTestModel()
	m.lib = libState{level: 0, entries: []string{"Nujabes"}, filter: "nuj"}

	next, _ := m.updateLibrary(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(model)
	if m.lib.filter != "nu" {
		t.Fatalf("filter after backspace = %q, want %q", m.lib.filter, "nu")
	}
}

func TestLibraryFilterResetsOnLevelChange(t *testing.T) {
	m := newTestModel()
	m.lib = libState{level: 0, entries: []string{"Nujabes"}, cursor: 0, filter: "nuj"}

	next, _ := m.libraryEnter()
	m = next.(model)
	if m.lib.filter != "" {
		t.Errorf("filter after drilling into a new level = %q, want empty", m.lib.filter)
	}
}

func TestLibraryQIsFilterableNotAShortcut(t *testing.T) {
	m := newTestModel()
	m.screen = screenLibrary // must be set explicitly: updateLibrary alone never changes m.screen except via libraryBack(), so this proves 'q' didn't trigger that path, not just that the screen was already screenMenu to begin with
	m.lib = libState{level: 0, entries: []string{"Queen"}}

	next, _ := m.updateLibrary(key("q"))
	m = next.(model)
	if m.screen != screenLibrary {
		t.Fatalf("'q' must not exit Library while type-ahead filtering is active, got screen = %v", m.screen)
	}
	if m.lib.filter != "q" {
		t.Errorf("filter = %q, want %q", m.lib.filter, "q")
	}
}

func TestLibraryEnterOnTrackReturnsStartPlaybackCmd(t *testing.T) {
	m := newTestModel()
	m.lib = libState{
		level:   2,
		artist:  "Nujabes",
		album:   "Modal Soul",
		cursor:  1,
		entries: []string{"01 Feather.opus", "02 Reflection Eternal.opus"},
	}

	_, cmd := m.libraryEnter()
	if cmd == nil {
		t.Fatal("expected a cmd starting playback, got nil")
	}
	msg := cmd()
	sp, ok := msg.(startPlaybackMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want startPlaybackMsg", msg)
	}
	if sp.index != 1 || len(sp.tracks) != 2 {
		t.Fatalf("startPlaybackMsg = %+v, want index=1 and 2 tracks", sp)
	}
	if sp.tracks[1].Title != "02 Reflection Eternal" {
		t.Errorf("tracks[1].Title = %q, want %q (extension stripped)", sp.tracks[1].Title, "02 Reflection Eternal")
	}
}

func TestInputAndLibraryViews(t *testing.T) {
	m := newTestModel()
	next, _ := m.openInput("search", "Enter query")
	m = next.(model)
	if m.screen != screenInput || !strings.Contains(m.viewInput(), "Enter query") {
		t.Error("input screen not shown")
	}

	next, _ = m.openLibrary()
	m = next.(model)
	if m.screen != screenLibrary {
		t.Fatalf("expected library screen, got %v", m.screen)
	}
	if v := m.viewLibrary(); !strings.Contains(v, "Artists") {
		t.Errorf("library view missing breadcrumb: %q", v)
	}

	next, _ = m.openDownloadLog()
	m = next.(model)
	if m.screen != screenDownloadLog {
		t.Fatalf("expected download log screen, got %v", m.screen)
	}
	if !strings.Contains(m.viewDownloadLog(), "Download log") {
		t.Error("download log view missing title")
	}
}

func TestGlobalPlaybackKeysDoNothingWithoutAPlayer(t *testing.T) {
	m := newTestModel()
	m.screen = screenMenu
	// No player loaded yet — space/n/p/s/arrows must not panic.
	for _, k := range []string{" ", "n", "p", "s"} {
		next, _ := m.handleKey(key(k))
		m = next.(model)
	}
	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	m = next.(model)
	_ = next
}

func TestSpacebarDoesNotTogglePauseOnLibraryScreen(t *testing.T) {
	m := newTestModel()
	m.screen = screenLibrary
	m.lib = libState{level: 0, entries: []string{"Nujabes"}}

	next, _ := m.handleKey(key(" "))
	m = next.(model)
	// " " must have been treated as a filter character, not a global pause.
	if m.lib.filter != " " {
		t.Errorf("filter = %q, want a single space (global pause must not intercept it here)", m.lib.filter)
	}
}

func TestSpacebarDoesNotTogglePauseOnInputScreen(t *testing.T) {
	m := newTestModel()
	next, _ := m.openInput("search", "Enter query")
	m = next.(model)

	valueBefore := m.input.Value()
	next, _ = m.updateInput(key(" "))
	m = next.(model)
	if m.input.Value() == valueBefore {
		t.Error("space must reach the text input on the Input screen, not be swallowed as global pause")
	}
}

func TestStartPlaybackMsgWithoutMpvSetsPlayerErr(t *testing.T) {
	if player.Available() {
		t.Skip("mpv is installed on this machine; this test only covers the not-installed path")
	}
	m := newTestModel()
	next, _ := m.Update(startPlaybackMsg{tracks: []player.Track{{Path: "/x.opus", Title: "X"}}, index: 0})
	m = next.(model)
	if m.playerErr == "" {
		t.Error("expected playerErr to be set when mpv is not installed")
	}
	if m.player != nil {
		t.Error("expected model.player to stay nil when mpv is not installed")
	}
}

func TestPlayerErrClearsOnNextKeyPress(t *testing.T) {
	m := newTestModel()
	m.playerErr = "mpv not found"
	next, _ := m.handleKey(key("x"))
	m = next.(model)
	if m.playerErr != "" {
		t.Error("expected playerErr to clear on the next keypress")
	}
}

func TestPlaybackLastErrorRendersOnMenuAndLibrary(t *testing.T) {
	m := newTestModel()
	m.playback = player.State{LastError: "playback error, skipping track"}

	if !strings.Contains(m.viewMenu(), "playback error, skipping track") {
		t.Error("expected LastError to render on the menu screen")
	}

	next, _ := m.openLibrary()
	m = next.(model)
	m.playback = player.State{LastError: "playback error, skipping track"}
	if !strings.Contains(m.viewLibrary(), "playback error, skipping track") {
		t.Error("expected LastError to render on the library screen")
	}
}

func TestLibraryArtistLevelShowsBareTracksAlongsideAlbums(t *testing.T) {
	dir := t.TempDir()
	artistDir := filepath.Join(dir, "Aimer")
	if err := os.MkdirAll(filepath.Join(artistDir, "Album"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artistDir, "Bare Track.opus"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	m.lib = libState{level: 0, entries: []string{"Aimer"}}

	next, _ := m.libraryEnter()
	m = next.(model)
	if m.lib.level != 1 {
		t.Fatalf("expected level 1, got %d", m.lib.level)
	}
	want := []string{"Album", "Bare Track.opus"}
	got := m.lib.visible()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries = %v, want %v (bare track alongside the album dir, both listed)", got, want)
	}
}

func TestLibraryEnterOnBareTrackStartsPlaybackDirectly(t *testing.T) {
	dir := t.TempDir()
	artistDir := filepath.Join(dir, "Aimer")
	if err := os.MkdirAll(artistDir, 0o755); err != nil {
		t.Fatal(err)
	}
	trackPath := filepath.Join(artistDir, "Bare Track.opus")
	if err := os.WriteFile(trackPath, []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	m.lib = libState{level: 1, artist: "Aimer", entries: []string{"Bare Track.opus"}, cursor: 0}

	_, cmd := m.libraryEnter()
	if cmd == nil {
		t.Fatal("expected a cmd starting playback directly, got nil (did it try to descend to a nonexistent album level instead?)")
	}
	msg := cmd()
	sp, ok := msg.(startPlaybackMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want startPlaybackMsg", msg)
	}
	if len(sp.tracks) != 1 || sp.tracks[0].Path != trackPath {
		t.Fatalf("startPlaybackMsg = %+v, want a single track at %q", sp, trackPath)
	}
	if sp.tracks[0].Title != "Bare Track" {
		t.Errorf("Title = %q, want %q (extension stripped)", sp.tracks[0].Title, "Bare Track")
	}
}

func TestLibraryEnterOnAlbumDirAtArtistLevelStillDescends(t *testing.T) {
	dir := t.TempDir()
	artistDir := filepath.Join(dir, "Aimer")
	if err := os.MkdirAll(filepath.Join(artistDir, "Album"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artistDir, "Album", "Track.opus"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	m.lib = libState{level: 1, artist: "Aimer", entries: []string{"Album"}, cursor: 0}

	next, cmd := m.libraryEnter()
	m = next.(model)
	if cmd != nil {
		t.Fatal("expected no playback cmd when entering a real album directory, got one")
	}
	if m.lib.level != 2 {
		t.Fatalf("expected to descend to level 2, got %d", m.lib.level)
	}
	if len(m.lib.entries) != 1 || m.lib.entries[0] != "Track.opus" {
		t.Fatalf("entries = %v, want [Track.opus]", m.lib.entries)
	}
}

func TestSettingsSetupWizardEntryOpensWizard(t *testing.T) {
	m := newTestModel()
	m.screen = screenSettings
	m.settingsCursor = len(settingItems) - 1 // the new entry, appended last

	next, _ := m.activateSetting()
	m = next.(model)
	if m.screen != screenSetupWizard {
		t.Fatalf("expected screenSetupWizard, got %v", m.screen)
	}
	if m.wizardReturnTo != screenSettings {
		t.Fatalf("expected wizardReturnTo = screenSettings, got %v", m.wizardReturnTo)
	}
}

func TestLibraryBackFromAlbumLevelStillShowsBareTracks(t *testing.T) {
	dir := t.TempDir()
	artistDir := filepath.Join(dir, "Aimer")
	if err := os.MkdirAll(filepath.Join(artistDir, "Album"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artistDir, "Bare Track.opus"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	// Simulate having descended into the Album, then backing out of it.
	m.lib = libState{level: 2, artist: "Aimer", album: "Album", entries: []string{}}

	next, _ := m.libraryBack()
	m = next.(model)
	if m.lib.level != 1 {
		t.Fatalf("expected level 1, got %d", m.lib.level)
	}
	want := []string{"Album", "Bare Track.opus"}
	got := m.lib.visible()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("entries after backing out = %v, want %v (bare track must still be visible)", got, want)
	}
}
