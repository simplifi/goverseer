package watcher

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewWatcher_DummyWatcher(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	watcherConfig := config.DummyWatcherConfig{
		PollMilliseconds: 100,
	}
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type:   "dummy",
			Config: watcherConfig,
		},
	}

	watcher, err := NewWatcher(cfg, log)
	assert.NoError(t, err)
	assert.IsType(t, &DummyWatcher{}, watcher)
}

func TestNewWatcher_FileWatcher(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	watcherConfig := config.FileWatcherConfig{
		Path:        "/path/to/file",
		PollSeconds: 5,
	}
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type:   "file",
			Config: watcherConfig,
		},
	}

	watcher, err := NewWatcher(cfg, log)
	assert.NoError(t, err)
	assert.IsType(t, &FileWatcher{}, watcher)
}

func TestNewWatcher_GceWatcher(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	watcherConfig := config.GceWatcherConfig{
		Source: "instance",
		Key:    "my-key",
	}
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type:   "gce",
			Config: watcherConfig,
		},
	}

	watcher, err := NewWatcher(cfg, log)
	assert.NoError(t, err)
	assert.IsType(t, &GCEWatcher{}, watcher)
}

func TestNewWatcher_Unknown(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := &config.Config{
		Watcher: config.DynamicWatcherConfig{
			Type: "foo",
		},
	}

	_, err := NewWatcher(cfg, log)
	assert.Error(t, err, "should throw an error for unknown watcher type")
}
