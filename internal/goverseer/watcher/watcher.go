package watcher

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Watcher is an interface for watching for changes
type Watcher interface {
	Watch(changes chan interface{})
	Stop()
}

type WatcherFactory struct {
	mu       sync.RWMutex
	creators map[string]func(interface{}, *slog.Logger) (Watcher, error)
}

var factory = &WatcherFactory{
	creators: make(map[string]func(interface{}, *slog.Logger) (Watcher, error)),
}

func (f *WatcherFactory) Register(watcherType string, creator func(interface{}, *slog.Logger) (Watcher, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[watcherType] = creator
}

func (f *WatcherFactory) Create(watcherType string, cfg interface{}, log *slog.Logger) (Watcher, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	creator, exists := f.creators[watcherType]
	if !exists {
		return nil, fmt.Errorf("unknown watcher type: %s", watcherType)
	}
	return creator(cfg, log)
}

// NewWatcher creates a new Watcher based on the config
// The config is the watcher configuration
func NewWatcher(cfg *config.Config) (Watcher, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("watcher", cfg.Watcher.Type)

	return factory.Create(cfg.Watcher.Type, cfg.Watcher.Config, log)
}
