package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for a watcher and executioner
type Config struct {
	// Name is the name of the configuration, this will show up in logs
	Name string `yaml:"name"`

	// Watcher is the configuration for the watcher
	// it is dynamic because the configuration can be different for each watcher
	Watcher DynamicWatcherConfig `yaml:"watcher"`

	// Executioner is the configuration for the executioner
	// it is dynamic because the configuration can be different for each executioner
	Executioner DynamicExecutionerConfig `yaml:"executioner"`

	// ChangeBuffer is the number of changes to buffer in the overseer's queue
	ChangeBuffer int `yaml:"change_buffer,omitempty"`
}

// ValidateAndSetDefaults validates the Config and sets default values
func (cfg *Config) ValidateAndSetDefaults() error {
	if cfg.ChangeBuffer == 0 {
		cfg.ChangeBuffer = 100 // default to 100
	}
	if cfg.ChangeBuffer < 1 {
		return fmt.Errorf("change_buffer must be greater than or equal to 1")
	}

	if err := cfg.Watcher.ValidateAndSetDefaults(); err != nil {
		return err
	}

	if err := cfg.Executioner.ValidateAndSetDefaults(); err != nil {
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

	// Validate and set defaults for the configuration
	if err := cfg.ValidateAndSetDefaults(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
