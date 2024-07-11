package executor

import (
	"fmt"
	"log/slog"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Executor interface
	_ Executor = (*DummyExecutor)(nil)
)

func init() {
	RegisterExecutor("dummy", func() Executor { return &DummyExecutor{} })
}

// DummyExecutor logs the data to stdout
type DummyExecutor struct {
	// log is the logger
	log *slog.Logger
}

func (e *DummyExecutor) Create(cfg config.ExecutorConfig, log *slog.Logger) error {
	_, ok := cfg.(*config.DummyExecutorConfig)
	if !ok {
		return fmt.Errorf("invalid config for dummy executor")
	}

	e.log = log

	return nil
}

// Execute logs the data to stdout
func (e *DummyExecutor) Execute(data interface{}) error {
	e.log.Info("received data", slog.String("data", fmt.Sprintf("%v", data)))
	return nil
}

// Stop signals the executor to stop
func (e *DummyExecutor) Stop() {
	// Nothing to do here really, it's just a dummy
	e.log.Info("shutting down executor")
}
