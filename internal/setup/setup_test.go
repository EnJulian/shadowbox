package setup

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func fakeLookPathFound(string) (string, error) { return "/usr/bin/x", nil }
func fakeLookPathNotFound(string) (string, error) {
	return "", errors.New("not found")
}

func TestInstallCommandDarwinWithBrew(t *testing.T) {
	res := installCommandFor("darwin", fakeLookPathFound, Tool{Name: "mpv", PackageName: "mpv"})
	if !res.Automatable {
		t.Fatal("expected automatable when brew is found")
	}
	want := []string{"brew", "install", "mpv"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandDarwinWithoutBrew(t *testing.T) {
	res := installCommandFor("darwin", fakeLookPathNotFound, Tool{Name: "mpv", PackageName: "mpv"})
	if res.Automatable {
		t.Fatal("expected not automatable when brew is missing")
	}
	if !strings.Contains(res.Instruction, "brew.sh") {
		t.Errorf("Instruction = %q, want a pointer to installing Homebrew", res.Instruction)
	}
}

func TestInstallCommandWindowsPrefersScoopOverWinget(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "scoop" {
			return "/scoop", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("windows", lookup, Tool{Name: "aria2c", PackageName: "aria2"})
	if !res.Automatable {
		t.Fatal("expected automatable via scoop")
	}
	want := []string{"scoop", "install", "aria2"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandWindowsFallsBackToWinget(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "winget" {
			return "/winget", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("windows", lookup, Tool{Name: "aria2c", PackageName: "aria2"})
	if !res.Automatable {
		t.Fatal("expected automatable via winget")
	}
	want := []string{"winget", "install", "aria2"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandWindowsNoManagerFound(t *testing.T) {
	res := installCommandFor("windows", fakeLookPathNotFound, Tool{Name: "mpv", PackageName: "mpv"})
	if res.Automatable {
		t.Fatal("expected not automatable when neither scoop nor winget is found")
	}
}

func TestInstallCommandLinuxAptGet(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "apt-get" {
			return "/usr/bin/apt-get", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("linux", lookup, Tool{Name: "mpv", PackageName: "mpv"})
	if !res.Automatable {
		t.Fatal("expected automatable via apt-get")
	}
	want := []string{"sudo", "apt-get", "install", "-y", "mpv"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandLinuxDnf(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "dnf" {
			return "/usr/bin/dnf", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("linux", lookup, Tool{Name: "aria2c", PackageName: "aria2"})
	if !res.Automatable {
		t.Fatal("expected automatable via dnf")
	}
	want := []string{"sudo", "dnf", "install", "-y", "aria2"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandLinuxPacman(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "pacman" {
			return "/usr/bin/pacman", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("linux", lookup, Tool{Name: "mpv", PackageName: "mpv"})
	if !res.Automatable {
		t.Fatal("expected automatable via pacman")
	}
	want := []string{"sudo", "pacman", "-S", "--noconfirm", "mpv"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandLinuxZypper(t *testing.T) {
	lookup := func(name string) (string, error) {
		if name == "zypper" {
			return "/usr/bin/zypper", nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("linux", lookup, Tool{Name: "mpv", PackageName: "mpv"})
	if !res.Automatable {
		t.Fatal("expected automatable via zypper")
	}
	want := []string{"sudo", "zypper", "install", "-y", "mpv"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v", res.Command, want)
	}
}

func TestInstallCommandLinuxPrefersEarlierManagerWhenMultipleFound(t *testing.T) {
	// apt-get and dnf both "found" — apt-get must win (priority order).
	lookup := func(name string) (string, error) {
		if name == "apt-get" || name == "dnf" {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
	res := installCommandFor("linux", lookup, Tool{Name: "mpv", PackageName: "mpv"})
	want := []string{"sudo", "apt-get", "install", "-y", "mpv"}
	if !reflect.DeepEqual(res.Command, want) {
		t.Errorf("Command = %v, want %v (apt-get should win over dnf)", res.Command, want)
	}
}

func TestInstallCommandLinuxNoManagerFoundFallsBackToInstructions(t *testing.T) {
	res := installCommandFor("linux", fakeLookPathNotFound, Tool{Name: "mpv", PackageName: "mpv"})
	if res.Automatable {
		t.Fatal("expected not automatable when no known package manager is found")
	}
	if !strings.Contains(res.Instruction, "mpv") {
		t.Errorf("Instruction = %q, want it to mention the package name", res.Instruction)
	}
}

func TestAvailableReturnsFalseForUnrecognizedTool(t *testing.T) {
	if Available("does-not-exist") {
		t.Error("expected false for an unrecognized tool name")
	}
}

func TestToolsListsMpvAndAria2(t *testing.T) {
	names := map[string]bool{}
	for _, tool := range Tools {
		names[tool.Name] = true
	}
	if !names["mpv"] || !names["aria2c"] {
		t.Errorf("Tools = %+v, want it to include mpv and aria2c", Tools)
	}
}
