package registry

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
	"github.com/rs/zerolog/log"
)

// Service handles agent registration and discovery
type Service struct {
	store        store.Store
	mu           sync.RWMutex
	watchers     map[string]*watcher
	watcherCount int
}

// WatchEvent represents a registry change event
type WatchEvent struct {
	Type  EventType
	Agent *core.Agent
}

// EventType represents the type of registry event
type EventType int

const (
	EventRegistered EventType = iota
	EventDeregistered
	EventUpdated
	EventHealthChanged
)

// NewService creates a new registry service
func NewService(s store.Store) *Service {
	return &Service{
		store:    s,
		watchers: make(map[string]*watcher),
	}
}

// Register registers a new agent
func (s *Service) Register(ctx context.Context, agent *core.Agent) (*core.Agent, error) {
	if err := agent.Validate(); err != nil {
		return nil, err
	}

	// Generate ID if not set
	if agent.ID == "" {
		newAgent := core.NewAgent(agent.Name, agent.Version, agent.Description)
		agent.ID = newAgent.ID
		agent.RegisteredAt = newAgent.RegisteredAt
		agent.LastHeartbeat = newAgent.LastHeartbeat
		agent.Status = newAgent.Status
	}

	if err := s.store.CreateAgent(ctx, agent); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, errors.New("agent already registered")
		}
		return nil, err
	}

	log.Info().
		Str("agent_id", agent.ID).
		Str("agent_name", agent.Name).
		Str("version", agent.Version).
		Int("capabilities", len(agent.Capabilities)).
		Msg("Agent registered")

	s.notifyWatchers(WatchEvent{Type: EventRegistered, Agent: agent})

	return agent, nil
}

// Deregister removes an agent from the registry
func (s *Service) Deregister(ctx context.Context, agentID string) error {
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAgent(ctx, agentID); err != nil {
		return err
	}

	log.Info().
		Str("agent_id", agentID).
		Str("agent_name", agent.Name).
		Msg("Agent deregistered")

	s.notifyWatchers(WatchEvent{Type: EventDeregistered, Agent: agent})

	return nil
}

// GetAgent retrieves an agent by ID
func (s *Service) GetAgent(ctx context.Context, agentID string) (*core.Agent, error) {
	return s.store.GetAgent(ctx, agentID)
}

// ListAgents returns all agents, optionally filtered
func (s *Service) ListAgents(ctx context.Context, filter *store.AgentFilter) ([]*core.Agent, error) {
	return s.store.ListAgents(ctx, filter)
}

// Discover finds agents with a specific capability
func (s *Service) Discover(ctx context.Context, capability string) ([]*core.Agent, error) {
	agents, err := s.store.FindAgentsByCapability(ctx, capability)
	if err != nil {
		return nil, err
	}

	// Filter to only healthy agents
	healthy := make([]*core.Agent, 0)
	for _, agent := range agents {
		if agent.Status == core.StatusHealthy {
			healthy = append(healthy, agent)
		}
	}

	return healthy, nil
}

// Heartbeat updates the last heartbeat time for an agent
func (s *Service) Heartbeat(ctx context.Context, agentID string) error {
	agent, err := s.store.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	oldStatus := agent.Status
	agent.UpdateHeartbeat()

	if err := s.store.UpdateAgent(ctx, agent); err != nil {
		return err
	}

	log.Debug().
		Str("agent_id", agentID).
		Time("heartbeat", agent.LastHeartbeat).
		Msg("Heartbeat received")

	if oldStatus != core.StatusHealthy {
		s.notifyWatchers(WatchEvent{Type: EventHealthChanged, Agent: agent})
	}

	return nil
}

// watcher holds a channel and optional capability filter
type watcher struct {
	ch           chan WatchEvent
	capabilities []string
}

// Watch registers a channel to receive registry events
// If capabilities is non-empty, only events for agents with matching capabilities will be sent
func (s *Service) Watch(capabilities []string) (string, <-chan WatchEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.watcherCount++
	id := "watcher-" + strconv.Itoa(s.watcherCount)
	ch := make(chan WatchEvent, 100)
	s.watchers[id] = &watcher{
		ch:           ch,
		capabilities: capabilities,
	}

	return id, ch
}

// Unwatch removes a watcher
func (s *Service) Unwatch(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, ok := s.watchers[id]; ok {
		close(w.ch)
		delete(s.watchers, id)
	}
}

func (s *Service) notifyWatchers(event WatchEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, w := range s.watchers {
		// If capabilities filter is set, only send events for agents with matching capabilities
		if len(w.capabilities) > 0 && event.Agent != nil {
			match := false
			for _, cap := range w.capabilities {
				if event.Agent.HasCapability(cap) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		select {
		case w.ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// UpdateAgent updates an agent's information
func (s *Service) UpdateAgent(ctx context.Context, agent *core.Agent) error {
	if err := agent.Validate(); err != nil {
		return err
	}

	existing, err := s.store.GetAgent(ctx, agent.ID)
	if err != nil {
		return err
	}

	// Preserve registration time
	agent.RegisteredAt = existing.RegisteredAt

	if err := s.store.UpdateAgent(ctx, agent); err != nil {
		return err
	}

	log.Info().
		Str("agent_id", agent.ID).
		Str("agent_name", agent.Name).
		Msg("Agent updated")

	s.notifyWatchers(WatchEvent{Type: EventUpdated, Agent: agent})

	return nil
}
