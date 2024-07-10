package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

var (
	// Ensure this implements the Executor interface
	_ Executor = (*CommandExecutor)(nil)
)

func init() {
	factory.Register("command", func(cfg interface{}, log *slog.Logger) (Executor, error) {
		v, ok := cfg.(config.CommandExecutorConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config for command executor")
		}
		return NewCommandExecutor(v.Command, log), nil
	})
}

// CommandExecutor logs the data to stdout
type CommandExecutor struct {
	// Command is the command to execute
	Command string

	// log is the logger
	log *slog.Logger

	// stop is a channel to signal the executor to stop
	stop chan struct{}
}

// NewCommandExecutor creates a new CommandExecutor
// The command is the command to execute
// The log is the logger
func NewCommandExecutor(Command string, log *slog.Logger) *CommandExecutor {
	return &CommandExecutor{
		Command: Command,
		log:     log,
		stop:    make(chan struct{}),
	}
}

// streamOutput streams the output of the pipe to the logger
func (e *CommandExecutor) streamOutput(name string, pipe io.ReadCloser) {
	go func() {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			e.log.Info("command output", slog.String(name, scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			e.log.Error(fmt.Sprintf("error reading %s", name), tint.Err(err))
		}
	}()
}

func (e *CommandExecutor) cacheData(data interface{}) (string, error) {
	// Create a temporary file in the temporary directory
	tempFile, err := os.CreateTemp("", "data")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if _, err := tempFile.Write([]byte(data.(string))); err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

// Execute runs the command with the given data
func (e *CommandExecutor) Execute(data interface{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.log.Info("starting executor")

	cache, err := e.cacheData(data)
	if err != nil {
		return err
	}
	defer os.Remove(cache)

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", e.Command)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOVERSEER_DATA=%s", cache))

	// TODO: Figure out how to merge these two pipes so we don't have to
	// do everything twice.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	// Stream stdout and stderr to the logger
	e.streamOutput("stdout", stdout)
	e.streamOutput("stderr", stderr)

	// Wait for the command to finish running, but don't block otherwise we'll
	// never be able to stop the executor if the command hangs
	wait := make(chan error, 1)
	go func() {
		wait <- cmd.Wait()
	}()

	select {
	case <-e.stop:
		cancel()
		return nil
	case err := <-wait:
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop signals the executor to stop
func (e *CommandExecutor) Stop() {
	e.log.Info("shutting down executor")
	close(e.stop)
}
