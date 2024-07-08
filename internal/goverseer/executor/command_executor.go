package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/lmittmann/tint"
)

var (
	// Ensure this implements the Executor interface
	_ Executor = (*CommandExecutor)(nil)
)

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

// Execute runs the command with the given data
func (e *CommandExecutor) Execute(data interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.log.Info("starting executor")

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", e.Command)

	// TODO: We should write the data to a file and pass the file path to the
	// command instead of passing the data directly. This will prevent command
	// injection attacks, and allow for larger data sizes.
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOVERSEER_DATA=%s", data))

	// TODO: Figure out how to merge these two pipes so we don't have to
	// do everything twice.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		e.log.Error("error creating stdout pipe", tint.Err(err))
		return
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		e.log.Error("error creating stderr pipe", tint.Err(err))
		return
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		e.log.Error("error starting command", tint.Err(err))
		return
	}

	// Stream stdout and stderr to the logger
	e.streamOutput("stdout", stdout)
	e.streamOutput("stderr", stderr)

	wait := make(chan error, 1)
	go func() {
		wait <- cmd.Wait()
	}()

	select {
	case <-e.stop:
		cancel()
		return
	case err := <-wait:
		if err != nil {
			e.log.Error("error waiting for command", tint.Err(err))
			return
		}
	}
}

// Stop signals the executor to stop
func (e *CommandExecutor) Stop() {
	e.log.Info("shutting down executor")
	close(e.stop)
}
