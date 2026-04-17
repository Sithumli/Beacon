package sdk

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskInfo_IsFinal(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"pending is not final", TaskPending, false},
		{"running is not final", TaskRunning, false},
		{"completed is final", TaskCompleted, true},
		{"failed is final", TaskFailed, true},
		{"cancelled is final", TaskCancelled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &TaskInfo{Status: tt.status}
			if got := task.IsFinal(); got != tt.want {
				t.Errorf("IsFinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskInfo_IsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
		want   bool
	}{
		{"pending is not success", TaskPending, false},
		{"running is not success", TaskRunning, false},
		{"completed is success", TaskCompleted, true},
		{"failed is not success", TaskFailed, false},
		{"cancelled is not success", TaskCancelled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &TaskInfo{Status: tt.status}
			if got := task.IsSuccess(); got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskInfo_Duration_WithCompletedAt(t *testing.T) {
	createdAt := time.Now().Add(-10 * time.Second)
	completedAt := time.Now()

	task := &TaskInfo{
		CreatedAt:   createdAt,
		CompletedAt: &completedAt,
	}

	duration := task.Duration()

	// Should be approximately 10 seconds
	if duration < 9*time.Second || duration > 11*time.Second {
		t.Errorf("Duration() = %v, expected ~10s", duration)
	}
}

func TestTaskInfo_Duration_WithoutCompletedAt(t *testing.T) {
	createdAt := time.Now().Add(-5 * time.Second)

	task := &TaskInfo{
		CreatedAt:   createdAt,
		CompletedAt: nil,
	}

	duration := task.Duration()

	// Should be approximately 5 seconds (since creation)
	if duration < 4*time.Second || duration > 6*time.Second {
		t.Errorf("Duration() = %v, expected ~5s", duration)
	}
}

func TestTaskInfo_GetResult(t *testing.T) {
	result := map[string]string{"output": "hello world"}
	resultJSON, _ := json.Marshal(result)

	task := &TaskInfo{
		Result: resultJSON,
	}

	var got map[string]string
	err := task.GetResult(&got)
	if err != nil {
		t.Fatalf("GetResult() error = %v", err)
	}

	if got["output"] != "hello world" {
		t.Errorf("GetResult() = %v, want %v", got, result)
	}
}

func TestTaskInfo_GetResult_InvalidJSON(t *testing.T) {
	task := &TaskInfo{
		Result: json.RawMessage("invalid json"),
	}

	var got map[string]string
	err := task.GetResult(&got)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestTaskInfo_GetPayload(t *testing.T) {
	payload := map[string]int{"count": 42}
	payloadJSON, _ := json.Marshal(payload)

	task := &TaskInfo{
		Payload: payloadJSON,
	}

	var got map[string]int
	err := task.GetPayload(&got)
	if err != nil {
		t.Fatalf("GetPayload() error = %v", err)
	}

	if got["count"] != 42 {
		t.Errorf("GetPayload() = %v, want %v", got, payload)
	}
}

func TestTaskInfo_GetPayload_InvalidJSON(t *testing.T) {
	task := &TaskInfo{
		Payload: json.RawMessage("invalid json"),
	}

	var got map[string]string
	err := task.GetPayload(&got)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNewTask(t *testing.T) {
	builder := NewTask("echo")

	if builder == nil {
		t.Fatal("builder should not be nil")
	}
	if builder.capability != "echo" {
		t.Errorf("capability = %s, want echo", builder.capability)
	}
}

func TestTaskBuilder_From(t *testing.T) {
	builder := NewTask("echo").From("sender-1")

	if builder.fromAgent != "sender-1" {
		t.Errorf("fromAgent = %s, want sender-1", builder.fromAgent)
	}
}

func TestTaskBuilder_To(t *testing.T) {
	builder := NewTask("echo").To("receiver-1")

	if builder.toAgent != "receiver-1" {
		t.Errorf("toAgent = %s, want receiver-1", builder.toAgent)
	}
}

func TestTaskBuilder_WithPayload(t *testing.T) {
	payload := map[string]string{"message": "hello"}
	builder := NewTask("echo").WithPayload(payload)

	if builder.payload == nil {
		t.Fatal("payload should not be nil")
	}
}

func TestTaskBuilder_Build(t *testing.T) {
	payload := map[string]string{"message": "hello"}

	request := NewTask("echo").
		From("sender-1").
		To("receiver-1").
		WithPayload(payload).
		Build()

	if request.FromAgent != "sender-1" {
		t.Errorf("FromAgent = %s, want sender-1", request.FromAgent)
	}
	if request.ToAgent != "receiver-1" {
		t.Errorf("ToAgent = %s, want receiver-1", request.ToAgent)
	}
	if request.Capability != "echo" {
		t.Errorf("Capability = %s, want echo", request.Capability)
	}
	if request.Payload == nil {
		t.Error("Payload should not be nil")
	}
}

func TestTaskBuilder_Chaining(t *testing.T) {
	// Test fluent interface
	request := NewTask("process").
		From("agent-a").
		To("agent-b").
		WithPayload(map[string]int{"value": 100}).
		Build()

	if request.FromAgent != "agent-a" {
		t.Error("chained From() failed")
	}
	if request.ToAgent != "agent-b" {
		t.Error("chained To() failed")
	}
	if request.Capability != "process" {
		t.Error("capability mismatch")
	}
}

func TestTaskStatus_Constants(t *testing.T) {
	if TaskPending != "pending" {
		t.Errorf("TaskPending = %s, want pending", TaskPending)
	}
	if TaskRunning != "running" {
		t.Errorf("TaskRunning = %s, want running", TaskRunning)
	}
	if TaskCompleted != "completed" {
		t.Errorf("TaskCompleted = %s, want completed", TaskCompleted)
	}
	if TaskFailed != "failed" {
		t.Errorf("TaskFailed = %s, want failed", TaskFailed)
	}
	if TaskCancelled != "cancelled" {
		t.Errorf("TaskCancelled = %s, want cancelled", TaskCancelled)
	}
}

func TestTaskRequest_Fields(t *testing.T) {
	request := TaskRequest{
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Capability: "echo",
		Payload:    map[string]string{"key": "value"},
	}

	if request.FromAgent != "sender" {
		t.Error("FromAgent mismatch")
	}
	if request.ToAgent != "receiver" {
		t.Error("ToAgent mismatch")
	}
	if request.Capability != "echo" {
		t.Error("Capability mismatch")
	}
	if request.Payload == nil {
		t.Error("Payload should not be nil")
	}
}

func TestTaskInfo_AllFields(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(5 * time.Second)

	task := TaskInfo{
		ID:          "task-123",
		FromAgent:   "sender",
		ToAgent:     "receiver",
		Capability:  "echo",
		Payload:     json.RawMessage(`{"input": "test"}`),
		Status:      TaskCompleted,
		Result:      json.RawMessage(`{"output": "result"}`),
		Error:       "",
		CreatedAt:   now,
		UpdatedAt:   now.Add(1 * time.Second),
		CompletedAt: &completedAt,
	}

	if task.ID != "task-123" {
		t.Error("ID mismatch")
	}
	if task.FromAgent != "sender" {
		t.Error("FromAgent mismatch")
	}
	if task.ToAgent != "receiver" {
		t.Error("ToAgent mismatch")
	}
	if task.Capability != "echo" {
		t.Error("Capability mismatch")
	}
	if task.Status != TaskCompleted {
		t.Error("Status mismatch")
	}
	if task.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}
