package gce_metadata_watcher

import (
	"context"
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

// TestParseConfig tests the ParseConfig function
func TestParseConfig(t *testing.T) {
	var parsedConfig *GceMetadataWatcherConfig

	testKey := "valid"
	parsedConfig, err := ParseConfig(map[string]interface{}{
		"key": testKey,
	})
	assert.NoError(t, err,
		"Parsing a valid config should not return an error")
	assert.Equal(t, testKey, parsedConfig.Key,
		"Key should be set to the value in the config")
	assert.Equal(t, DefaultSource, parsedConfig.Source,
		"Source should be set to the default")
	assert.Equal(t, DefaultRecursive, parsedConfig.Recursive,
		"Recursive should be set to the default")
	assert.Equal(t, DefaultMetadataUrl, parsedConfig.MetadataUrl,
		"MetadataUrl should be set to the default")
	assert.Equal(t, DefaultMetadataErrorWaitSeconds, parsedConfig.MetadataErrorWaitSeconds,
		"MetadataErrorWaitSeconds should be set to the default")

	// Test setting the source
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"key":    testKey,
		"source": "instance",
	})
	assert.NoError(t, err,
		"Parsing a config with a valid source should not return an error")
	assert.Equal(t, "instance", parsedConfig.Source,
		"Source should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"key":    testKey,
		"source": "this-is-wrong",
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect source value should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"key":    testKey,
		"source": 1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect source type should return an error")

	// Test setting recursive
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"key":       testKey,
		"recursive": true,
	})
	assert.NoError(t, err,
		"Parsing a config with a valid recursive value should not return an error")
	assert.Equal(t, true, parsedConfig.Recursive,
		"Recursive should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"key":       testKey,
		"recursive": 1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect recursive type should return an error")

	// Test setting key
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"key": testKey,
	})
	assert.NoError(t, err,
		"Parsing a config with a valid key should not return an error")
	assert.Equal(t, testKey, parsedConfig.Key,
		"Key should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"key": nil,
	})
	assert.Error(t, err,
		"Parsing a config with no key should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"key": "",
	})
	assert.Error(t, err,
		"Parsing a config with an empty key should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"key": 1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect key type should return an error")

	// Test setting the metadata_url
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"key":          testKey,
		"metadata_url": "http://localhost:8888/computeMetadata/v1",
	})
	assert.NoError(t, err,
		"Parsing a config with a valid metadata_url should not return an error")
	assert.Equal(t, "http://localhost:8888/computeMetadata/v1", parsedConfig.MetadataUrl,
		"MetadataUrl should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"key":          testKey,
		"metadata_url": "",
	})
	assert.Error(t, err,
		"Parsing a config with an empty metadata_url should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"key":          testKey,
		"metadata_url": 1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect metadata_url type should return an error")

	// Test setting the metadata_error_wait_seconds
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"key":                         testKey,
		"metadata_error_wait_seconds": 10,
	})
	assert.NoError(t, err,
		"Parsing a config with a valid metadata_error_wait_seconds should not return an error")
	assert.Equal(t, 10, parsedConfig.MetadataErrorWaitSeconds,
		"MetadataUrl should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"key":          testKey,
		"metadata_url": "",
	})
	assert.Error(t, err,
		"Parsing a config with an empty metadata_url should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"key":          testKey,
		"metadata_url": 1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect metadata_url type should return an error")
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	var cfg config.Config
	cfg = config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "gce_metadata",
			Config: map[string]interface{}{
				"key": "test",
			},
		},
	}
	watcher, err := New(cfg, nil)
	assert.NoError(t, err,
		"Creating a new GceMetadataWatcher should not return an error")
	assert.NotNil(t, watcher,
		"Creating a new GceMetadataWatcher should return a watcher")

	cfg = config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "gce_metadata",
			Config: map[string]interface{}{
				"key": nil,
			},
		},
	}
	watcher, err = New(cfg, nil)
	assert.Error(t, err,
		"Creating a new GceMetadataWatcher with an invalid config should return an error")
	assert.Nil(t, watcher,
		"Creating a new GceMetadataWatcher with an invalid config should not return a watcher")
}

func TestGceMetadataWatcher_Watch(t *testing.T) {
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

	ctx, cancel := context.WithCancel(context.Background())

	watcher := GceMetadataWatcher{
		Key:               "test",
		Recursive:         true,
		MetadataUrl:       mockServer.URL,
		MetadataErrorWait: 1 * time.Second,
		log:               log,
		ctx:               ctx,
		cancel:            cancel,
	}

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

func TestGceMetadataWatcher_Stop(t *testing.T) {
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

	ctx, cancel := context.WithCancel(context.Background())

	watcher := GceMetadataWatcher{
		Key:               "test",
		Recursive:         true,
		MetadataUrl:       mockServer.URL,
		MetadataErrorWait: 1 * time.Second,
		log:               log,
		ctx:               ctx,
		cancel:            cancel,
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

	// Simulate a change by closing the mock response channel, which will unblock
	// the request
	close(mockResponseChan)

	// Assert that the change was NOT received
	select {
	case <-changes:
		assert.Fail(t, "Received change after stopping watcher")
	case <-time.After(1 * time.Second):
		// Success
	}
}
