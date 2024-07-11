package overseer

import (
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/executioner"
	"github.com/simplifi/goverseer/internal/goverseer/watcher"
)

const (
	// DefaultChangeBuffer is the default size of the change buffer
	// This is the number of changes that can be buffered before the watcher blocks
	DefaultChangeBuffer = 100
)

// Runs a Watcher and listens on channel for change, triggers Action when that happens
type Overseer struct {
	// watcher is the watcher
	watcher watcher.Watcher

	// executioner is the executioner
	executioner executioner.Executioner

	// log is the logger
	log *slog.Logger

	// changes is the channel through which we send changes from the watcher to the executioner
	changes chan interface{}

	// stop is a channel to signal the overseer to stop
	stop chan struct{}

	// wg is the wait group for all overseer goroutines
	wg sync.WaitGroup
}

// New creates a new Overseer
func New(cfg *config.Config, stop chan struct{}) (*Overseer, error) {
	// Setup the logger
	log := slog.
		New(tint.NewHandler(os.Stdout, nil)).
		With("overseer", cfg.Name)

	if cfg.ChangeBuffer == 0 {
		cfg.ChangeBuffer = DefaultChangeBuffer
	}
	changes := make(chan interface{}, cfg.ChangeBuffer)

	watcher, err := watcher.New(cfg)
	if err != nil {
		return nil, err
	}

	executioner, err := executioner.New(cfg)
	if err != nil {
		return nil, err
	}

	o := &Overseer{
		watcher:     *watcher,
		executioner: *executioner,
		log:         log,
		changes:     changes,
		stop:        stop,
	}

	return o, nil
}

// Run starts the overseer
func (o *Overseer) Run() {
	// Send data to the change channel for processing
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		o.watcher.Watch(o.changes)
	}()

	for {
		select {
		case <-o.stop:
			o.Stop()
			return
		case data := <-o.changes:
			o.wg.Add(1)
			go func() {
				defer o.wg.Done()
				if err := o.executioner.Execute(data); err != nil {
					o.log.Error("error running executioner", tint.Err(err))
				}
			}()
		}
	}
}

// Stop signals the overseer to stop
func (o *Overseer) Stop() {
	o.log.Info("shutting down overseer")
	o.watcher.Stop()
	o.executioner.Stop()

	o.log.Info("waiting for overseer to finish")
	// Wait here so we don't close the changes channel before the executioner is done
	o.wg.Wait()
	o.log.Info("done")
	close(o.changes)
}
