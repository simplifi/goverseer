package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDynamicExecutionerConfig_UnmarshalYAML(t *testing.T) {
	// Unmarshalling a valid config should not return an error
	yamlConfigSnippet := []byte(`
dummy:
`)
	var cfg DynamicExecutionerConfig
	err := yaml.Unmarshal(yamlConfigSnippet, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "dummy", cfg.Type)
	assert.IsType(t, &DummyExecutionerConfig{}, cfg.Config)
}
