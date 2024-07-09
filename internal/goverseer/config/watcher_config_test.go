package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestDummyWatcherConfig(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Valid command executor configuration
	yamlDataValid := []byte(`
type: dummy
config:
  poll_milliseconds: 100
`)
	var dynamicConfigValid DynamicWatcherConfig
	err := yaml.Unmarshal(yamlDataValid, &dynamicConfigValid)
	assert.NoError(t, err)

	err = validate.Struct(dynamicConfigValid)
	assert.NoError(t, err)
	assert.Equal(t, "dummy", dynamicConfigValid.Type)
	assert.Equal(t, DummyWatcherConfig{PollMilliseconds: 100}, dynamicConfigValid.Config)

	// Invalid command executor configuration
	yamlDataInvalid := []byte(`
type: dummy
config: {}
`)
	var dynamicConfigInvalid DynamicWatcherConfig
	yaml.Unmarshal(yamlDataInvalid, &dynamicConfigInvalid)
	err = validate.Struct(dynamicConfigInvalid)
	assert.Error(t, err)
}

func TestFileWatcherConfig(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Valid command executor configuration
	yamlDataValid := []byte(`
type: file
config:
  path: /tmp/foo/bar
  poll_seconds: 5
`)
	var dynamicConfigValid DynamicWatcherConfig
	err := yaml.Unmarshal(yamlDataValid, &dynamicConfigValid)
	assert.NoError(t, err)

	err = validate.Struct(dynamicConfigValid)
	assert.NoError(t, err)
	assert.Equal(t, "file", dynamicConfigValid.Type)
	assert.Equal(t, FileWatcherConfig{Path: "/tmp/foo/bar", PollSeconds: 5}, dynamicConfigValid.Config)

	// Invalid command executor configuration
	yamlDataInvalid := []byte(`
type: file
config: {}
`)
	var dynamicConfigInvalid DynamicExecutorConfig
	yaml.Unmarshal(yamlDataInvalid, &dynamicConfigInvalid)
	err = validate.Struct(dynamicConfigInvalid)
	assert.Error(t, err)
}

func TestGceWatcherConfig(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Valid command executor configuration
	yamlDataValid := []byte(`
type: gce
config:
  source: instance
  key: my-key
`)
	var dynamicConfigValid DynamicWatcherConfig
	err := yaml.Unmarshal(yamlDataValid, &dynamicConfigValid)
	assert.NoError(t, err)

	err = validate.Struct(dynamicConfigValid)
	assert.NoError(t, err)
	assert.Equal(t, "gce", dynamicConfigValid.Type)
	assert.Equal(t, GceWatcherConfig{Source: "instance", Key: "my-key"}, dynamicConfigValid.Config)

	// Invalid command executor configuration
	yamlDataInvalid := []byte(`
type: gce
config: {}
`)
	var dynamicConfigInvalid DynamicExecutorConfig
	yaml.Unmarshal(yamlDataInvalid, &dynamicConfigInvalid)
	err = validate.Struct(dynamicConfigInvalid)
	assert.Error(t, err)
}
