package download

import (
	"strings"
	"testing"

	"github.com/EnJulian/shadowbox/internal/progress"
)

func TestScanYTDLPProgress(t *testing.T) {
	var got []progress.Update
	d := &Downloader{
		Progress: func(u progress.Update) {
			got = append(got, u)
		},
	}
	input := strings.NewReader(
		"[download] Downloading item 2 of 5\n" +
			"[download] 100% of 3.45MiB\n" +
			"[download] Downloading video 3 of 5\n",
	)
	d.scanYTDLPProgress(input)

	if len(got) != 2 {
		t.Fatalf("expected 2 updates, got %d: %+v", len(got), got)
	}
	if got[0].Current != 2 || got[0].Total != 5 || got[0].Heading != "Downloading playlist track 2/5" {
		t.Fatalf("unexpected first update: %+v", got[0])
	}
	if got[1].Current != 3 || got[1].Total != 5 {
		t.Fatalf("unexpected second update: %+v", got[1])
	}
}
