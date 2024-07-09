package executor

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutor_DummyExecutor(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	executorConfig := config.DummyExecutorConfig{}
	cfg := &config.Config{
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: executorConfig,
		},
	}

	watcher, err := NewExecutor(cfg, log)
	assert.NoError(t, err)
	assert.IsType(t, &DummyExecutor{}, watcher)
}

func TestNewExecutor_CommandExecutor(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	executorConfig := config.CommandExecutorConfig{
		Command: "echo 'Hello, World!'",
	}
	cfg := &config.Config{
		Executor: config.DynamicExecutorConfig{
			Type:   "command",
			Config: executorConfig,
		},
	}

	watcher, err := NewExecutor(cfg, log)
	assert.NoError(t, err)
	assert.IsType(t, &CommandExecutor{}, watcher)
}

func TestNewExecutor_Unknown(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := &config.Config{
		Executor: config.DynamicExecutorConfig{
			Type: "foo",
		},
	}

	_, err := NewExecutor(cfg, log)
	assert.Error(t, err, "should throw an error for unknown executor type")
}
