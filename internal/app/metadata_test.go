package app

import (
	"testing"

	"github.com/EnJulian/shadowbox/internal/apis/itunes"
	"github.com/EnJulian/shadowbox/internal/apis/musicbrainz"
)

func TestFilterITunesCandidatesByArtist(t *testing.T) {
	candidates := []itunes.TrackCandidate{
		{Title: "Isometric", Artist: "Zane Alexander"},
		{Title: "Other Song", Artist: "Someone Else"},
	}
	filtered := filterITunesCandidates(candidates, "Zane Alexander")
	if len(filtered) != 1 {
		t.Fatalf("len = %d, want 1", len(filtered))
	}
	if filtered[0].Artist != "Zane Alexander" {
		t.Errorf("artist = %q", filtered[0].Artist)
	}
}

func TestFilterMBCandidatesByArtist(t *testing.T) {
	candidates := []musicbrainz.RecordingCandidate{
		{Title: "Isometric", Artist: "Zane Alexander"},
		{Title: "Recording the Recording", Artist: "陳奕迅"},
	}
	filtered := filterMBCandidates(candidates, "Zane Alexander")
	if len(filtered) != 1 {
		t.Fatalf("len = %d, want 1", len(filtered))
	}
	if filtered[0].Artist != "Zane Alexander" {
		t.Errorf("artist = %q", filtered[0].Artist)
	}
}

func TestFilterMBCandidatesByArtistNoMatch(t *testing.T) {
	candidates := []musicbrainz.RecordingCandidate{
		{Title: "Recording the Recording", Artist: "陳奕迅"},
	}
	filtered := filterMBCandidates(candidates, "Zane Alexander")
	if len(filtered) != 0 {
		t.Fatalf("len = %d, want 0", len(filtered))
	}
}

func TestFilterITunesCandidatesUnknownPassthrough(t *testing.T) {
	candidates := []itunes.TrackCandidate{
		{Title: "A", Artist: "X"},
		{Title: "B", Artist: "Y"},
	}
	filtered := filterITunesCandidates(candidates, "Unknown")
	if len(filtered) != 2 {
		t.Fatalf("len = %d, want 2", len(filtered))
	}
}
