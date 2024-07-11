package executioner

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestNewExecutioner_DummyExecutioner(t *testing.T) {
	cfg := &config.Config{
		Executioner: config.DynamicExecutionerConfig{
			Type:   "dummy",
			Config: &config.DummyExecutionerConfig{},
		},
		Watcher: config.DynamicWatcherConfig{
			Type:   "dummy",
			Config: &config.DummyWatcherConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	executioner, err := New(cfg)
	assert.NoError(t, err)
	assert.IsType(t, &DummyExecutioner{}, *executioner)
}

func TestNewExecutioner_Unknown(t *testing.T) {
	cfg := &config.Config{
		Executioner: config.DynamicExecutionerConfig{
			Type: "foo",
		},
		Watcher: config.DynamicWatcherConfig{
			Type:   "dummy",
			Config: &config.DummyWatcherConfig{},
		},
	}

	_, err := New(cfg)
	assert.Error(t, err, "should throw an error for unknown executioner type")
}
