package store

import (
	"context"
	"errors"

	"github.com/Sithumli/Beacon/internal/core"
)

// Common errors
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
)

// AgentFilter defines criteria for filtering agents
type AgentFilter struct {
	Status     *core.AgentStatus
	Capability *string
	Tags       []string
}

// TaskFilter defines criteria for filtering tasks
type TaskFilter struct {
	Status     *core.TaskStatus
	FromAgent  *string
	ToAgent    *string
	Capability *string
	Limit      *int
	Offset     *int
}

// Store defines the interface for data persistence
type Store interface {
	AgentStore
	TaskStore
	Close() error
}

// AgentStore defines operations for agent persistence
type AgentStore interface {
	// CreateAgent stores a new agent
	CreateAgent(ctx context.Context, agent *core.Agent) error

	// GetAgent retrieves an agent by ID
	GetAgent(ctx context.Context, id string) (*core.Agent, error)

	// UpdateAgent updates an existing agent
	UpdateAgent(ctx context.Context, agent *core.Agent) error

	// DeleteAgent removes an agent by ID
	DeleteAgent(ctx context.Context, id string) error

	// ListAgents returns all agents matching the filter
	ListAgents(ctx context.Context, filter *AgentFilter) ([]*core.Agent, error)

	// FindAgentsByCapability returns agents with the given capability
	FindAgentsByCapability(ctx context.Context, capability string) ([]*core.Agent, error)
}

// TaskStore defines operations for task persistence
type TaskStore interface {
	// CreateTask stores a new task
	CreateTask(ctx context.Context, task *core.Task) error

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, id string) (*core.Task, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *core.Task) error

	// DeleteTask removes a task by ID
	DeleteTask(ctx context.Context, id string) error

	// ListTasks returns all tasks matching the filter
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*core.Task, error)

	// GetPendingTasksForAgent returns pending tasks for a specific agent
	GetPendingTasksForAgent(ctx context.Context, agentID string) ([]*core.Task, error)
}
