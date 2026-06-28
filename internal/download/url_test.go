package download

import (
	"errors"
	"strings"
	"testing"
)

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

func TestValidateInput(t *testing.T) {
	cases := []struct {
		input string
		ok    bool
	}{
		{"Imagine Dragons Believer", true},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://artist.bandcamp.com/album/foo", true},
		{"http://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://evil.example/video", false},
		{"ftp://www.youtube.com/watch?v=x", false},
		{"", false},
		{"track\x00name", false},
	}
	for _, c := range cases {
		err := ValidateInput(c.input)
		if c.ok && err != nil {
			t.Errorf("ValidateInput(%q) = %v, want nil", c.input, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ValidateInput(%q) = nil, want error", c.input)
		}
	}
}

func TestValidateInputDisallowedHost(t *testing.T) {
	err := ValidateInput("https://evil.example/video")
	if !errors.Is(err, ErrDisallowedHost) {
		t.Fatalf("ValidateInput() = %v, want ErrDisallowedHost", err)
	}
}

func TestStrategyArgsUseOptionTerminator(t *testing.T) {
	d := New("opus")
	target := "ytsearch1:test query"
	for _, s := range d.strategies() {
		args := appendTarget(s.args(d.Format, "out.%(ext)s"), target)
		if len(args) < 2 || args[len(args)-2] != "--" || args[len(args)-1] != target {
			t.Errorf("strategy %q args missing -- terminator: %v", s.name, args)
		}
		if strings.Contains(strings.Join(args[:len(args)-2], " "), target) {
			t.Errorf("strategy %q embeds target before terminator", s.name)
		}
	}
}

func TestPlaylistArgsUseOptionTerminator(t *testing.T) {
	args := []string{"-f", "bestaudio", "-o", "out.%(ext)s"}
	args = appendTarget(args, "https://www.youtube.com/playlist?list=PL123")
	if args[len(args)-2] != "--" {
		t.Fatalf("playlist args missing -- terminator: %v", args)
	}
}
