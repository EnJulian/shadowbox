package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/EnJulian/shadowbox/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "View and edit Shadowbox configuration",
	}
	c.AddCommand(newConfigGetCmd(), newConfigSetCmd(), newConfigListCmd(), newConfigPathCmd())
	return c
}

// configKeys maps dotted keys to getters/setters over the Config struct.
// Secret values are masked when listed.
var configKeys = map[string]struct {
	secret bool
	get    func(c *config.Config) string
	set    func(c *config.Config, v string) error
}{
	"audio_format": {
		get: func(c *config.Config) string { return c.AudioFormat },
		set: func(c *config.Config, v string) error { c.AudioFormat = v; return nil },
	},
	"music_directory": {
		get: func(c *config.Config) string { return c.MusicDirectory },
		set: func(c *config.Config, v string) error { c.MusicDirectory = v; return nil },
	},
	"use_genius": {
		get: func(c *config.Config) string { return strconv.FormatBool(c.UseGenius) },
		set: func(c *config.Config, v string) error { return setBool(v, &c.UseGenius) },
	},
	"verbose": {
		get: func(c *config.Config) string { return strconv.FormatBool(c.Verbose) },
		set: func(c *config.Config, v string) error { return setBool(v, &c.Verbose) },
	},
	"theme": {
		get: func(c *config.Config) string { return c.Theme },
		set: func(c *config.Config, v string) error { c.Theme = v; return nil },
	},
	"genius.access_token": {
		secret: true,
		get:    func(c *config.Config) string { return c.Genius.AccessToken },
		set:    func(c *config.Config, v string) error { c.Genius.AccessToken = v; return nil },
	},
}

func sortedKeys() []string {
	keys := make([]string, 0, len(configKeys))
	for k := range configKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func setBool(v string, dst *bool) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fmt.Errorf("expected a boolean (true/false), got %q", v)
	}
	*dst = b
	return nil
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value of a configuration key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := configKeys[args[0]]
			if !ok {
				return fmt.Errorf("unknown key %q (run 'shadowbox config list')", args[0])
			}
			fmt.Println(entry.get(cfg))
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration key and persist it",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			entry, ok := configKeys[args[0]]
			if !ok {
				return fmt.Errorf("unknown key %q (run 'shadowbox config list')", args[0])
			}
			if err := entry.set(cfg, args[1]); err != nil {
				return err
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Printf("set %s\n", args[0])
			return nil
		},
	}
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration keys and values",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			for _, k := range sortedKeys() {
				entry := configKeys[k]
				val := entry.get(cfg)
				if entry.secret && val != "" {
					val = maskSecret(val)
				}
				fmt.Printf("%-22s %s\n", k, val)
			}
		},
	}
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the path to the config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Println(p)
			return nil
		},
	}
}

// maskSecret reveals only the last few characters of a secret value.
func maskSecret(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
}
