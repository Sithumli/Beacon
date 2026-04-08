package store

import (
	"context"
	"sync"

	"github.com/Sithumli/Beacon/internal/core"
)

// MemoryStore implements Store interface with in-memory storage
type MemoryStore struct {
	mu     sync.RWMutex
	agents map[string]*core.Agent
	tasks  map[string]*core.Task
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		agents: make(map[string]*core.Agent),
		tasks:  make(map[string]*core.Task),
	}
}

// Close implements Store.Close
func (m *MemoryStore) Close() error {
	return nil
}

// CreateAgent stores a new agent
func (m *MemoryStore) CreateAgent(ctx context.Context, agent *core.Agent) error {
	if agent == nil {
		return ErrInvalidInput
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agent.ID]; exists {
		return ErrAlreadyExists
	}

	// Store a copy to prevent external modifications
	agentCopy := *agent
	agentCopy.Capabilities = make([]core.Capability, len(agent.Capabilities))
	copy(agentCopy.Capabilities, agent.Capabilities)
	m.agents[agent.ID] = &agentCopy

	return nil
}

// GetAgent retrieves an agent by ID
func (m *MemoryStore) GetAgent(ctx context.Context, id string) (*core.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agent, exists := m.agents[id]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy
	agentCopy := *agent
	agentCopy.Capabilities = make([]core.Capability, len(agent.Capabilities))
	copy(agentCopy.Capabilities, agent.Capabilities)
	return &agentCopy, nil
}

// UpdateAgent updates an existing agent
func (m *MemoryStore) UpdateAgent(ctx context.Context, agent *core.Agent) error {
	if agent == nil {
		return ErrInvalidInput
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agent.ID]; !exists {
		return ErrNotFound
	}

	agentCopy := *agent
	agentCopy.Capabilities = make([]core.Capability, len(agent.Capabilities))
	copy(agentCopy.Capabilities, agent.Capabilities)
	m.agents[agent.ID] = &agentCopy

	return nil
}

// DeleteAgent removes an agent by ID
func (m *MemoryStore) DeleteAgent(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[id]; !exists {
		return ErrNotFound
	}

	delete(m.agents, id)
	return nil
}

// ListAgents returns all agents matching the filter
func (m *MemoryStore) ListAgents(ctx context.Context, filter *AgentFilter) ([]*core.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*core.Agent, 0)
	for _, agent := range m.agents {
		if m.matchesAgentFilter(agent, filter) {
			agentCopy := *agent
			agentCopy.Capabilities = make([]core.Capability, len(agent.Capabilities))
			copy(agentCopy.Capabilities, agent.Capabilities)
			result = append(result, &agentCopy)
		}
	}
	return result, nil
}

func (m *MemoryStore) matchesAgentFilter(agent *core.Agent, filter *AgentFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Status != nil && agent.Status != *filter.Status {
		return false
	}

	if filter.Capability != nil && !agent.HasCapability(*filter.Capability) {
		return false
	}

	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			found := false
			for _, agentTag := range agent.Metadata.Tags {
				if agentTag == tag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}

// FindAgentsByCapability returns agents with the given capability
func (m *MemoryStore) FindAgentsByCapability(ctx context.Context, capability string) ([]*core.Agent, error) {
	filter := &AgentFilter{Capability: &capability}
	return m.ListAgents(ctx, filter)
}

// CreateTask stores a new task
func (m *MemoryStore) CreateTask(ctx context.Context, task *core.Task) error {
	if task == nil {
		return ErrInvalidInput
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; exists {
		return ErrAlreadyExists
	}

	taskCopy := *task
	m.tasks[task.ID] = &taskCopy

	return nil
}

// GetTask retrieves a task by ID
func (m *MemoryStore) GetTask(ctx context.Context, id string) (*core.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[id]
	if !exists {
		return nil, ErrNotFound
	}

	taskCopy := *task
	return &taskCopy, nil
}

// UpdateTask updates an existing task
func (m *MemoryStore) UpdateTask(ctx context.Context, task *core.Task) error {
	if task == nil {
		return ErrInvalidInput
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; !exists {
		return ErrNotFound
	}

	taskCopy := *task
	m.tasks[task.ID] = &taskCopy

	return nil
}

// DeleteTask removes a task by ID
func (m *MemoryStore) DeleteTask(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[id]; !exists {
		return ErrNotFound
	}

	delete(m.tasks, id)
	return nil
}

// ListTasks returns all tasks matching the filter
func (m *MemoryStore) ListTasks(ctx context.Context, filter *TaskFilter) ([]*core.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*core.Task, 0)
	for _, task := range m.tasks {
		if m.matchesTaskFilter(task, filter) {
			taskCopy := *task
			result = append(result, &taskCopy)
		}
	}
	return result, nil
}

func (m *MemoryStore) matchesTaskFilter(task *core.Task, filter *TaskFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Status != nil && task.Status != *filter.Status {
		return false
	}

	if filter.FromAgent != nil && task.FromAgent != *filter.FromAgent {
		return false
	}

	if filter.ToAgent != nil && task.ToAgent != *filter.ToAgent {
		return false
	}

	if filter.Capability != nil && task.Capability != *filter.Capability {
		return false
	}

	return true
}

// GetPendingTasksForAgent returns pending tasks for a specific agent
func (m *MemoryStore) GetPendingTasksForAgent(ctx context.Context, agentID string) ([]*core.Task, error) {
	status := core.TaskPending
	filter := &TaskFilter{
		ToAgent: &agentID,
		Status:  &status,
	}
	return m.ListTasks(ctx, filter)
}

// Ensure MemoryStore implements Store
var _ Store = (*MemoryStore)(nil)
