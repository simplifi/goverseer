package config

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testConfig = `
name: TestConfig
watcher:
  type: file
  config:
    path: /tmp/test1
executor:
  type: dummy
`
)

// createTempConfig creates a temporary configuration file and returns its path.
// The cleanup function will delete the file when called.
func createTempConfig(t *testing.T, content string) (string, func()) {
	t.Helper()

	// Create a temporary directory for the config file
	tempDir := t.TempDir()

	// Create the temporary configuration file
	configFile := filepath.Join(tempDir, "test-config.yaml")
	err := os.WriteFile(configFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary configuration file: %v", err)
	}

	// Return the path and a cleanup function
	return configFile, func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("Failed to cleanup temporary directory: %v", err)
		}
	}
}

func TestFromPath(t *testing.T) {
	// Create a temporary configuration file
	configFile, cleanup := createTempConfig(t, testConfig)
	defer cleanup()

	// Call the FromPath function
	configs, err := FromPath(filepath.Dir(configFile))
	if err != nil {
		t.Fatalf("FromPath returned an error: %v", err)
	}

	// Check the number of configurations
	if len(configs) != 1 {
		t.Fatalf("Expected 1 configurations, got %d", len(configs))
	}

	// Check the configuration values
	expectedName := "TestConfig"
	if configs[0].Name != expectedName {
		t.Errorf("Expected configuration name %q, got %q", expectedName, configs[0].Name)
	}
}

func TestFromFile(t *testing.T) {
	// Create a temporary configuration file
	configFile, cleanup := createTempConfig(t, testConfig)
	defer cleanup()

	// Call the FromFile function
	config, err := FromFile(configFile)
	if err != nil {
		t.Fatalf("FromFile returned an error: %v", err)
	}

	// Check the configuration values
	expectedName := "TestConfig"
	if config.Name != expectedName {
		t.Errorf("Expected configuration name %q, got %q", expectedName, config.Name)
	}

	expectedWatcherType := "file"
	if config.Watcher.Type != expectedWatcherType {
		t.Errorf("Expected watcher type %q, got %q", expectedWatcherType, config.Watcher.Type)
	}

	fileWatcherConfig, ok := config.Watcher.Config.(FileWatcherConfig)
	if !ok {
		t.Fatalf("Watcher config is not of expected type FileWatcherConfig")
	}

	expectedWatcherConfigPath := "/tmp/test1"
	if fileWatcherConfig.Path != expectedWatcherConfigPath {
		t.Errorf("Expected watcher config path %q, got %q", expectedWatcherConfigPath, fileWatcherConfig.Path)
	}

	expectedExecutorType := "dummy"
	if config.Executor.Type != expectedExecutorType {
		t.Errorf("Expected executor type %q, got %q", expectedExecutorType, config.Executor.Type)
	}

	if _, ok := config.Executor.Config.(DummyExecutorConfig); !ok {
		t.Fatalf("Executor config is not of expected type DummyExecutorConfig")
	}
}
