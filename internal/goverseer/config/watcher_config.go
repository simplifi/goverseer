package config

import "fmt"

func init() {
	RegisterWatcherConfig("dummy", func() WatcherConfig { return &DummyWatcherConfig{} })
	RegisterWatcherConfig("file", func() WatcherConfig { return &FileWatcherConfig{} })
	RegisterWatcherConfig("gce", func() WatcherConfig { return &GCEWatcherConfig{} })
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

// FileWatcherConfig is the configuration for a file watcher
type FileWatcherConfig struct {
	Path        string `yaml:"path"`
	PollSeconds int    `yaml:"poll_seconds"`
}

// ValidateAndSetDefaults validates the FileWatcherConfig and sets default values
func (cfg *FileWatcherConfig) ValidateAndSetDefaults() error {
	if cfg.PollSeconds == 0 {
		cfg.PollSeconds = 1 // default to 1 second
	}
	if cfg.PollSeconds < 1 {
		return fmt.Errorf("poll_seconds must be greater than or equal to 1")
	}
	if cfg.Path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

const (
	// GCEDefaultMetadataUrl is the default URL for GCE metadata
	// it can be overridden by setting the metadata_url in the config
	// this can be useful for testing
	// e.g. http://localhost:8888/computeMetadata/v1
	GCEDefaultMetadataUrl = "http://metadata.google.internal/computeMetadata/v1"

	// GCEValidSourceInstance is the instance metadata source
	GCEValidSourceInstance = "instance"

	// GCEValidSourceProject is the project metadata source
	GCEValidSourceProject = "project"
)

// GCEWatcherConfig is the configuration for a GCE watcher
type GCEWatcherConfig struct {
	Source      string `yaml:"source"`
	Key         string `yaml:"key"`
	Recursive   bool   `yaml:"recurse,omitempty"`
	MetadataUrl string `yaml:"metadata_url,omitempty"`
}

// ValidateAndSetDefaults validates the GCEWatcherConfig and sets default values
func (cfg *GCEWatcherConfig) ValidateAndSetDefaults() error {
	if cfg.MetadataUrl == "" {
		cfg.MetadataUrl = GCEDefaultMetadataUrl
	}
	if cfg.Source != GCEValidSourceInstance && cfg.Source != GCEValidSourceProject {
		return fmt.Errorf("source must be one of %s or %s", GCEValidSourceInstance, GCEValidSourceProject)
	}
	if cfg.Key == "" {
		return fmt.Errorf("key is required")
	}
	return nil
}
