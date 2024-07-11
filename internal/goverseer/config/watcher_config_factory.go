package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// WatcherConfig is the interface that all watcher configurations should implement
type WatcherConfig interface {
	ValidateAndSetDefaults() error
}

// WatcherConfigFactory is a function that creates a new instance of a watcher config
type WatcherConfigFactory func() WatcherConfig

// WatcherConfigRegistry is a global registry for watcher config factories
var WatcherConfigRegistry = make(map[string]WatcherConfigFactory)

// RegisterWatcherConfig registers a watcher config factory with the global registry
func RegisterWatcherConfig(watcherType string, factory WatcherConfigFactory) {
	WatcherConfigRegistry[watcherType] = factory
}

// DynamicWatcherConfig is a custom type that handles dynamic unmarshalling
type DynamicWatcherConfig struct {
	Type   string
	Config WatcherConfig
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Watcher
func (w *DynamicWatcherConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if len(raw) != 1 {
		return fmt.Errorf("invalid watcher configuration")
	}

	for key, value := range raw {
		w.Type = key

		// Get the registered factory function
		factory, found := WatcherConfigRegistry[key]
		if !found {
			return fmt.Errorf("unknown watcher type: %s", key)
		}

		// Create an instance of the config using the factory function
		config := factory()

		// Unmarshal the value into the config instance
		configBytes, err := yaml.Marshal(value)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(configBytes, config); err != nil {
			return err
		}

		w.Config = config
	}

	return nil
}

// ValidateAndSetDefaults validates the Watcher configuration and sets default values
func (w *DynamicWatcherConfig) ValidateAndSetDefaults() error {
	return w.Config.ValidateAndSetDefaults()
}
