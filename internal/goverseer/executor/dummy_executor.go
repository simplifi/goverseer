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
	factory.Register("dummy", func(cfg interface{}, log *slog.Logger) (Executor, error) {
		_, ok := cfg.(config.DummyExecutorConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for dummy executor")
		}
		return NewDummyExecutor(log), nil
	})
}

// DummyExecutor logs the data to stdout
type DummyExecutor struct {
	// log is the logger
	log *slog.Logger
}

// NewDummyExecutor creates a new DummyExecutor
// The log is the logger
func NewDummyExecutor(log *slog.Logger) *DummyExecutor {
	return &DummyExecutor{
		log: log,
	}
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
