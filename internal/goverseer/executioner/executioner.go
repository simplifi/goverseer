package executioner

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

// Executioner is an interface for executing actions
type Executioner interface {
	Execute(data interface{}) error
	Create(cfg config.ExecutionerConfig, log *slog.Logger) error
	Stop()
}

// New creates a new Executioner based on the config
func New(cfg *config.Config) (Executioner, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("executioner", cfg.Executioner.Type)

	switch cfg.Executioner.Type {
	case "log":
		exec := LogExecutioner{}
		err := exec.Create(cfg.Executioner.Config, log)
		return &exec, err
	default:
		return nil, fmt.Errorf("unknown executioner type: %s", cfg.Executioner.Type)
	}
}
