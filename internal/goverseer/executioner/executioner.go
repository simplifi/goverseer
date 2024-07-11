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

type ExecutionerFactory func() Executioner

// ExecutionerRegistry is a global registry for executioner factories
var ExecutionerRegistry = make(map[string]ExecutionerFactory)

// RegisterExecutioner registers a executioner factory with the global registry
func RegisterExecutioner(executionerType string, factory ExecutionerFactory) {
	ExecutionerRegistry[executionerType] = factory
}

// New creates a new Executioner based on the config
func New(cfg *config.Config) (*Executioner, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name).
		With("executioner", cfg.Executioner.Type)

	// Get the registered factory function
	factory, found := ExecutionerRegistry[cfg.Executioner.Type]
	if !found {
		return nil, fmt.Errorf("unknown executioner type: %s", cfg.Executioner.Type)
	}

	// Create an instance of the executioner using the factory function
	exec := factory()
	err := exec.Create(cfg.Executioner.Config, log)

	return &exec, err
}
