package executor

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Executor is an interface for executing actions
type Executor interface {
	Execute(data interface{}) error
	Stop()
}

type ExecutorFactory struct {
	mu       sync.RWMutex
	creators map[string]func(interface{}, *slog.Logger) (Executor, error)
}

var factory = &ExecutorFactory{
	creators: make(map[string]func(interface{}, *slog.Logger) (Executor, error)),
}

func (f *ExecutorFactory) Register(executorType string, creator func(interface{}, *slog.Logger) (Executor, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.creators[executorType] = creator
}

func (f *ExecutorFactory) Create(executorType string, cfg interface{}, log *slog.Logger) (Executor, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	creator, exists := f.creators[executorType]
	if !exists {
		return nil, fmt.Errorf("unknown executor type: %s", executorType)
	}
	return creator(cfg, log)
}

// NewExecutor creates a new Executor based on the config
func NewExecutor(cfg *config.Config) (Executor, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("executor", cfg.Executor.Type)

	return factory.Create(cfg.Executor.Type, cfg.Executor.Config, log)
}
