package executioner

import (
	"fmt"
	"log/slog"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Executioner interface
	_ Executioner = (*DummyExecutioner)(nil)
)

func init() {
	RegisterExecutioner("dummy", func() Executioner { return &DummyExecutioner{} })
}

// DummyExecutioner logs the data to stdout
type DummyExecutioner struct {
	// log is the logger
	log *slog.Logger
}

func (e *DummyExecutioner) Create(cfg config.ExecutionerConfig, log *slog.Logger) error {
	_, ok := cfg.(*config.DummyExecutionerConfig)
	if !ok {
		return fmt.Errorf("invalid config for dummy executioner")
	}

	e.log = log

	return nil
}

// Execute logs the data to stdout
func (e *DummyExecutioner) Execute(data interface{}) error {
	e.log.Info("received data", slog.String("data", fmt.Sprintf("%v", data)))
	return nil
}

// Stop signals the executioner to stop
func (e *DummyExecutioner) Stop() {
	// Nothing to do here really, it's just a dummy
	e.log.Info("shutting down executioner")
}
