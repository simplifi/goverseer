package watcher

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/lmittmann/tint"
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

func TestFileWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	// Create a temp file we can watch
	testFilePath := filepath.Join(t.TempDir(), "test.txt")
	touchFile(t, testFilePath)

	// Create a channel to receive file changes
	changes := make(chan interface{})
	wg := &sync.WaitGroup{}

	// Create a new FileWatcher
	fileWatcher, err := NewFileWatcher(testFilePath, 1, log)
	assert.NoError(t, err)

	// Start watching the file
	wg.Add(1)
	go fileWatcher.Watch(changes, wg)

	// Touch the file to trigger a change
	touchFile(t, testFilePath)

	// Assert that the file change was detected
	// We limit the time to avoid hanging tests
	select {
	case path := <-changes:
		assert.Equal(t, testFilePath, path)
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timed out waiting for file change")
	}

	// Stop the watcher
	fileWatcher.Stop()
	wg.Wait()
}

func TestNewFileWatcher(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	path := "/path/to/file.txt"
	pollSeconds := 1
	watcher, err := NewFileWatcher(path, pollSeconds, log)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.Equal(t, path, watcher.Path)
	assert.Equal(t, time.Duration(pollSeconds)*time.Second, watcher.PollInterval)
}
