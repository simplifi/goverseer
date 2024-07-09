package manager

import (
	"testing"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

func TestManager(t *testing.T) {
	// Create some sample overseer configurations
	configs := []*config.Config{
		{
			Name: "TestManager",
			Watcher: config.DynamicWatcherConfig{
				Type: "dummy",
				Config: config.DummyWatcherConfig{
					PollMilliseconds: 100,
				},
			},
			Executor: config.DynamicExecutorConfig{
				Type:   "dummy",
				Config: config.DummyExecutorConfig{},
			},
		},
	}

	// Create a new Manager instance
	manager := NewManager(configs)

	// Run the manager
	err := manager.Run()

	// Check if there was an error
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	manager.Stop()
}
