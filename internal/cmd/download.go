package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/spf13/cobra"
)

func newDownloadCmd() *cobra.Command {
	var (
		query     string
		directory string
		output    string
		format    string
		spotify   bool
	)
	c := &cobra.Command{
		Use:   "download",
		Short: "Download a track or playlist and tag it",
		Long: "Download audio from a YouTube/Bandcamp/KHInsider URL or a search query, then\n" +
			"organise and tag it with metadata, cover art, and lyrics.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if query == "" && len(args) > 0 {
				query = strings.Join(args, " ")
			}
			if query == "" {
				q, err := prompt("Enter song title and artist or URL: ")
				if err != nil {
					return err
				}
				query = q
			}
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("a query or URL is required")
			}

			a := app.New(cfg)
			opts := app.Options{
				MusicDir:   directory,
				Output:     output,
				Format:     format,
				UseSpotify: spotify || cfg.UseSpotify,
			}
			return a.Run(cmd.Context(), query, opts)
		},
	}
	c.Flags().StringVarP(&query, "query", "q", "", "song title and artist or URL")
	c.Flags().StringVarP(&directory, "directory", "d", "", "base music directory (default ~/Music)")
	c.Flags().StringVarP(&output, "output", "o", "", "output filename override (no extension)")
	c.Flags().StringVarP(&format, "format", "f", "", "audio format (opus, m4a, mp3, flac, wav)")
	c.Flags().BoolVarP(&spotify, "spotify", "s", false, "use Spotify for metadata")
	return c
}

// prompt reads a single line of input from stdin.
func prompt(label string) (string, error) {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
