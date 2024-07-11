package executioner

import (
	"log/slog"
	"os"
	"testing"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/stretchr/testify/assert"
)

func TestDummyExecutioner_Execute(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))
	cfg := &config.DummyExecutionerConfig{}
	cfg.ValidateAndSetDefaults()

	executioner := DummyExecutioner{}
	err := executioner.Create(cfg, log)
	assert.NoError(t, err)

	err = executioner.Execute("foo")
	assert.NoError(t, err)
}
