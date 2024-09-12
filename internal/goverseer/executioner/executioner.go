package executioner

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/executioner/log_executioner"
	"github.com/simplifi/goverseer/internal/goverseer/executioner/shell_executioner"
)

// Executioner is an interface for executing actions
type Executioner interface {
	Execute(data interface{}) error
	Stop()
}

// New creates a new Executioner based on the config
// It returns an Executioner based on the config or an error
func New(cfg *config.Config) (Executioner, error) {
	// Setup the logger
	logger := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("executioner", cfg.Executioner.Type)

	switch cfg.Executioner.Type {
	case "log":
		return log_executioner.New(*cfg, logger)
	case "shell":
		return shell_executioner.New(*cfg, logger)
	default:
		return nil, fmt.Errorf("unknown executioner type: %s", cfg.Executioner.Type)
	}
}
