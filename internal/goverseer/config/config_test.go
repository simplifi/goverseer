package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testConfigWatcherToLog = `
name: WatcherToLog
watcher:
  time:
    poll_seconds: 1
executioner:
  log:
`
	testConfigGceToCommand = `
name: GceToCommand
watcher:
  time:
    poll_seconds: 1
executioner:
  log:
`
)

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

func TestFromFile(t *testing.T) {
	_, testConfig := writeTestConfigs(t, testConfigWatcherToLog)

	// Call the FromFile function
	config, err := FromFile(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, "WatcherToLog", config.Name)

	// Check the watcher config
	assert.Equal(t, "time", config.Watcher.Type)
	assert.IsType(t, &TimeWatcherConfig{}, config.Watcher.Config)

	// Check the executioner config
	assert.Equal(t, "log", config.Executioner.Type)
	assert.IsType(t, &LogExecutionerConfig{}, config.Executioner.Config)
}

func TestValidateAndSetDefaults(t *testing.T) {
	// A basic valid configuration
	config := &Config{
		Name: "TestConfig",
		Watcher: DynamicWatcherConfig{
			Type:   "time",
			Config: &TimeWatcherConfig{},
		},
		Executioner: DynamicExecutionerConfig{
			Type:   "log",
			Config: &LogExecutionerConfig{},
		},
	}
	err := config.ValidateAndSetDefaults()
	assert.NoError(t, err)
}
