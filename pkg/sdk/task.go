package sdk

import (
	"encoding/json"
	"time"
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

// TaskInfo contains information about a task
type TaskInfo struct {
	ID          string
	FromAgent   string
	ToAgent     string
	Capability  string
	Payload     json.RawMessage
	Status      TaskStatus
	Result      json.RawMessage
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// IsFinal returns true if the task is in a final state
func (t *TaskInfo) IsFinal() bool {
	return t.Status == TaskCompleted || t.Status == TaskFailed || t.Status == TaskCancelled
}

// IsSuccess returns true if the task completed successfully
func (t *TaskInfo) IsSuccess() bool {
	return t.Status == TaskCompleted
}

// Duration returns the time taken for the task
func (t *TaskInfo) Duration() time.Duration {
	if t.CompletedAt == nil {
		return time.Since(t.CreatedAt)
	}
	return t.CompletedAt.Sub(t.CreatedAt)
}

// GetResult unmarshals the result into the provided struct
func (t *TaskInfo) GetResult(v interface{}) error {
	return json.Unmarshal(t.Result, v)
}

// GetPayload unmarshals the payload into the provided struct
func (t *TaskInfo) GetPayload(v interface{}) error {
	return json.Unmarshal(t.Payload, v)
}

// TaskBuilder helps build task requests
type TaskBuilder struct {
	fromAgent  string
	toAgent    string
	capability string
	payload    interface{}
}

// NewTask creates a new task builder
func NewTask(capability string) *TaskBuilder {
	return &TaskBuilder{
		capability: capability,
	}
}

// From sets the source agent
func (b *TaskBuilder) From(agentID string) *TaskBuilder {
	b.fromAgent = agentID
	return b
}

// To sets the target agent
func (b *TaskBuilder) To(agentID string) *TaskBuilder {
	b.toAgent = agentID
	return b
}

// WithPayload sets the task payload
func (b *TaskBuilder) WithPayload(payload interface{}) *TaskBuilder {
	b.payload = payload
	return b
}

// TaskRequest represents a task request ready to send
type TaskRequest struct {
	FromAgent  string
	ToAgent    string
	Capability string
	Payload    interface{}
}

// Build creates the task request
func (b *TaskBuilder) Build() TaskRequest {
	return TaskRequest{
		FromAgent:  b.fromAgent,
		ToAgent:    b.toAgent,
		Capability: b.capability,
		Payload:    b.payload,
	}
}
