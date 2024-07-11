package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDummyWatcherConfig_ValidateAndSetDefaults(t *testing.T) {
	config := &DummyWatcherConfig{
		PollSeconds: 10,
	}

	// Valid config should not return an error
	err := config.ValidateAndSetDefaults()
	assert.NoError(t, err)

	// Invalid configuration, PollSeconds less than 0
	config.PollSeconds = -1
	err = config.ValidateAndSetDefaults()
	assert.Error(t, err)

	// Check defaults are set, PollSeconds missing (0) so it should get the default
	config.PollSeconds = 0
	err = config.ValidateAndSetDefaults()
	assert.NoError(t, err)
	assert.Equal(t, 1, config.PollSeconds)
}

func TestFileWatcherConfig_ValidateAndSetDefaults(t *testing.T) {
	config := &FileWatcherConfig{
		Path:        "/",
		PollSeconds: 10,
	}

	// Valid config should not return an error
	err := config.ValidateAndSetDefaults()
	assert.NoError(t, err)

	// Invalid configuration, PollSeconds less than 0
	config.PollSeconds = -1
	err = config.ValidateAndSetDefaults()
	assert.Error(t, err)

	// Check defaults are set, PollSeconds missing (0) so it should get the default
	config.PollSeconds = 0
	err = config.ValidateAndSetDefaults()
	assert.NoError(t, err)
	assert.Equal(t, 1, config.PollSeconds)

	// Path missing ("") so it should error
	config.PollSeconds = 0
	config.Path = ""
	err = config.ValidateAndSetDefaults()
	assert.Error(t, err)
}

func TestGceWatcherConfig_ValidateAndSetDefaults(t *testing.T) {
	config := &GCEWatcherConfig{
		Source:      "instance",
		Key:         "node-json",
		Recursive:   true,
		MetadataUrl: "http://metadata.google.internal/computeMetadata/v1/",
	}

	// Valid config should not return an error
	err := config.ValidateAndSetDefaults()
	assert.NoError(t, err)

	// Invalid configuration, Source not valid
	config.Source = "foobar"
	err = config.ValidateAndSetDefaults()
	assert.Error(t, err)

	// Invalid configuration, Key missing
	config.Source = "instance"
	config.Key = ""
	err = config.ValidateAndSetDefaults()
	assert.Error(t, err)
}
