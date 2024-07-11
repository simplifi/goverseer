package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testConfigWatcherToDummy = `
name: WatcherToDummy
watcher:
  file:
    path: /tmp/test1
executioner:
  dummy:
`
	testConfigGceToCommand = `
name: GceToCommand
watcher:
  gce:
    source: instance
    key: foo
executioner:
  command:
    command: echo "Hello, World!"
`
)

func writeTestConfigs(t *testing.T, files ...string) (string, []string) {
	t.Helper()

	testConfigDir := t.TempDir()
	testConfigFiles := make([]string, 0, len(files))

	// Write out the temporary configuration files
	for n, content := range files {
		configFile := filepath.Join(testConfigDir, fmt.Sprintf("test%d.yaml", n))
		testConfigFiles = append(testConfigFiles, configFile)
		err := os.WriteFile(configFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary configuration file: %v", err)
		}
	}

	// Return the path list of files
	return testConfigDir, testConfigFiles
}

func TestFromPath(t *testing.T) {
	testConfigDir, testConfigs := writeTestConfigs(t,
		testConfigWatcherToDummy,
		testConfigGceToCommand,
	)

	// Call the FromPath function
	configs, err := FromPath(testConfigDir)
	assert.NoError(t, err)
	assert.Len(t, configs, len(testConfigs))
}

func TestFromFile(t *testing.T) {
	_, testConfigs := writeTestConfigs(t, testConfigWatcherToDummy)
	configFile := testConfigs[0]

	// Call the FromFile function
	config, err := FromFile(configFile)
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
		ChangeBuffer: 10,
	}
	err := config.ValidateAndSetDefaults()
	assert.NoError(t, err)

	// Invalid configuration, ChangeBuffer less than 0
	config.ChangeBuffer = -1
	err = config.ValidateAndSetDefaults()
	assert.Error(t, err)

	// Check defaults are set, ChangeBuffer missing (0) so it should get the default
	config.ChangeBuffer = 0
	err = config.ValidateAndSetDefaults()
	assert.NoError(t, err)
	assert.Equal(t, 100, config.ChangeBuffer)
}
