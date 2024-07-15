package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDynamicWatcherConfig_UnmarshalYAML(t *testing.T) {
	// Unmarshalling a valid config should not return an error
	yamlConfigSnippet := []byte(`
time:
`)
	var cfg DynamicWatcherConfig
	err := yaml.Unmarshal(yamlConfigSnippet, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "time", cfg.Type)
	assert.IsType(t, &TimeWatcherConfig{}, cfg.Config)
}
