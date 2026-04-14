package broker

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/registry"
	"github.com/Sithumli/Beacon/internal/store"
	"github.com/rs/zerolog/log"
)

// Service handles task routing between agents
type Service struct {
	store        store.Store
	registry     *registry.Service
	mu           sync.RWMutex
	subscribers  map[string]chan TaskEvent
	subCount     int
}

// TaskEvent represents a task-related event
type TaskEvent struct {
	Type EventType
	Task *core.Task
}

// EventType represents the type of task event
type EventType int

const (
	EventNewTask EventType = iota
	EventTaskUpdated
	EventTaskCancelled
)

// NewService creates a new broker service
func NewService(s store.Store, reg *registry.Service) *Service {
	return &Service{
		store:       s,
		registry:    reg,
		subscribers: make(map[string]chan TaskEvent),
	}
}

// SendTask sends a task to a specific agent
func (s *Service) SendTask(ctx context.Context, fromAgent, toAgent, capability string, payload json.RawMessage) (*core.Task, error) {
	// Verify target agent exists and has the capability
	agent, err := s.registry.GetAgent(ctx, toAgent)
	if err != nil {
		return nil, errors.New("target agent not found")
	}

	if !agent.HasCapability(capability) {
		return nil, errors.New("agent does not have capability: " + capability)
	}

	if agent.Status != core.StatusHealthy {
		return nil, errors.New("agent is not healthy")
	}

	// Create and store task
	task := core.NewTask(fromAgent, toAgent, capability, payload)
	if err := task.Validate(); err != nil {
		return nil, err
	}

	if err := s.store.CreateTask(ctx, task); err != nil {
		return nil, err
	}

	log.Info().
		Str("task_id", task.ID).
		Str("from", fromAgent).
		Str("to", toAgent).
		Str("capability", capability).
		Msg("Task created")

	s.notifySubscribers(toAgent, TaskEvent{Type: EventNewTask, Task: task})

	return task, nil
}

// RouteTask routes a task to any available agent with the capability
func (s *Service) RouteTask(ctx context.Context, fromAgent, capability string, payload json.RawMessage) (*core.Task, error) {
	// Find agents with the capability
	agents, err := s.registry.Discover(ctx, capability)
	if err != nil {
		return nil, err
	}

	if len(agents) == 0 {
		return nil, errors.New("no agents available with capability: " + capability)
	}

	// Simple selection: pick the first healthy agent
	// In production, you might use load balancing or other strategies
	var selectedAgent *core.Agent
	for _, agent := range agents {
		if agent.Status == core.StatusHealthy {
			selectedAgent = agent
			break
		}
	}

	if selectedAgent == nil {
		return nil, errors.New("no healthy agents available")
	}

	return s.SendTask(ctx, fromAgent, selectedAgent.ID, capability, payload)
}

// GetTask retrieves a task by ID
func (s *Service) GetTask(ctx context.Context, taskID string) (*core.Task, error) {
	return s.store.GetTask(ctx, taskID)
}

// UpdateTask updates a task's status and optionally result/error
func (s *Service) UpdateTask(ctx context.Context, taskID string, status core.TaskStatus, result json.RawMessage, errMsg string) (*core.Task, error) {
	task, err := s.store.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	switch status {
	case core.TaskRunning:
		if err := task.Start(); err != nil {
			return nil, err
		}
	case core.TaskCompleted:
		if err := task.Complete(result); err != nil {
			return nil, err
		}
	case core.TaskFailed:
		if err := task.Fail(errMsg); err != nil {
			return nil, err
		}
	case core.TaskCancelled:
		if err := task.Cancel(); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid status")
	}

	if err := s.store.UpdateTask(ctx, task); err != nil {
		return nil, err
	}

	log.Info().
		Str("task_id", taskID).
		Str("status", string(task.Status)).
		Msg("Task updated")

	eventType := EventTaskUpdated
	if status == core.TaskCancelled {
		eventType = EventTaskCancelled
	}
	s.notifySubscribers(task.ToAgent, TaskEvent{Type: eventType, Task: task})

	return task, nil
}

// ListTasks returns tasks matching the filter
func (s *Service) ListTasks(ctx context.Context, filter *store.TaskFilter) ([]*core.Task, error) {
	return s.store.ListTasks(ctx, filter)
}

// CancelTask cancels a pending or running task
func (s *Service) CancelTask(ctx context.Context, taskID string) (*core.Task, error) {
	return s.UpdateTask(ctx, taskID, core.TaskCancelled, nil, "")
}

// Subscribe registers an agent to receive task events
func (s *Service) Subscribe(agentID string) (string, <-chan TaskEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subCount++
	id := agentID + "-sub-" + strconv.Itoa(s.subCount)
	ch := make(chan TaskEvent, 100)
	s.subscribers[id] = ch

	return id, ch
}

// Unsubscribe removes a subscription
func (s *Service) Unsubscribe(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[id]; ok {
		close(ch)
		delete(s.subscribers, id)
	}
}

func (s *Service) notifySubscribers(agentID string, event TaskEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for subID, ch := range s.subscribers {
		// Match subscriptions for this agent
		if len(subID) >= len(agentID) && subID[:len(agentID)] == agentID {
			select {
			case ch <- event:
			default:
				// Channel full, skip
			}
		}
	}
}

// GetPendingTasks returns pending tasks for an agent
func (s *Service) GetPendingTasks(ctx context.Context, agentID string) ([]*core.Task, error) {
	return s.store.GetPendingTasksForAgent(ctx, agentID)
}
