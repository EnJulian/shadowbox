// internal/ui/workspace/input.go
package workspace

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/suggest"
)

// Input is the shared workspace.Workspace behind URL, Playlist, and Enhance:
// a single text field, an optional clipboard hint, and a submit action that
// starts a background task.
type Input struct {
	app         *app.App
	st          style.Styles
	title       string
	summary     string
	clipboard   bool // offer a clipboard-URL hint on Activate
	input       textinput.Model
	clipHint    string
	submit      func(a *app.App, value string) func(ctx context.Context, opts app.Options) error
}

// NewURL builds the URL-download Input workspace.
func NewURL(a *app.App, st style.Styles) *Input {
	return newInput(a, st, "Enter a YouTube or Bandcamp URL", "Download", true,
		func(a *app.App, value string) func(context.Context, app.Options) error {
			return func(ctx context.Context, opts app.Options) error { return a.Run(ctx, value, opts) }
		})
}

// NewPlaylist builds the playlist-download Input workspace.
func NewPlaylist(a *app.App, st style.Styles) *Input {
	return newInput(a, st, "Enter a YouTube playlist URL", "Playlist download", true,
		func(a *app.App, value string) func(context.Context, app.Options) error {
			return func(ctx context.Context, opts app.Options) error { return a.RunPlaylist(ctx, value, opts) }
		})
}

// NewEnhance builds the directory-enhance Input workspace.
func NewEnhance(a *app.App, st style.Styles) *Input {
	return newInput(a, st, "Enter a directory of audio files to enhance", "Enhancement", false,
		func(a *app.App, value string) func(context.Context, app.Options) error {
			return func(ctx context.Context, opts app.Options) error {
				return a.EnhanceDir(ctx, value, true, []string{"opus", "mp3", "m4a", "flac"}, false, opts)
			}
		})
}

func newInput(a *app.App, st style.Styles, title, summary string, clipboard bool, submit func(*app.App, string) func(context.Context, app.Options) error) *Input {
	ti := textinput.New()
	ti.CharLimit = 512
	ti.Width = 60
	return &Input{app: a, st: st, title: title, summary: summary, clipboard: clipboard, input: ti, submit: submit}
}

func (in *Input) Activate() Workspace {
	in.input.SetValue("")
	in.input.Focus()
	in.clipHint = ""
	if in.clipboard {
		in.clipHint = suggest.ClipboardURL()
	}
	return in
}

func (in *Input) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return in, nil
	}
	switch keyMsg.String() {
	case "tab":
		if in.clipHint != "" {
			in.input.SetValue(in.clipHint)
			in.input.CursorEnd()
		}
		return in, nil
	case "enter":
		value := strings.TrimSpace(in.input.Value())
		if value == "" {
			return in, nil
		}
		run := in.submit(in.app, value)
		return in, tea.Batch(StartTask(in.summary, run), SwitchSection(SectionDownloads))
	}
	var cmd tea.Cmd
	in.input, cmd = in.input.Update(keyMsg)
	return in, cmd
}

func (in *Input) View(width, height int) string {
	var b strings.Builder
	b.WriteString(in.st.Subtitle.Render(in.title) + "\n\n")
	b.WriteString(in.input.View() + "\n")
	if in.clipHint != "" {
		b.WriteString("\n" + in.st.Help.Render("Paste from clipboard: "+in.clipHint+" — Tab to accept"))
	}
	return b.String()
}
