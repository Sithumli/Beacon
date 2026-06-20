package registry

import (
	"context"
	"testing"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
)

func setupRegistryService(t *testing.T) (*Service, *store.MemoryStore) {
	t.Helper()
	memStore := store.NewMemoryStore()
	service := NewService(memStore)
	return service, memStore
}

func createTestAgentData(id string) *core.Agent {
	return &core.Agent{
		ID:      id,
		Name:    "Test Agent " + id,
		Version: "1.0.0",
		Endpoint: core.Endpoint{
			Host:     "localhost",
			Port:     50051,
			Protocol: "grpc",
		},
		Capabilities: []core.Capability{
			{Name: "echo", Description: "Echo capability"},
		},
		Metadata: core.Metadata{
			Author: "Test Author",
			Tags:   []string{"test"},
		},
		Status: core.StatusHealthy,
	}
}

func TestNewService(t *testing.T) {
	service, _ := setupRegistryService(t)

	if service == nil {
		t.Fatal("service should not be nil")
	}
	if service.watchers == nil {
		t.Fatal("watchers map should be initialized")
	}
}

func TestService_Register(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	registered, err := service.Register(ctx, agent)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if registered == nil {
		t.Fatal("registered agent should not be nil")
	}
	if registered.ID != "agent-1" {
		t.Errorf("agent ID mismatch: got %s", registered.ID)
	}
}

func TestService_Register_GeneratesID(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := &core.Agent{
		Name:    "Test Agent",
		Version: "1.0.0",
		Endpoint: core.Endpoint{
			Host:     "localhost",
			Port:     50051,
			Protocol: "grpc",
		},
	}

	registered, err := service.Register(ctx, agent)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if registered.ID == "" {
		t.Error("should generate ID when not provided")
	}
}

func TestService_Register_ValidationError(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	// Missing required fields
	agent := &core.Agent{
		Name: "Test Agent",
		// Missing version and endpoint
	}

	_, err := service.Register(ctx, agent)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestService_Register_Duplicate(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	_, err := service.Register(ctx, agent)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestService_Deregister(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	err := service.Deregister(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Deregister() error = %v", err)
	}

	// Verify agent is gone
	_, err = service.GetAgent(ctx, "agent-1")
	if err == nil {
		t.Error("agent should not exist after deregistration")
	}
}

func TestService_Deregister_NotFound(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	err := service.Deregister(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestService_GetAgent(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	retrieved, err := service.GetAgent(ctx, "agent-1")
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}

	if retrieved.ID != "agent-1" {
		t.Errorf("ID mismatch: got %s", retrieved.ID)
	}
}

func TestService_GetAgent_NotFound(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	_, err := service.GetAgent(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestService_ListAgents(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	// Register multiple agents
	for i := 1; i <= 3; i++ {
		agent := createTestAgentData("agent-" + string(rune('0'+i)))
		service.Register(ctx, agent)
	}

	agents, err := service.ListAgents(ctx, nil)
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) != 3 {
		t.Errorf("expected 3 agents, got %d", len(agents))
	}
}

func TestService_ListAgents_WithFilter(t *testing.T) {
	service, memStore := setupRegistryService(t)
	ctx := context.Background()

	// Create healthy agent
	agent1 := createTestAgentData("agent-1")
	service.Register(ctx, agent1)

	// Create unhealthy agent
	agent2 := createTestAgentData("agent-2")
	service.Register(ctx, agent2)
	agent2.Status = core.StatusUnhealthy
	memStore.UpdateAgent(ctx, agent2)

	// Filter by status
	status := core.StatusHealthy
	agents, err := service.ListAgents(ctx, &store.AgentFilter{Status: &status})
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("expected 1 healthy agent, got %d", len(agents))
	}
}

func TestService_Discover(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	// Register agent with echo capability
	agent1 := createTestAgentData("agent-1")
	agent1.Capabilities = []core.Capability{{Name: "echo", Description: "Echo"}}
	service.Register(ctx, agent1)

	// Register agent with process capability
	agent2 := createTestAgentData("agent-2")
	agent2.Capabilities = []core.Capability{{Name: "process", Description: "Process"}}
	service.Register(ctx, agent2)

	// Discover echo capability
	agents, err := service.Discover(ctx, "echo")
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("expected 1 agent with echo, got %d", len(agents))
	}
	if agents[0].ID != "agent-1" {
		t.Errorf("expected agent-1, got %s", agents[0].ID)
	}
}

func TestService_Discover_OnlyHealthy(t *testing.T) {
	service, memStore := setupRegistryService(t)
	ctx := context.Background()

	// Register healthy agent
	agent1 := createTestAgentData("agent-1")
	service.Register(ctx, agent1)

	// Register unhealthy agent
	agent2 := createTestAgentData("agent-2")
	service.Register(ctx, agent2)
	agent2.Status = core.StatusUnhealthy
	memStore.UpdateAgent(ctx, agent2)

	agents, err := service.Discover(ctx, "echo")
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}

	if len(agents) != 1 {
		t.Errorf("expected 1 healthy agent, got %d", len(agents))
	}
}

func TestService_Heartbeat(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	err := service.Heartbeat(ctx, "agent-1")
	if err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}

	// Verify heartbeat was updated
	retrieved, _ := service.GetAgent(ctx, "agent-1")
	if !retrieved.LastHeartbeat.After(agent.LastHeartbeat) {
		t.Error("heartbeat should be updated")
	}
}

func TestService_Heartbeat_NotFound(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	err := service.Heartbeat(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestService_Heartbeat_RecoveryNotification(t *testing.T) {
	service, memStore := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	// Mark agent unhealthy
	agent.Status = core.StatusUnhealthy
	memStore.UpdateAgent(ctx, agent)

	// Watch for events
	_, ch := service.Watch(nil)

	// Send heartbeat (should recover to healthy)
	service.Heartbeat(ctx, "agent-1")

	// Should receive health changed event
	select {
	case event := <-ch:
		if event.Type != EventHealthChanged {
			t.Errorf("expected EventHealthChanged, got %d", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected health changed event")
	}
}

func TestService_Watch(t *testing.T) {
	service, _ := setupRegistryService(t)

	id, ch := service.Watch(nil)

	if id == "" {
		t.Error("watch ID should not be empty")
	}
	if ch == nil {
		t.Error("channel should not be nil")
	}
}

func TestService_Watch_WithCapabilities(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	// Watch for specific capability
	_, ch := service.Watch([]string{"process"})

	// Register agent with different capability
	agent1 := createTestAgentData("agent-1")
	agent1.Capabilities = []core.Capability{{Name: "echo", Description: "Echo"}}
	service.Register(ctx, agent1)

	// Should NOT receive event
	select {
	case <-ch:
		t.Error("should not receive event for non-matching capability")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Register agent with matching capability
	agent2 := createTestAgentData("agent-2")
	agent2.Capabilities = []core.Capability{{Name: "process", Description: "Process"}}
	service.Register(ctx, agent2)

	// Should receive event
	select {
	case event := <-ch:
		if event.Agent.ID != "agent-2" {
			t.Errorf("expected agent-2, got %s", event.Agent.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected event for matching capability")
	}
}

func TestService_Unwatch(t *testing.T) {
	service, _ := setupRegistryService(t)

	id, ch := service.Watch(nil)
	service.Unwatch(id)

	// Channel should be closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after unwatch")
		}
	default:
		// May not be closed yet
	}
}

func TestService_WatcherNotification_Register(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	_, ch := service.Watch(nil)

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	select {
	case event := <-ch:
		if event.Type != EventRegistered {
			t.Errorf("expected EventRegistered, got %d", event.Type)
		}
		if event.Agent.ID != "agent-1" {
			t.Errorf("expected agent-1, got %s", event.Agent.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected registration event")
	}
}

func TestService_WatcherNotification_Deregister(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	_, ch := service.Watch(nil)

	service.Deregister(ctx, "agent-1")

	select {
	case event := <-ch:
		if event.Type != EventDeregistered {
			t.Errorf("expected EventDeregistered, got %d", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected deregistration event")
	}
}

func TestService_UpdateAgent(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	// Update agent
	agent.Name = "Updated Agent"
	agent.Version = "2.0.0"
	err := service.UpdateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("UpdateAgent() error = %v", err)
	}

	// Verify update
	retrieved, _ := service.GetAgent(ctx, "agent-1")
	if retrieved.Name != "Updated Agent" {
		t.Errorf("name not updated: got %s", retrieved.Name)
	}
	if retrieved.Version != "2.0.0" {
		t.Errorf("version not updated: got %s", retrieved.Version)
	}
}

func TestService_UpdateAgent_PreservesRegisteredAt(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	registered, _ := service.Register(ctx, agent)
	originalRegisteredAt := registered.RegisteredAt

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Update agent
	agent.Name = "Updated Agent"
	agent.RegisteredAt = time.Now() // Try to change it
	service.UpdateAgent(ctx, agent)

	// Verify RegisteredAt is preserved
	retrieved, _ := service.GetAgent(ctx, "agent-1")
	if !retrieved.RegisteredAt.Equal(originalRegisteredAt) {
		t.Error("RegisteredAt should be preserved")
	}
}

func TestService_UpdateAgent_ValidationError(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	// Try to update with invalid data
	agent.Name = ""
	err := service.UpdateAgent(ctx, agent)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestService_UpdateAgent_NotFound(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("nonexistent")
	err := service.UpdateAgent(ctx, agent)
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestService_UpdateAgent_Notification(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	_, ch := service.Watch(nil)

	agent.Name = "Updated Agent"
	service.UpdateAgent(ctx, agent)

	select {
	case event := <-ch:
		if event.Type != EventUpdated {
			t.Errorf("expected EventUpdated, got %d", event.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected update event")
	}
}

func TestService_MultipleWatchers(t *testing.T) {
	service, _ := setupRegistryService(t)
	ctx := context.Background()

	_, ch1 := service.Watch(nil)
	_, ch2 := service.Watch(nil)

	agent := createTestAgentData("agent-1")
	service.Register(ctx, agent)

	// Both watchers should receive event
	received := 0
	timeout := time.After(100 * time.Millisecond)

	for received < 2 {
		select {
		case <-ch1:
			received++
		case <-ch2:
			received++
		case <-timeout:
			break
		}
		if received == 2 {
			break
		}
	}

	if received != 2 {
		t.Errorf("expected 2 watchers to receive event, got %d", received)
	}
}

func TestEventType_Constants(t *testing.T) {
	if EventRegistered != 0 {
		t.Errorf("EventRegistered should be 0, got %d", EventRegistered)
	}
	if EventDeregistered != 1 {
		t.Errorf("EventDeregistered should be 1, got %d", EventDeregistered)
	}
	if EventUpdated != 2 {
		t.Errorf("EventUpdated should be 2, got %d", EventUpdated)
	}
	if EventHealthChanged != 3 {
		t.Errorf("EventHealthChanged should be 3, got %d", EventHealthChanged)
	}
}
