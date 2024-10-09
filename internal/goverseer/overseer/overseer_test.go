package overseer

import (
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/logger"
	"github.com/stretchr/testify/assert"
)

// TestOverseer_Run tests the Overseer's Run function
func TestOverseer_Run(t *testing.T) {
	cfg := &config.Config{
		Name: "TestManager",
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}{
				"poll_seconds": 1,
			},
		},
		Executioner: config.ExecutionerConfig{
			Type: "log",
			Config: map[string]interface{}{
				"tag": "test",
			},
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

func TestOverseer_New(t *testing.T) {
	cfg := &config.Config{
		Name: "TestManager",
		Watcher: config.WatcherConfig{
			Type: "time",
			Config: map[string]interface{}{
				"poll_seconds": 1,
			},
		},
		Executioner: config.ExecutionerConfig{
			Type: "log",
			Config: map[string]interface{}{
				"tag": "test",
			},
		},
	}

	_, err := New(cfg)
	assert.NoError(t, err,
		"Creating a new Overseer with no logger config should not error")
	assert.Equal(t, logger.DefaultLogLevel, logger.Log.GetLevel(),
		"Creating a new Overseer with no logger config should set the default log level")

	cfg.Logger = config.LoggerConfig{
		Level: "debug",
	}
	_, err = New(cfg)
	assert.NoError(t, err,
		"Creating a new Overseer from a valid config should not error")
	assert.Equal(t, log.DebugLevel, logger.Log.GetLevel(),
		"Creating a new Overseer should set the configured log level")
}
