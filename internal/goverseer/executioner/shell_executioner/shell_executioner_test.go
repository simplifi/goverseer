package shell_executioner

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	var parsedConfig *Config

	parsedConfig, err := ParseConfig(map[string]interface{}{
		"command": "echo 123",
	})
	assert.NoError(t, err,
		"Parsing a valid config should not return an error")
	assert.Equal(t, "echo 123", parsedConfig.Command,
		"Command should be set to the value in the config")
	assert.Equal(t, DefaultShell, parsedConfig.Shell,
		"Shell should be set to the default value")

	// Test setting the shell
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"command": "echo 123",
		"shell":   "/bin/bash",
	})
	assert.NoError(t, err,
		"Parsing a valid config should not return an error")
	assert.Equal(t, "/bin/bash", parsedConfig.Shell,
		"Shell should be set to the value in the config")

	parsedConfig, err = ParseConfig(map[string]interface{}{
		"command": "echo 123",
		"shell":   nil,
	})
	assert.NoError(t, err,
		"Parsing a config with a nil shell should not return an error")
	assert.Equal(t, DefaultShell, parsedConfig.Shell,
		"Parsing a config with a nil shell should set default value")

	_, err = ParseConfig(map[string]interface{}{
		"command": "echo 123",
		"shell":   1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect shell type should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"command": "echo 123",
		"shell":   "",
	})
	assert.Error(t, err,
		"Parsing a config with an empty shell should return an error")

	// Test setting the command
	parsedConfig, err = ParseConfig(map[string]interface{}{
		"command": "echo 123",
	})
	assert.NoError(t, err,
		"Parsing a config with a valid command should not return an error")
	assert.Equal(t, "echo 123", parsedConfig.Command,
		"Command should be set to the value in the config")

	_, err = ParseConfig(map[string]interface{}{
		"command": 1,
	})
	assert.Error(t, err,
		"Parsing a config with an incorrect command type should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"command": "",
	})
	assert.Error(t, err,
		"Parsing a config with an empty command should return an error")

	_, err = ParseConfig(map[string]interface{}{
		"command": nil,
	})
	assert.Error(t, err,
		"Parsing a config with a nil command should return an error")

	_, err = ParseConfig(map[string]interface{}{})
	assert.Error(t, err,
		"Parsing a config with no command should return an error")
}

func TestNew(t *testing.T) {
	var cfg config.Config
	cfg = config.Config{
		Name: "TestConfig",
		Executioner: config.ExecutionerConfig{
			Type: "shell",
			Config: map[string]interface{}{
				"command": "echo 123",
			},
		},
	}
	executioner, err := New(cfg, slog.Default())
	assert.NoError(t, err,
		"Creating a new ShellExecutioner should not return an error")
	assert.NotNil(t, executioner,
		"Creating a new ShellExecutioner should return an executioner")

	cfg = config.Config{
		Name: "TestConfig",
		Executioner: config.ExecutionerConfig{
			Type: "shell",
		},
	}
	executioner, err = New(cfg, slog.Default())
	assert.Error(t, err,
		"Creating a new ShellExecutioner with an invalid config should return an error")
	assert.Nil(t, executioner,
		"Creating a new ShellExecutioner with an invalid config should not return an executioner")
}

func TestShellExecutioner_Execute(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	tempDir, _ := os.MkdirTemp("", "goverseer-test")
	ctx, cancel := context.WithCancel(context.Background())
	executioner := ShellExecutioner{
		Config: Config{
			Command: "echo ${GOVERSEER_DATA}",
			Shell:   DefaultShell,
		},
		workDir: tempDir,
		log:     log,
		stop:    make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}

	err := executioner.Execute("test_data")
	assert.NoError(t, err,
		"Executing a valid command should not return an error")
}

func TestShellExecutioner_Stop(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	tempDir, _ := os.MkdirTemp("", "goverseer-test")
	ctx, cancel := context.WithCancel(context.Background())
	executioner := ShellExecutioner{
		Config: Config{
			Command: "for i in {1..1000}; do echo $i; sleep 1; done",
			Shell:   DefaultShell,
		},
		workDir: tempDir,
		log:     log,
		stop:    make(chan struct{}),
		ctx:     ctx,
		cancel:  cancel,
	}

	go func() {
		executioner.Execute("test_data")
	}()

	executioner.Stop()

	// Sleep for a second to allow the executioner time to stop
	time.Sleep(1 * time.Second)

	assert.Equal(t, 0, len(executioner.stop),
		"Stopping the executioner should close the stop channel")

	assert.Equal(t, context.Canceled, executioner.ctx.Err(),
		"Stopping the executioner should cancel the context")
}
