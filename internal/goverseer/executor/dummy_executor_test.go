package executor

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestDummyExecutor_Execute(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := &config.DummyExecutorConfig{}
	cfg.ValidateAndSetDefaults()

	executor := DummyExecutor{}
	err := executor.Create(cfg, log)
	assert.NoError(t, err)

	err = executor.Execute("foo")
	assert.NoError(t, err)
}
