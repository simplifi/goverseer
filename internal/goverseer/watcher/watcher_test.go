package watcher

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

// TestWatcher_New tests the New function
func TestWatcher_New(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}(map[string]interface{}{
				"poll_seconds": 1,
			}),
		},
		Executioner: config.ExecutionerConfig{
			Type: "log",
			Config: map[string]interface{}(map[string]interface{}{
				"tag": "test",
			}),
		},
	}
	// A valid configuration should not return an error
	watcher, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &TimeWatcher{}, watcher)

	// An invalid configuration should return an error
	cfg.Watcher.Type = "foo"
	_, err = New(cfg)
	assert.Error(t, err, "should throw an error for unknown watcher type")
}
