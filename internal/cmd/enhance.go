package cmd

import (
	"strings"

	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/spf13/cobra"
)

func newEnhanceCmd() *cobra.Command {
	var (
		recursive  bool
		extensions string
		dryRun     bool
	)
	c := &cobra.Command{
		Use:   "enhance <directory>",
		Short: "Batch-enhance existing audio files with online metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := app.New(cfg)
			exts := strings.Split(extensions, ",")
			return a.EnhanceDir(cmd.Context(), args[0], recursive, exts, dryRun)
		},
	}
	c.Flags().BoolVarP(&recursive, "recursive", "r", false, "recurse into subdirectories")
	c.Flags().StringVarP(&extensions, "extensions", "e", "opus,mp3,m4a,flac", "comma-separated audio extensions")
	c.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change without writing")
	return c
}
