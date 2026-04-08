package core

import (
	"encoding/json"
	"testing"
)

func TestNewTask(t *testing.T) {
	payload := json.RawMessage(`{"message": "hello"}`)
	task := NewTask("agent-a", "agent-b", "echo", payload)

	if task.FromAgent != "agent-a" {
		t.Errorf("expected FromAgent 'agent-a', got '%s'", task.FromAgent)
	}
	if task.ToAgent != "agent-b" {
		t.Errorf("expected ToAgent 'agent-b', got '%s'", task.ToAgent)
	}
	if task.Capability != "echo" {
		t.Errorf("expected Capability 'echo', got '%s'", task.Capability)
	}
	if task.Status != TaskPending {
		t.Errorf("expected Status 'pending', got '%s'", task.Status)
	}
	if task.ID == "" {
		t.Error("expected ID to be generated")
	}
}

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: &Task{
				FromAgent:  "agent-a",
				ToAgent:    "agent-b",
				Capability: "echo",
				Payload:    json.RawMessage(`{}`),
			},
			wantErr: false,
		},
		{
			name: "missing from_agent",
			task: &Task{
				ToAgent:    "agent-b",
				Capability: "echo",
				Payload:    json.RawMessage(`{}`),
			},
			wantErr: true,
		},
		{
			name: "missing to_agent",
			task: &Task{
				FromAgent:  "agent-a",
				Capability: "echo",
				Payload:    json.RawMessage(`{}`),
			},
			wantErr: true,
		},
		{
			name: "missing capability",
			task: &Task{
				FromAgent: "agent-a",
				ToAgent:   "agent-b",
				Payload:   json.RawMessage(`{}`),
			},
			wantErr: true,
		},
		{
			name: "missing payload",
			task: &Task{
				FromAgent:  "agent-a",
				ToAgent:    "agent-b",
				Capability: "echo",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskStatusTransitions(t *testing.T) {
	task := NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))

	// pending -> running
	if err := task.Start(); err != nil {
		t.Errorf("failed to start task: %v", err)
	}
	if task.Status != TaskRunning {
		t.Errorf("expected status 'running', got '%s'", task.Status)
	}

	// running -> completed
	if err := task.Complete(json.RawMessage(`{"result": "done"}`)); err != nil {
		t.Errorf("failed to complete task: %v", err)
	}
	if task.Status != TaskCompleted {
		t.Errorf("expected status 'completed', got '%s'", task.Status)
	}
	if task.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestTaskFailure(t *testing.T) {
	task := NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))

	task.Start()
	if err := task.Fail("something went wrong"); err != nil {
		t.Errorf("failed to fail task: %v", err)
	}

	if task.Status != TaskFailed {
		t.Errorf("expected status 'failed', got '%s'", task.Status)
	}
	if task.Error != "something went wrong" {
		t.Errorf("expected error message, got '%s'", task.Error)
	}
}

func TestTaskCancel(t *testing.T) {
	task := NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))

	if err := task.Cancel(); err != nil {
		t.Errorf("failed to cancel task: %v", err)
	}

	if task.Status != TaskCancelled {
		t.Errorf("expected status 'cancelled', got '%s'", task.Status)
	}
}

func TestTaskInvalidTransition(t *testing.T) {
	task := NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))

	// Can't complete a pending task
	err := task.Complete(nil)
	if err == nil {
		t.Error("expected error for invalid transition")
	}

	// Complete properly
	task.Start()
	task.Complete(nil)

	// Can't start a completed task
	err = task.Start()
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestTaskIsFinal(t *testing.T) {
	task := NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))

	if task.IsFinal() {
		t.Error("pending task should not be final")
	}

	task.Start()
	if task.IsFinal() {
		t.Error("running task should not be final")
	}

	task.Complete(nil)
	if !task.IsFinal() {
		t.Error("completed task should be final")
	}
}
