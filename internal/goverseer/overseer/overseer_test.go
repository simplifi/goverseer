package overseer

import (
	"sync"
	"testing"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

func TestOverseer(t *testing.T) {
	cfg := &config.Config{
		Name: "TestManager",
		Watcher: config.DynamicWatcherConfig{
			Type: "time",
			Config: &config.TimeWatcherConfig{
				PollSeconds: 1,
			},
		},
		Executioner: config.DynamicExecutionerConfig{
			Type:   "log",
			Config: &config.LogExecutionerConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	overseer, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create Overseer: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		overseer.Run()
	}()

	// Wait for a short time to let the overseer run
	time.Sleep(200 * time.Millisecond)

	// Stop the overseer and wait
	overseer.Stop()
	wg.Wait()
}
