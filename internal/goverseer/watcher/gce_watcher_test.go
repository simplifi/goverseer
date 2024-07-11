package watcher

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestGCEWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	mockResponseChan := make(chan struct{})
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait for a signal, this is used to simulate a change in the metadata
		// value on GCE during tests
		<-mockResponseChan
		w.Header().Add("ETag", "mock-etag")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock response"))
	}))
	defer mockServer.Close()

	cfg := &config.GCEWatcherConfig{
		Source:      "instance",
		Key:         "test-key",
		MetadataUrl: mockServer.URL,
		Recursive:   true,
	}
	cfg.ValidateAndSetDefaults()

	watcher := GCEWatcher{}
	err := watcher.Create(cfg, log)
	assert.NoError(t, err)

	changes := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watcher.Watch(changes)
	}()

	// Simulate a change by closing the mock response channel, which will unblock
	// the request
	close(mockResponseChan)

	// Assert that the change was received
	select {
	case value := <-changes:
		assert.Equal(t, "mock response", value)
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Timed out waiting for change")
	}

	// Stop the watcher and wait
	watcher.Stop()
	wg.Wait()
}
