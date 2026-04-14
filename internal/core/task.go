package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
	TaskCancelled TaskStatus = "cancelled"
)

// ValidTransitions defines which status transitions are allowed
var ValidTransitions = map[TaskStatus][]TaskStatus{
	TaskPending:   {TaskRunning, TaskCancelled},
	TaskRunning:   {TaskCompleted, TaskFailed, TaskCancelled},
	TaskCompleted: {},
	TaskFailed:    {},
	TaskCancelled: {},
}

// Task represents a unit of work sent between agents
type Task struct {
	ID          string          `json:"task_id"`
	FromAgent   string          `json:"from_agent"`
	ToAgent     string          `json:"to_agent"`
	Capability  string          `json:"capability"`
	Payload     json.RawMessage `json:"payload"`
	Status      TaskStatus      `json:"status"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       string          `json:"error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// NewTask creates a new task with the given parameters
func NewTask(fromAgent, toAgent, capability string, payload json.RawMessage) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:         uuid.New().String(),
		FromAgent:  fromAgent,
		ToAgent:    toAgent,
		Capability: capability,
		Payload:    payload,
		Status:     TaskPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate checks if the task is valid
func (t *Task) Validate() error {
	if t.FromAgent == "" {
		return errors.New("from_agent is required")
	}
	if t.ToAgent == "" {
		return errors.New("to_agent is required")
	}
	if t.Capability == "" {
		return errors.New("capability is required")
	}
	if len(t.Payload) == 0 {
		return errors.New("payload is required")
	}
	// Validate payload is valid JSON
	var js interface{}
	if err := json.Unmarshal(t.Payload, &js); err != nil {
		return errors.New("payload must be valid JSON")
	}
	return nil
}

// CanTransitionTo checks if the task can transition to the given status
func (t *Task) CanTransitionTo(status TaskStatus) bool {
	allowed, ok := ValidTransitions[t.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == status {
			return true
		}
	}
	return false
}

// SetStatus updates the task status with validation
func (t *Task) SetStatus(status TaskStatus) error {
	if !t.CanTransitionTo(status) {
		return errors.New("invalid status transition from " + string(t.Status) + " to " + string(status))
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Start marks the task as running
func (t *Task) Start() error {
	return t.SetStatus(TaskRunning)
}

// Complete marks the task as completed with a result
func (t *Task) Complete(result json.RawMessage) error {
	if err := t.SetStatus(TaskCompleted); err != nil {
		return err
	}
	t.Result = result
	now := time.Now().UTC()
	t.CompletedAt = &now
	return nil
}

// Fail marks the task as failed with an error message
func (t *Task) Fail(errMsg string) error {
	if err := t.SetStatus(TaskFailed); err != nil {
		return err
	}
	t.Error = errMsg
	now := time.Now().UTC()
	t.CompletedAt = &now
	return nil
}

// Cancel marks the task as cancelled
func (t *Task) Cancel() error {
	return t.SetStatus(TaskCancelled)
}

// IsFinal returns true if the task is in a final state
func (t *Task) IsFinal() bool {
	return t.Status == TaskCompleted || t.Status == TaskFailed || t.Status == TaskCancelled
}

// Duration returns the time taken to complete the task
func (t *Task) Duration() time.Duration {
	if t.CompletedAt == nil {
		return time.Since(t.CreatedAt)
	}
	return t.CompletedAt.Sub(t.CreatedAt)
}
