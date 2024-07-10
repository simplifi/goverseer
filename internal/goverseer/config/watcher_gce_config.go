package config

import "gopkg.in/yaml.v3"

func init() {
	watcherConfigFactory.Register("gce", func(data []byte) (interface{}, error) {
		var config GceWatcherConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
		return config, nil
	})
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
