package config

import (
	"gopkg.in/yaml.v3"
)

var executorConfigFactory = &ConfigFactory{
	creators: make(map[string]func([]byte) (interface{}, error)),
}

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

	config, err := executorConfigFactory.Create(dc.Type, configData)
	if err != nil {
		return err
	}
	dc.Config = config

	return nil
}
