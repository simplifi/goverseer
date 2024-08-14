package watcher

import (
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestTimeWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := &config.TimeWatcherConfig{}
	cfg.ValidateAndSetDefaults()

	// Create a channel to receive changes
	changes := make(chan interface{})
	wg := &sync.WaitGroup{}

	// Create a new TimeWatcher
	watcher, err := newTimeWatcher(cfg, log)
	assert.NoError(t, err)
	t.Log(watcher.PollInterval)
	// Start watching the file
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Watch(changes)
	}()

	// Assert that the tick was detected
	// We limit the time to avoid hanging tests
	select {
	case value := <-changes:
		assert.NotEmpty(t, value)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "Timed out waiting for file change")
	}

	// Stop the watcher
	watcher.Stop()
	wg.Wait()
}
