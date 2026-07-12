// internal/ui/workspace/log_test.go
package workspace

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/ui/style"
)

func TestLogActivateWithNoLogsShowsHint(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	l := &Log{st: st}
	view := l.View(80, 20)
	if !strings.Contains(view, "no download logs yet") {
		t.Fatalf("expected empty-log hint, got %q", view)
	}
}

func TestLogScrollClampsToBounds(t *testing.T) {
	st := style.NewStyles(style.ThemeByName("hacker"))
	l := NewLog(st)
	l.lines = []string{"a", "b", "c", "d", "e"}
	l.scroll = 0

	ws, _ := l.Update(key("up")) // already at top, should not go negative
	if ws.(*Log).scroll != 0 {
		t.Fatalf("scroll = %d, want 0", ws.(*Log).scroll)
	}

	ws, _ = ws.Update(key("end"))
	if ws.(*Log).scroll < 0 {
		t.Fatalf("scroll went negative: %d", ws.(*Log).scroll)
	}
}
