package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/EnJulian/shadowbox/internal/player"
)

func TestRenderBannerWithPlaybackIdleMatchesPlainBanner(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")

	got := renderBannerWithPlayback(st, theme, player.State{})
	want := st.title.Render(banner)
	if got != want {
		t.Errorf("idle banner should render exactly like the plain banner, got a different string")
	}
}

func TestRenderBannerWithPlaybackShowsNowPlayingLine(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")
	state := player.State{Title: "Feather", Elapsed: 90 * time.Second, Duration: 222 * time.Second, Playing: true}

	got := renderBannerWithPlayback(st, theme, state)
	if !strings.Contains(got, "Feather") {
		t.Error("expected the now-playing line to contain the track title")
	}
	if !strings.Contains(got, "1:30") || !strings.Contains(got, "3:42") {
		t.Errorf("expected elapsed/total time as m:ss, got:\n%s", got)
	}
}

func TestRenderBannerWithPlaybackSweepsProportionally(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")

	early := renderBannerWithPlayback(st, theme, player.State{Title: "X", Elapsed: 1 * time.Second, Duration: 100 * time.Second, Playing: true})
	late := renderBannerWithPlayback(st, theme, player.State{Title: "X", Elapsed: 90 * time.Second, Duration: 100 * time.Second, Playing: true})

	if early == late {
		t.Error("banner rendering should differ between 1% and 90% progress")
	}
}

func TestRenderBannerWithPlaybackHandlesZeroDuration(t *testing.T) {
	st := newStyles(themeByName("hacker"))
	theme := themeByName("hacker")
	// Duration is briefly 0 right after Load(), before the first poll — must
	// not panic or divide by zero.
	got := renderBannerWithPlayback(st, theme, player.State{Title: "X", Playing: true})
	if !strings.Contains(got, "X") {
		t.Error("expected the title to still render with a zero duration")
	}
}
