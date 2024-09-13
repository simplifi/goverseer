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

	// Check the executioner config
	assert.Equal(t, "log", config.Executioner.Type)
	assert.IsType(t, map[string]interface{}{"tag": "test"}, config.Executioner.Config)
}
