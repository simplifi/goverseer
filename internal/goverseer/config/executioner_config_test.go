package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogExecutionerConfig_ValidateAndSetDefaults(t *testing.T) {
	cfg := &LogExecutionerConfig{}

	// Valid config should not return an error
	err := cfg.ValidateAndSetDefaults()
	assert.NoError(t, err)
}
