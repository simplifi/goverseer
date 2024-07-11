package watcher

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Watcher interface
	_ Watcher = (*DummyWatcher)(nil)
)

func init() {
	RegisterWatcher("dummy", func() Watcher { return &DummyWatcher{} })
}

// DummyWatcher is a dummy watcher that ticks at a regular interval, not
// watching anything, like a dummy
type DummyWatcher struct {
	// PollInterval is the interval to to wait between ticks
	PollInterval time.Duration

	// log is the logger
	log *slog.Logger

	// stop is a channel to signal the watcher to stop
	stop chan struct{}
}

func (w *DummyWatcher) Create(cfg config.WatcherConfig, log *slog.Logger) error {
	v, ok := cfg.(*config.DummyWatcherConfig)
	if !ok {
		return fmt.Errorf("invalid config for dummy watcher, got %T", cfg)
	}

	w.PollInterval = time.Duration(v.PollSeconds) * time.Second
	w.log = log
	w.stop = make(chan struct{})

	return nil
}

// Watch watches the file for changes and sends the path to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *DummyWatcher) Watch(changes chan interface{}) {
	w.log.Info("starting watcher")
	for {
		select {
		case <-w.stop:
			return
		case value := <-time.After(w.PollInterval):
			w.log.Info("dummy watcher tick", slog.Time("value", value))
			changes <- value
		}
	}
}

// Stop signals the watcher to stop
func (w *DummyWatcher) Stop() {
	w.log.Info("shutting down watcher")
	close(w.stop)
}
