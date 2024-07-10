package config

// TODO: Better validation, handle defaults more gracefully

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// ConfigFactory is a factory for creating configuration objects
type ConfigFactory struct {
	mu       sync.RWMutex
	creators map[string]func([]byte) (interface{}, error)
}

// Register registers a creator function for a configuration type
func (cf *ConfigFactory) Register(configType string, creator func([]byte) (interface{}, error)) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	cf.creators[configType] = creator
}

// Create creates a configuration object from the given data
func (cf *ConfigFactory) Create(configType string, data []byte) (interface{}, error) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	creator, exists := cf.creators[configType]
	if !exists {
		return nil, fmt.Errorf("unknown config type: %s", configType)
	}
	return creator(data)
}

// Config is the configuration for a watcher and executor
type Config struct {
	// Name is the name of the configuration, this will show up in logs
	Name string `yaml:"name" validate:"required"`

	// Watcher is the configuration for the watcher
	// it is dynamic because the configuration can be different for each watcher
	Watcher DynamicWatcherConfig `yaml:"watcher"`

	// Executor is the configuration for the executor
	// it is dynamic because the configuration can be different for each executor
	Executor DynamicExecutorConfig `yaml:"executor"`

	// ChangeBuffer is the number of changes to buffer in the overseer's queue
	ChangeBuffer int `yaml:"change_buffer,omitempty"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

// FromFile reads a configuration file
func FromFile(path string) (*Config, error) {
	cfgFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(cfgFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// configsInPath returns all configuration files in the given directory
func configsInPath(path string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// FromPath reads all configuration files in the given directory
func FromPath(path string) ([]*Config, error) {
	configs := []*Config{}
	files, err := configsInPath(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		config, err := FromFile(file)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}

	return configs, nil
}
