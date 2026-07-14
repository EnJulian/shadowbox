package ui

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/player"
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
	screenSetupWizard
	screenHelp
	screenEnhancePicker
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

	// enhance folder picker
	enhancePicker enhancePickerState

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

	// playback
	playback  player.State
	player    *player.Player
	playerErr string

	// setup wizard
	wizardReturnTo screen
	wizardCursor   int
	wizardItems    []wizardItem

	// help
	helpReturnTo screen
	helpScroll   int
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

	m := initialModel(cfg, theme, ti, sp)

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if fm, ok := final.(model); ok && fm.player != nil {
		_ = fm.player.Close()
	}
	return err
}

// initialModel builds the model runProgram starts with, triggering the
// setup wizard automatically the very first time Shadowbox ever runs (no
// config file on disk yet). If the check itself errors for a reason other
// than "file does not exist," this fails safe and skips the auto-trigger —
// the wizard is still reachable manually from Settings either way.
func initialModel(cfg *config.Config, theme Theme, ti textinput.Model, sp spinner.Model) model {
	m := model{
		cfg:     cfg,
		app:     app.New(cfg),
		theme:   theme,
		st:      newStyles(theme),
		input:   ti,
		spinner: sp,
		screen:  screenMenu,
	}
	if path, err := config.Path(); err == nil {
		if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
			next, _ := m.openWizard(screenMenu)
			m = next.(model)
		}
	}
	return m
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
		if m.player != nil {
			m.playback = m.player.State()
		}
		return m, cmd

	case startPlaybackMsg:
		if m.player == nil {
			if !player.Available() {
				m.playerErr = "mpv not found — install it to enable playback"
				return m, nil
			}
			p, err := player.New()
			if err != nil {
				m.playerErr = "failed to start mpv: " + err.Error()
				return m, nil
			}
			m.player = p
		}
		if err := m.player.Load(msg.tracks, msg.index); err != nil {
			m.playerErr = "failed to play: " + err.Error()
		}
		return m, nil

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

	case wizardInstallDoneMsg:
		m = m.handleWizardInstallDone(msg)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// screenCapturesText reports whether the given screen consumes free-text
// keystrokes (a live filter or text field), and must not have global
// playback keys (space/n/p/s/arrows) stolen out from under it.
func screenCapturesText(s screen) bool {
	switch s {
	case screenLibrary, screenInput, screenSettingEdit, screenEnhancePicker:
		return true
	}
	return false
}

// handlePlaybackKey handles the global playback keys. handled is false for
// any other key, so the caller falls through to the screen's own handling.
func (m model) handlePlaybackKey(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.player == nil {
		switch msg.String() {
		case " ", "n", "p", "s":
			return m, nil, true // consume the key, but there's nothing to control yet
		}
		if msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight {
			return m, nil, true
		}
		return m, nil, false
	}

	switch msg.String() {
	case " ":
		_ = m.player.TogglePause()
		return m, nil, true
	case "n":
		_ = m.player.Next()
		return m, nil, true
	case "p":
		_ = m.player.Prev()
		return m, nil, true
	case "s":
		_ = m.player.Stop()
		return m, nil, true
	}
	switch msg.Type {
	case tea.KeyLeft:
		_ = m.player.SeekBy(-10 * time.Second)
		return m, nil, true
	case tea.KeyRight:
		_ = m.player.SeekBy(10 * time.Second)
		return m, nil, true
	}
	return m, nil, false
}

// handleKey dispatches key events to the active screen.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.playerErr = ""

	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.screen == screenMenu && msg.String() == "q" {
		return m, tea.Quit
	}

	if !screenCapturesText(m.screen) {
		if next, cmd, handled := m.handlePlaybackKey(msg); handled {
			return next, cmd
		}
		if msg.String() == "?" && m.screen != screenHelp && m.screen != screenRunning && m.screen != screenResult {
			return m.openHelp(m.screen)
		}
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
	case screenEnhancePicker:
		return m.updateEnhancePicker(msg)
	case screenDownloadLog:
		return m.updateDownloadLog(msg)
	case screenPicker:
		return m.updatePicker(msg)
	case screenSetupWizard:
		return m.updateWizard(msg)
	case screenHelp:
		return m.updateHelp(msg)
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
	case screenEnhancePicker:
		return m.viewEnhancePicker()
	case screenDownloadLog:
		return m.viewDownloadLog()
	case screenRunning:
		return m.viewRunning()
	case screenPicker:
		return m.viewPicker()
	case screenSetupWizard:
		return m.viewWizard()
	case screenHelp:
		return m.viewHelp()
	case screenResult:
		return m.viewResult()
	}
	return ""
}
