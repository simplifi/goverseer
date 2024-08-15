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
	Config interface{}
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for DynamicWatcherConfig
// Because we want to have the type parsed from the yaml node rather than having
// to specify a watcher.type node in the config we need custom unmarshalling
func (d *WatcherConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}

	for k, v := range raw {
		d.Type = k
		d.Config = v
		break
	}

	return nil
}

// Config is the configuration for a watcher and executioner
type Config struct {
	// Name is the name of the configuration, this will show up in logs
	Name string

	// Watcher is the configuration for the watcher
	// it is dynamic because the configuration can be different for each watcher
	Watcher WatcherConfig

	// Executioner is the configuration for the executioner
	// it is dynamic because the configuration can be different for each executioner
	Executioner DynamicExecutionerConfig
}

// ValidateAndSetDefaults validates the Config and sets default values
func (cfg *Config) ValidateAndSetDefaults() error {
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
