package config

import "fmt"

func init() {
	RegisterExecutorConfig("command", func() ExecutorConfig { return &CommandExecutorConfig{} })
	RegisterExecutorConfig("dummy", func() ExecutorConfig { return &DummyExecutorConfig{} })
}

// CommandExecutorConfig is the configuration for a command executor
type CommandExecutorConfig struct {
	// Command is the command to execute
	Command string `yaml:"command"`
}

// ValidateAndSetDefaults validates the CommandExecutorConfig and sets default values
func (cfg *CommandExecutorConfig) ValidateAndSetDefaults() error {
	if cfg.Command == "" {
		return fmt.Errorf("command is required")
	}
	return nil
}

// DummyExecutorConfig is the configuration for a dummy executor
// this is used for testing and has no configuration
type DummyExecutorConfig struct{}

// ValidateAndSetDefaults validates the DummyExecutorConfig and sets default values
func (cfg *DummyExecutorConfig) ValidateAndSetDefaults() error {
	return nil
}
