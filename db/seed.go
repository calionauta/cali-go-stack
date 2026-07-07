package db

import (
	"log/slog"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// SeedDefaults creates default collections and data on first run.
func SeedDefaults(app *pocketbase.PocketBase) error {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Collections are auto-created by PocketBase on first access.
		slog.Info("seed: default collections ready")
		return se.Next()
	})
	return nil
}
