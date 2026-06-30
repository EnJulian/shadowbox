// Package progress defines structured pipeline progress updates for the UI.
package progress

// Update describes a single pipeline stage, optionally with numeric position.
type Update struct {
	Stage   string
	Heading string // running-screen title; empty keeps the current heading
	Current int    // 1-based; 0 when unknown
	Total   int    // 0 when unknown
}
