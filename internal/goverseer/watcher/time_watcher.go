package watcher

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Watcher interface
	_ Watcher = (*TimeWatcher)(nil)
)

// TimeWatcher is a time watcher that ticks at a regular interval
type TimeWatcher struct {
	// PollInterval is the interval to to wait between ticks
	PollInterval time.Duration

	// log is the logger
	log *slog.Logger

	// stop is a channel to signal the watcher to stop
	stop chan struct{}
}

func (w *TimeWatcher) Create(cfg config.WatcherConfig, log *slog.Logger) error {
	v, ok := cfg.(*config.TimeWatcherConfig)
	if !ok {
		return fmt.Errorf("invalid config for time watcher, got %T", cfg)
	}

	w.PollInterval = time.Duration(v.PollSeconds) * time.Second
	w.log = log
	w.stop = make(chan struct{})

	return nil
}

// Watch ticks at regular intervals, sending the time to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *TimeWatcher) Watch(change chan interface{}) {
	w.log.Info("starting watcher")
	for {
		select {
		case <-w.stop:
			return
		case value := <-time.After(w.PollInterval):
			w.log.Info("time watcher tick", slog.Time("value", value))
			change <- value
		}
	}
}

// Stop signals the watcher to stop
func (w *TimeWatcher) Stop() {
	w.log.Info("shutting down watcher")
	close(w.stop)
}
