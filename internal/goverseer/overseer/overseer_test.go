package overseer

import (
	"sync"
	"testing"
	"time"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// TestOverseer tests the Overseer
func TestOverseer(t *testing.T) {
	cfg := &config.Config{
		Name: "TestManager",
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}(map[string]interface{}{
				"poll_seconds": 1,
			}),
		},
		Executioner: config.ExecutionerConfig{
			Type: "log",
			Config: map[string]interface{}(map[string]interface{}{
				"tag": "test",
			}),
		},
	}

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
