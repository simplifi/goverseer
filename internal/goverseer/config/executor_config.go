package config

func init() {
	RegisterExecutorConfig("dummy", func() ExecutorConfig { return &DummyExecutorConfig{} })
}

// DummyExecutorConfig is the configuration for a dummy executor
// this is used for testing and has no configuration
type DummyExecutorConfig struct{}

// ValidateAndSetDefaults validates the DummyExecutorConfig and sets default values
func (cfg *DummyExecutorConfig) ValidateAndSetDefaults() error {
	return nil
}
