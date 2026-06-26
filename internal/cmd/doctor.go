package cmd

import (
	"fmt"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/EnJulian/shadowbox/internal/download"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check that required external tools and credentials are available",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Shadowbox environment check")
			fmt.Println()

			fmt.Println("External tools:")
			allRequiredFound := true
			for _, dep := range download.Dependencies() {
				status := "missing"
				mark := "x"
				if dep.Found {
					status = dep.Path
					mark = "+"
				} else if dep.Required {
					allRequiredFound = false
				}
				req := "optional"
				if dep.Required {
					req = "required"
				}
				fmt.Printf("  [%s] %-8s (%s) - %s\n", mark, dep.Name, req, status)
			}

			fmt.Println()
			fmt.Println("Credentials:")
			printCredStatus("Spotify client ID", cfg.Spotify.ClientID != "")
			printCredStatus("Spotify client secret", cfg.Spotify.ClientSecret != "")
			printCredStatus("Genius access token", cfg.Genius.AccessToken != "")

			fmt.Println()
			path, _ := config.Path()
			fmt.Printf("Config file: %s\n", path)
			fmt.Printf("Music directory: %s\n", cfg.MusicDirectory)

			if !allRequiredFound {
				fmt.Println()
				return fmt.Errorf("one or more required tools are missing; install them and re-run 'shadowbox doctor'")
			}
			return nil
		},
	}
}

func printCredStatus(label string, set bool) {
	mark := "x"
	status := "not set"
	if set {
		mark = "+"
		status = "configured"
	}
	fmt.Printf("  [%s] %-22s %s\n", mark, label, status)
}
