// Package setup detects Shadowbox's optional external tools (mpv, aria2)
// and figures out how to install whichever ones are missing, for the
// first-run and Settings-triggered setup wizard.
package setup

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/EnJulian/shadowbox/internal/player"
)

// Tool describes an optional external dependency the setup wizard can help
// install. Name is the binary Shadowbox actually looks for on PATH;
// PackageName is what package managers call it — not always the same
// (aria2's binary is aria2c, but Homebrew/Scoop's package is "aria2").
type Tool struct {
	Name        string
	PackageName string
	Purpose     string
}

// Tools is the fixed list of optional dependencies the wizard offers.
var Tools = []Tool{
	{Name: "mpv", PackageName: "mpv", Purpose: "in-app audio playback"},
	{Name: "aria2c", PackageName: "aria2", Purpose: "accelerated multi-connection downloads"},
}

// Available reports whether the named tool (Tool.Name) is already usable.
func Available(name string) bool {
	switch name {
	case "mpv":
		return player.Available()
	case "aria2c":
		return download.HasAria2()
	}
	return false
}

// InstallResult describes how to get a tool installed on this machine.
type InstallResult struct {
	// Automatable is true if Command can be run directly, without elevated
	// permissions.
	Automatable bool
	// Command is the argv to execute; only meaningful when Automatable.
	Command []string
	// Instruction is a human-readable line to show either way.
	Instruction string
}

// InstallCommand returns the install approach for tool on this machine.
func InstallCommand(tool Tool) InstallResult {
	return installCommandFor(runtime.GOOS, exec.LookPath, tool)
}

// installCommandFor is InstallCommand's testable core: goos and lookPath
// are injected so every platform branch is exercisable without real
// package managers installed or being locked to the test runner's own OS.
func installCommandFor(goos string, lookPath func(string) (string, error), tool Tool) InstallResult {
	switch goos {
	case "darwin":
		cmd := []string{"brew", "install", tool.PackageName}
		if _, err := lookPath("brew"); err == nil {
			return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
		}
		return InstallResult{Instruction: "install Homebrew (https://brew.sh), then run: " + strings.Join(cmd, " ")}
	case "windows":
		if _, err := lookPath("scoop"); err == nil {
			cmd := []string{"scoop", "install", tool.PackageName}
			return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
		}
		if _, err := lookPath("winget"); err == nil {
			cmd := []string{"winget", "install", tool.PackageName}
			return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
		}
		return InstallResult{Instruction: "install via Scoop (https://scoop.sh) or winget — package name: " + tool.PackageName}
	default:
		// Linux: check known package managers in priority order. Unlike
		// macOS/Windows, most of these need sudo — that's fine, since
		// tea.ExecProcess hands the real terminal to the subprocess for its
		// whole run, so a sudo password prompt behaves exactly as it would
		// running the command directly in a terminal.
		linuxManagers := []struct {
			bin string
			cmd func(pkg string) []string
		}{
			{"apt-get", func(pkg string) []string { return []string{"sudo", "apt-get", "install", "-y", pkg} }},
			{"dnf", func(pkg string) []string { return []string{"sudo", "dnf", "install", "-y", pkg} }},
			{"pacman", func(pkg string) []string { return []string{"sudo", "pacman", "-S", "--noconfirm", pkg} }},
			{"zypper", func(pkg string) []string { return []string{"sudo", "zypper", "install", "-y", pkg} }},
		}
		for _, mgr := range linuxManagers {
			if _, err := lookPath(mgr.bin); err == nil {
				cmd := mgr.cmd(tool.PackageName)
				return InstallResult{Automatable: true, Command: cmd, Instruction: strings.Join(cmd, " ")}
			}
		}
		return InstallResult{Instruction: "install `" + tool.PackageName + "` using your distribution's package manager, then ensure it's on your PATH"}
	}
}
