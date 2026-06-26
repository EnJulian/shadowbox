// Package log provides styled, tagged terminal output that mirrors the look of
// the original Python Shadowbox UI (e.g. "[SUCCESS]", "[META]", "[ERROR]").
//
// Errors and warnings are always shown. Informational output is gated behind
// verbose mode, matching the behaviour of the Python implementation.
package log

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	verbose bool

	outW io.Writer = os.Stdout
	errW io.Writer = os.Stderr

	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	red     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	gray    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	magenta = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
)

// SetVerbose toggles whether informational messages are printed.
func SetVerbose(v bool) { verbose = v }

// Verbose reports whether verbose logging is enabled.
func Verbose() bool { return verbose }

// SetWriters overrides the destinations for normal and error output. Passing nil
// for either keeps the current writer. Used by the TUI to silence direct stdout
// writes while Bubble Tea controls the screen.
func SetWriters(out, err io.Writer) {
	if out != nil {
		outW = out
	}
	if err != nil {
		errW = err
	}
}

// Reset restores output to the standard streams.
func Reset() {
	outW = os.Stdout
	errW = os.Stderr
}

func emit(w io.Writer, style lipgloss.Style, tag, msg string) {
	fmt.Fprintf(w, "%s %s\n", style.Render("["+tag+"]"), msg)
}

// Success prints a success message (verbose only).
func Success(format string, a ...any) {
	if verbose {
		emit(outW, green, "SUCCESS", fmt.Sprintf(format, a...))
	}
}

// Successf prints a success message with a custom tag (verbose only).
func Successf(tag, format string, a ...any) {
	if verbose {
		emit(outW, green, tag, fmt.Sprintf(format, a...))
	}
}

// Info prints an informational message (verbose only).
func Info(format string, a ...any) {
	if verbose {
		emit(outW, cyan, "INFO", fmt.Sprintf(format, a...))
	}
}

// Infof prints an informational message with a custom tag (verbose only).
func Infof(tag, format string, a ...any) {
	if verbose {
		emit(outW, cyan, tag, fmt.Sprintf(format, a...))
	}
}

// System prints a system message (verbose only).
func System(format string, a ...any) {
	if verbose {
		emit(outW, gray, "SYSTEM", fmt.Sprintf(format, a...))
	}
}

// Systemf prints a system message with a custom tag (verbose only).
func Systemf(tag, format string, a ...any) {
	if verbose {
		emit(outW, gray, tag, fmt.Sprintf(format, a...))
	}
}

// Step prints a high-level progress message that is always shown.
func Step(tag, format string, a ...any) {
	emit(outW, magenta, tag, fmt.Sprintf(format, a...))
}

// Warning prints a warning message (always shown).
func Warning(format string, a ...any) {
	emit(errW, yellow, "WARNING", fmt.Sprintf(format, a...))
}

// Warningf prints a warning message with a custom tag (always shown).
func Warningf(tag, format string, a ...any) {
	emit(errW, yellow, tag, fmt.Sprintf(format, a...))
}

// Error prints an error message (always shown).
func Error(format string, a ...any) {
	emit(errW, red, "ERROR", fmt.Sprintf(format, a...))
}

// Errorf prints an error message with a custom tag (always shown).
func Errorf(tag, format string, a ...any) {
	emit(errW, red, tag, fmt.Sprintf(format, a...))
}
