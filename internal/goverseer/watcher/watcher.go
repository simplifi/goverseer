package watcher

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/watcher/time"
)

// Watcher is an interface for watching for changes
type Watcher interface {
	Watch(change chan interface{})
	Stop()
}

// New creates a new Watcher based on the config
// The config is the watcher configuration
func New(cfg *config.Config) (Watcher, error) {
	// Setup the logger
	logger := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("watcher", cfg.Watcher.Type)

	switch cfg.Watcher.Type {
	case "time":
		return time.New(*cfg, logger)
	default:
		return nil, fmt.Errorf("unknown watcher type: %s", cfg.Watcher.Type)
	}
}
