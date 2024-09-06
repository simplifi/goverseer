package log

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

// TestLogExecutioner_ParseConfig tests the ParseConfig function
func TestLogExecutioner_ParseConfig(t *testing.T) {
	// Unmarshalling a valid config should not return an error
	validConfig := map[string]interface{}{
		"tag": "test",
	}

	cfg, err := ParseConfig(validConfig)
	assert.NoError(t, err)
	assert.Equal(t, "test", cfg.Tag)

	// Unmarshalling a config without tag should return a default value
	emptyConfig := map[string]interface{}{}
	cfg, err = ParseConfig(emptyConfig)
	assert.NoError(t, err)
	assert.Equal(t, DefaultTag, cfg.Tag)
}

// TestLogExecutioner_Execute tests the Execute method of LogExecutioner
func TestLogExecutioner_Execute(t *testing.T) {
	logger := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := config.Config{
		Name: "TestConfig",
		Watcher: config.WatcherConfig{
			Type: "time",
		},
		Executioner: config.ExecutionerConfig{
			Type: "log",
			Config: map[string]interface{}{
				"tag": "test",
			},
		},
	}

	executioner, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create LogExecutioner: %v", err)
	}

	// Executing a valid config should not return an error
	err = executioner.Execute("foo")
	assert.NoError(t, err)
}
