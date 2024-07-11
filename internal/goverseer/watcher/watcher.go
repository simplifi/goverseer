package watcher

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Watcher is an interface for watching for changes
type Watcher interface {
	Watch(changes chan interface{})
	Create(cfg config.WatcherConfig, log *slog.Logger) error
	Stop()
}

type WatcherFactory func() Watcher

// WatcherRegistry is a global registry for watcher factories
var WatcherRegistry = make(map[string]WatcherFactory)

// RegisterWatcher registers a watcher factory with the global registry
func RegisterWatcher(watcherType string, factory WatcherFactory) {
	WatcherRegistry[watcherType] = factory
}

// New creates a new Watcher based on the config
// The config is the watcher configuration
func New(cfg *config.Config) (*Watcher, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("watcher", cfg.Watcher.Type)

	// Get the registered factory function
	factory, found := WatcherRegistry[cfg.Watcher.Type]
	if !found {
		return nil, fmt.Errorf("unknown watcher type: %s", cfg.Watcher.Type)
	}

	// Create an instance of the executor using the factory function
	exec := factory()
	err := exec.Create(cfg.Watcher.Config, log)

	return &exec, err
}
