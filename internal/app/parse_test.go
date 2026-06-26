package app

import "testing"

func TestParseFromQuery(t *testing.T) {
	cases := []struct {
		query, title, artist string
	}{
		{"Imagine Dragons - Believer", "Believer", "Imagine Dragons"},
		{"Adele - Hello (Official Video)", "Hello", "Adele"},
		{"Imagine Dragons Believer", "Believer", "Imagine Dragons"},
		{"Adele Hello", "Hello", "Adele"},
		{"Believer", "Believer", "Unknown"},
		{"Pink Floyd Time", "Time", "Pink Floyd"},
	}
	for _, c := range cases {
		title, artist := parseFromQuery(c.query)
		if title != c.title || artist != c.artist {
			t.Errorf("parseFromQuery(%q) = (%q, %q), want (%q, %q)", c.query, title, artist, c.title, c.artist)
		}
	}
}

func TestParseFromFilename(t *testing.T) {
	title, artist := parseFromFilename("/music/01 - Imagine Dragons - Believer.opus")
	if artist != "Imagine Dragons" || title != "Believer" {
		t.Errorf("got (%q, %q)", title, artist)
	}
	title2, artist2 := parseFromFilename("/music/Solo Song.mp3")
	if title2 != "Solo Song" || artist2 != "Unknown" {
		t.Errorf("got (%q, %q)", title2, artist2)
	}
}

func TestCleanTitle(t *testing.T) {
	if got := cleanTitle("Hello [Official Music Video]"); got != "Hello" {
		t.Errorf("cleanTitle = %q, want Hello", got)
	}
}
