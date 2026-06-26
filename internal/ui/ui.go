// Package ui implements the interactive terminal interface for Shadowbox using
// the Bubble Tea framework.
package ui

import (
	"github.com/EnJulian/shadowbox/internal/config"
)

// Run launches the interactive interface. It is implemented in the TUI phase.
func Run(cfg *config.Config) error {
	return runProgram(cfg)
}
