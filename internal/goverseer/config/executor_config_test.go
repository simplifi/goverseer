package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyExecutorConfig_ValidateAndSetDefaults(t *testing.T) {
	cfg := &DummyExecutorConfig{}

	// Valid config should not return an error
	err := cfg.ValidateAndSetDefaults()
	assert.NoError(t, err)
}
