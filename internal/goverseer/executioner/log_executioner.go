package executioner

import (
	"fmt"
	"log/slog"

	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Executioner interface
	_ Executioner = (*LogExecutioner)(nil)
)

// LogExecutioner logs the data to stdout
type LogExecutioner struct {
	// log is the logger
	log *slog.Logger
}

func newLogExecutioner(cfg config.ExecutionerConfig, log *slog.Logger) (*LogExecutioner, error) {
	_, ok := cfg.(*config.LogExecutionerConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for log executioner")
	}

	return &LogExecutioner{
		log: log,
	}, nil
}

// Execute logs the data to stdout
func (e *LogExecutioner) Execute(data interface{}) error {
	e.log.Info("received data", slog.String("data", fmt.Sprintf("%v", data)))
	return nil
}

// Stop signals the executioner to stop
func (e *LogExecutioner) Stop() {
	e.log.Info("shutting down executioner")
}
