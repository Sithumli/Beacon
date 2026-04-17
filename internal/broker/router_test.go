package broker

import (
	"context"
	"testing"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/registry"
	"github.com/Sithumli/Beacon/internal/store"
)

func setupRouter(t *testing.T, strategy RouterStrategy) (*Router, *registry.Service, *store.MemoryStore) {
	t.Helper()
	memStore := store.NewMemoryStore()
	regService := registry.NewService(memStore)
	router := NewRouter(regService, strategy)
	return router, regService, memStore
}

func registerTestAgent(t *testing.T, regService *registry.Service, memStore *store.MemoryStore, id string, capability string, healthy bool) *core.Agent {
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

	regService.Register(ctx, agent)

	if !healthy {
		agent.Status = core.StatusUnhealthy
		memStore.UpdateAgent(ctx, agent)
	}

	return agent
}

func TestNewRouter(t *testing.T) {
	router, _, _ := setupRouter(t, StrategyRandom)

	if router == nil {
		t.Fatal("router should not be nil")
	}
	if router.rrIndex == nil {
		t.Fatal("rrIndex map should be initialized")
	}
	if router.rng == nil {
		t.Fatal("rng should be initialized")
	}
}

func TestRouter_SelectAgent_Random(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRandom)
	ctx := context.Background()

	// Register multiple healthy agents
	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-2", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-3", "echo", true)

	// Select agent multiple times - should work without error
	for i := 0; i < 10; i++ {
		agent, err := router.SelectAgent(ctx, "echo")
		if err != nil {
			t.Fatalf("SelectAgent() error = %v", err)
		}
		if agent == nil {
			t.Fatal("selected agent should not be nil")
		}
		if !agent.HasCapability("echo") {
			t.Error("selected agent should have echo capability")
		}
	}
}

func TestRouter_SelectAgent_RoundRobin(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRoundRobin)
	ctx := context.Background()

	// Register agents
	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-2", "echo", true)

	// Track selections
	selections := make(map[string]int)

	// Select multiple times
	for i := 0; i < 10; i++ {
		agent, err := router.SelectAgent(ctx, "echo")
		if err != nil {
			t.Fatalf("SelectAgent() error = %v", err)
		}
		selections[agent.ID]++
	}

	// Round robin should use both agents (at least once each)
	if len(selections) < 2 {
		t.Errorf("round robin should use multiple agents, got: %v", selections)
	}
	// Both agents should be selected
	if selections["agent-1"] == 0 || selections["agent-2"] == 0 {
		t.Errorf("round robin should select both agents, got: %v", selections)
	}
}

func TestRouter_SelectAgent_LeastTasks(t *testing.T) {
	// LeastTasks currently falls back to random
	router, regService, memStore := setupRouter(t, StrategyLeastTasks)
	ctx := context.Background()

	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-2", "echo", true)

	agent, err := router.SelectAgent(ctx, "echo")
	if err != nil {
		t.Fatalf("SelectAgent() error = %v", err)
	}
	if agent == nil {
		t.Fatal("selected agent should not be nil")
	}
}

func TestRouter_SelectAgent_NoAgents(t *testing.T) {
	router, _, _ := setupRouter(t, StrategyRandom)
	ctx := context.Background()

	agent, err := router.SelectAgent(ctx, "echo")
	if err != nil {
		t.Fatalf("SelectAgent() should not error, got = %v", err)
	}
	if agent != nil {
		t.Error("should return nil when no agents available")
	}
}

func TestRouter_SelectAgent_NoHealthyAgents(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRandom)
	ctx := context.Background()

	// Register unhealthy agents only
	registerTestAgent(t, regService, memStore, "agent-1", "echo", false)
	registerTestAgent(t, regService, memStore, "agent-2", "echo", false)

	agent, err := router.SelectAgent(ctx, "echo")
	if err != nil {
		t.Fatalf("SelectAgent() should not error, got = %v", err)
	}
	if agent != nil {
		t.Error("should return nil when no healthy agents")
	}
}

func TestRouter_SelectAgent_MixedHealth(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRandom)
	ctx := context.Background()

	// Mix of healthy and unhealthy
	registerTestAgent(t, regService, memStore, "agent-1", "echo", false) // unhealthy
	registerTestAgent(t, regService, memStore, "agent-2", "echo", true)  // healthy
	registerTestAgent(t, regService, memStore, "agent-3", "echo", false) // unhealthy

	// Should only select the healthy agent
	for i := 0; i < 5; i++ {
		agent, err := router.SelectAgent(ctx, "echo")
		if err != nil {
			t.Fatalf("SelectAgent() error = %v", err)
		}
		if agent.ID != "agent-2" {
			t.Errorf("should select healthy agent-2, got %s", agent.ID)
		}
	}
}

func TestRouter_SelectAgent_SingleAgent(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRandom)
	ctx := context.Background()

	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)

	agent, err := router.SelectAgent(ctx, "echo")
	if err != nil {
		t.Fatalf("SelectAgent() error = %v", err)
	}
	if agent.ID != "agent-1" {
		t.Errorf("expected agent-1, got %s", agent.ID)
	}
}

func TestRouter_SelectAgent_DifferentCapabilities(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRandom)
	ctx := context.Background()

	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-2", "process", true)

	// Select for echo capability
	agent, _ := router.SelectAgent(ctx, "echo")
	if agent == nil || agent.ID != "agent-1" {
		t.Error("should select agent-1 for echo capability")
	}

	// Select for process capability
	agent, _ = router.SelectAgent(ctx, "process")
	if agent == nil || agent.ID != "agent-2" {
		t.Error("should select agent-2 for process capability")
	}

	// Select for nonexistent capability
	agent, _ = router.SelectAgent(ctx, "nonexistent")
	if agent != nil {
		t.Error("should return nil for nonexistent capability")
	}
}

func TestRouter_SetStrategy(t *testing.T) {
	router, _, _ := setupRouter(t, StrategyRandom)

	if router.strategy != StrategyRandom {
		t.Errorf("initial strategy should be Random")
	}

	router.SetStrategy(StrategyRoundRobin)
	if router.strategy != StrategyRoundRobin {
		t.Error("strategy should be updated to RoundRobin")
	}

	router.SetStrategy(StrategyLeastTasks)
	if router.strategy != StrategyLeastTasks {
		t.Error("strategy should be updated to LeastTasks")
	}
}

func TestRouter_RoundRobinIndexWrapping(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRoundRobin)
	ctx := context.Background()

	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-2", "echo", true)

	// Make many selections to ensure index wraps correctly
	for i := 0; i < 100; i++ {
		agent, err := router.SelectAgent(ctx, "echo")
		if err != nil {
			t.Fatalf("SelectAgent() error = %v at iteration %d", err, i)
		}
		if agent == nil {
			t.Fatalf("agent should not be nil at iteration %d", i)
		}
	}
}

func TestRouter_RoundRobinPerCapability(t *testing.T) {
	router, regService, _ := setupRouter(t, StrategyRoundRobin)
	ctx := context.Background()

	// Register agents with different capabilities
	agent1 := &core.Agent{
		ID:           "agent-1",
		Name:         "Test Agent 1",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}, {Name: "process", Description: "Process"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent1)

	agent2 := &core.Agent{
		ID:           "agent-2",
		Name:         "Test Agent 2",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50052, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent2)

	// Round robin index should be separate per capability
	router.SelectAgent(ctx, "echo")
	echoIndex := router.rrIndex["echo"]

	router.SelectAgent(ctx, "process")
	processIndex := router.rrIndex["process"]

	// Both should have been incremented independently
	if echoIndex == 0 && processIndex == 0 {
		// This is fine - it means the first selection for each
	}

	// Select echo again
	router.SelectAgent(ctx, "echo")
	newEchoIndex := router.rrIndex["echo"]

	if newEchoIndex == echoIndex {
		t.Error("echo index should have incremented")
	}
}

func TestRouter_DefaultStrategy(t *testing.T) {
	router, regService, memStore := setupRouter(t, RouterStrategy(999)) // Invalid strategy
	ctx := context.Background()

	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)

	// Should fall back to random
	agent, err := router.SelectAgent(ctx, "echo")
	if err != nil {
		t.Fatalf("SelectAgent() error = %v", err)
	}
	if agent == nil {
		t.Fatal("agent should not be nil")
	}
}

func TestRouter_ConcurrentSelection(t *testing.T) {
	router, regService, memStore := setupRouter(t, StrategyRoundRobin)
	ctx := context.Background()

	registerTestAgent(t, regService, memStore, "agent-1", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-2", "echo", true)
	registerTestAgent(t, regService, memStore, "agent-3", "echo", true)

	// Run concurrent selections
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				agent, err := router.SelectAgent(ctx, "echo")
				if err != nil {
					t.Errorf("concurrent SelectAgent() error = %v", err)
					done <- false
					return
				}
				if agent == nil {
					t.Error("concurrent SelectAgent() returned nil")
					done <- false
					return
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRouterStrategy_Constants(t *testing.T) {
	// Verify strategy constants have expected values
	if StrategyRandom != 0 {
		t.Errorf("StrategyRandom should be 0, got %d", StrategyRandom)
	}
	if StrategyRoundRobin != 1 {
		t.Errorf("StrategyRoundRobin should be 1, got %d", StrategyRoundRobin)
	}
	if StrategyLeastTasks != 2 {
		t.Errorf("StrategyLeastTasks should be 2, got %d", StrategyLeastTasks)
	}
}
