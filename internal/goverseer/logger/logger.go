package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

// Global logger instance
var Log *log.Logger

const (
	// DefaultLogLevel is the default log level for the logger
	DefaultLogLevel = log.InfoLevel
)

// init initializes the global logger instance. It sets the output to stdout
// and the log level to the DefaultLogLevel.
func init() {
	Log = log.New(os.Stdout)
	Log.SetLevel(DefaultLogLevel)
}

// SetLevel sets the logging level for the global logger.
// It accepts a string representation of the log level,
// such as "debug", "info", "warn", "error", or "fatal".
// If an invalid level is provided, it defaults to the DefaultLogLevel
// and logs a warning message.
func SetLevel(level string) {
	// Attempt to parse the provided log level string into a log.Level value.
	lvl, err := log.ParseLevel(level)

	// If the parsing fails, it indicates an invalid log level was provided.
	if err != nil {
		Log.Warn("Invalid log level provided, using default instead",
			"level", level,
			"default", DefaultLogLevel,
			"error", err)

		// Set the log level to the default level.
		lvl = DefaultLogLevel
	}

	// Set the log level of the global logger to the determined level.
	Log.SetLevel(lvl)
}
