package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// WatcherConfig is a custom type that handles dynamic unmarshalling
type WatcherConfig struct {
	// Type is the type of watcher
	Type string `mapstructure:"type" validate:"required,oneof=file time gce_metadata gcp_secrets"`

	// Config is the configuration for the watcher
	// The config values will be parsed by the watcher
	Config map[string]interface{} `mapstructure:"config"`
}

// ExecutionerConfig is a custom type that handles dynamic unmarshalling
type ExecutionerConfig struct {
	// Type is the type of executioner
	Type string `mapstructure:"type" validate:"required,oneof=log shell"`

	// Config is the configuration for the watcher
	// The config values will be parsed by the watcher
	Config map[string]interface{} `mapstructure:"config"`
}

// LoggerConfig is the configuration for the global logger
type LoggerConfig struct {
	// Level is the log level
	Level string `mapstructure:"level"`
}

// Config is the configuration for a watcher and executioner
type Config struct {
	// Name is the name of the configuration, this will show up in logs
	Name string `mapstructure:"name"`

	// Logger is the configuration for the logger
	Logger LoggerConfig `mapstructure:"logger"`

	// Watcher is the configuration for the watcher
	// it is dynamic because the configuration can be different for each watcher
	Watcher WatcherConfig `mapstructure:"watcher" validate:"required"`

	// Executioner is the configuration for the executioner
	// it is dynamic because the configuration can be different for each executioner
	Executioner ExecutionerConfig `mapstructure:"executioner" validate:"required"`
}

// FromFile reads a configuration file and unmarshals it into a Config struct
func FromFile(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfgMap map[string]interface{}
	if err := yaml.Unmarshal(raw, &cfgMap); err != nil {
		return nil, err
	}

	var cfg Config
	if err := Decode(cfgMap, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
