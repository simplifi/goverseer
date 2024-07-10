package config

import "gopkg.in/yaml.v3"

func init() {
	executorConfigFactory.Register("dummy", func(data []byte) (interface{}, error) {
		var config DummyExecutorConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
		return config, nil
	})
}

// DummyExecutorConfig is the configuration for a dummy executor
// this is used for testing and has no configuration
type DummyExecutorConfig struct{}
