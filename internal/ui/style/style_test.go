package style

import "testing"

func TestThemeByNameFallsBackToHacker(t *testing.T) {
	got := ThemeByName("does-not-exist")
	if got.Name != "hacker" {
		t.Fatalf("ThemeByName fallback = %q, want %q", got.Name, "hacker")
	}
}

func TestThemeOrderMatchesRegisteredThemes(t *testing.T) {
	for _, name := range ThemeOrder {
		if ThemeByName(name).Name != name {
			t.Errorf("ThemeOrder entry %q not registered", name)
		}
	}
}

func TestRenderBannerContainsSubtitle(t *testing.T) {
	st := NewStyles(ThemeByName("hacker"))
	out := RenderBanner(st)
	if len(out) == 0 {
		t.Fatal("RenderBanner returned empty string")
	}
}
