package watcher

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/assert"
)

func TestGCEWatcher_Watch(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	mockResponseChan := make(chan struct{})
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If query param has `wait_for_change` then wait for a signal
		// This is used to simulate a change in the metadata value on GCE
		if r.URL.Query().Get("wait_for_change") == "true" {
			<-mockResponseChan
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock response"))
	}))
	defer mockServer.Close()

	// Create a Resty client with the mock server's URL
	mockClient := resty.New().SetBaseURL(mockServer.URL)
	ctx, cancel := context.WithCancel(context.Background())

	w := &GCEWatcher{
		Key:       "test-key",
		Recursive: true,
		client:    mockClient,
		lastValue: "",
		log:       log,
		ctx:       ctx,
		cancel:    cancel,
	}

	changes := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		w.Watch(changes, &wg)
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
