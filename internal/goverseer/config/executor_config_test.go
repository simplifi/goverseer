package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandExecutorConfig_ValidateAndSetDefaults(t *testing.T) {
	cfg := &CommandExecutorConfig{
		Command: "ls -l",
	}

	// Valid config should not return an error
	err := cfg.ValidateAndSetDefaults()
	assert.NoError(t, err)

	// Invalid config should return an error
	cfg.Command = ""
	err = cfg.ValidateAndSetDefaults()
	assert.Error(t, err)
}

func TestDummyExecutorConfig_ValidateAndSetDefaults(t *testing.T) {
	cfg := &DummyExecutorConfig{}

	// Valid config should not return an error
	err := cfg.ValidateAndSetDefaults()
	assert.NoError(t, err)
}
