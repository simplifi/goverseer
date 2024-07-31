package watcher

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewWatcher_TimeWatcher(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "time",
			Config: &config.TimeWatcherConfig{
				PollSeconds: 100,
			},
		},
		Executioner: config.DynamicExecutionerConfig{
			Type:   "log",
			Config: &config.LogExecutionerConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	watcher, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &TimeWatcher{}, watcher)
}

func TestNewWatcher_Unknown(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "foo",
		},
		Executioner: config.DynamicExecutionerConfig{
			Type:   "log",
			Config: &config.LogExecutionerConfig{},
		},
	}

	_, err := New(cfg)
	assert.Error(t, err, "should throw an error for unknown watcher type")
}
