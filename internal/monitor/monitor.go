package monitor

import (
	"log"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/portcheck"
	"github.com/user/portwatch/internal/runner"
)

// Monitor polls configured ports and triggers actions on state changes.
type Monitor struct {
	cfg     *config.Config
	checker *portcheck.Checker
	states  map[string]portcheck.State
	stop    chan struct{}
}

// New creates a new Monitor from the given config.
func New(cfg *config.Config) *Monitor {
	return &Monitor{
		cfg:     cfg,
		checker: portcheck.NewChecker(2 * time.Second),
		states:  make(map[string]portcheck.State),
		stop:    make(chan struct{}),
	}
}

// Start begins polling all configured ports at the configured interval.
func (m *Monitor) Start() {
	interval := time.Duration(m.cfg.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run an immediate check before waiting for the first tick.
	m.checkAll()

	for {
		select {
		case <-ticker.C:
			m.checkAll()
		case <-m.stop:
			return
		}
	}
}

// Stop signals the monitor to cease polling.
func (m *Monitor) Stop() {
	close(m.stop)
}

func (m *Monitor) checkAll() {
	for _, entry := range m.cfg.Ports {
		key := entry.Address()
		newState := m.checker.Check(entry.Host, entry.Port)

		prev, seen := m.states[key]
		if !seen || prev != newState {
			m.states[key] = newState
			if seen {
				log.Printf("[portwatch] %s state changed: %s -> %s", key, prev, newState)
				m.trigger(entry, newState)
			} else {
				log.Printf("[portwatch] %s initial state: %s", key, newState)
			}
		}
	}
}

func (m *Monitor) trigger(entry config.PortEntry, state portcheck.State) {
	for _, action := range entry.Actions {
		if action.On != string(state) {
			continue
		}
		p := runner.Payload{Host: entry.Host, Port: entry.Port, State: string(state)}
		if action.Webhook != "" {
			if err := runner.RunWebhook(action.Webhook, p); err != nil {
				log.Printf("[portwatch] webhook error for %s: %v", entry.Address(), err)
			}
		}
		if action.Command != "" {
			if err := runner.RunCommand(action.Command, p); err != nil {
				log.Printf("[portwatch] command error for %s: %v", entry.Address(), err)
			}
		}
	}
}
