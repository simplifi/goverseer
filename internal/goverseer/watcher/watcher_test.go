package watcher

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewWatcher_DummyWatcher(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "dummy",
			Config: &config.DummyWatcherConfig{
				PollSeconds: 100,
			},
		},
		Executioner: config.DynamicExecutionerConfig{
			Type:   "dummy",
			Config: &config.DummyExecutionerConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	watcher, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &DummyWatcher{}, *watcher)
}

func TestNewWatcher_Unknown(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "foo",
		},
		Executioner: config.DynamicExecutionerConfig{
			Type:   "dummy",
			Config: &config.DummyExecutionerConfig{},
		},
	}

	_, err := New(cfg)
	assert.Error(t, err, "should throw an error for unknown watcher type")
}
