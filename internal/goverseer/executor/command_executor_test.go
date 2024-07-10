package executor

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/assert"
)

// NOTE: This tests that the command executor can run a command
// it does NOT test that the command itself succeeds
// this is intentional as we don't care if the command succeeds or fails as long
// as it runs and outputs to stderr/stdout
func TestCommandExecutor_Execute(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	command := "echo 'Hello, World!'"
	executor := NewCommandExecutor(command, log)
	err := executor.Execute("foo")
	assert.NoError(t, err)
}
