package config

import (
	"gopkg.in/yaml.v3"
)

var watcherConfigFactory = &ConfigFactory{
	creators: make(map[string]func([]byte) (interface{}, error)),
}

// DynamicWatcherConfig is a dynamic configuration for a watcher
type DynamicWatcherConfig struct {
	// Type is the type of watcher
	Type string `yaml:"type" validate:"required"`

	// Config is the configuration for the watcher
	// this is dynamic based on the type defined above
	// See UnmarshalYAML below
	Config interface{} `yaml:"config" validate:"required"`
}

// UnmarshalYAML unmarshals the watcher configuration using the dynamic types
func (dc *DynamicWatcherConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	dc.Type = raw["type"].(string)

	configData, err := yaml.Marshal(raw["config"])
	if err != nil {
		return err
	}

	config, err := watcherConfigFactory.Create(dc.Type, configData)
	if err != nil {
		return err
	}
	dc.Config = config

	return nil
}
