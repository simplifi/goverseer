package executioner

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutioner_LogExecutioner(t *testing.T) {
	cfg := &config.Config{
		Executioner: config.DynamicExecutionerConfig{
			Type:   "log",
			Config: &config.LogExecutionerConfig{},
		},
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}(map[string]interface{}{
				"poll_seconds": 1,
			}),
		},
	}
	cfg.ValidateAndSetDefaults()

	executioner, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &LogExecutioner{}, executioner)
}

func TestNewExecutioner_Unknown(t *testing.T) {
	cfg := &config.Config{
		Executioner: config.DynamicExecutionerConfig{
			Type: "foo",
		},
		Watcher: config.WatcherConfig{
			Type:   "time",
			Config: map[string]interface{}(map[string]interface{}{"poll_seconds": 1}),
		},
	}

	_, err := New(cfg)
	assert.Error(t, err, "should throw an error for unknown executioner type")
}
