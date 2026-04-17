package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
)

func setupSQLiteStore(t *testing.T) (*SQLiteStore, func()) {
	t.Helper()

	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "beacon-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestNewSQLiteStore(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()

	if store == nil {
		t.Fatal("store should not be nil")
	}
	if store.db == nil {
		t.Fatal("db should not be nil")
	}
}

func TestNewSQLiteStore_InvalidPath(t *testing.T) {
	// Try to create store in a non-existent directory with invalid path
	_, err := NewSQLiteStore("/nonexistent/path/that/cannot/exist/test.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestSQLiteStore_Close(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()

	err := store.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// Agent tests

func TestSQLiteStore_CreateAgent(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	agent := createTestAgent("agent-1")
	err := store.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent() error = %v", err)
	}

	// Verify agent was created
	retrieved, err := store.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, agent.ID)
	}
	if retrieved.Name != agent.Name {
		t.Errorf("Name mismatch: got %s, want %s", retrieved.Name, agent.Name)
	}
}

func TestSQLiteStore_CreateAgent_Duplicate(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	agent := createTestAgent("agent-1")
	err := store.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("first CreateAgent() error = %v", err)
	}

	// Try to create duplicate
	err = store.CreateAgent(ctx, agent)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestSQLiteStore_GetAgent_NotFound(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	_, err := store.GetAgent(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteStore_UpdateAgent(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	agent := createTestAgent("agent-1")
	err := store.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent() error = %v", err)
	}

	// Update agent
	agent.Name = "Updated Agent"
	agent.Status = core.StatusUnhealthy
	err = store.UpdateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("UpdateAgent() error = %v", err)
	}

	// Verify update
	retrieved, err := store.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}

	if retrieved.Name != "Updated Agent" {
		t.Errorf("Name not updated: got %s", retrieved.Name)
	}
	if retrieved.Status != core.StatusUnhealthy {
		t.Errorf("Status not updated: got %s", retrieved.Status)
	}
}

func TestSQLiteStore_UpdateAgent_NotFound(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	agent := createTestAgent("nonexistent")
	err := store.UpdateAgent(ctx, agent)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteStore_DeleteAgent(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	agent := createTestAgent("agent-1")
	err := store.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent() error = %v", err)
	}

	err = store.DeleteAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("DeleteAgent() error = %v", err)
	}

	// Verify deletion
	_, err = store.GetAgent(ctx, agent.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestSQLiteStore_DeleteAgent_NotFound(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	err := store.DeleteAgent(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteStore_ListAgents(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create multiple agents
	for i := 1; i <= 3; i++ {
		agent := createTestAgent("agent-" + string(rune('0'+i)))
		if err := store.CreateAgent(ctx, agent); err != nil {
			t.Fatalf("CreateAgent() error = %v", err)
		}
	}

	// List all agents
	agents, err := store.ListAgents(ctx, nil)
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) != 3 {
		t.Errorf("expected 3 agents, got %d", len(agents))
	}
}

func TestSQLiteStore_ListAgents_WithStatusFilter(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create healthy agent
	agent1 := createTestAgent("agent-1")
	agent1.Status = core.StatusHealthy
	store.CreateAgent(ctx, agent1)

	// Create unhealthy agent
	agent2 := createTestAgent("agent-2")
	agent2.Status = core.StatusUnhealthy
	store.CreateAgent(ctx, agent2)

	// Filter by healthy status
	status := core.StatusHealthy
	agents, err := store.ListAgents(ctx, &AgentFilter{Status: &status})
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("expected 1 healthy agent, got %d", len(agents))
	}
}

func TestSQLiteStore_ListAgents_WithCapabilityFilter(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create agent with echo capability
	agent1 := createTestAgent("agent-1")
	agent1.Capabilities = []core.Capability{{Name: "echo", Description: "Echo"}}
	store.CreateAgent(ctx, agent1)

	// Create agent with process capability
	agent2 := createTestAgent("agent-2")
	agent2.Capabilities = []core.Capability{{Name: "process", Description: "Process"}}
	store.CreateAgent(ctx, agent2)

	// Filter by echo capability
	cap := "echo"
	agents, err := store.ListAgents(ctx, &AgentFilter{Capability: &cap})
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("expected 1 agent with echo capability, got %d", len(agents))
	}
	if agents[0].ID != "agent-1" {
		t.Errorf("expected agent-1, got %s", agents[0].ID)
	}
}

func TestSQLiteStore_FindAgentsByCapability(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create agents
	agent1 := createTestAgent("agent-1")
	agent1.Capabilities = []core.Capability{{Name: "echo", Description: "Echo"}}
	store.CreateAgent(ctx, agent1)

	agent2 := createTestAgent("agent-2")
	agent2.Capabilities = []core.Capability{{Name: "echo", Description: "Echo"}, {Name: "process", Description: "Process"}}
	store.CreateAgent(ctx, agent2)

	agents, err := store.FindAgentsByCapability(ctx, "echo")
	if err != nil {
		t.Fatalf("FindAgentsByCapability() error = %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("expected 2 agents with echo capability, got %d", len(agents))
	}
}

// Task tests

func TestSQLiteStore_CreateTask(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	task := createTestTask("task-1")
	err := store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Verify task was created
	retrieved, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("ID mismatch: got %s, want %s", retrieved.ID, task.ID)
	}
}

func TestSQLiteStore_CreateTask_Duplicate(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	task := createTestTask("task-1")
	err := store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("first CreateTask() error = %v", err)
	}

	err = store.CreateTask(ctx, task)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestSQLiteStore_GetTask_NotFound(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	_, err := store.GetTask(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteStore_UpdateTask(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	task := createTestTask("task-1")
	err := store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Update task
	task.Status = core.TaskCompleted
	task.Result = json.RawMessage(`{"output": "done"}`)
	completedAt := time.Now().UTC()
	task.CompletedAt = &completedAt

	err = store.UpdateTask(ctx, task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	// Verify update
	retrieved, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrieved.Status != core.TaskCompleted {
		t.Errorf("Status not updated: got %s", retrieved.Status)
	}
	if retrieved.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestSQLiteStore_UpdateTask_NotFound(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	task := createTestTask("nonexistent")
	err := store.UpdateTask(ctx, task)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteStore_DeleteTask(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	task := createTestTask("task-1")
	store.CreateTask(ctx, task)

	err := store.DeleteTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	_, err = store.GetTask(ctx, task.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got %v", err)
	}
}

func TestSQLiteStore_DeleteTask_NotFound(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	err := store.DeleteTask(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSQLiteStore_ListTasks(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create multiple tasks
	for i := 1; i <= 5; i++ {
		task := createTestTask("task-" + string(rune('0'+i)))
		store.CreateTask(ctx, task)
	}

	tasks, err := store.ListTasks(ctx, nil)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks, got %d", len(tasks))
	}
}

func TestSQLiteStore_ListTasks_WithFilters(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create tasks with different statuses
	task1 := createTestTask("task-1")
	task1.Status = core.TaskPending
	task1.FromAgent = "sender-1"
	task1.ToAgent = "receiver-1"
	store.CreateTask(ctx, task1)

	task2 := createTestTask("task-2")
	task2.Status = core.TaskCompleted
	task2.FromAgent = "sender-1"
	task2.ToAgent = "receiver-2"
	store.CreateTask(ctx, task2)

	task3 := createTestTask("task-3")
	task3.Status = core.TaskPending
	task3.FromAgent = "sender-2"
	task3.ToAgent = "receiver-1"
	store.CreateTask(ctx, task3)

	// Test status filter
	status := core.TaskPending
	tasks, _ := store.ListTasks(ctx, &TaskFilter{Status: &status})
	if len(tasks) != 2 {
		t.Errorf("expected 2 pending tasks, got %d", len(tasks))
	}

	// Test FromAgent filter
	from := "sender-1"
	tasks, _ = store.ListTasks(ctx, &TaskFilter{FromAgent: &from})
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks from sender-1, got %d", len(tasks))
	}

	// Test ToAgent filter
	to := "receiver-1"
	tasks, _ = store.ListTasks(ctx, &TaskFilter{ToAgent: &to})
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks to receiver-1, got %d", len(tasks))
	}

	// Test combined filters
	tasks, _ = store.ListTasks(ctx, &TaskFilter{
		Status:    &status,
		FromAgent: &from,
	})
	if len(tasks) != 1 {
		t.Errorf("expected 1 task with combined filters, got %d", len(tasks))
	}
}

func TestSQLiteStore_ListTasks_Pagination(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create 10 tasks
	for i := 1; i <= 10; i++ {
		task := createTestTask("task-" + string(rune('0'+i)))
		time.Sleep(1 * time.Millisecond) // Ensure different created_at
		store.CreateTask(ctx, task)
	}

	// Test limit
	limit := 5
	tasks, _ := store.ListTasks(ctx, &TaskFilter{Limit: &limit})
	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks with limit, got %d", len(tasks))
	}

	// Test offset
	offset := 3
	limit = 5
	tasks, _ = store.ListTasks(ctx, &TaskFilter{Offset: &offset, Limit: &limit})
	if len(tasks) != 5 {
		t.Errorf("expected 5 tasks with offset, got %d", len(tasks))
	}
}

func TestSQLiteStore_GetPendingTasksForAgent(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create pending task for agent-1
	task1 := createTestTask("task-1")
	task1.ToAgent = "agent-1"
	task1.Status = core.TaskPending
	store.CreateTask(ctx, task1)

	// Create completed task for agent-1
	task2 := createTestTask("task-2")
	task2.ToAgent = "agent-1"
	task2.Status = core.TaskCompleted
	store.CreateTask(ctx, task2)

	// Create pending task for agent-2
	task3 := createTestTask("task-3")
	task3.ToAgent = "agent-2"
	task3.Status = core.TaskPending
	store.CreateTask(ctx, task3)

	tasks, err := store.GetPendingTasksForAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetPendingTasksForAgent() error = %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 pending task for agent-1, got %d", len(tasks))
	}
}

func TestSQLiteStore_TaskWithNullableFields(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	// Create task with minimal fields
	task := &core.Task{
		ID:         "task-minimal",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Capability: "echo",
		Payload:    json.RawMessage(`{}`),
		Status:     core.TaskPending,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
		// Result, Error, CompletedAt are intentionally nil/empty
	}

	err := store.CreateTask(ctx, task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	retrieved, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrieved.Result != nil && len(retrieved.Result) > 0 {
		t.Error("Result should be nil/empty")
	}
	if retrieved.Error != "" {
		t.Error("Error should be empty")
	}
	if retrieved.CompletedAt != nil {
		t.Error("CompletedAt should be nil")
	}
}

func TestSQLiteStore_AgentWithMetadata(t *testing.T) {
	store, cleanup := setupSQLiteStore(t)
	defer cleanup()
	ctx := context.Background()

	agent := createTestAgent("agent-1")
	agent.Metadata = core.Metadata{
		Author: "Test Author",
		Tags:   []string{"tag1", "tag2", "tag3"},
	}

	err := store.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("CreateAgent() error = %v", err)
	}

	retrieved, err := store.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}

	if retrieved.Metadata.Author != "Test Author" {
		t.Errorf("Author mismatch: got %s", retrieved.Metadata.Author)
	}
	if len(retrieved.Metadata.Tags) != 3 {
		t.Errorf("Tags count mismatch: got %d", len(retrieved.Metadata.Tags))
	}
}

// Helper functions

func createTestAgent(id string) *core.Agent {
	return &core.Agent{
		ID:          id,
		Name:        "Test Agent " + id,
		Version:     "1.0.0",
		Description: "A test agent",
		Endpoint: core.Endpoint{
			Host:     "localhost",
			Port:     50051,
			Protocol: "grpc",
		},
		Capabilities: []core.Capability{
			{Name: "echo", Description: "Echo capability"},
		},
		Metadata: core.Metadata{
			Author: "Test",
			Tags:   []string{"test"},
		},
		Status:        core.StatusHealthy,
		RegisteredAt:  time.Now().UTC(),
		LastHeartbeat: time.Now().UTC(),
	}
}

func createTestTask(id string) *core.Task {
	return &core.Task{
		ID:         id,
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Capability: "echo",
		Payload:    json.RawMessage(`{"message": "hello"}`),
		Status:     core.TaskPending,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
}

// Test helper functions
func TestNullableHelpers(t *testing.T) {
	// Test nullableJSON
	t.Run("nullableJSON with data", func(t *testing.T) {
		data := json.RawMessage(`{"key": "value"}`)
		result := nullableJSON(data)
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("nullableJSON with empty data", func(t *testing.T) {
		var data json.RawMessage
		result := nullableJSON(data)
		if result != nil {
			t.Error("expected nil result")
		}
	})

	// Test nullableString
	t.Run("nullableString with value", func(t *testing.T) {
		result := nullableString("hello")
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("nullableString with empty", func(t *testing.T) {
		result := nullableString("")
		if result != nil {
			t.Error("expected nil result")
		}
	})

	// Test nullableTime
	t.Run("nullableTime with value", func(t *testing.T) {
		now := time.Now()
		result := nullableTime(&now)
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("nullableTime with nil", func(t *testing.T) {
		result := nullableTime(nil)
		if result != nil {
			t.Error("expected nil result")
		}
	})
}
