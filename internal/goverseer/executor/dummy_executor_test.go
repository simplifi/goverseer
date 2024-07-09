package executor

import (
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/lmittmann/tint"
)

// TODO: We don't have a good way to test this right now. We need to refactor
// the code to make it testable. Perhaps we could add a channel to the Execute
// method that we can use to send to the test?
func TestDummyExecutor_Execute(t *testing.T) {
	log := slog.New(tint.NewHandler(os.Stderr, &tint.Options{Level: slog.LevelError}))

	executor := NewDummyExecutor(log)

	// Create a wait group to wait for the execution to finish
	var wg sync.WaitGroup
	wg.Add(1)

	// Execute the command
	executor.Execute("foo", &wg)

	// Wait for the execution to finish
	wg.Wait()
}
