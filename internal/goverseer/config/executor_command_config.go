package config

import (
	"gopkg.in/yaml.v3"
)

func init() {
	executorConfigFactory.Register("command", func(data []byte) (interface{}, error) {
		var config CommandExecutorConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
		return config, nil
	})
}

// CommandExecutorConfig is the configuration for a command executor
type CommandExecutorConfig struct {
	// Command is the command to execute
	Command string `yaml:"command" validate:"required"`
}
