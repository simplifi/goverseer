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
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: &config.DummyExecutorConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	watcher, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &DummyWatcher{}, *watcher)
}

func TestNewWatcher_FileWatcher(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "file",
			Config: &config.FileWatcherConfig{
				Path:        "/path/to/file",
				PollSeconds: 5,
			},
		},
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: &config.DummyExecutorConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	watcher, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &FileWatcher{}, *watcher)
}

func TestNewWatcher_GceWatcher(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "gce",
			Config: &config.GCEWatcherConfig{
				Source: "instance",
				Key:    "my-key",
			},
		},
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: &config.DummyExecutorConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	watcher, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &GCEWatcher{}, *watcher)
}

func TestNewWatcher_Unknown(t *testing.T) {
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "foo",
		},
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: &config.DummyExecutorConfig{},
		},
	}

	_, err := New(cfg)
	assert.Error(t, err, "should throw an error for unknown watcher type")
}
