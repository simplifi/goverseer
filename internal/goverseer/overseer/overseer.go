package overseer

import (
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/executor"
	"github.com/simplifi/goverseer/internal/goverseer/watcher"
)

const (
	// DefaultChangeBuffer is the default size of the change buffer
	DefaultChangeBuffer = 100
)

// Runs a Watcher and listens on channel for change, triggers Action when that happens
type Overseer struct {
	// watcher is the watcher
	watcher watcher.Watcher

	// executor is the executor
	executor executor.Executor

	// log is the logger
	log *slog.Logger

	// changes is the channel through which we send changes from the watcher to the executor
	changes chan interface{}

	// stop is a channel to signal the overseer to stop
	stop chan struct{}

	// wg is the wait group for all overseer goroutines
	wg sync.WaitGroup
}

// NewOverseer creates a new Overseer
func NewOverseer(cfg *config.Config, stop chan struct{}) (*Overseer, error) {
	log := slog.New(tint.NewHandler(os.Stderr, nil)).With("overseer", cfg.Name)
	if cfg.ChangeBuffer == 0 {
		cfg.ChangeBuffer = DefaultChangeBuffer
	}
	changes := make(chan interface{}, cfg.ChangeBuffer)

	watcher, err := watcher.NewWatcher(cfg, log)
	if err != nil {
		return nil, err
	}

	executor, err := executor.NewExecutor(cfg, log)
	if err != nil {
		return nil, err
	}

	o := &Overseer{
		watcher:  watcher,
		executor: executor,
		log:      log,
		changes:  changes,
		stop:     stop,
	}

	return o, nil
}

// Run starts the overseer
func (o *Overseer) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	// Send data to the change channel for processing
	o.wg.Add(1)
	go o.watcher.Watch(o.changes, &o.wg)

	for {
		select {
		case <-o.stop:
			o.Stop()
			return
		case data := <-o.changes:
			o.wg.Add(1)
			go o.executor.Execute(data, &o.wg)
		}
	}
}

// Stop signals the overseer to stop
func (o *Overseer) Stop() {
	o.log.Info("shutting down overseer")
	o.watcher.Stop()
	o.executor.Stop()

	o.log.Info("waiting for overseer to finish")
	// Wait here so we don't close the changes channel before the executor is done
	o.wg.Wait()
	close(o.changes)
}
