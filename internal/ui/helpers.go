package ui

import (
	"github.com/EnJulian/shadowbox/internal/app"
	"github.com/EnJulian/shadowbox/internal/config"
)

// rebuildApp recreates the pipeline App after credentials or format change.
func rebuildApp(cfg *config.Config) *app.App {
	return app.New(cfg)
}
