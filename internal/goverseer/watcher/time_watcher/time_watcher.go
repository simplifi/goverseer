package time_watcher

import (
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/simplifi/goverseer/internal/goverseer/config"
)

const (
	// DefaultPollSeconds is the default number of seconds to wait between ticks
	DefaultPollSeconds = 1
)

// TimeWatcherConfig is the configuration for a time watcher
type Config struct {
	// PollSeconds is the number of seconds to wait between ticks
	PollSeconds int
}

// ParseConfig parses the config for a time watcher
// It validates the config, sets defaults if missing, and returns the config
func ParseConfig(config interface{}) (*Config, error) {
	cfgMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config")
	}

	twc := &Config{
		PollSeconds: DefaultPollSeconds,
	}

	if pollSeconds, ok := cfgMap["poll_seconds"].(int); ok {
		if pollSeconds < 1 {
			return nil, fmt.Errorf("poll_seconds must be greater than or equal to 1")
		}
		twc.PollSeconds = pollSeconds
	} else if cfgMap["poll_seconds"] != nil {
		return nil, fmt.Errorf("poll_seconds must be an integer")
	}

	return twc, nil
}

// TimeWatcher is a time watcher that ticks at a regular interval
type TimeWatcher struct {
	Config

	// stop is a channel to signal the watcher to stop
	stop chan struct{}
}

// New creates a new TimeWatcher based on the config
func New(cfg config.Config) (*TimeWatcher, error) {
	tcfg, err := ParseConfig(cfg.Watcher.Config)
	if err != nil {
		return nil, err
	}

	return &TimeWatcher{
		Config: Config{
			PollSeconds: tcfg.PollSeconds,
		},
		stop: make(chan struct{}),
	}, nil
}

// Watch ticks at regular intervals, sending the time to the changes channel
// The changes channel is where the path to the file is sent when it changes
func (w *TimeWatcher) Watch(change chan interface{}) {
	log.Info("starting watcher")
	for {
		select {
		case <-w.stop:
			return
		case value := <-time.After(time.Duration(w.PollSeconds) * time.Second):
			log.Info("time watcher tick", "value", value)
			change <- value
		}
	}
}

// Stop signals the watcher to stop
func (w *TimeWatcher) Stop() {
	log.Info("shutting down watcher")
	close(w.stop)
}
