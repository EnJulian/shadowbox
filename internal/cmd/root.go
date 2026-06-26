// Package cmd wires up the Shadowbox command-line interface using Cobra.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/EnJulian/shadowbox/internal/config"
	applog "github.com/EnJulian/shadowbox/internal/log"
	"github.com/EnJulian/shadowbox/internal/ui"
	"github.com/spf13/cobra"
)

// cfg is the loaded configuration shared across commands.
var cfg *config.Config

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "shadowbox",
		Short: "Music acquisition CLI for YouTube and Bandcamp",
		Long: "Shadowbox rips audio from YouTube and Bandcamp, converts it to Opus\n" +
			"(or your chosen format), and injects rich metadata, cover art, and lyrics.\n\n" +
			"Run without arguments to launch the interactive interface.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			loaded, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			cfg = loaded

			verbose, _ := cmd.Flags().GetBool("verbose")
			applog.SetVerbose(verbose || cfg.Verbose)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// No subcommand: launch the interactive TUI.
			return ui.Run(cfg)
		},
	}

	root.PersistentFlags().BoolP("verbose", "v", false, "enable verbose logging")

	root.AddCommand(
		newDownloadCmd(),
		newTagCmd(),
		newEnhanceCmd(),
		newConfigCmd(),
		newDoctorCmd(),
		newVersionCmd(),
	)
	return root
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := newRootCmd().ExecuteContext(ctx); err != nil {
		applog.Error("%v", err)
		return 1
	}
	return 0
}
