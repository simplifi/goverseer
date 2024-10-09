package file_watcher

import (
	"fmt"
	"os"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/logger"
)

const (
	// DefaultPollSeconds is the default number of seconds to wait between polls
	DefaultPollSeconds = 5
)

// Config is the configuration for a file watcher
type Config struct {
	// Path is the path to the file to watch
	Path string

	// PollSeconds is the number of seconds to wait between ticks
	PollSeconds int
}

// ParseConfig parses the config for a file watcher
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(config interface{}) (*Config, error) {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	cfg := &Config{
		PollSeconds: DefaultPollSeconds,
	}

	// Path is required and must be a string
	if path, ok := cfgMap["path"].(string); ok {
		if path == "" {
			return nil, fmt.Errorf("path must not be empty")
		}
		cfg.Path = path
	} else if cfgMap["path"] != nil {
		return nil, fmt.Errorf("path must be a string")
	} else {
		return nil, fmt.Errorf("path is required")
	}

	// If PollSeconds is set, it must be a positive number
	if pollSeconds, ok := cfgMap["poll_seconds"].(int); ok {
		if pollSeconds < 1 {
			return nil, fmt.Errorf("poll_seconds must be greater than or equal to 1")
		}
		cfg.PollSeconds = pollSeconds
	} else if cfgMap["poll_seconds"] != nil {
		return nil, fmt.Errorf("poll_seconds must be an integer")
	}

	return cfg, nil
}

// FileWatcher watches a file for changes and sends the path thru change channel
type FileWatcher struct {
	Config

	// lastValue is the last time the file was modified
	lastValue time.Time

	// stop is a channel to signal the watcher to stop
	stop chan struct{}
}

// New creates a new FileWatcher based on the config
func New(cfg config.Config) (*FileWatcher, error) {
	tcfg, err := ParseConfig(cfg.Watcher.Config)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		Config: Config{
			Path:        tcfg.Path,
			PollSeconds: tcfg.PollSeconds,
		},
		lastValue: time.Now(),
		stop:      make(chan struct{}),
	}, nil
}

// Watch watches the file for changes and sends the path to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *FileWatcher) Watch(changes chan interface{}) {
	logger.Log.Info("starting watcher")

	for {
		select {
		case <-w.stop:
			return
		case <-time.After(time.Duration(w.PollSeconds) * time.Second):
			info, err := os.Stat(w.Path)
			if err != nil {
				logger.Log.Error("error getting file info",
					"path", w.Path,
					"err", err)
			}
			if err == nil && info.ModTime().After(w.lastValue) {
				logger.Log.Info("file changed",
					"path", w.Path,
					"mod_time", info.ModTime())
				w.lastValue = info.ModTime()
				changes <- w.Path
			}
		}
	}
}

// Stop signals the watcher to stop
func (w *FileWatcher) Stop() {
	logger.Log.Info("shutting down watcher")
	close(w.stop)
}
