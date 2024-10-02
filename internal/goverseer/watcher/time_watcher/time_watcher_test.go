package time_watcher

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

// TestTimeWatcher_ParseConfig tests the ParseConfig function
func TestTimeWatcher_ParseConfig(t *testing.T) {
	// Unmarshalling a valid config should not return an error
	validConfig := map[string]interface{}{
		"poll_seconds": 1,
	}

	cfg, err := ParseConfig(validConfig)
	assert.NoError(t, err)
	assert.Equal(t, 1, cfg.PollSeconds)

	// Unmarshalling an invalid config should return an error
	invalidConfig := map[string]interface{}{
		"poll_seconds": "invalid",
	}

	_, err = ParseConfig(invalidConfig)
	assert.Error(t, err)

	// Unmarshalling a config with a poll_seconds less than 1 should return an error
	invalidConfig = map[string]interface{}{
		"poll_seconds": 0,
	}

	_, err = ParseConfig(invalidConfig)
	assert.Error(t, err)

	// Unmarshalling a config without poll_seconds should return a default value
	emptyConfig := map[string]interface{}{}
	cfg, err = ParseConfig(emptyConfig)
	assert.NoError(t, err)
	assert.Equal(t, DefaultPollSeconds, cfg.PollSeconds)
}

// TestTimeWatcher_Watch tests the Watch function
func TestTimeWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}{
				"poll_seconds": 1,
			},
		},
		Executioner: config.ExecutionerConfig{},
	}

	// Create a channel to receive changes
	changes := make(chan interface{})
	wg := &sync.WaitGroup{}

	// Create a new TimeWatcher
	watcher, err := New(cfg, log)
	assert.NoError(t, err)
	t.Log(watcher.PollSeconds)
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
