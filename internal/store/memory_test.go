package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Sithumli/Beacon/internal/core"
)

func TestMemoryStoreAgent(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create agent
	agent := core.NewAgent("TestAgent", "1.0.0", "A test agent")
	agent.Endpoint = core.Endpoint{
		Host:     "localhost",
		Port:     50051,
		Protocol: "grpc",
	}

	err := store.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	// Get agent
	retrieved, err := store.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("failed to get agent: %v", err)
	}
	if retrieved.Name != agent.Name {
		t.Errorf("expected name '%s', got '%s'", agent.Name, retrieved.Name)
	}

	// List agents
	agents, err := store.ListAgents(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list agents: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}

	// Update agent
	agent.Version = "2.0.0"
	err = store.UpdateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("failed to update agent: %v", err)
	}

	retrieved, _ = store.GetAgent(ctx, agent.ID)
	if retrieved.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got '%s'", retrieved.Version)
	}

	// Delete agent
	err = store.DeleteAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("failed to delete agent: %v", err)
	}

	_, err = store.GetAgent(ctx, agent.ID)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound after deletion")
	}
}

func TestMemoryStoreAgentDuplicate(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	agent := core.NewAgent("TestAgent", "1.0.0", "A test agent")
	agent.Endpoint = core.Endpoint{Host: "localhost", Port: 50051}

	store.CreateAgent(ctx, agent)

	// Try to create again
	err := store.CreateAgent(ctx, agent)
	if err != ErrAlreadyExists {
		t.Error("expected ErrAlreadyExists")
	}
}

func TestMemoryStoreFindByCapability(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create agent with capability
	agent1 := core.NewAgent("Agent1", "1.0.0", "Agent with echo")
	agent1.Endpoint = core.Endpoint{Host: "localhost", Port: 50051}
	agent1.Capabilities = []core.Capability{
		{Name: "echo", Description: "Echo capability"},
	}
	store.CreateAgent(ctx, agent1)

	// Create agent without capability
	agent2 := core.NewAgent("Agent2", "1.0.0", "Agent without echo")
	agent2.Endpoint = core.Endpoint{Host: "localhost", Port: 50052}
	store.CreateAgent(ctx, agent2)

	// Find by capability
	agents, err := store.FindAgentsByCapability(ctx, "echo")
	if err != nil {
		t.Fatalf("failed to find agents: %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
	if agents[0].Name != "Agent1" {
		t.Errorf("expected Agent1, got %s", agents[0].Name)
	}
}

func TestMemoryStoreTask(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create task
	task := core.NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{"msg":"hello"}`))

	err := store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Get task
	retrieved, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if retrieved.Capability != task.Capability {
		t.Errorf("expected capability '%s', got '%s'", task.Capability, retrieved.Capability)
	}

	// List tasks
	tasks, err := store.ListTasks(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	// Update task
	task.Status = core.TaskRunning
	err = store.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}

	retrieved, _ = store.GetTask(ctx, task.ID)
	if retrieved.Status != core.TaskRunning {
		t.Errorf("expected status 'running', got '%s'", retrieved.Status)
	}

	// Delete task
	err = store.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	_, err = store.GetTask(ctx, task.ID)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound after deletion")
	}
}

func TestMemoryStoreGetPendingTasks(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create pending task
	task1 := core.NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))
	store.CreateTask(ctx, task1)

	// Create running task
	task2 := core.NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))
	task2.Status = core.TaskRunning
	store.CreateTask(ctx, task2)

	// Get pending tasks for agent-b
	pending, err := store.GetPendingTasksForAgent(ctx, "agent-b")
	if err != nil {
		t.Fatalf("failed to get pending tasks: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 pending task, got %d", len(pending))
	}
}

func TestMemoryStoreTaskFilter(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create multiple tasks
	task1 := core.NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))
	store.CreateTask(ctx, task1)

	task2 := core.NewTask("agent-a", "agent-c", "compute", json.RawMessage(`{}`))
	store.CreateTask(ctx, task2)

	// Filter by capability
	capability := "echo"
	filter := &TaskFilter{Capability: &capability}
	tasks, _ := store.ListTasks(ctx, filter)
	if len(tasks) != 1 {
		t.Errorf("expected 1 task with echo capability, got %d", len(tasks))
	}

	// Filter by to_agent
	toAgent := "agent-c"
	filter = &TaskFilter{ToAgent: &toAgent}
	tasks, _ = store.ListTasks(ctx, filter)
	if len(tasks) != 1 {
		t.Errorf("expected 1 task to agent-c, got %d", len(tasks))
	}
}

func TestMemoryStore_Close(t *testing.T) {
	store := NewMemoryStore()
	err := store.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestMemoryStore_CreateAgent_Nil(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.CreateAgent(ctx, nil)
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestMemoryStore_UpdateAgent_Nil(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.UpdateAgent(ctx, nil)
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestMemoryStore_UpdateAgent_NotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	agent := core.NewAgent("Test", "1.0.0", "desc")
	err := store.UpdateAgent(ctx, agent)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_DeleteAgent_NotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.DeleteAgent(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_GetAgent_NotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	_, err := store.GetAgent(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_CreateTask_Nil(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.CreateTask(ctx, nil)
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestMemoryStore_CreateTask_Duplicate(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	task := core.NewTask("a", "b", "echo", json.RawMessage(`{}`))
	store.CreateTask(ctx, task)

	err := store.CreateTask(ctx, task)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestMemoryStore_UpdateTask_Nil(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.UpdateTask(ctx, nil)
	if err != ErrInvalidInput {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestMemoryStore_UpdateTask_NotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	task := core.NewTask("a", "b", "echo", json.RawMessage(`{}`))
	err := store.UpdateTask(ctx, task)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_DeleteTask_NotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	err := store.DeleteTask(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_GetTask_NotFound(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	_, err := store.GetTask(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_ListAgents_WithStatusFilter(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	agent1 := core.NewAgent("Agent1", "1.0.0", "desc")
	agent1.Endpoint = core.Endpoint{Host: "localhost", Port: 50051}
	agent1.Status = core.StatusHealthy
	store.CreateAgent(ctx, agent1)

	agent2 := core.NewAgent("Agent2", "1.0.0", "desc")
	agent2.Endpoint = core.Endpoint{Host: "localhost", Port: 50052}
	agent2.Status = core.StatusUnhealthy
	store.CreateAgent(ctx, agent2)

	status := core.StatusHealthy
	agents, _ := store.ListAgents(ctx, &AgentFilter{Status: &status})
	if len(agents) != 1 {
		t.Errorf("expected 1 healthy agent, got %d", len(agents))
	}
}

func TestMemoryStore_ListAgents_WithTagsFilter(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	agent1 := core.NewAgent("Agent1", "1.0.0", "desc")
	agent1.Endpoint = core.Endpoint{Host: "localhost", Port: 50051}
	agent1.Metadata.Tags = []string{"production", "api"}
	store.CreateAgent(ctx, agent1)

	agent2 := core.NewAgent("Agent2", "1.0.0", "desc")
	agent2.Endpoint = core.Endpoint{Host: "localhost", Port: 50052}
	agent2.Metadata.Tags = []string{"staging"}
	store.CreateAgent(ctx, agent2)

	agents, _ := store.ListAgents(ctx, &AgentFilter{Tags: []string{"production"}})
	if len(agents) != 1 {
		t.Errorf("expected 1 agent with production tag, got %d", len(agents))
	}
}

func TestMemoryStore_ListTasks_WithStatusFilter(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	task1 := core.NewTask("a", "b", "echo", json.RawMessage(`{}`))
	task1.Status = core.TaskPending
	store.CreateTask(ctx, task1)

	task2 := core.NewTask("a", "b", "echo", json.RawMessage(`{}`))
	task2.Status = core.TaskCompleted
	store.CreateTask(ctx, task2)

	status := core.TaskPending
	tasks, _ := store.ListTasks(ctx, &TaskFilter{Status: &status})
	if len(tasks) != 1 {
		t.Errorf("expected 1 pending task, got %d", len(tasks))
	}
}

func TestMemoryStore_ListTasks_WithFromAgentFilter(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	task1 := core.NewTask("agent-a", "agent-b", "echo", json.RawMessage(`{}`))
	store.CreateTask(ctx, task1)

	task2 := core.NewTask("agent-x", "agent-b", "echo", json.RawMessage(`{}`))
	store.CreateTask(ctx, task2)

	from := "agent-a"
	tasks, _ := store.ListTasks(ctx, &TaskFilter{FromAgent: &from})
	if len(tasks) != 1 {
		t.Errorf("expected 1 task from agent-a, got %d", len(tasks))
	}
}

func TestMemoryStore_ListTasks_Pagination(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create 10 tasks
	for i := 0; i < 10; i++ {
		task := core.NewTask("a", "b", "echo", json.RawMessage(`{}`))
		store.CreateTask(ctx, task)
	}

	// Test limit
	limit := 5
	tasks, _ := store.ListTasks(ctx, &TaskFilter{Limit: &limit})
	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks with limit, got %d", len(tasks))
	}

	// Test offset
	offset := 7
	tasks, _ = store.ListTasks(ctx, &TaskFilter{Offset: &offset})
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks with offset 7, got %d", len(tasks))
	}

	// Test offset beyond length
	offset = 20
	tasks, _ = store.ListTasks(ctx, &TaskFilter{Offset: &offset})
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks with offset beyond length, got %d", len(tasks))
	}
}

func TestMemoryStore_DeepCopy_Agent(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	agent := core.NewAgent("Agent", "1.0.0", "desc")
	agent.Endpoint = core.Endpoint{Host: "localhost", Port: 50051}
	agent.Capabilities = []core.Capability{{Name: "echo", Description: "Echo"}}
	agent.Metadata.Tags = []string{"test"}
	store.CreateAgent(ctx, agent)

	// Get and modify
	retrieved, _ := store.GetAgent(ctx, agent.ID)
	retrieved.Capabilities[0].Name = "modified"
	retrieved.Metadata.Tags[0] = "modified"

	// Original should be unchanged
	original, _ := store.GetAgent(ctx, agent.ID)
	if original.Capabilities[0].Name == "modified" {
		t.Error("original capability was modified - deep copy failed")
	}
	if original.Metadata.Tags[0] == "modified" {
		t.Error("original tags were modified - deep copy failed")
	}
}

func TestMemoryStore_DeepCopy_Task(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	task := core.NewTask("a", "b", "echo", json.RawMessage(`{"key":"value"}`))
	task.Result = json.RawMessage(`{"result":"data"}`)
	store.CreateTask(ctx, task)

	// Get and modify
	retrieved, _ := store.GetTask(ctx, task.ID)
	retrieved.Payload[0] = 'X'

	// Original should be unchanged
	original, _ := store.GetTask(ctx, task.ID)
	if original.Payload[0] == 'X' {
		t.Error("original payload was modified - deep copy failed")
	}
}
