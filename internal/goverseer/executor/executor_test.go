package executor

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutor_DummyExecutor(t *testing.T) {
	cfg := &config.Config{
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: &config.DummyExecutorConfig{},
		},
		Watcher: config.DynamicWatcherConfig{
			Type:   "dummy",
			Config: &config.DummyWatcherConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	executor, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &DummyExecutor{}, *executor)
}

func TestNewExecutor_CommandExecutor(t *testing.T) {
	cfg := &config.Config{
		Executor: config.DynamicExecutorConfig{
			Type: "command",
			Config: &config.CommandExecutorConfig{
				Command: "echo 'Hello, World!'",
			},
		},
		Watcher: config.DynamicWatcherConfig{
			Type:   "dummy",
			Config: &config.DummyWatcherConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	executor, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &CommandExecutor{}, *executor)
}

func TestNewExecutor_Unknown(t *testing.T) {
	cfg := &config.Config{
		Executor: config.DynamicExecutorConfig{
			Type: "foo",
		},
		Watcher: config.DynamicWatcherConfig{
			Type:   "dummy",
			Config: &config.DummyWatcherConfig{},
		},
	}

	_, err := New(cfg)
	assert.Error(t, err, "should throw an error for unknown executor type")
}
