package file_watcher

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func touchFile(t *testing.T, filename string) {
	t.Helper()

	currentTime := time.Now()

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		if err := file.Close(); err != nil {
			t.Fatalf("Failed to close temporary file: %v", err)
		}
	}

	if err := os.Chtimes(filename, currentTime, currentTime); err != nil {
		t.Fatalf("Failed to change file times: %v", err)
	}
}

func TestParseConfig(t *testing.T) {
	var parsedConfig *Config
	var err error

	parsedConfig, err = ParseConfig(map[string]interface{}{
		"path":         "/tmp/test",
		"poll_seconds": DefaultPollSeconds,
	})
	assert.NoError(t, err,
		"Parsing a valid config should not return an error")
	assert.Equal(t, DefaultPollSeconds, parsedConfig.PollSeconds,
		"PollSeconds should be set to the value in the config")

	// Test setting the path
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"path": "/tmp/test",
	})
	assert.NoError(t, err,
		"Parsing a config with a valid path should not return an error")
	assert.Equal(t, "/tmp/test", parsedConfig.Path,
		"Path should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"path": 9,
	})
	assert.Error(t, err,
		"Parsing a config with an invalid path should return an error")

	// Test setting PollSeconds
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"path":         "/tmp/test",
		"poll_seconds": 10,
	})
	assert.NoError(t, err,
		"Parsing a config with valid poll_seconds should not return an error")
	assert.Equal(t, 10, parsedConfig.PollSeconds,
		"PollSeconds should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"path":         "/tmp/test",
		"poll_seconds": 0,
	})
	assert.Error(t, err,
		"Parsing a config with poll_seconds less than 1 should return an error")
}

func TestNew(t *testing.T) {
	var cfg config.Config
	cfg = config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "file",
			Config: map[string]interface{}{
				"path": "/tmp/test",
			},
		},
	}
	watcher, err := New(cfg, nil)
	assert.NoError(t, err,
		"Creating a new FileWatcher should not return an error")
	assert.NotNil(t, watcher,
		"Creating a new FileWatcher should return a watcher")

	cfg = config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "file",
			Config: map[string]interface{}{
				"path": nil,
			},
		},
	}
	watcher, err = New(cfg, nil)
	assert.Error(t, err,
		"Creating a new FileWatcher with an invalid config should return an error")
	assert.Nil(t, watcher,
		"Creating a new FileWatcher with an invalid config should not return a watcher")
}

func TestFileWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	// Create a temp file we can watch
	testFilePath := filepath.Join(t.TempDir(), "test.txt")
	touchFile(t, testFilePath)

	cfg := config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "file",
			Config: map[string]interface{}{
				"path":         testFilePath,
				"poll_seconds": 1, // Set a short poll interval for testing
			},
		},
		Executioner: config.ExecutionerConfig{},
	}

	// Create a channel to receive changes
	changes := make(chan interface{})
	wg := &sync.WaitGroup{}

	// Create a new FileWatcher
	watcher, err := New(cfg, log)
	assert.NoError(t, err)

	// Start watching the file
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Watch(changes)
	}()

	// Touch the file to trigger a change
	touchFile(t, testFilePath)

	// Assert that the tick was detected
	// We limit the time to avoid hanging tests
	select {
	case path := <-changes:
		assert.Equal(t, testFilePath, path)
	case <-time.After(2 * time.Second):
		assert.Fail(t, "Timed out waiting for file change")
	}

	// Stop the watcher
	watcher.Stop()
	wg.Wait()
}

func TestFileWatcher_Stop(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	// Create a temp file we can watch
	testFilePath := filepath.Join(t.TempDir(), "test.txt")
	touchFile(t, testFilePath)

	watcher := FileWatcher{
		Config: Config{
			Path:        testFilePath,
			PollSeconds: 1,
		},
		lastValue: time.Now(),
		log:       log,
		stop:      make(chan struct{}),
	}

	changes := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Watch(changes)
	}()

	// Stop the watcher and wait
	watcher.Stop()
	wg.Wait()

	// Trigger a change by touching the test file
	touchFile(t, testFilePath)

	// Assert that the change was NOT received
	select {
	case <-changes:
		assert.Fail(t, "Received change after stopping watcher")
	case <-time.After(1 * time.Second):
		// Success
	}
}
