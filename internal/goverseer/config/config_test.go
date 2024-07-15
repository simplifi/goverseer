package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testConfigWatcherToDummy = `
name: WatcherToDummy
watcher:
  dummy:
    poll_seconds: 1
executioner:
  dummy:
`
	testConfigGceToCommand = `
name: GceToCommand
watcher:
  dummy:
    poll_seconds: 1
executioner:
  dummy:
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
	_, testConfig := writeTestConfigs(t, testConfigWatcherToDummy)

	// Call the FromFile function
	config, err := FromFile(testConfig)
	assert.NoError(t, err)
	assert.Equal(t, "WatcherToDummy", config.Name)

	// Check the watcher config
	assert.Equal(t, "dummy", config.Watcher.Type)
	assert.IsType(t, &DummyWatcherConfig{}, config.Watcher.Config)

	// Check the executioner config
	assert.Equal(t, "dummy", config.Executioner.Type)
	assert.IsType(t, &DummyExecutionerConfig{}, config.Executioner.Config)
}

func TestValidateAndSetDefaults(t *testing.T) {
	// A basic valid configuration
	config := &Config{
		Name: "TestConfig",
		Watcher: DynamicWatcherConfig{
			Type:   "dummy",
			Config: &DummyWatcherConfig{},
		},
		Executioner: DynamicExecutionerConfig{
			Type:   "dummy",
			Config: &DummyExecutionerConfig{},
		},
	}
	err := config.ValidateAndSetDefaults()
	assert.NoError(t, err)
}
