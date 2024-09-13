package file_watcher

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

const (
	// DefaultPollSeconds is the default number of seconds to wait between polls
	DefaultPollSeconds = 5
)

// FileWatcherConfig is the configuration for a file watcher
type FileWatcherConfig struct {
	// Path is the path to the file to watch
	Path string

	// PollSeconds is the number of seconds to wait between ticks
	PollSeconds int
}

// ParseConfig parses the config for a file watcher
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(config interface{}) (*FileWatcherConfig, error) {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	cfg := &FileWatcherConfig{
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
	// Path is the path to the file to watch
	Path string

	// PollInterval is the interval to poll the file for changes
	PollInterval time.Duration

	// lastValue is the last time the file was modified
	lastValue time.Time

	// log is the logger
	log *slog.Logger

	// stop is a channel to signal the watcher to stop
	stop chan struct{}
}

// New creates a new FileWatcher based on the config
func New(cfg config.Config, log *slog.Logger) (*FileWatcher, error) {
	tcfg, err := ParseConfig(cfg.Watcher.Config)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		Path:         tcfg.Path,
		PollInterval: time.Duration(tcfg.PollSeconds) * time.Second,
		lastValue:    time.Now(),
		log:          log,
		stop:         make(chan struct{}),
	}, nil
}

// Watch watches the file for changes and sends the path to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *FileWatcher) Watch(changes chan interface{}) {
	w.log.Info("starting watcher")

	for {
		select {
		case <-w.stop:
			return
		case <-time.After(w.PollInterval):
			info, err := os.Stat(w.Path)
			if err != nil {
				w.log.Error("error getting file info",
					slog.String("path", w.Path), tint.Err(err))
			}
			if err == nil && info.ModTime().After(w.lastValue) {
				w.log.Info("file changed",
					slog.String("path", w.Path),
					slog.Time("mod_time", info.ModTime()))
				w.lastValue = info.ModTime()
				changes <- w.Path
			}
		}
	}
}

// Stop signals the watcher to stop
func (w *FileWatcher) Stop() {
	w.log.Info("shutting down watcher")
	close(w.stop)
}
