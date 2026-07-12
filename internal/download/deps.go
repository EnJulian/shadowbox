package download

import (
	"os/exec"
)

// Dependency describes an external command Shadowbox relies on.
type Dependency struct {
	Name     string // command name on PATH
	Purpose  string // human-readable description
	Required bool   // whether the program cannot function without it
	Path     string // resolved path, empty if not found
	Found    bool
}

// Dependencies returns the external tools Shadowbox uses, resolving each on the
// current PATH.
func Dependencies() []Dependency {
	specs := []Dependency{
		{Name: "yt-dlp", Purpose: "downloading audio from YouTube/Bandcamp/KHInsider", Required: true},
		{Name: "ffmpeg", Purpose: "audio extraction and conversion", Required: true},
		{Name: "aria2c", Purpose: "accelerated multi-connection downloads", Required: false},
		{Name: "mpv", Purpose: "in-app audio playback", Required: false},
	}
	for i := range specs {
		if p, err := exec.LookPath(specs[i].Name); err == nil {
			specs[i].Path = p
			specs[i].Found = true
		}
	}
	return specs
}

// MissingRequired returns the names of required dependencies that are not on PATH.
func MissingRequired() []string {
	var missing []string
	for _, d := range Dependencies() {
		if d.Required && !d.Found {
			missing = append(missing, d.Name)
		}
	}
	return missing
}

// HasAria2 reports whether aria2c is available for accelerated downloads.
func HasAria2() bool {
	_, err := exec.LookPath("aria2c")
	return err == nil
}
