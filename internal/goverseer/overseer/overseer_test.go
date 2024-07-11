package overseer

import (
	"sync"
	"testing"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

func TestOverseer(t *testing.T) {
	stop := make(chan struct{})

	cfg := &config.Config{
		Name: "TestManager",
		Watcher: config.DynamicWatcherConfig{
			Type: "dummy",
			Config: &config.DummyWatcherConfig{
				PollSeconds: 1,
			},
		},
		Executor: config.DynamicExecutorConfig{
			Type:   "dummy",
			Config: &config.DummyExecutorConfig{},
		},
	}
	cfg.ValidateAndSetDefaults()

	overseer, err := New(cfg, stop)
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
	close(stop)
	wg.Wait()
}
