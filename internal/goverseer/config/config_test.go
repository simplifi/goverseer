package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// testConfigWatcherToLog is a basic test configuration for a time watcher and
	// a log executioner
	testConfigWatcherToLog = `
name: WatcherToLog
watcher:
  type: time
  config:
    poll_seconds: 1
executioner:
  type: log
  config:
    tag: test
`
	// testConfigWatcherToLogNoConfig is a basic test configuration for a time
	// watcher and a log executioner with no configuration provided
	testConfigWatcherToLogNoConfig = `
name: WatcherToLog
watcher:
  type: time
executioner:
  type: log
`
	// testConfigLogger is a basic test configuration for testing
	// the main logger configuration
	testConfigLogger = `
name: Logger
logger:
  level: debug
watcher:
  type: time
executioner:
  type: log
`
	testConfigInvalidWatcherType = `
name: InvalidWatcher
watcher:
  type: unknown
executioner:
  type: log
`
	testConfigUnknownField = `
name: UnknownField
unknown: true
watcher:
  type: time
executioner:
  type: log
`
)

// writeTestConfigs writes test configurations to a temporary directory
// It returns the path to the directory and the path to the configuration file
func writeTestConfigs(t *testing.T, content string) (string, string) {
	t.Helper()

	testConfigDir := t.TempDir()
	configFile := filepath.Join(testConfigDir, "test.yaml")
	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary configuration file: %v", err)
	}

	// Return the path list of files
	return testConfigDir, configFile
}

// TestFromFile tests the FromFile function
func TestFromFile(t *testing.T) {
	_, testConfig := writeTestConfigs(t, testConfigWatcherToLog)

	// Call the FromFile function
	config, err := FromFile(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, "WatcherToLog", config.Name)

	// Check the watcher config
	assert.Equal(t, "time", config.Watcher.Type)
	assert.IsType(t, map[string]interface{}{"poll_seconds": 1}, config.Watcher.Config)
	assert.Equal(t, 1, config.Watcher.Config["poll_seconds"])

	// Check the executioner config
	assert.Equal(t, "log", config.Executioner.Type)
	assert.IsType(t, map[string]interface{}{"tag": "test"}, config.Executioner.Config)
	assert.Equal(t, "test", config.Executioner.Config["tag"])

	// Test with a config that's missing non-required Configs
	_, testConfig = writeTestConfigs(t, testConfigWatcherToLogNoConfig)
	config, err = FromFile(testConfig)
	assert.NoError(t, err,
		"Parsing a config file with no Configs should not error")
	assert.Equal(t, map[string]interface{}(nil), config.Executioner.Config,
		"An executioner with no Config should have an empty map for the value")
	assert.Equal(t, map[string]interface{}(nil), config.Watcher.Config,
		"A watcher with no Config should have an empty map for the value")

	// Test with a config with logger configuration
	_, testConfig = writeTestConfigs(t, testConfigLogger)
	config, err = FromFile(testConfig)
	assert.NoError(t, err,
		"Parsing a config file with a valid logger config should not error")
	assert.Equal(t, "debug", config.Logger.Level,
		"An config file with logger configuration should parse correctly")

	// Test with an invalid watcher type
	_, testConfig = writeTestConfigs(t, testConfigInvalidWatcherType)
	config, err = FromFile(testConfig)
	assert.Error(t, err,
		"Parsing a config file with an invalid watcher type should error")
	assert.Nil(t, config,
		"Parsing a config file with an invalid watcher type should not return config")

	// Test with an unknown top-level field
	_, testConfig = writeTestConfigs(t, testConfigUnknownField)
	config, err = FromFile(testConfig)
	assert.Error(t, err,
		"Parsing a config file with an unknown top-level field should error")
	assert.Nil(t, config,
		"Parsing a config file with an unknown top-level field should not return config")
}

func TestDecode(t *testing.T) {
	type decodeConfig struct {
		Name    string `mapstructure:"name" validate:"required"`
		Count   int    `mapstructure:"count" validate:"gte=1"`
		Enabled bool   `mapstructure:"enabled"`
	}

	cfg := &decodeConfig{
		Count: 1,
	}

	err := Decode(map[string]interface{}{
		"name":    "test",
		"enabled": true,
	}, cfg)
	assert.NoError(t, err,
		"Decoding a valid dynamic config should not error")
	assert.Equal(t, "test", cfg.Name,
		"Decode should set configured fields")
	assert.Equal(t, 1, cfg.Count,
		"Decode should preserve default values for missing fields")
	assert.Equal(t, true, cfg.Enabled,
		"Decode should set configured booleans")

	cfg = &decodeConfig{
		Count: 1,
	}
	err = Decode(map[string]interface{}{
		"name": 9,
	}, cfg)
	assert.Error(t, err,
		"Decoding a config with an invalid field type should error")
	assert.Contains(t, err.Error(), "name must be a string")

	cfg = &decodeConfig{
		Count: 1,
	}
	err = Decode(map[string]interface{}{
		"name":  "",
		"count": 0,
	}, cfg)
	assert.Error(t, err,
		"Decoding a config with an empty required string should error")
	assert.Contains(t, err.Error(), "name must not be empty")
}
