package executioner

import (
	"fmt"

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
	switch cfg.Executioner.Type {
	case "log":
		return log_executioner.New(*cfg)
	case "shell":
		return shell_executioner.New(*cfg)
	default:
		return nil, fmt.Errorf("unknown executioner type: %s", cfg.Executioner.Type)
	}
}
