// Package config loads and persists Shadowbox user preferences and credentials.
//
// Configuration lives at ~/.config/shadowbox/config.yaml (or the platform
// equivalent) and is backed by Viper. Credentials may also be supplied through
// environment variables, which always take precedence over the config file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Spotify holds the Spotify Web API client credentials.
type Spotify struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

// Genius holds the Genius API access token used for lyrics lookups.
type Genius struct {
	AccessToken string `mapstructure:"access_token"`
}

// Config is the fully resolved Shadowbox configuration.
type Config struct {
	AudioFormat    string  `mapstructure:"audio_format"`
	MusicDirectory string  `mapstructure:"music_directory"`
	UseSpotify     bool    `mapstructure:"use_spotify"`
	UseGenius      bool    `mapstructure:"use_genius"`
	Verbose        bool    `mapstructure:"verbose"`
	Theme          string  `mapstructure:"theme"`
	Spotify        Spotify `mapstructure:"spotify"`
	Genius         Genius  `mapstructure:"genius"`
}

const (
	// EnvPrefix is the prefix for Shadowbox-specific environment variables,
	// e.g. SHADOWBOX_AUDIO_FORMAT.
	EnvPrefix = "SHADOWBOX"
)

// defaultMusicDir returns ~/Music, falling back to the working directory.
func defaultMusicDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "Music"
	}
	return filepath.Join(home, "Music")
}

// Dir returns the directory that holds config.yaml.
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "shadowbox"), nil
}

// Path returns the full path to config.yaml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("audio_format", "opus")
	v.SetDefault("music_directory", defaultMusicDir())
	v.SetDefault("use_spotify", false)
	v.SetDefault("use_genius", true)
	v.SetDefault("verbose", false)
	v.SetDefault("theme", "hacker")
	v.SetDefault("spotify.client_id", "")
	v.SetDefault("spotify.client_secret", "")
	v.SetDefault("genius.access_token", "")
}

// newViper builds a Viper instance wired to the config file, defaults, and the
// environment variable bindings (including backward-compatible names).
func newViper() (*viper.Viper, error) {
	v := viper.New()
	setDefaults(v)

	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Backward-compatible environment variables from the Python version.
	_ = v.BindEnv("spotify.client_id", "SHADOWBOX_SPOTIFY_CLIENT_ID", "SPOTIFY_CLIENT_ID")
	_ = v.BindEnv("spotify.client_secret", "SHADOWBOX_SPOTIFY_CLIENT_SECRET", "SPOTIFY_CLIENT_SECRET")
	_ = v.BindEnv("genius.access_token", "SHADOWBOX_GENIUS_ACCESS_TOKEN", "GENIUS_ACCESS_TOKEN")

	return v, nil
}

// Load reads configuration from disk, applies defaults and environment
// overrides, and migrates a legacy ~/.shadowbox_settings.json if present.
func Load() (*Config, error) {
	if err := migrateLegacy(); err != nil {
		// Migration failures are non-fatal; continue with whatever config exists.
		fmt.Fprintf(os.Stderr, "warning: legacy settings migration failed: %v\n", err)
	}

	v, err := newViper()
	if err != nil {
		return nil, err
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	cfg.MusicDirectory = expandHome(cfg.MusicDirectory)
	return &cfg, nil
}

// Save writes the provided configuration to config.yaml, creating the config
// directory if necessary.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	v, err := newViper()
	if err != nil {
		return err
	}
	v.Set("audio_format", cfg.AudioFormat)
	v.Set("music_directory", cfg.MusicDirectory)
	v.Set("use_spotify", cfg.UseSpotify)
	v.Set("use_genius", cfg.UseGenius)
	v.Set("verbose", cfg.Verbose)
	v.Set("theme", cfg.Theme)
	v.Set("spotify.client_id", cfg.Spotify.ClientID)
	v.Set("spotify.client_secret", cfg.Spotify.ClientSecret)
	v.Set("genius.access_token", cfg.Genius.AccessToken)

	path, err := Path()
	if err != nil {
		return err
	}
	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("securing config file: %w", err)
	}
	return nil
}

// expandHome expands a leading ~ to the user's home directory.
func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
