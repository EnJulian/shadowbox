package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// legacySettings mirrors the schema of the Python version's
// ~/.shadowbox_settings.json file.
type legacySettings struct {
	AudioFormat         string `json:"audio_format"`
	MusicDirectory      string `json:"music_directory"`
	UseSpotify          bool   `json:"use_spotify"`
	UseGenius           bool   `json:"use_genius"`
	VerboseLogging      bool   `json:"verbose_logging"`
	Theme               string `json:"theme"`
	GeniusAccessToken   string `json:"genius_access_token"`
	SpotifyClientID     string `json:"spotify_client_id"`
	SpotifyClientSecret string `json:"spotify_client_secret"`
}

// legacyPath returns the path to the Python version's settings file.
func legacyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".shadowbox_settings.json"), nil
}

// migrateLegacy imports a legacy JSON settings file into the new YAML config the
// first time Shadowbox runs. It is a no-op when the new config already exists or
// when no legacy file is present.
func migrateLegacy() error {
	newPath, err := Path()
	if err != nil {
		return err
	}
	if _, err := os.Stat(newPath); err == nil {
		return nil // New config already exists; nothing to migrate.
	}

	oldPath, err := legacyPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(oldPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var legacy legacySettings
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fmt.Errorf("parsing legacy settings: %w", err)
	}

	cfg := &Config{
		AudioFormat:    firstNonEmpty(legacy.AudioFormat, "opus"),
		MusicDirectory: firstNonEmpty(legacy.MusicDirectory, defaultMusicDir()),
		UseSpotify:     legacy.UseSpotify,
		UseGenius:      legacy.UseGenius,
		Verbose:        legacy.VerboseLogging,
		Theme:          firstNonEmpty(legacy.Theme, "hacker"),
		Spotify: Spotify{
			ClientID:     legacy.SpotifyClientID,
			ClientSecret: legacy.SpotifyClientSecret,
		},
		Genius: Genius{
			AccessToken: legacy.GeniusAccessToken,
		},
	}
	return Save(cfg)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
