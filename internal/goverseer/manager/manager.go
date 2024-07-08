package manager

import (
	"sync"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/overseer"
)

// Manager manages overseers
type Manager struct {
	// configs is the list of overseer configurations
	configs []*config.Config

	// stop is a channel to signal the manager to stop
	stop chan struct{}

	// wg is the wait group for all overseers
	wg sync.WaitGroup
}

// NewManager creates a new Manager
// The configs are the overseer configurations
func NewManager(configs []*config.Config) *Manager {
	return &Manager{
		configs: configs,
		stop:    make(chan struct{}),
	}
}

// Run starts all overseers
func (m *Manager) Run() error {
	for _, cfg := range m.configs {
		overseer, err := overseer.NewOverseer(cfg, m.stop)
		if err != nil {
			return err
		}
		m.wg.Add(1)
		go overseer.Run(&m.wg)
	}

	return nil
}

// Stop signals all overseers to stop
func (m *Manager) Stop() {
	close(m.stop)
	m.wg.Wait()
}
