package suggest

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestHistoryAddDedupesAndPrepends(t *testing.T) {
	h := &History{}
	h.Add("nujabes feather")
	h.Add("kanye")
	h.Add("nujabes feather") // re-adding moves it to the front, no duplicate

	got := h.Matches("", 10)
	want := []string{"nujabes feather", "kanye"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Matches = %v, want %v", got, want)
	}
}

func TestHistoryMatchesFiltersByPrefixCaseInsensitive(t *testing.T) {
	h := &History{}
	h.Add("Nujabes Feather")
	h.Add("Kanye Stronger")

	got := h.Matches("nu", 10)
	if len(got) != 1 || got[0] != "Nujabes Feather" {
		t.Fatalf("Matches(\"nu\") = %v, want [Nujabes Feather]", got)
	}
}

func TestHistorySaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "search_history")

	h := &History{}
	h.Add("first query")
	h.Add("second query")
	if err := h.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}
	got := loaded.Matches("", 10)
	want := []string{"second query", "first query"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded Matches = %v, want %v", got, want)
	}
}

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	h, err := LoadHistory(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("LoadHistory() error = %v, want nil for a missing file", err)
	}
	if len(h.Matches("", 10)) != 0 {
		t.Fatalf("expected empty history, got %v", h.Matches("", 10))
	}
}
