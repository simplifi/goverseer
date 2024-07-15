package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeWatcherConfig_ValidateAndSetDefaults(t *testing.T) {
	config := &TimeWatcherConfig{
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
