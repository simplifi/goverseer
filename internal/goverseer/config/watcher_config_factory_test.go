package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDynamicWatcherConfig_UnmarshalYAML(t *testing.T) {
	// Unmarshalling a valid config should not return an error
	yamlConfigSnippet := []byte(`
dummy:
`)
	var cfg DynamicWatcherConfig
	err := yaml.Unmarshal(yamlConfigSnippet, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "dummy", cfg.Type)
	assert.IsType(t, &DummyWatcherConfig{}, cfg.Config)
}
