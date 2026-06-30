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
	"github.com/EnJulian/shadowbox/internal/progress"
)

type screen int

const (
	screenMenu screen = iota
	screenInput
	screenSettings
	screenSettingEdit
	screenThemePicker
	screenLibrary
	screenDownloadLog
	screenRunning
	screenPicker
	screenResult
)

// taskDoneMsg is emitted when a background pipeline operation completes.
type taskDoneMsg struct {
	summary string
	err     error
}

// progressMsg carries a pipeline stage update to the running screen.
type progressMsg progress.Update

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
	progress      progress.Update
	progressCh    chan progress.Update
	runningHeading string
	taskSummary   string
	taskCancel    context.CancelFunc
	promptReqCh   chan promptOutgoing

	// interactive picker
	picker pickerState

	// result
	result    string
	resultErr error

	// download log viewer
	logLines  []string
	logScroll int
}

var mainMenu = []string{
	"Search & Download",
	"Download from URL",
	"Download Playlist",
	"Enhance Existing Files",
	"Library",
	"Download Log",
	"Settings",
	"Exit",
}

// runProgram builds and runs the Bubble Tea program.
func runProgram(cfg *config.Config) error {
	capture := applog.DownloadLogWriter()
	_ = applog.LoadDownloadLog()
	applog.SetWriters(io.MultiWriter(io.Discard, capture), io.MultiWriter(io.Discard, capture))
	applog.SetVerbose(true)
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
		m.progress = progress.Update(msg)
		if m.progress.Heading != "" {
			m.runningHeading = m.progress.Heading
		}
		return m, m.taskListenCmd()

	case promptRequestMsg:
		m.screen = screenPicker
		m.picker = pickerState{
			title:   msg.out.req.Title,
			options: msg.out.req.Options,
			cursor:  0,
			pending: &msg.out,
		}
		return m, m.taskListenCmd()

	case taskDoneMsg:
		m.screen = screenResult
		m.result = msg.summary
		m.resultErr = msg.err
		m.progress = progress.Update{}
		m.taskCancel = nil
		m.promptReqCh = nil
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
	case screenDownloadLog:
		return m.updateDownloadLog(msg)
	case screenPicker:
		return m.updatePicker(msg)
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
func (m *model) startTask(summary string, fn func(ctx context.Context, opts app.Options) error) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.taskCancel = cancel
	m.taskSummary = summary
	m.runningHeading = "Initializing"
	m.screen = screenRunning
	m.progress = progress.Update{}
	applog.BeginDownloadSession("Initializing")

	ch := make(chan progress.Update, 32)
	m.progressCh = ch
	promptCh := make(chan promptOutgoing, 1)
	m.promptReqCh = promptCh

	report := func(u progress.Update) {
		select {
		case ch <- u:
		default:
		}
	}

	run := func() tea.Msg {
		opts := app.Options{
			Progress: report,
			Select:   makeSelectFunc(promptCh),
		}
		err := fn(ctx, opts)
		cancel()
		close(promptCh)
		close(ch)
		summaryMsg := summary + " complete"
		if err != nil {
			summaryMsg = summary + " failed"
		}
		return taskDoneMsg{summary: summaryMsg, err: err}
	}

	return tea.Batch(m.spinner.Tick, run, m.taskListenCmd())
}

func (m model) taskListenCmd() tea.Cmd {
	var cmds []tea.Cmd
	if m.progressCh != nil {
		cmds = append(cmds, waitForProgress(m.progressCh))
	}
	if m.promptReqCh != nil {
		cmds = append(cmds, waitForPrompt(m.promptReqCh))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// waitForProgress blocks on the next stage description from the channel and
// re-subscribes after each one. It stops (returns nil) when the channel closes.
func waitForProgress(ch chan progress.Update) tea.Cmd {
	return func() tea.Msg {
		u, ok := <-ch
		if !ok {
			return nil
		}
		return progressMsg(u)
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
	case screenDownloadLog:
		return m.viewDownloadLog()
	case screenRunning:
		return m.viewRunning()
	case screenPicker:
		return m.viewPicker()
	case screenResult:
		return m.viewResult()
	}
	return ""
}
