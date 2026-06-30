package download

import (
	"testing"
)

const sampleSearchJSON = `{
  "entries": [
    {
      "id": "abc123",
      "title": "Believer",
      "uploader": "ImagineDragonsVEVO",
      "duration": 204.0,
      "webpage_url": "https://www.youtube.com/watch?v=abc123"
    },
    {
      "id": "def456",
      "title": "Believer (Live)",
      "uploader": "Imagine Dragons",
      "duration": 310.5,
      "url": "https://www.youtube.com/watch?v=def456"
    }
  ]
}`

func TestParseSearchJSON(t *testing.T) {
	results, err := parseSearchJSON([]byte(sampleSearchJSON))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
	if results[0].Title != "Believer" || results[0].Uploader != "ImagineDragonsVEVO" {
		t.Errorf("first result: %+v", results[0])
	}
	if results[0].URL != "https://www.youtube.com/watch?v=abc123" {
		t.Errorf("url = %q", results[0].URL)
	}
	if results[0].Duration != "3:24" {
		t.Errorf("duration = %q, want 3:24", results[0].Duration)
	}
	if results[1].Duration != "5:11" {
		t.Errorf("duration = %q, want 5:11", results[1].Duration)
	}
}

func TestParseSearchJSONEmpty(t *testing.T) {
	_, err := parseSearchJSON([]byte(`{"entries":[]}`))
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestFormatDuration(t *testing.T) {
	if got := formatDuration(0); got != "" {
		t.Errorf("zero duration = %q", got)
	}
	if got := formatDuration(65); got != "1:05" {
		t.Errorf("65s = %q, want 1:05", got)
	}
}
