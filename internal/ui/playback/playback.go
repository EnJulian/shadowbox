// Package playback is a Phase 1 stub. It exists so the shell's reserved
// playback bar slot (shell.Layout.PlaybarHeight) and the root model have a
// concrete type to hold, without any decoding/IPC logic yet — that ships in
// a later update (see docs/superpowers/specs/2026-07-11-terminal-ui-redesign-v2-design.md).
package playback

// State describes the current playback state. Active is always false this
// phase.
type State struct {
	Active bool
}
