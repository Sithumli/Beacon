package broker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/registry"
	"github.com/Sithumli/Beacon/internal/store"
)

func setupBrokerService(t *testing.T) (*Service, *registry.Service, *store.MemoryStore) {
	t.Helper()
	memStore := store.NewMemoryStore()
	regService := registry.NewService(memStore)
	brokerService := NewService(memStore, regService)
	return brokerService, regService, memStore
}

func createAndRegisterAgent(t *testing.T, regService *registry.Service, id, capability string) *core.Agent {
	t.Helper()
	ctx := context.Background()

	agent := &core.Agent{
		ID:      id,
		Name:    "Test Agent " + id,
		Version: "1.0.0",
		Endpoint: core.Endpoint{
			Host:     "localhost",
			Port:     50051,
			Protocol: "grpc",
		},
		Capabilities: []core.Capability{
			{Name: capability, Description: "Test capability"},
		},
		Status: core.StatusHealthy,
	}

	_, err := regService.Register(ctx, agent)
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	return agent
}

func TestNewService(t *testing.T) {
	service, _, _ := setupBrokerService(t)

	if service == nil {
		t.Fatal("service should not be nil")
	}
	if service.subscribers == nil {
		t.Fatal("subscribers map should be initialized")
	}
}

func TestService_SendTask(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	// Register target agent
	createAndRegisterAgent(t, regService, "receiver-1", "echo")

	// Send task
	payload := json.RawMessage(`{"message": "hello"}`)
	task, err := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)
	if err != nil {
		t.Fatalf("SendTask() error = %v", err)
	}

	if task == nil {
		t.Fatal("task should not be nil")
	}
	if task.ID == "" {
		t.Error("task ID should not be empty")
	}
	if task.Status != core.TaskPending {
		t.Errorf("task status should be pending, got %s", task.Status)
	}
}

func TestService_SendTask_AgentNotFound(t *testing.T) {
	service, _, _ := setupBrokerService(t)
	ctx := context.Background()

	payload := json.RawMessage(`{"message": "hello"}`)
	_, err := service.SendTask(ctx, "sender-1", "nonexistent", "echo", payload)
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestService_SendTask_MissingCapability(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	// Register agent with different capability
	createAndRegisterAgent(t, regService, "receiver-1", "process")

	payload := json.RawMessage(`{"message": "hello"}`)
	_, err := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)
	if err == nil {
		t.Error("expected error for missing capability")
	}
}

func TestService_SendTask_UnhealthyAgent(t *testing.T) {
	service, regService, memStore := setupBrokerService(t)
	ctx := context.Background()

	// Register agent and mark unhealthy
	agent := createAndRegisterAgent(t, regService, "receiver-1", "echo")
	agent.Status = core.StatusUnhealthy
	memStore.UpdateAgent(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	_, err := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)
	if err == nil {
		t.Error("expected error for unhealthy agent")
	}
}

func TestService_RouteTask(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	// Register multiple agents with same capability
	createAndRegisterAgent(t, regService, "agent-1", "echo")
	createAndRegisterAgent(t, regService, "agent-2", "echo")

	payload := json.RawMessage(`{"message": "hello"}`)
	task, err := service.RouteTask(ctx, "sender", "echo", payload)
	if err != nil {
		t.Fatalf("RouteTask() error = %v", err)
	}

	if task == nil {
		t.Fatal("task should not be nil")
	}
	// Task should be routed to one of the agents
	if task.ToAgent != "agent-1" && task.ToAgent != "agent-2" {
		t.Errorf("task should be routed to agent-1 or agent-2, got %s", task.ToAgent)
	}
}

func TestService_RouteTask_NoAgents(t *testing.T) {
	service, _, _ := setupBrokerService(t)
	ctx := context.Background()

	payload := json.RawMessage(`{"message": "hello"}`)
	_, err := service.RouteTask(ctx, "sender", "echo", payload)
	if err == nil {
		t.Error("expected error when no agents available")
	}
}

func TestService_RouteTask_NoHealthyAgents(t *testing.T) {
	service, regService, memStore := setupBrokerService(t)
	ctx := context.Background()

	// Register agent and mark unhealthy
	agent := createAndRegisterAgent(t, regService, "agent-1", "echo")
	agent.Status = core.StatusUnhealthy
	memStore.UpdateAgent(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	_, err := service.RouteTask(ctx, "sender", "echo", payload)
	if err == nil {
		t.Error("expected error when no healthy agents")
	}
}

func TestService_GetTask(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")

	payload := json.RawMessage(`{"message": "hello"}`)
	created, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	task, err := service.GetTask(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if task.ID != created.ID {
		t.Errorf("ID mismatch: got %s, want %s", task.ID, created.ID)
	}
}

func TestService_GetTask_NotFound(t *testing.T) {
	service, _, _ := setupBrokerService(t)
	ctx := context.Background()

	_, err := service.GetTask(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestService_UpdateTask_Start(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	updated, err := service.UpdateTask(ctx, task.ID, core.TaskRunning, nil, "")
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updated.Status != core.TaskRunning {
		t.Errorf("status should be running, got %s", updated.Status)
	}
}

func TestService_UpdateTask_Complete(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// First start the task
	service.UpdateTask(ctx, task.ID, core.TaskRunning, nil, "")

	// Then complete it
	result := json.RawMessage(`{"output": "world"}`)
	updated, err := service.UpdateTask(ctx, task.ID, core.TaskCompleted, result, "")
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updated.Status != core.TaskCompleted {
		t.Errorf("status should be completed, got %s", updated.Status)
	}
	if updated.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestService_UpdateTask_Fail(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// Start then fail
	service.UpdateTask(ctx, task.ID, core.TaskRunning, nil, "")
	updated, err := service.UpdateTask(ctx, task.ID, core.TaskFailed, nil, "something went wrong")
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updated.Status != core.TaskFailed {
		t.Errorf("status should be failed, got %s", updated.Status)
	}
	if updated.Error != "something went wrong" {
		t.Errorf("error message mismatch: got %s", updated.Error)
	}
}

func TestService_UpdateTask_Cancel(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	updated, err := service.UpdateTask(ctx, task.ID, core.TaskCancelled, nil, "")
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if updated.Status != core.TaskCancelled {
		t.Errorf("status should be cancelled, got %s", updated.Status)
	}
}

func TestService_UpdateTask_InvalidTransition(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// Complete the task
	service.UpdateTask(ctx, task.ID, core.TaskRunning, nil, "")
	service.UpdateTask(ctx, task.ID, core.TaskCompleted, nil, "")

	// Try to update again (invalid transition from completed)
	_, err := service.UpdateTask(ctx, task.ID, core.TaskRunning, nil, "")
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestService_UpdateTask_InvalidStatus(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	_, err := service.UpdateTask(ctx, task.ID, "invalid-status", nil, "")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestService_ListTasks(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)

	// Create multiple tasks
	for i := 0; i < 3; i++ {
		service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)
	}

	tasks, err := service.ListTasks(ctx, nil)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestService_CancelTask(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	cancelled, err := service.CancelTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CancelTask() error = %v", err)
	}

	if cancelled.Status != core.TaskCancelled {
		t.Errorf("status should be cancelled, got %s", cancelled.Status)
	}
}

func TestService_Subscribe(t *testing.T) {
	service, _, _ := setupBrokerService(t)

	id, ch := service.Subscribe("agent-1")

	if id == "" {
		t.Error("subscription ID should not be empty")
	}
	if ch == nil {
		t.Error("channel should not be nil")
	}
	if !contains(id, "agent-1") {
		t.Error("subscription ID should contain agent ID")
	}
}

func TestService_Unsubscribe(t *testing.T) {
	service, _, _ := setupBrokerService(t)

	id, ch := service.Subscribe("agent-1")
	service.Unsubscribe(id)

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after unsubscribe")
		}
	default:
		// Channel might be empty but not closed yet, that's fine
	}
}

func TestService_SubscriberNotification(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")

	// Subscribe for receiver-1
	_, ch := service.Subscribe("receiver-1")

	// Send task to receiver-1
	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// Should receive notification
	select {
	case event := <-ch:
		if event.Type != EventNewTask {
			t.Errorf("expected EventNewTask, got %d", event.Type)
		}
		if event.Task.ID != task.ID {
			t.Errorf("task ID mismatch")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected to receive notification")
	}
}

func TestService_SubscriberNotification_DifferentAgent(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	createAndRegisterAgent(t, regService, "receiver-2", "echo")

	// Subscribe for receiver-2
	_, ch := service.Subscribe("receiver-2")

	// Send task to receiver-1 (different agent)
	payload := json.RawMessage(`{"message": "hello"}`)
	service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// Should NOT receive notification
	select {
	case <-ch:
		t.Error("should not receive notification for different agent")
	case <-time.After(50 * time.Millisecond):
		// Expected - no notification
	}
}

func TestService_GetPendingTasks(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")
	payload := json.RawMessage(`{"message": "hello"}`)

	// Create pending tasks
	service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)
	task2, _ := service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// Complete one task
	service.UpdateTask(ctx, task2.ID, core.TaskRunning, nil, "")
	service.UpdateTask(ctx, task2.ID, core.TaskCompleted, nil, "")

	pending, err := service.GetPendingTasks(ctx, "receiver-1")
	if err != nil {
		t.Fatalf("GetPendingTasks() error = %v", err)
	}

	if len(pending) != 1 {
		t.Errorf("expected 1 pending task, got %d", len(pending))
	}
}

func TestService_MultipleSubscribers(t *testing.T) {
	service, regService, _ := setupBrokerService(t)
	ctx := context.Background()

	createAndRegisterAgent(t, regService, "receiver-1", "echo")

	// Create multiple subscriptions for same agent
	_, ch1 := service.Subscribe("receiver-1")
	_, ch2 := service.Subscribe("receiver-1")

	payload := json.RawMessage(`{"message": "hello"}`)
	service.SendTask(ctx, "sender-1", "receiver-1", "echo", payload)

	// Both should receive notification
	received1 := false
	received2 := false

	for i := 0; i < 2; i++ {
		select {
		case <-ch1:
			received1 = true
		case <-ch2:
			received2 = true
		case <-time.After(100 * time.Millisecond):
		}
	}

	if !received1 || !received2 {
		t.Error("both subscribers should receive notification")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
