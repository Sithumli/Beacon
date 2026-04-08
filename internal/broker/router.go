package broker

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/registry"
)

// RouterStrategy defines how tasks are routed to agents
type RouterStrategy int

const (
	// StrategyRandom selects a random healthy agent
	StrategyRandom RouterStrategy = iota
	// StrategyRoundRobin cycles through agents
	StrategyRoundRobin
	// StrategyLeastTasks selects the agent with fewest pending tasks
	StrategyLeastTasks
)

// Router handles intelligent task routing
type Router struct {
	registry *registry.Service
	strategy RouterStrategy
	mu       sync.Mutex
	rrIndex  map[string]int // round-robin index per capability
	rng      *rand.Rand
}

// NewRouter creates a new router with the given strategy
func NewRouter(reg *registry.Service, strategy RouterStrategy) *Router {
	return &Router{
		registry: reg,
		strategy: strategy,
		rrIndex:  make(map[string]int),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SelectAgent selects an agent for the given capability
func (r *Router) SelectAgent(ctx context.Context, capability string) (*core.Agent, error) {
	agents, err := r.registry.Discover(ctx, capability)
	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return nil, nil
	}

	// Filter to healthy agents only
	healthy := make([]*core.Agent, 0)
	for _, agent := range agents {
		if agent.Status == core.StatusHealthy {
			healthy = append(healthy, agent)
		}
	}

	if len(healthy) == 0 {
		return nil, nil
	}

	switch r.strategy {
	case StrategyRandom:
		return r.selectRandom(healthy), nil
	case StrategyRoundRobin:
		return r.selectRoundRobin(healthy, capability), nil
	case StrategyLeastTasks:
		// For now, fall back to random - implementing least tasks
		// would require tracking task counts
		return r.selectRandom(healthy), nil
	default:
		return r.selectRandom(healthy), nil
	}
}

func (r *Router) selectRandom(agents []*core.Agent) *core.Agent {
	r.mu.Lock()
	defer r.mu.Unlock()
	return agents[r.rng.Intn(len(agents))]
}

func (r *Router) selectRoundRobin(agents []*core.Agent, capability string) *core.Agent {
	r.mu.Lock()
	defer r.mu.Unlock()

	idx := r.rrIndex[capability]
	if idx >= len(agents) {
		idx = 0
	}

	selected := agents[idx]
	r.rrIndex[capability] = (idx + 1) % len(agents)

	return selected
}

// SetStrategy updates the routing strategy
func (r *Router) SetStrategy(strategy RouterStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategy = strategy
}
