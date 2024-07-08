package watcher

import (
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/lmittmann/tint"
)

var (
	// Ensure this implements the Watcher interface
	_ Watcher = (*FileWatcher)(nil)
)

// FileWatcher watches a file for changes and sends the path to the changeCh
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

// NewFileWatcher creates a new FileWatcher
// The Path is the path to the file to watch
// The PollSeconds is the interval to poll the file for changes
// The log is the logger
func NewFileWatcher(Path string, PollSeconds int, log *slog.Logger) (*FileWatcher, error) {
	w := &FileWatcher{
		Path:         Path,
		PollInterval: time.Duration(PollSeconds) * time.Second,
		lastValue:    time.Now(),
		log:          log,
		stop:         make(chan struct{}),
	}

	return w, nil
}

// Watch watches the file for changes and sends the path to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *FileWatcher) Watch(changes chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	w.log.Info("starting watcher")
	for {
		select {
		case <-w.stop:
			return
		default:
			info, err := os.Stat(w.Path)
			if err != nil {
				w.log.Error("error getting file info", slog.String("path", w.Path), tint.Err(err))
			}
			if err == nil && info.ModTime().After(w.lastValue) {
				w.log.Info("file changed", slog.String("path", w.Path), slog.Time("mod_time", info.ModTime()))
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
