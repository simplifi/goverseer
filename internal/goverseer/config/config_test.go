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
  type: file
  config:
    path: /tmp/test1
executor:
  type: dummy
`
	testConfigGceToCommand = `
name: GceToCommand
watcher:
  type: gce
  config:
    source: instance
    key: foo
executor:
  type: command
  config:
    command: echo "Hello, World!"
`
)

func writeTestConfigs(t *testing.T, files map[string]string) (string, []string) {
	t.Helper()

	// Create a temporary directory for the config file
	testConfigDir := t.TempDir()

	// Create the temporary configuration files
	testConfigFiles := make([]string, 0, len(files))
	for name, content := range files {
		configFile := filepath.Join(testConfigDir, name)
		testConfigFiles = append(testConfigFiles, configFile)
		err := os.WriteFile(configFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create temporary configuration file: %v", err)
		}
	}

	// Return the path and a cleanup function
	return testConfigDir, testConfigFiles
}

func TestFromPath(t *testing.T) {
	testConfigDir, testConfigs := writeTestConfigs(t, map[string]string{
		"watcher-to-dummy.yaml": testConfigWatcherToDummy,
		"gce-to-command.yaml":   testConfigGceToCommand,
	})

	// Call the FromPath function
	configs, err := FromPath(testConfigDir)
	assert.NoError(t, err)
	assert.Len(t, configs, len(testConfigs))
}

func TestFromFile(t *testing.T) {
	_, testConfigs := writeTestConfigs(t, map[string]string{
		"watcher-to-dummy.yaml": testConfigWatcherToDummy,
	})
	configFile := testConfigs[0]

	// Call the FromFile function
	config, err := FromFile(configFile)
	assert.NoError(t, err)
	assert.Equal(t, "WatcherToDummy", config.Name)

	// Check the watcher config
	assert.Equal(t, "file", config.Watcher.Type)
	assert.IsType(t, FileWatcherConfig{}, config.Watcher.Config)

	// Check the executor config
	assert.Equal(t, "dummy", config.Executor.Type)
	assert.IsType(t, DummyExecutorConfig{}, config.Executor.Config)
}

func TestValidate(t *testing.T) {
	// A basic valid configuration
	config := &Config{
		Name: "TestConfig",
		Watcher: DynamicWatcherConfig{
			Type:   "file",
			Config: "",
		},
		Executor: DynamicExecutorConfig{
			Type:   "dummy",
			Config: "",
		},
		ChangeBuffer: 10,
	}
	err := config.Validate()
	assert.NoError(t, err)

	// Missing required field
	config = &Config{
		Watcher:      DynamicWatcherConfig{},
		Executor:     DynamicExecutorConfig{},
		ChangeBuffer: 10,
	}
	err = config.Validate()
	assert.Error(t, err)
}
