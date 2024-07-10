package config

import "gopkg.in/yaml.v3"

func init() {
	watcherConfigFactory.Register("file", func(data []byte) (interface{}, error) {
		var config FileWatcherConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
		return config, nil
	})
}

// FileWatcherConfig is the configuration for a file watcher
type FileWatcherConfig struct {
	// Path is the path to the file to watch
	Path string `yaml:"path" validate:"required"`

	// PollSeconds is the number of seconds to wait between polling the file
	PollSeconds int `yaml:"poll_seconds" validate:"gte=1"`
}
