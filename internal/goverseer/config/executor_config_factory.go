package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ExecutorConfig is the interface that all executor configurations should implement
type ExecutorConfig interface {
	ValidateAndSetDefaults() error
}

// ExecutorConfigFactory is a function that creates a new instance of a executor config
type ExecutorConfigFactory func() ExecutorConfig

// ExecutorConfigRegistry is a global registry for executor config factories
var ExecutorConfigRegistry = make(map[string]ExecutorConfigFactory)

// RegisterExecutorConfig registers a executor config factory with the global registry
func RegisterExecutorConfig(executorType string, factory ExecutorConfigFactory) {
	ExecutorConfigRegistry[executorType] = factory
}

// DynamicExecutorConfig is a custom type that handles dynamic unmarshalling
type DynamicExecutorConfig struct {
	Type   string
	Config ExecutorConfig
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Executor
func (w *DynamicExecutorConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if len(raw) != 1 {
		return fmt.Errorf("invalid executor configuration")
	}

	for key, value := range raw {
		w.Type = key

		// Get the registered factory function
		factory, found := ExecutorConfigRegistry[key]
		if !found {
			return fmt.Errorf("unknown executor type: %s", key)
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

// ValidateAndSetDefaults validates the Executor configuration and sets default values
func (w *DynamicExecutorConfig) ValidateAndSetDefaults() error {
	return w.Config.ValidateAndSetDefaults()
}
