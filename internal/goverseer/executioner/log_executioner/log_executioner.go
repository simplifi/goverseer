package log_executioner

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

const (
	// DefaultTag is the default tag to add to the logs
	DefaultTag = ""
)

// LogExecutionerConfig is the configuration for a log executioner
type Config struct {
	// Tag is a tag to add to the logs, by default it is empty
	Tag string
}

// ParseConfig parses the config for a log executioner
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(config interface{}) (*Config, error) {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	lec := &Config{
		Tag: DefaultTag,
	}

	if tag, ok := cfgMap["tag"].(string); ok {
		lec.Tag = tag
	}

	return lec, nil
}

// LogExecutioner logs the data to stdout
// It implements the Executioner interface
type LogExecutioner struct {
	Config

	// log is the logger
	log *log.Logger
}

// New creates a new LogExecutioner based on the config
func New(cfg config.Config) (*LogExecutioner, error) {
	lcfg, err := ParseConfig(cfg.Executioner.Config)
	if err != nil {
		return nil, err
	}

	return &LogExecutioner{
		log: log.New(os.Stdout).With("tag", lcfg.Tag),
	}, nil
}

// Execute logs the data to stdout
func (e *LogExecutioner) Execute(data interface{}) error {
	e.log.Info("received data", "data", fmt.Sprintf("%v", data))
	return nil
}

// Stop signals the executioner to stop
func (e *LogExecutioner) Stop() {
	log.Info("shutting down executioner")
}
