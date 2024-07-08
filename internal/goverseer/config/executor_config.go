package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// CommandExecutorConfig is the configuration for a command executor
type CommandExecutorConfig struct {
	// Command is the command to execute
	Command string `yaml:"command" validate:"required"`
}

// DummyExecutorConfig is the configuration for a dummy executor
// this is used for testing and has no configuration
type DummyExecutorConfig struct{}

// DynamicExecutorConfig is a dynamic configuration for an executor
type DynamicExecutorConfig struct {
	// Type is the type of executor
	Type string `yaml:"type" validate:"required"`

	// Config is the configuration for the executor
	// this is dynamic based on the type defined above
	// See UnmarshalYAML below
	Config interface{} `yaml:"config" validate:"required"`
}

// UnmarshalYAML unmarshals the executor configuration using the dynamic types
func (dc *DynamicExecutorConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	dc.Type = raw["type"].(string)

	configData, err := yaml.Marshal(raw["config"])
	if err != nil {
		return err
	}

	switch dc.Type {
	case "command":
		var config CommandExecutorConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return err
		}
		dc.Config = config
	case "dummy":
		var config DummyExecutorConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return err
		}
		dc.Config = config
	default:
		return fmt.Errorf("unknown config type: %s", dc.Type)
	}

	return nil
}
