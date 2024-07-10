package executor

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/assert"
)

func TestDummyExecutor_Execute(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	executor := NewDummyExecutor(log)
	err := executor.Execute("foo")
	assert.NoError(t, err)
}
