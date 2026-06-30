package app

import "testing"

func TestParseFromQuery(t *testing.T) {
	cases := []struct {
		query, title, artist string
	}{
		{"Believer by Imagine Dragons", "Believer", "Imagine Dragons"},
		{"High Speed Chasing by BØRNS", "High Speed Chasing", "BØRNS"},
		{"believer BY imagine dragons", "believer", "imagine dragons"},
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
	if got := cleanTitle("Isometric (Lo-Fi Synthwave)"); got != "Isometric" {
		t.Errorf("cleanTitle = %q, want Isometric", got)
	}
}

func TestParseYouTubeTitle(t *testing.T) {
	title, artist, ok := parseYouTubeTitle("Zane Alexander - Isometric")
	if !ok || title != "Isometric" || artist != "Zane Alexander" {
		t.Errorf("parseYouTubeTitle = (%q, %q, %v)", title, artist, ok)
	}
	title2, _, ok2 := parseYouTubeTitle("Zane Alexander - Isometric (Lo-Fi Synthwave)")
	if !ok2 || title2 != "Isometric" {
		t.Errorf("parseYouTubeTitle with suffix = (%q, %v)", title2, ok2)
	}
}

func TestApplySelectedTrackMetaKeepsByFormatArtist(t *testing.T) {
	opts := Options{
		searchMeta:     true,
		searchTitle:    "Isometric",
		searchArtist:   "Zane Alexander",
		searchByFormat: true,
	}
	applySelectedTrackMeta(&opts, "Zane Alexander - Isometric (Lo-Fi Synthwave)")
	if opts.searchArtist != "Zane Alexander" {
		t.Errorf("artist = %q, want Zane Alexander", opts.searchArtist)
	}
	if opts.searchTitle != "Isometric" {
		t.Errorf("title = %q, want Isometric (unchanged from query)", opts.searchTitle)
	}
}

func TestApplySelectedTrackMetaFillsUnknownArtist(t *testing.T) {
	opts := Options{
		searchMeta:     true,
		searchTitle:    "Isometric",
		searchArtist:   "Unknown",
		searchByFormat: false,
	}
	applySelectedTrackMeta(&opts, "Zane Alexander - Isometric")
	if opts.searchArtist != "Zane Alexander" {
		t.Errorf("artist = %q, want Zane Alexander", opts.searchArtist)
	}
}
