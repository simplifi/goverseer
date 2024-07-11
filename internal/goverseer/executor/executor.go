package executor

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Executor is an interface for executing actions
type Executor interface {
	Execute(data interface{}) error
	Create(cfg config.ExecutorConfig, log *slog.Logger) error
	Stop()
}

type ExecutorFactory func() Executor

// ExecutorRegistry is a global registry for executor factories
var ExecutorRegistry = make(map[string]ExecutorFactory)

// RegisterExecutor registers a executor factory with the global registry
func RegisterExecutor(executorType string, factory ExecutorFactory) {
	ExecutorRegistry[executorType] = factory
}

// New creates a new Executor based on the config
func New(cfg *config.Config) (*Executor, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("executor", cfg.Executor.Type)

	// Get the registered factory function
	factory, found := ExecutorRegistry[cfg.Executor.Type]
	if !found {
		return nil, fmt.Errorf("unknown executor type: %s", cfg.Executor.Type)
	}

	// Create an instance of the executor using the factory function
	exec := factory()
	err := exec.Create(cfg.Executor.Config, log)

	return &exec, err
}
