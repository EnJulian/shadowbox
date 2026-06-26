package cmd

import (
	"fmt"
	"strings"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
	var (
		file   string
		title  string
		artist string
		output string
	)
	c := &cobra.Command{
		Use:   "tag [query|url]",
		Short: "Tag an existing file, or download and tag using Spotify metadata",
		Long: "With --file, tags an existing audio file in place using online metadata\n" +
			"(optionally overriding title/artist). With a query or URL, downloads the\n" +
			"track using Spotify metadata enrichment.",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := app.New(cfg)

			if file != "" {
				return a.Enhance(cmd.Context(), file, title, artist)
			}

			query := strings.Join(args, " ")
			if strings.TrimSpace(query) == "" {
				return fmt.Errorf("provide --file to tag an existing file, or a query/URL to download")
			}
			return a.Run(cmd.Context(), query, app.Options{
				Output:     output,
				UseSpotify: true,
			})
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "existing audio file to tag in place")
	c.Flags().StringVarP(&title, "title", "t", "", "track title override")
	c.Flags().StringVarP(&artist, "artist", "a", "", "artist override")
	c.Flags().StringVarP(&output, "output", "o", "", "output filename for downloads")
	return c
}
