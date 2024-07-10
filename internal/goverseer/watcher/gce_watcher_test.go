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

	w, _ := NewGCEWatcher("instance", "test-key", mockServer.URL, true, log)

	changes := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Watch(changes)
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
	w.Stop()
	wg.Wait()
}

func TestNewGCEWatcher(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	// Test with valid source
	w, err := NewGCEWatcher("instance", "test-key", "", true, log)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	// Test with invalid source
	w, err = NewGCEWatcher("invalid-source", "test-key", "", true, log)
	assert.Error(t, err)
	assert.Nil(t, w)
	assert.EqualError(t, err, "source must be either 'instance' or 'project'")
}
