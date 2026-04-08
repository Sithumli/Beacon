package health

import (
	"context"
	"sync"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
	"github.com/rs/zerolog/log"
)

// Config holds health monitor configuration
type Config struct {
	CheckInterval time.Duration // How often to check agent health
	HeartbeatTTL  time.Duration // How long before an agent is marked unhealthy
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		CheckInterval: 10 * time.Second,
		HeartbeatTTL:  30 * time.Second,
	}
}

// Monitor watches agent health and marks stale agents as unhealthy
type Monitor struct {
	config   Config
	store    store.Store
	mu       sync.RWMutex
	watchers []chan<- HealthEvent
	stopCh   chan struct{}
	running  bool
}

// HealthEvent represents a change in agent health
type HealthEvent struct {
	AgentID   string
	OldStatus core.AgentStatus
	NewStatus core.AgentStatus
	Timestamp time.Time
}

// NewMonitor creates a new health monitor
func NewMonitor(s store.Store, cfg Config) *Monitor {
	return &Monitor{
		config:   cfg,
		store:    s,
		watchers: make([]chan<- HealthEvent, 0),
		stopCh:   make(chan struct{}),
	}
}

// Start begins the health monitoring loop
func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}
	m.running = true
	m.mu.Unlock()

	log.Info().
		Dur("interval", m.config.CheckInterval).
		Dur("ttl", m.config.HeartbeatTTL).
		Msg("Starting health monitor")

	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.stopCh:
			return nil
		case <-ticker.C:
			m.checkAgents(ctx)
		}
	}
}

// Stop stops the health monitor
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}
	m.running = false
	close(m.stopCh)
}

// Watch registers a channel to receive health events
func (m *Monitor) Watch(ch chan<- HealthEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.watchers = append(m.watchers, ch)
}

// Unwatch removes a channel from receiving health events
func (m *Monitor) Unwatch(ch chan<- HealthEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, w := range m.watchers {
		if w == ch {
			m.watchers = append(m.watchers[:i], m.watchers[i+1:]...)
			return
		}
	}
}

func (m *Monitor) checkAgents(ctx context.Context) {
	agents, err := m.store.ListAgents(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list agents for health check")
		return
	}

	for _, agent := range agents {
		if agent.IsExpired(m.config.HeartbeatTTL) && agent.Status == core.StatusHealthy {
			oldStatus := agent.Status
			agent.MarkUnhealthy()

			if err := m.store.UpdateAgent(ctx, agent); err != nil {
				log.Error().
					Err(err).
					Str("agent_id", agent.ID).
					Msg("Failed to update agent status")
				continue
			}

			log.Warn().
				Str("agent_id", agent.ID).
				Str("agent_name", agent.Name).
				Time("last_heartbeat", agent.LastHeartbeat).
				Msg("Agent marked unhealthy due to missed heartbeat")

			m.notifyWatchers(HealthEvent{
				AgentID:   agent.ID,
				OldStatus: oldStatus,
				NewStatus: core.StatusUnhealthy,
				Timestamp: time.Now(),
			})
		}
	}
}

func (m *Monitor) notifyWatchers(event HealthEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.watchers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}
