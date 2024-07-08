package executor

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Executor is an interface for executing actions
type Executor interface {
	Execute(data interface{}, wg *sync.WaitGroup)
	Stop()
}

// NewExecutor creates a new Executor based on the config
func NewExecutor(cfg *config.Config, log *slog.Logger) (Executor, error) {
	switch v := cfg.Executor.Config.(type) {
	case config.DummyExecutorConfig:
		return NewDummyExecutor(log), nil
	case config.CommandExecutorConfig:
		return NewCommandExecutor(v.Command, log), nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", cfg.Watcher.Type)
	}
}
