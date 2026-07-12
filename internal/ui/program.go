package ui

import (
	"context"
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/progress"
	"github.com/EnJulian/shadowbox/internal/ui/overlay"
	"github.com/EnJulian/shadowbox/internal/ui/playback"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/workspace"
)

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayPicker
	overlayHelp
	overlayResult
	overlayTheme
)

// promptOutgoing/promptResult wire app.SelectFunc through to the picker overlay.
type promptOutgoing struct {
	req  app.PromptRequest
	resp chan promptResult
}

type promptResult struct {
	idx int
	err error
}

type promptRequestMsg struct{ out promptOutgoing }

// taskDoneMsg is emitted when a background pipeline operation completes.
type taskDoneMsg struct {
	summary string
	err     error
}

type model struct {
	cfg   *config.Config
	app   *app.App
	theme style.Theme
	st    style.Styles

	width, height int

	pane          shell.Pane
	activeSection workspace.Section
	workspaces    map[workspace.Section]workspace.Workspace
	downloads     *workspace.Downloads

	spinner  spinner.Model
	playback playback.State

	taskCancel  context.CancelFunc
	progressCh  chan progress.Update
	promptReqCh chan promptOutgoing

	ov          overlayKind
	confirmQuit bool
	picker      overlay.Picker
	pending     *promptOutgoing
	themeOv     overlay.Theme
	result      overlay.Result
}

func historyPath(cfg *config.Config) string {
	dir, err := config.Dir()
	if err != nil {
		return ""
	}
	return dir + "/search_history"
}

func (m *model) buildWorkspaces() {
	downloads := workspace.NewDownloads(m.st)
	m.downloads = downloads
	m.workspaces = map[workspace.Section]workspace.Workspace{
		workspace.SectionSearch:    workspace.NewSearch(m.app, m.cfg, m.st, historyPath(m.cfg)),
		workspace.SectionURL:       workspace.NewURL(m.app, m.st),
		workspace.SectionPlaylist:  workspace.NewPlaylist(m.app, m.st),
		workspace.SectionLibrary:   workspace.NewLibrary(m.cfg, m.st),
		workspace.SectionDownloads: downloads,
		workspace.SectionEnhance:   workspace.NewEnhance(m.app, m.st),
		workspace.SectionLog:       workspace.NewLog(m.st),
		workspace.SectionSettings:  workspace.NewSettings(m.cfg, m.st),
	}
}

// runProgram builds and runs the Bubble Tea program.
func runProgram(cfg *config.Config) error {
	capture := applog.DownloadLogWriter()
	_ = applog.LoadDownloadLog()
	applog.SetWriters(io.MultiWriter(io.Discard, capture), io.MultiWriter(io.Discard, capture))
	applog.SetVerbose(true)
	defer applog.Reset()

	theme := style.ThemeByName(cfg.Theme)
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := &model{
		cfg:           cfg,
		app:           app.New(cfg),
		theme:         theme,
		st:            style.NewStyles(theme),
		spinner:       sp,
		activeSection: workspace.SectionSearch,
	}
	m.buildWorkspaces()
	m.workspaces[workspace.SectionSearch] = m.workspaces[workspace.SectionSearch].Activate()

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m *model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.downloads.SetSpinnerFrame(m.spinner.View())
		return m, cmd

	case progress.Update:
		m.downloads.Update(msg)
		return m, m.taskListenCmd()

	case promptRequestMsg:
		m.ov = overlayPicker
		m.picker = overlay.Picker{Title: msg.out.req.Title, Options: msg.out.req.Options}
		m.pending = &msg.out
		return m, m.taskListenCmd()

	case taskDoneMsg:
		m.ov = overlayResult
		m.result = overlay.Result{Summary: msg.summary, Err: msg.err}
		m.downloads.Finish(msg.summary, msg.err)
		m.taskCancel = nil
		m.promptReqCh = nil
		return m, nil

	case workspace.StartTaskMsg:
		return m, m.startTask(msg.Summary, msg.Run)

	case workspace.SwitchSectionMsg:
		m.switchSection(msg.Section)
		return m, nil

	case workspace.CancelTaskMsg:
		if m.taskCancel != nil {
			m.taskCancel()
		}
		return m, nil

	case workspace.SettingsThemeRequestMsg:
		m.ov = overlayTheme
		m.themeOv = overlay.Theme{Cursor: themeIndex(m.cfg.Theme)}
		return m, nil

	case workspace.SettingsChangedMsg:
		m.applyTheme(m.cfg.Theme)
		m.app = rebuildApp(m.cfg)
		m.buildWorkspaces()
		return m, nil

	case shell.FocusNavMsg:
		m.pane = shell.PaneNav
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func themeIndex(name string) int {
	for i, n := range style.ThemeOrder {
		if n == name {
			return i
		}
	}
	return 0
}

func (m *model) applyTheme(name string) {
	m.theme = style.ThemeByName(name)
	m.st = style.NewStyles(m.theme)
}

func (m *model) switchSection(s workspace.Section) {
	m.activeSection = s
	m.pane = shell.PaneContent
	m.workspaces[s] = m.workspaces[s].Activate()
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// A pending quit confirmation is modal: it consumes the very next
	// keypress as a yes/no answer rather than letting it flow through to
	// overlays, text input, or normal shortcuts.
	if m.confirmQuit {
		m.confirmQuit = false
		if msg.String() == "q" {
			if m.taskCancel != nil {
				m.taskCancel()
			}
			return m, tea.Quit
		}
		return m, nil
	}

	if m.ov != overlayNone {
		return m.handleOverlayKey(msg)
	}

	// Tab always cycles panes, even while a workspace has a live text
	// cursor — it's never a printable rune a user would type into a text
	// field, so (unlike q/?// /digits below) it must not be gated behind
	// the TextFocused check. Handled here, before that check, so it keeps
	// working from every state: Nav, or any Content workspace regardless
	// of TextFocused.
	switch msg.String() {
	case "tab", "shift+tab":
		m.pane = m.pane.Toggle()
		if m.pane == shell.PaneContent {
			m.workspaces[m.activeSection] = m.workspaces[m.activeSection].Activate()
		}
		return m, nil
	}

	// If the Content pane's active workspace currently has a live text
	// cursor (query input, URL/Playlist/Enhance field, Settings inline edit,
	// Library's type-ahead filter), single-character global shortcuts below
	// must not steal the keystroke — e.g. typing "Queen" into Search must
	// not quit the app. Skip straight to workspace dispatch in that case.
	if m.pane == shell.PaneContent {
		if tf, ok := m.workspaces[m.activeSection].(workspace.TextFocused); ok && tf.TextFocused() {
			return m.handleContentKey(msg)
		}
	}

	switch msg.String() {
	case "q":
		if m.taskCancel != nil {
			m.confirmQuit = true
			return m, nil
		}
		return m, tea.Quit
	case "?":
		m.ov = overlayHelp
		return m, nil
	case "/":
		m.switchSection(workspace.SectionSearch)
		return m, nil
	}

	if n := sectionForDigit(msg.String()); n >= 0 {
		m.switchSection(workspace.Section(n))
		return m, nil
	}

	if m.pane == shell.PaneNav {
		return m.handleNavKey(msg)
	}
	return m.handleContentKey(msg)
}

func sectionForDigit(s string) int {
	if len(s) == 1 && s[0] >= '1' && s[0] <= '8' {
		return int(s[0] - '1')
	}
	return -1
}

func (m *model) handleNavKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.activeSection > 0 {
			m.activeSection--
		}
	case "down", "j":
		if int(m.activeSection) < len(workspace.Order)-1 {
			m.activeSection++
		}
	case "right", "l", "enter":
		m.switchSection(m.activeSection)
	}
	return m, nil
}

func (m *model) handleContentKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ws, cmd := m.workspaces[m.activeSection].Update(msg)
	m.workspaces[m.activeSection] = ws
	return m, cmd
}

func (m *model) handleOverlayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.ov {
	case overlayPicker:
		return m.handlePickerKey(msg)
	case overlayHelp:
		if msg.String() == "?" || msg.String() == "esc" {
			m.ov = overlayNone
		}
	case overlayResult:
		m.ov = overlayNone
	case overlayTheme:
		return m.handleThemeKey(msg)
	}
	return m, nil
}

func (m *model) handlePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pending == nil {
		m.ov = overlayNone
		return m, nil
	}
	switch msg.String() {
	case "esc":
		m.pending.resp <- promptResult{idx: -1, err: app.ErrSelectionCancelled}
		m.pending = nil
		m.ov = overlayNone
		if m.taskCancel != nil {
			m.taskCancel()
		}
	case "up", "k":
		m.picker.MoveUp()
	case "down", "j":
		m.picker.MoveDown()
	case "enter":
		idx := m.picker.Cursor
		pending := m.pending
		m.pending = nil
		m.ov = overlayNone
		pending.resp <- promptResult{idx: idx, err: nil}
		return m, m.taskListenCmd()
	}
	return m, nil
}

func (m *model) handleThemeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.applyTheme(m.cfg.Theme) // revert live preview
		m.ov = overlayNone
	case "up", "k":
		m.themeOv.MoveUp()
		m.applyTheme(m.themeOv.Selected())
	case "down", "j":
		m.themeOv.MoveDown()
		m.applyTheme(m.themeOv.Selected())
	case "enter":
		m.cfg.Theme = m.themeOv.Selected()
		_ = config.Save(m.cfg)
		m.applyTheme(m.cfg.Theme)
		m.ov = overlayNone
	}
	return m, nil
}

// startTask transitions Downloads to the running state and runs fn in the background.
func (m *model) startTask(summary string, fn func(ctx context.Context, opts app.Options) error) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.taskCancel = cancel
	m.downloads.SetRunning("Initializing")
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
		opts := app.Options{Progress: report, Select: makeSelectFunc(promptCh)}
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

func (m *model) taskListenCmd() tea.Cmd {
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

func waitForProgress(ch chan progress.Update) tea.Cmd {
	return func() tea.Msg {
		u, ok := <-ch
		if !ok {
			return nil
		}
		return u
	}
}

func waitForPrompt(reqCh chan promptOutgoing) tea.Cmd {
	return func() tea.Msg {
		out, ok := <-reqCh
		if !ok {
			return nil
		}
		return promptRequestMsg{out: out}
	}
}

func makeSelectFunc(reqCh chan promptOutgoing) app.SelectFunc {
	return func(ctx context.Context, req app.PromptRequest) (int, error) {
		respCh := make(chan promptResult, 1)
		out := promptOutgoing{req: req, resp: respCh}
		select {
		case reqCh <- out:
		case <-ctx.Done():
			return -1, ctx.Err()
		}
		select {
		case r := <-respCh:
			return r.idx, r.err
		case <-ctx.Done():
			return -1, ctx.Err()
		}
	}
}

func (m *model) navBody() string {
	var b string
	for i, item := range workspace.Order {
		cursor := "  "
		label := m.st.Item.Render(item.Label)
		if workspace.Section(i) == m.activeSection {
			label = m.st.Selected.Render(item.Label)
			if m.pane == shell.PaneNav {
				cursor = m.st.Accent.Render("> ")
			} else {
				cursor = "  "
			}
		}
		b += cursor + label + "\n"
	}
	return b
}

func (m *model) statusBody() string {
	if m.confirmQuit {
		return m.st.Help.Render("Press q again to quit and cancel the download, or any other key to stay")
	}
	if m.ov != overlayNone {
		return m.st.Help.Render("esc: close")
	}
	if m.pane == shell.PaneNav {
		return m.st.Help.Render(workspace.Order[m.activeSection].Label + " │ up/down: navigate │ enter/→: open │ tab: content │ ?: help │ q: quit")
	}
	return m.st.Help.Render(workspace.Order[m.activeSection].Label + " │ tab: nav │ ←: back │ ?: help │ q: quit")
}

func (m *model) View() string {
	layout := shell.Compute(m.width, m.height)
	content := m.workspaces[m.activeSection].View(layout.ContentWidth-2, layout.ContentHeight-2)

	if m.ov != overlayNone {
		content = m.overlayBody(layout)
	}

	return shell.Render(m.st, m.theme, layout, m.pane == shell.PaneNav, m.navBody(), content, m.playbarBody(), m.statusBody())
}

// playbarBody renders the reserved playback bar row. playback.State is a
// Phase 1 stub (Active is always false) so this is currently always empty;
// shell.Layout.PlaybarHeight is also always 0 this phase, so the row never
// actually occupies space regardless. Real rendering ships in a later phase.
func (m *model) playbarBody() string {
	if !m.playback.Active {
		return ""
	}
	return ""
}

func (m *model) overlayBody(layout shell.Layout) string {
	switch m.ov {
	case overlayPicker:
		return m.picker.View(m.st)
	case overlayHelp:
		return overlay.Help{}.View(m.st)
	case overlayResult:
		return m.result.View(m.st)
	case overlayTheme:
		return m.themeOv.View(m.st)
	}
	return ""
}
