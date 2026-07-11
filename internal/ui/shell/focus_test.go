package shell

import "testing"

func TestPaneToggle(t *testing.T) {
	if PaneNav.Toggle() != PaneContent {
		t.Fatalf("PaneNav.Toggle() = %v, want PaneContent", PaneNav.Toggle())
	}
	if PaneContent.Toggle() != PaneNav {
		t.Fatalf("PaneContent.Toggle() = %v, want PaneNav", PaneContent.Toggle())
	}
}

func TestRequestNavFocusReturnsFocusNavMsg(t *testing.T) {
	cmd := RequestNavFocus()
	msg := cmd()
	if _, ok := msg.(FocusNavMsg); !ok {
		t.Fatalf("RequestNavFocus() cmd returned %T, want FocusNavMsg", msg)
	}
}
