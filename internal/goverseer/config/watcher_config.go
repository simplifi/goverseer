package config

import "fmt"

func init() {
	RegisterWatcherConfig("time", func() WatcherConfig { return &TimeWatcherConfig{} })
}

// TimeWatcherConfig is the configuration for a time watcher
type TimeWatcherConfig struct {
	// PollSeconds is the number of seconds to wait between ticks
	PollSeconds int `yaml:"poll_seconds"`
}

// ValidateAndSetDefaults validates the TimeWatcherConfig and sets default values
func (cfg *TimeWatcherConfig) ValidateAndSetDefaults() error {
	if cfg.PollSeconds == 0 {
		cfg.PollSeconds = 1 // default to 1 second
	}
	if cfg.PollSeconds < 1 {
		return fmt.Errorf("poll_seconds must be greater than or equal to 1")
	}
	return nil
}
