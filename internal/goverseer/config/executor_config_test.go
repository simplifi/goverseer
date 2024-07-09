package config

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCommandExecutorConfig(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Valid command executor configuration
	yamlDataValid := []byte(`
type: command
config:
  command: echo 'Hello, World!'
`)
	var dynamicConfigValid DynamicExecutorConfig
	err := yaml.Unmarshal(yamlDataValid, &dynamicConfigValid)
	assert.NoError(t, err)

	err = validate.Struct(dynamicConfigValid)
	assert.NoError(t, err)
	assert.Equal(t, "command", dynamicConfigValid.Type)
	assert.Equal(t, CommandExecutorConfig{Command: "echo 'Hello, World!'"}, dynamicConfigValid.Config)

	// Invalid command executor configuration
	yamlDataInvalid := []byte(`
type: command
config:
  foo: bar
`)
	var dynamicConfigInvalid DynamicExecutorConfig
	yaml.Unmarshal(yamlDataInvalid, &dynamicConfigInvalid)
	err = validate.Struct(dynamicConfigInvalid)
	assert.Error(t, err)
}

func TestDummyExecutorConfig(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Valid dummy executor configuration
	yamlDataValid := []byte(`
type: dummy
config: {}
`)
	var dynamicConfigValid DynamicExecutorConfig
	err := yaml.Unmarshal(yamlDataValid, &dynamicConfigValid)
	assert.NoError(t, err)

	err = validate.Struct(dynamicConfigValid)
	assert.NoError(t, err)
	assert.Equal(t, DummyExecutorConfig{}, dynamicConfigValid.Config)
}
