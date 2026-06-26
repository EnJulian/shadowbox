package download

import "testing"

func TestURLClassifiers(t *testing.T) {
	cases := []struct {
		url        string
		isURL      bool
		isYT       bool
		isPlaylist bool
		isBandcamp bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true, true, false, false},
		{"https://youtu.be/dQw4w9WgXcQ", true, true, false, false},
		{"https://www.youtube.com/playlist?list=PL123", true, true, true, false},
		{"https://music.youtube.com/watch?v=x&list=RDABC", true, true, true, false},
		{"https://artist.bandcamp.com/album/foo", true, false, false, true},
		{"Imagine Dragons Believer", false, false, false, false},
		{"  https://youtu.be/abc  ", true, true, false, false},
	}
	for _, c := range cases {
		if got := IsURL(c.url); got != c.isURL {
			t.Errorf("IsURL(%q) = %v, want %v", c.url, got, c.isURL)
		}
		if got := IsYouTube(c.url); got != c.isYT {
			t.Errorf("IsYouTube(%q) = %v, want %v", c.url, got, c.isYT)
		}
		if got := IsYouTubePlaylist(c.url); got != c.isPlaylist {
			t.Errorf("IsYouTubePlaylist(%q) = %v, want %v", c.url, got, c.isPlaylist)
		}
		if got := IsBandcamp(c.url); got != c.isBandcamp {
			t.Errorf("IsBandcamp(%q) = %v, want %v", c.url, got, c.isBandcamp)
		}
	}
}

func TestNewestAudioFilePrefersFormat(t *testing.T) {
	dir := t.TempDir()
	mk := func(name string) {
		writeFile(t, dir, name)
	}
	mk("a.webm")
	mk("b.opus")
	got := newestAudioFile(dir, "opus")
	if got == "" {
		t.Fatal("expected to find a file")
	}
	if filepathBase(got) != "b.opus" {
		t.Errorf("got %q, want b.opus (preferred format)", got)
	}
}
