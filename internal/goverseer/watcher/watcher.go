package watcher

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Watcher is an interface for watching for changes
type Watcher interface {
	Watch(changes chan interface{}, wg *sync.WaitGroup)
	Stop()
}

// NewWatcher creates a new Watcher based on the config
// The config is the watcher configuration
// The log is the logger
func NewWatcher(cfg *config.Config, log *slog.Logger) (Watcher, error) {
	switch v := cfg.Watcher.Config.(type) {
	case config.FileWatcherConfig:
		return NewFileWatcher(v.Path, v.PollSeconds, log)
	case config.GceWatcherConfig:
		return NewGCEWatcher(v.Source, v.Key, v.MetadataUrl, log)
	default:
		return nil, fmt.Errorf("unknown watcher type: %s", cfg.Watcher.Type)
	}
}
