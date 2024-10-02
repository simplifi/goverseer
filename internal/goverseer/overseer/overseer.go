package overseer

import (
	"sync"

	"github.com/charmbracelet/log"
	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/executioner"
	"github.com/simplifi/goverseer/internal/goverseer/watcher"
)

// Runs a Watcher and listens on channel for change, triggers Action when that happens
// NOTE: The change channel does not have a buffer, so it will block until the
// executioner is ready to process the change.
type Overseer struct {
	// watcher is the watcher
	watcher watcher.Watcher

	// executioner is the executioner
	executioner executioner.Executioner

	// change is the channel through which we send changes from the watcher to the executioner
	change chan interface{}

	// stop is a channel to signal the overseer to stop
	stop chan struct{}

	// waitGroup is the wait group for all overseer goroutines
	waitGroup sync.WaitGroup
}

// New creates a new Overseer
func New(cfg *config.Config) (*Overseer, error) {
	watcher, err := watcher.New(cfg)
	if err != nil {
		return nil, err
	}

	executioner, err := executioner.New(cfg)
	if err != nil {
		return nil, err
	}

	o := &Overseer{
		watcher:     watcher,
		executioner: executioner,
		change:      make(chan interface{}),
		stop:        make(chan struct{}),
	}

	return o, nil
}

// Run starts the overseer
func (o *Overseer) Run() {
	// Send data to the change channel for processing
	o.waitGroup.Add(1)
	go func() {
		defer o.waitGroup.Done()
		o.watcher.Watch(o.change)
	}()

	for {
		select {
		case <-o.stop:
			return
		case data := <-o.change:
			o.waitGroup.Add(1)
			go func() {
				defer o.waitGroup.Done()
				if err := o.executioner.Execute(data); err != nil {
					log.Error("error running executioner", "err", err)
				}
			}()
		}
	}
}

// Stop signals the overseer to stop
func (o *Overseer) Stop() {
	log.Info("shutting down overseer")
	close(o.stop)
	o.watcher.Stop()
	o.executioner.Stop()

	log.Info("waiting for overseer to finish")
	// Wait here so we don't close the changes channel before the executioner is done
	o.waitGroup.Wait()
	log.Info("done")
	close(o.change)
}
