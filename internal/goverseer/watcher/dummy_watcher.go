package watcher

import (
	"log/slog"
	"sync"
	"time"
)

var (
	// Ensure this implements the Watcher interface
	_ Watcher = (*DummyWatcher)(nil)
)

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

// NewDummyWatcher creates a new DummyWatcher
// The log is the logger
func NewDummyWatcher(PollMilliseconds int, log *slog.Logger) (*DummyWatcher, error) {
	w := &DummyWatcher{
		PollInterval: time.Duration(PollMilliseconds) * time.Millisecond,
		log:          log,
		stop:         make(chan struct{}),
	}

	return w, nil
}

// Watch watches the file for changes and sends the path to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *DummyWatcher) Watch(changes chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

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
