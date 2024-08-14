package config

func init() {
	RegisterExecutionerConfig("log", func() ExecutionerConfig { return &LogExecutionerConfig{} })
}

// LogExecutionerConfig is the configuration for a log executioner
// this is used for testing and has no configuration
type LogExecutionerConfig struct{}

// ValidateAndSetDefaults validates the LogExecutionerConfig and sets default values
func (cfg *LogExecutionerConfig) ValidateAndSetDefaults() error {
	return nil
}
