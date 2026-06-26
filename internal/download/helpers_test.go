package download

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeFile creates a small file with a slightly increasing mtime so ordering is
// deterministic in tests.
func writeFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, make([]byte, 2048), 0o644); err != nil {
		t.Fatal(err)
	}
	// Stagger modification times by name so the later-created file is "newest".
	mod := time.Now().Add(time.Duration(len(name)) * time.Second)
	_ = os.Chtimes(path, mod, mod)
}

func filepathBase(p string) string { return filepath.Base(p) }
