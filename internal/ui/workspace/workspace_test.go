package workspace

import (
	"context"
	"testing"

	"github.com/EnJulian/shadowbox/internal/app"
)

func TestOrderCoversEverySection(t *testing.T) {
	seen := map[Section]bool{}
	for _, m := range Order {
		seen[m.Section] = true
		if m.Label == "" {
			t.Errorf("section %v has empty label", m.Section)
		}
	}
	want := []Section{SectionSearch, SectionURL, SectionPlaylist, SectionLibrary, SectionDownloads, SectionEnhance, SectionLog, SectionSettings}
	if len(seen) != len(want) {
		t.Fatalf("Order has %d unique sections, want %d", len(seen), len(want))
	}
	for _, s := range want {
		if !seen[s] {
			t.Errorf("Order missing section %v", s)
		}
	}
}

func TestStartTaskCmdCarriesSummary(t *testing.T) {
	called := false
	cmd := StartTask("Download", func(ctx context.Context, opts app.Options) error {
		called = true
		return nil
	})
	msg := cmd()
	stMsg, ok := msg.(StartTaskMsg)
	if !ok {
		t.Fatalf("StartTask() cmd returned %T, want StartTaskMsg", msg)
	}
	if stMsg.Summary != "Download" {
		t.Fatalf("Summary = %q, want %q", stMsg.Summary, "Download")
	}
	_ = stMsg.Run(context.Background(), app.Options{})
	if !called {
		t.Fatal("StartTaskMsg.Run was not the closure passed to StartTask")
	}
}

func TestSwitchSectionCmd(t *testing.T) {
	msg := SwitchSection(SectionDownloads)()
	ss, ok := msg.(SwitchSectionMsg)
	if !ok || ss.Section != SectionDownloads {
		t.Fatalf("SwitchSection cmd = %#v, want SwitchSectionMsg{SectionDownloads}", msg)
	}
}

func TestCancelTaskCmd(t *testing.T) {
	if _, ok := CancelTask()().(CancelTaskMsg); !ok {
		t.Fatal("CancelTask() cmd did not return CancelTaskMsg")
	}
}
