package config

func init() {
	RegisterExecutionerConfig("dummy", func() ExecutionerConfig { return &DummyExecutionerConfig{} })
}

// DummyExecutionerConfig is the configuration for a dummy executioner
// this is used for testing and has no configuration
type DummyExecutionerConfig struct{}

// ValidateAndSetDefaults validates the DummyExecutionerConfig and sets default values
func (cfg *DummyExecutionerConfig) ValidateAndSetDefaults() error {
	return nil
}
