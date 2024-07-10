package config

import "gopkg.in/yaml.v3"

func init() {
	watcherConfigFactory.Register("dummy", func(data []byte) (interface{}, error) {
		var config DummyWatcherConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
		return config, nil
	})
}

// DummyWatcherConfig is the configuration for a dummy watcher
type DummyWatcherConfig struct {
	// PollSeconds is the number of seconds to wait between ticks
	PollSeconds int `yaml:"poll_seconds" validate:"gte=1"`
}
