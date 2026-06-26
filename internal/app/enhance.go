package app

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	applog "github.com/EnJulian/shadowbox/internal/log"
)

// EnhanceDir tags every matching audio file under dir. When recursive is false,
// only the top level is scanned. When dryRun is true, matching files are listed
// but not modified.
func (a *App) EnhanceDir(ctx context.Context, dir string, recursive bool, exts []string, dryRun bool) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	extSet := make(map[string]bool, len(exts))
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(e, ".")))
		if e != "" {
			extSet["."+e] = true
		}
	}

	var files []string
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if !recursive && path != dir {
				return fs.SkipDir
			}
			return nil
		}
		if extSet[strings.ToLower(filepath.Ext(path))] {
			files = append(files, path)
		}
		return nil
	}
	if err := filepath.WalkDir(dir, walk); err != nil {
		return err
	}

	if len(files) == 0 {
		applog.Warning("No matching audio files found in %s", dir)
		return nil
	}

	var ok int
	for _, f := range files {
		if dryRun {
			fmt.Printf("would enhance: %s\n", f)
			ok++
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := a.Enhance(ctx, f, "", ""); err != nil {
			applog.Error("Failed to enhance %s: %v", filepath.Base(f), err)
			continue
		}
		ok++
	}
	applog.Step("SUMMARY", "Enhanced %d of %d files", ok, len(files))
	return nil
}
