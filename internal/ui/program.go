package ui

import (
	"context"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	applog "github.com/EnJulian/shadowbox/internal/log"
)

type screen int

const (
	screenMenu screen = iota
	screenInput
	screenSettings
	screenSettingEdit
	screenThemePicker
	screenLibrary
	screenRunning
	screenResult
)

// taskDoneMsg is emitted when a background pipeline operation completes.
type taskDoneMsg struct {
	summary string
	err     error
}

// progressMsg carries a human-readable pipeline stage to the running screen.
type progressMsg string

type model struct {
	cfg   *config.Config
	app   *app.App
	theme Theme
	st    styles

	width  int
	height int

	screen     screen
	menuCursor int

	input        textinput.Model
	inputContext string
	inputTitle   string

	spinner spinner.Model

	// settings
	settingsCursor int
	editKey        string

	// theme picker
	themeCursor int

	// library navigation
	lib libState

	// running progress
	progress   string
	progressCh chan string

	// result
	result    string
	resultErr error
}

var mainMenu = []string{
	"Search & Download",
	"Download from URL",
	"Download Playlist",
	"Enhance Existing Files",
	"Library",
	"Settings",
	"Exit",
}

// runProgram builds and runs the Bubble Tea program.
func runProgram(cfg *config.Config) error {
	// Bubble Tea owns the screen; silence direct log writes for the session.
	applog.SetWriters(io.Discard, io.Discard)
	defer applog.Reset()

	theme := themeByName(cfg.Theme)
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      newStyles(theme),
		input:   ti,
		spinner: sp,
		screen:  screenMenu,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progressMsg:
		m.progress = string(msg)
		return m, waitForProgress(m.progressCh)

	case taskDoneMsg:
		m.screen = screenResult
		m.result = msg.summary
		m.resultErr = msg.err
		m.progress = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey dispatches key events to the active screen.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit on the menu screen.
	if m.screen == screenMenu {
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	} else if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case screenMenu:
		return m.updateMenu(msg)
	case screenInput:
		return m.updateInput(msg)
	case screenSettings:
		return m.updateSettings(msg)
	case screenSettingEdit:
		return m.updateSettingEdit(msg)
	case screenThemePicker:
		return m.updateThemePicker(msg)
	case screenLibrary:
		return m.updateLibrary(msg)
	case screenResult:
		// Any key returns to the menu.
		m.screen = screenMenu
		return m, nil
	case screenRunning:
		return m, nil
	}
	return m, nil
}

// startTask transitions to the running screen and runs fn in the background.
// fn receives a report callback it can use to surface live stage descriptions
// on the running screen.
func (m *model) startTask(label string, fn func(ctx context.Context, report func(string)) error) tea.Cmd {
	m.screen = screenRunning
	m.result = label
	m.progress = ""

	ch := make(chan string, 32)
	m.progressCh = ch
	report := func(s string) {
		// Non-blocking: never let progress reporting stall the pipeline.
		select {
		case ch <- s:
		default:
		}
	}

	run := func() tea.Msg {
		err := fn(context.Background(), report)
		close(ch)
		summary := label + " complete"
		if err != nil {
			summary = label + " failed"
		}
		return taskDoneMsg{summary: summary, err: err}
	}

	return tea.Batch(m.spinner.Tick, run, waitForProgress(ch))
}

// waitForProgress blocks on the next stage description from the channel and
// re-subscribes after each one. It stops (returns nil) when the channel closes.
func waitForProgress(ch chan string) tea.Cmd {
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return nil
		}
		return progressMsg(s)
	}
}

func (m model) View() string {
	switch m.screen {
	case screenMenu:
		return m.viewMenu()
	case screenInput:
		return m.viewInput()
	case screenSettings:
		return m.viewSettings()
	case screenSettingEdit:
		return m.viewSettingEdit()
	case screenThemePicker:
		return m.viewThemePicker()
	case screenLibrary:
		return m.viewLibrary()
	case screenRunning:
		return m.viewRunning()
	case screenResult:
		return m.viewResult()
	}
	return ""
}
