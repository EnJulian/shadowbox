// internal/ui/workspace/search.go
package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/ui/shell"
	"github.com/EnJulian/shadowbox/internal/ui/style"
	"github.com/EnJulian/shadowbox/internal/ui/suggest"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type searchFocus int

const (
	searchFocusInput searchFocus = iota
	searchFocusResults
)

// searchResultsMsg carries browsable results back from an async SearchTracks call.
type searchResultsMsg struct {
	results []download.SearchResult
	err     error
}

// Search is the workspace.Workspace for the Search nav section: a query
// input with autocomplete, then a browsable results list.
type Search struct {
	app         *app.App
	cfg         *config.Config
	st          style.Styles
	historyPath string
	history     *suggest.History

	input   textinput.Model
	focus   searchFocus
	results []download.SearchResult
	cursor  int
	loading bool
	errMsg  string
}

// NewSearch builds the Search workspace.
func NewSearch(a *app.App, cfg *config.Config, st style.Styles, historyPath string) *Search {
	ti := textinput.New()
	ti.CharLimit = 256
	ti.Width = 60
	ti.Placeholder = "Enter title by artist (e.g. High Speed Chasing by BØRNS)"

	history, _ := suggest.LoadHistory(historyPath)

	return &Search{app: a, cfg: cfg, st: st, historyPath: historyPath, history: history, input: ti}
}

// Activate resets focus to the query input.
func (s *Search) Activate() Workspace {
	s.focus = searchFocusInput
	s.input.Focus()
	return s
}

func (s *Search) suggestions() []string {
	q := strings.TrimSpace(s.input.Value())
	if q == "" {
		return nil
	}
	out := s.history.Matches(q, 5)
	for _, m := range suggest.LibraryMatches(s.cfg.MusicDirectory, q, 3) {
		out = append(out, m+" (in library)")
	}
	return out
}

// Update handles async search results and key input for both the query
// input and the results list, depending on current focus.
func (s *Search) Update(msg tea.Msg) (Workspace, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultsMsg:
		s.loading = false
		s.results = msg.results
		s.cursor = 0
		if msg.err != nil {
			s.errMsg = msg.err.Error()
		} else {
			s.errMsg = ""
		}
		return s, nil

	case tea.KeyMsg:
		return s.handleKey(msg)
	}
	return s, nil
}

func (s *Search) handleKey(msg tea.KeyMsg) (Workspace, tea.Cmd) {
	if s.focus == searchFocusInput {
		switch msg.String() {
		case "enter":
			query := strings.TrimSpace(s.input.Value())
			if query == "" {
				return s, nil
			}
			s.history.Add(query)
			_ = s.history.Save(s.historyPath)
			s.loading = true
			a := s.app
			return s, func() tea.Msg {
				results, err := a.SearchTracks(context.Background(), query, 10)
				return searchResultsMsg{results: results, err: err}
			}
		case "down":
			if len(s.results) > 0 {
				s.focus = searchFocusResults
				s.cursor = 0
				s.input.Blur()
			}
			return s, nil
		}
		var cmd tea.Cmd
		s.input, cmd = s.input.Update(msg)
		return s, cmd
	}

	// focus == searchFocusResults
	switch msg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		} else {
			s.focus = searchFocusInput
			s.input.Focus()
		}
	case "down", "j":
		if s.cursor < len(s.results)-1 {
			s.cursor++
		}
	case "left", "h":
		s.focus = searchFocusInput
		s.input.Focus()
		return s, shell.RequestNavFocus()
	case "enter":
		if len(s.results) == 0 {
			return s, nil
		}
		selected := s.results[s.cursor]
		return s, tea.Batch(
			StartTask("Download", func(ctx context.Context, opts app.Options) error {
				return s.app.Run(ctx, selected.URL, opts)
			}),
			SwitchSection(SectionDownloads),
		)
	}
	return s, nil
}

// View renders the query input, autocomplete suggestions, and any results.
func (s *Search) View(width, height int) string {
	var b strings.Builder
	b.WriteString(s.st.Subtitle.Render("Query:") + "\n")
	b.WriteString(s.input.View() + "\n")

	if s.loading {
		b.WriteString("\n" + s.st.Item.Render("searching…"))
		return b.String()
	}

	if suggestions := s.suggestions(); s.focus == searchFocusInput && len(suggestions) > 0 {
		b.WriteString("\n" + s.st.Help.Render("Suggestions") + "\n")
		for _, sug := range suggestions {
			b.WriteString("  " + s.st.Item.Render(sug) + "\n")
		}
	}

	if s.errMsg != "" {
		b.WriteString("\n" + s.st.Danger.Render(s.errMsg) + "\n")
	}

	if len(s.results) > 0 {
		b.WriteString("\n" + s.st.Subtitle.Render("Results") + "\n")
		for i, r := range s.results {
			cursor := "  "
			label := s.st.Item.Render(r.Title)
			if s.focus == searchFocusResults && i == s.cursor {
				cursor = s.st.Accent.Render("> ")
				label = s.st.Selected.Render(r.Title)
			}
			detail := fmt.Sprintf("%s · %s", r.Uploader, r.Duration)
			b.WriteString(cursor + label + "\n")
			b.WriteString("      " + s.st.Help.Render(detail) + "\n")
		}
	}
	return b.String()
}
