package watcher

import (
	"fmt"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/watcher/file_watcher"
	"github.com/simplifi/goverseer/internal/goverseer/watcher/gce_metadata_watcher"
	"github.com/simplifi/goverseer/internal/goverseer/watcher/time_watcher"
	"github.com/simplifi/goverseer/internal/goverseer/watcher/gcp_secrets_watcher"
)

// Watcher is an interface for watching for changes
type Watcher interface {
	Watch(change chan interface{})
	Stop()
}

// New creates a new Watcher based on the config
// The config is the watcher configuration
func New(cfg *config.Config) (Watcher, error) {
	switch cfg.Watcher.Type {
	case "file":
		return file_watcher.New(*cfg)
	case "time":
		return time_watcher.New(*cfg)
	case "gce_metadata":
		return gce_metadata_watcher.New(*cfg)
	case "gcp_secrets":
		return gcp_secrets_watcher.New(cfg.Watcher.Config)
	default:
		return nil, fmt.Errorf("unknown watcher type: %s", cfg.Watcher.Type)
	}
}
