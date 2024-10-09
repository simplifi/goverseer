package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// WatcherConfig is a custom type that handles dynamic unmarshalling
type WatcherConfig struct {
	// Type is the type of watcher
	Type string

	// Config is the configuration for the watcher
	// The config values will be parsed by the watcher
	Config map[string]interface{}
}

// ExecutionerConfig is a custom type that handles dynamic unmarshalling
type ExecutionerConfig struct {
	// Type is the type of executioner
	Type string

	// Config is the configuration for the watcher
	// The config values will be parsed by the watcher
	Config map[string]interface{}
}

// LoggerConfig is the configuration for the global logger
type LoggerConfig struct {
	// Level is the log level
	Level string
}

// Config is the configuration for a watcher and executioner
type Config struct {
	// Name is the name of the configuration, this will show up in logs
	Name string

	// Logger is the configuration for the logger
	Logger LoggerConfig

	// Watcher is the configuration for the watcher
	// it is dynamic because the configuration can be different for each watcher
	Watcher WatcherConfig

	// Executioner is the configuration for the executioner
	// it is dynamic because the configuration can be different for each executioner
	Executioner ExecutionerConfig
}

// FromFile reads a configuration file and unmarshals it into a Config struct
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
