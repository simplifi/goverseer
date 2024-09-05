package executioner

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

// TestExecutioner_New tests the New function
func TestExecutioner_New(t *testing.T) {
	cfg := &config.Config{
		Executioner: config.ExecutionerConfig{
			Type: "log",
			Config: map[string]interface{}(map[string]interface{}{
				"tag": "test",
			}),
		},
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}(map[string]interface{}{
				"poll_seconds": 1,
			}),
		},
	}

	// A valid configuration should not return an error
	executioner, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &LogExecutioner{}, executioner)

	// An invalid configuration should return an error
	cfg.Executioner.Type = "foo"
	_, err = New(cfg)
	assert.Error(t, err, "should throw an error for unknown executioner type")
}
