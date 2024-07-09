package watcher

import (
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/assert"
)

func TestDummyWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	// Create a channel to receive changes
	changes := make(chan interface{})
	wg := &sync.WaitGroup{}

	// Create a new DummyWatcher
	dummyWatcher, err := NewDummyWatcher(1, log)
	assert.NoError(t, err)

	// Start watching the file
	wg.Add(1)
	go dummyWatcher.Watch(changes, wg)

	// Assert that the tick was detected
	// We limit the time to avoid hanging tests
	select {
	case value := <-changes:
		assert.NotEmpty(t, value)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "Timed out waiting for file change")
	}

	// Stop the watcher
	dummyWatcher.Stop()
	wg.Wait()
}

func TestNewDummyWatcher(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	pollMilliseconds := 100
	watcher, err := NewDummyWatcher(pollMilliseconds, log)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.Equal(t, time.Duration(pollMilliseconds)*time.Millisecond, watcher.PollInterval)
}
