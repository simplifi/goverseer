package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyExecutionerConfig_ValidateAndSetDefaults(t *testing.T) {
	cfg := &DummyExecutionerConfig{}

	// Valid config should not return an error
	err := cfg.ValidateAndSetDefaults()
	assert.NoError(t, err)
}
