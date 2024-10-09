package shell_executioner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/logger"
)

const (
	// DataEnvVarName is the name of the environment variable that will be set
	// to the path of the file containing the data
	DataEnvVarName = "GOVERSEER_DATA"

	// DefaultShell is the default shell to use when executing a command
	DefaultShell = "/bin/sh -ec"

	// DefaultWorkDir is the default value for the work directory
	DefaultWorkDir = "/tmp"

	// DefaultPersistData is the default value for whether the command and data
	// will persist after completion
	DefaultPersistData = false
)

// Config is the configuration for a shell executioner
type Config struct {
	// Command is the command to execute
	Command string

	// Shell is the shell to use when executing the command
	// Options can also be passed to the shell here
	Shell string

	// WorkDir is the directory in which the ShellExecutioner will store
	// the data to pass into the command
	WorkDir string

	// PersistWorkDir determines whether the data will persist after completion
	// This can be useful to enable when troubleshooting configured commands but
	// should generally remain disabled
	PersistData bool
}

// ParseConfig parses the config for a log executioner
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(config interface{}) (*Config, error) {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	cfg := &Config{
		Shell:       DefaultShell,
		PersistData: DefaultPersistData,
		WorkDir:     DefaultWorkDir,
	}

	// Command is required and must be a string
	if command, ok := cfgMap["command"].(string); ok {
		if command == "" {
			return nil, fmt.Errorf("command must not be empty")
		}
		cfg.Command = command
	} else if cfgMap["command"] != nil {
		return nil, fmt.Errorf("command must be a string")
	} else {
		return nil, fmt.Errorf("command is required")
	}

	// If shell is set, it should be a string
	if cfgMap["shell"] != nil {
		if shell, ok := cfgMap["shell"].(string); ok {
			if shell == "" {
				return nil, fmt.Errorf("shell must not be empty")
			}
			cfg.Shell = shell
		} else if cfgMap["shell"] != nil {
			return nil, fmt.Errorf("shell must be a string")
		}
	}

	// If persist_data is set, it should be a string
	if cfgMap["persist_data"] != nil {
		if persistData, ok := cfgMap["persist_data"].(bool); ok {
			cfg.PersistData = persistData
		} else if cfgMap["persist_data"] != nil {
			return nil, fmt.Errorf("persist_data must be a boolean")
		}
	}

	// If work_dir is set, it should be a string
	if cfgMap["work_dir"] != nil {
		if workDir, ok := cfgMap["work_dir"].(string); ok {
			if workDir == "" {
				return nil, fmt.Errorf("work_dir must not be empty")
			}
			cfg.WorkDir = workDir
		} else if cfgMap["work_dir"] != nil {
			return nil, fmt.Errorf("work_dir must be a string")
		}
	}

	return cfg, nil
}

// ShellExecutioner runs a shell command
// It implements the Executioner interface
type ShellExecutioner struct {
	Config

	// stop is a channel to signal the executor to stop
	stop chan struct{}

	// ctx is the context used to control the command and output scanner
	ctx context.Context

	// cancel is the function to cancel the context
	cancel context.CancelFunc
}

// New creates a new ShellExecutioner based on the config
func New(cfg config.Config) (*ShellExecutioner, error) {
	pcfg, err := ParseConfig(cfg.Executioner.Config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ShellExecutioner{
		Config: Config{
			Command:     pcfg.Command,
			Shell:       pcfg.Shell,
			PersistData: pcfg.PersistData,
			WorkDir:     pcfg.WorkDir,
		},
		stop:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (e *ShellExecutioner) enableOutputStreaming(cmd *exec.Cmd) error {
	// Stream stdout of the command to the logger
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}
	go e.streamOutput(stdOut, "info")

	// Stream stderr of the command to the logger
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}
	go e.streamOutput(stdErr, "error")

	return nil
}

// streamOutput streams the output of a pipe to the logger
func (e *ShellExecutioner) streamOutput(pipe io.ReadCloser, outputType string) {
	scanner := bufio.NewScanner(pipe)
	for {
		select {
		case <-e.ctx.Done():
			logger.Log.Info("stopping output scanner")
			return
		default:
			if scanner.Scan() {
				if outputType == "error" {
					logger.Log.Error("command", "output", scanner.Text())
				} else {
					logger.Log.Info("command", "output", scanner.Text())
				}
			} else {
				if err := scanner.Err(); err != nil {
					// Avoid logging errors if the context was canceled mid-scan
					// This will happen when the executioner is being stopped
					if e.ctx.Err() == context.Canceled {
						continue
					}
					logger.Log.Error("error reading output", "err", err)
				}
				return
			}
		}
	}
}

// writeTempData writes the data to a file in the temporary work directory
// It returns the path to the file and an error if the data could not be written
func (e *ShellExecutioner) writeTempData(data interface{}) (string, error) {
	tempDataFile, err := os.CreateTemp(e.WorkDir, "goverseer")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer tempDataFile.Close()

	if _, err := tempDataFile.WriteString(data.(string)); err != nil {
		return "", fmt.Errorf("error writing data to temp file: %w", err)
	}

	logger.Log.Info("wrote data to work dir", "path", tempDataFile.Name())
	return tempDataFile.Name(), nil
}

// Execute runs the command with the given data
// It returns an error if the command could not be started or if the command
// returned an error.
// The data is written to a temp file and the path is passed to the command via
// the DataEnvVarName environment variable.
// The command is run in the configured shell.
func (e *ShellExecutioner) Execute(data interface{}) error {
	var tempDataPath string
	var err error

	// Write the data passed in from the watcher to a temp file in the work dir
	if tempDataPath, err = e.writeTempData(data); err != nil {
		return fmt.Errorf("error writing data: %w", err)
	}

	if e.PersistData {
		logger.Log.Warn("persisting data", "path", tempDataPath)
	} else {
		defer os.Remove(tempDataPath)
	}

	// Build the command to run
	// Split the Shell so we can pass the args to exec.Command the way it expects
	// Pass the path to the data file via the DataEnvVarName environment variable
	shellParts := strings.Split(e.Shell, " ")
	cmd := exec.CommandContext(e.ctx, shellParts[0], append(shellParts[1:], e.Command)...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", DataEnvVarName, tempDataPath))

	// Stream command output to the logger
	if err := e.enableOutputStreaming(cmd); err != nil {
		return fmt.Errorf("error enabling output streaming: %w", err)
	}

	// Start the command running
	// This does not block and depends on the caller to call cmd.Wait()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Wait for the command to finish running, but don't block otherwise we'll
	// never be able to stop the executor if the command hangs
	wait := make(chan error, 1)
	go func() {
		wait <- cmd.Wait()
	}()

	// Block here waiting for the command to complete or for the executor to stop
	select {
	case <-e.stop:
		e.cancel()
		return nil
	case err := <-wait:
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}
	}

	return nil
}

// Stop signals the executioner to stop
func (e *ShellExecutioner) Stop() {
	logger.Log.Info("shutting down executor")
	close(e.stop)
}
