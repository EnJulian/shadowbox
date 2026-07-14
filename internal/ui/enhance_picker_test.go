package ui

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestOpenEnhancePickerStartsAtMusicDirectory(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Artist A", "keep.txt"))
	mustWrite(t, filepath.Join(dir, "Artist B", "keep.txt"))

	m := newTestModel()
	m.cfg.MusicDirectory = dir

	next, _ := m.openEnhancePicker()
	m = next.(model)

	if m.screen != screenEnhancePicker {
		t.Fatalf("screen = %v, want screenEnhancePicker", m.screen)
	}
	if m.enhancePicker.cwd != dir {
		t.Fatalf("cwd = %q, want %q", m.enhancePicker.cwd, dir)
	}
	want := []string{"Artist A", "Artist B"}
	if !reflect.DeepEqual(m.enhancePicker.entries, want) {
		t.Fatalf("entries = %v, want %v", m.enhancePicker.entries, want)
	}
	visible := m.enhancePicker.visible()
	if len(visible) == 0 || visible[0] != enhancePickerUseThisFolder {
		t.Fatalf("visible[0] = %q, want pinned %q first", visible[0], enhancePickerUseThisFolder)
	}
}

func TestEnhancePickerDescendsIntoSubdirectory(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Artist A", "Album", "keep.txt"))

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	next, _ := m.openEnhancePicker()
	m = next.(model)

	m.enhancePicker.cursor = 1 // "Artist A" (index 0 is the pinned entry)
	next, _ = m.updateEnhancePicker(key("enter"))
	m = next.(model)

	wantCwd := filepath.Join(dir, "Artist A")
	if m.enhancePicker.cwd != wantCwd {
		t.Fatalf("cwd = %q, want %q", m.enhancePicker.cwd, wantCwd)
	}
	if !reflect.DeepEqual(m.enhancePicker.entries, []string{"Album"}) {
		t.Fatalf("entries = %v, want [Album]", m.enhancePicker.entries)
	}
	if m.enhancePicker.cursor != 0 {
		t.Fatalf("cursor = %d, want reset to 0 after descending", m.enhancePicker.cursor)
	}
}

func TestEnhancePickerEscGoesToParentDirectory(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Artist A", "Album", "keep.txt"))

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	next, _ := m.openEnhancePicker()
	m = next.(model)
	m.enhancePicker.cursor = 1
	next, _ = m.updateEnhancePicker(key("enter")) // descend into "Artist A"
	m = next.(model)

	next, _ = m.updateEnhancePicker(key("esc"))
	m = next.(model)

	if m.screen != screenEnhancePicker {
		t.Fatalf("screen = %v, want screenEnhancePicker (still browsing after going up)", m.screen)
	}
	if m.enhancePicker.cwd != dir {
		t.Fatalf("cwd = %q, want %q (back at the start)", m.enhancePicker.cwd, dir)
	}
	if !reflect.DeepEqual(m.enhancePicker.entries, []string{"Artist A"}) {
		t.Fatalf("entries = %v, want [Artist A]", m.enhancePicker.entries)
	}
}

func TestEnhancePickerEscAtFilesystemRootReturnsToMenu(t *testing.T) {
	root := string(filepath.Separator)

	m := newTestModel()
	m.screen = screenEnhancePicker
	m.enhancePicker = enhancePickerState{cwd: root, entries: listSubdirs(root)}

	next, _ := m.updateEnhancePicker(key("esc"))
	m = next.(model)

	if m.screen != screenMenu {
		t.Fatalf("screen = %v, want screenMenu", m.screen)
	}
}

func TestEnhancePickerFilterNarrowsSubdirectoriesButKeepsPinnedEntry(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "Imagine Dragons", "keep.txt"))
	mustWrite(t, filepath.Join(dir, "Adele", "keep.txt"))

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	next, _ := m.openEnhancePicker()
	m = next.(model)

	next, _ = m.updateEnhancePicker(key("imagine"))
	m = next.(model)

	visible := m.enhancePicker.visible()
	want := []string{enhancePickerUseThisFolder, "Imagine Dragons"}
	if !reflect.DeepEqual(visible, want) {
		t.Fatalf("visible = %v, want %v", visible, want)
	}
}

func TestEnhancePickerSelectUseThisFolderStartsEnhancementTask(t *testing.T) {
	dir := t.TempDir()

	m := newTestModel()
	m.cfg.MusicDirectory = dir
	next, _ := m.openEnhancePicker()
	m = next.(model)
	// cursor is already 0: "[ Use this folder ]"

	next, cmd := m.updateEnhancePicker(key("enter"))
	m = next.(model)

	if cmd == nil {
		t.Fatal("expected a task command when confirming the folder")
	}
	if m.screen != screenRunning {
		t.Fatalf("screen = %v, want screenRunning", m.screen)
	}
	if m.enhancePicker.cwd != dir {
		t.Fatalf("cwd = %q, want unchanged %q", m.enhancePicker.cwd, dir)
	}
}
