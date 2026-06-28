package tag

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReadKHInsiderEmbeddedTags(t *testing.T) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		t.Skip("yt-dlp not available")
	}
	dir := t.TempDir()
	mp3 := "https://jetta.vgmtreasurechest.com/soundtracks/kirby-and-the-forgotten-land-the-complete-soundtrack-2024/jxlmejhj/1-01.%20Ready%20to%20Go%21.mp3"
	out := filepath.Join(dir, "out.%(ext)s")
	cmd := exec.Command("yt-dlp", "-x", "--audio-format", "opus", "--embed-metadata", "-o", out, mp3)
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("download failed: %v %s", err, outBytes)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "out.*"))
	if len(files) == 0 {
		t.Fatal("no output file")
	}
	m, err := Read(files[0])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("metadata: %+v", m)
	if m.Artist != "Yuuta Ogasawara" {
		t.Errorf("artist = %q", m.Artist)
	}
	if m.Title != "Ready to Go!" {
		t.Errorf("title = %q", m.Title)
	}
	if m.TrackNumber == "" {
		t.Errorf("track number empty")
	}
	if m.DiscNumber == "" {
		t.Errorf("disc number empty")
	}
}
