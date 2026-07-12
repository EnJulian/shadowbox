package download

import "testing"

func TestDependenciesIncludesMpvAsOptional(t *testing.T) {
	for _, d := range Dependencies() {
		if d.Name == "mpv" {
			if d.Required {
				t.Error(`mpv dependency has Required = true, want false (playback-only, not needed to run Shadowbox)`)
			}
			return
		}
	}
	t.Fatal("Dependencies() does not include mpv")
}
