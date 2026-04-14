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
