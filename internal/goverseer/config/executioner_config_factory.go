package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ExecutionerConfig is the interface that all executioner configurations should implement
type ExecutionerConfig interface {
	ValidateAndSetDefaults() error
}

// ExecutionerConfigFactory is a function that creates a new instance of a executioner config
type ExecutionerConfigFactory func() ExecutionerConfig

// ExecutionerConfigRegistry is a global registry for executioner config factories
var ExecutionerConfigRegistry = make(map[string]ExecutionerConfigFactory)

// RegisterExecutionerConfig registers a executioner config factory with the global registry
func RegisterExecutionerConfig(executionerType string, factory ExecutionerConfigFactory) {
	ExecutionerConfigRegistry[executionerType] = factory
}

// DynamicExecutionerConfig is a custom type that handles dynamic unmarshalling
type DynamicExecutionerConfig struct {
	Type   string
	Config ExecutionerConfig
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Executioner
func (dec *DynamicExecutionerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	if len(raw) != 1 {
		return fmt.Errorf("invalid executioner configuration")
	}

	for key, value := range raw {
		dec.Type = key

		// Get the registered factory function
		factory, found := ExecutionerConfigRegistry[key]
		if !found {
			return fmt.Errorf("unknown executioner type: %s", key)
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

		dec.Config = config
	}

	return nil
}

// ValidateAndSetDefaults validates the Executioner configuration and sets default values
func (dec *DynamicExecutionerConfig) ValidateAndSetDefaults() error {
	return dec.Config.ValidateAndSetDefaults()
}
