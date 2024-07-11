package config

import "fmt"

func init() {
	RegisterWatcherConfig("dummy", func() WatcherConfig { return &DummyWatcherConfig{} })
}

// DummyWatcherConfig is the configuration for a GCE watcher
type DummyWatcherConfig struct {
	// PollSeconds is the number of seconds to wait between ticks
	PollSeconds int `yaml:"poll_seconds"`
}

// ValidateAndSetDefaults validates the GceWatcherConfig and sets default values
func (cfg *DummyWatcherConfig) ValidateAndSetDefaults() error {
	if cfg.PollSeconds == 0 {
		cfg.PollSeconds = 1 // default to 1 second
	}
	if cfg.PollSeconds < 1 {
		return fmt.Errorf("poll_seconds must be greater than or equal to 1")
	}
	return nil
}
