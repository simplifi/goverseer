package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// DummyWatcherConfig is the configuration for a dummy watcher
type DummyWatcherConfig struct {
	// PollMilliseconds is the number of milliseconds to wait between ticks
	PollMilliseconds int `yaml:"poll_milliseconds" validate:"gte=100"`
}

// FileWatcherConfig is the configuration for a file watcher
type FileWatcherConfig struct {
	// Path is the path to the file to watch
	Path string `yaml:"path" validate:"required"`

	// PollSeconds is the number of seconds to wait between polling the file
	PollSeconds int `yaml:"poll_seconds" validate:"gte=1"`
}

// GceWatcherConfig is the configuration for a GCE metadata watcher
type GceWatcherConfig struct {
	// Source is the source of the metadata, either 'instance' or 'project'
	Source string `yaml:"source" validate:"oneof=instance project"`

	// Key is the key to watch in the GCE metadata
	Key string `yaml:"key" validate:"required"`

	// Recurse is whether to recurse the metadata keys
	// See: https://cloud.google.com/compute/docs/metadata/querying-metadata#aggcontents
	Recursive bool `yaml:"recurse,omitempty"`

	// MetadataUrl is the URL to the GCE metadata server
	// The default is the GCE metadata server's default URL
	// You only need to set this if you are running the watcher outside of GCE
	// during testing
	// e.g. http://localhost:8888/computeMetadata/v1/
	MetadataUrl string `yaml:"metadata_url,omitempty"`
}

// DynamicWatcherConfig is a dynamic configuration for a watcher
type DynamicWatcherConfig struct {
	// Type is the type of watcher
	Type string `yaml:"type" validate:"required"`

	// Config is the configuration for the watcher
	// this is dynamic based on the type defined above
	// See UnmarshalYAML below
	Config interface{} `yaml:"config" validate:"required"`
}

// UnmarshalYAML unmarshals the watcher configuration using the dynamic types
func (dc *DynamicWatcherConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	dc.Type = raw["type"].(string)

	configData, err := yaml.Marshal(raw["config"])
	if err != nil {
		return err
	}

	switch dc.Type {
	case "dummy":
		var config DummyWatcherConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return err
		}
		dc.Config = config
	case "file":
		var config FileWatcherConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return err
		}
		dc.Config = config
	case "gce":
		var config GceWatcherConfig
		if err := yaml.Unmarshal(configData, &config); err != nil {
			return err
		}
		dc.Config = config
	default:
		return fmt.Errorf("unknown config type: %s", dc.Type)
	}

	return nil
}
