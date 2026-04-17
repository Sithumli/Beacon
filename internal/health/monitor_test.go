package health

import (
	"context"
	"testing"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
)

func setupMonitor(t *testing.T, cfg Config) (*Monitor, *store.MemoryStore) {
	t.Helper()
	memStore := store.NewMemoryStore()
	monitor := NewMonitor(memStore, cfg)
	return monitor, memStore
}

func createHealthyAgent(t *testing.T, memStore *store.MemoryStore, id string) *core.Agent {
	t.Helper()
	ctx := context.Background()

	agent := &core.Agent{
		ID:            id,
		Name:          "Test Agent " + id,
		Version:       "1.0.0",
		Endpoint:      core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Status:        core.StatusHealthy,
		LastHeartbeat: time.Now().UTC(),
	}

	err := memStore.CreateAgent(ctx, agent)
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	return agent
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CheckInterval != 10*time.Second {
		t.Errorf("expected CheckInterval 10s, got %s", cfg.CheckInterval)
	}
	if cfg.HeartbeatTTL != 30*time.Second {
		t.Errorf("expected HeartbeatTTL 30s, got %s", cfg.HeartbeatTTL)
	}
}

func TestNewMonitor(t *testing.T) {
	monitor, _ := setupMonitor(t, DefaultConfig())

	if monitor == nil {
		t.Fatal("monitor should not be nil")
	}
	if monitor.watchers == nil {
		t.Fatal("watchers should be initialized")
	}
	if monitor.stopCh == nil {
		t.Fatal("stopCh should be initialized")
	}
}

func TestMonitor_StartStop(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  100 * time.Millisecond,
	}
	monitor, _ := setupMonitor(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start in goroutine
	done := make(chan error)
	go func() {
		done <- monitor.Start(ctx)
	}()

	// Wait for it to start
	time.Sleep(20 * time.Millisecond)

	// Stop it
	monitor.Stop()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("monitor did not stop in time")
	}
}

func TestMonitor_StartAlreadyRunning(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  100 * time.Millisecond,
	}
	monitor, _ := setupMonitor(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start monitor
	go monitor.Start(ctx)
	time.Sleep(20 * time.Millisecond)

	// Try to start again (should return immediately)
	done := make(chan error, 1)
	go func() {
		done <- monitor.Start(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("second start should return nil, got %v", err)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("second start should return immediately")
	}

	monitor.Stop()
}

func TestMonitor_StopNotRunning(t *testing.T) {
	monitor, _ := setupMonitor(t, DefaultConfig())

	// Should not panic
	monitor.Stop()
}

func TestMonitor_ContextCancellation(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  100 * time.Millisecond,
	}
	monitor, _ := setupMonitor(t, cfg)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- monitor.Start(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("monitor did not respond to context cancellation")
	}
}

func TestMonitor_Watch(t *testing.T) {
	monitor, _ := setupMonitor(t, DefaultConfig())

	ch := make(chan HealthEvent, 10)
	monitor.Watch(ch)

	if len(monitor.watchers) != 1 {
		t.Errorf("expected 1 watcher, got %d", len(monitor.watchers))
	}
}

func TestMonitor_Unwatch(t *testing.T) {
	monitor, _ := setupMonitor(t, DefaultConfig())

	ch := make(chan HealthEvent, 10)
	monitor.Watch(ch)
	monitor.Unwatch(ch)

	if len(monitor.watchers) != 0 {
		t.Errorf("expected 0 watchers, got %d", len(monitor.watchers))
	}
}

func TestMonitor_UnwatchNonexistent(t *testing.T) {
	monitor, _ := setupMonitor(t, DefaultConfig())

	ch := make(chan HealthEvent, 10)
	// Should not panic
	monitor.Unwatch(ch)
}

func TestMonitor_CheckAgents_MarksUnhealthy(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  50 * time.Millisecond, // Very short TTL for testing
	}
	monitor, memStore := setupMonitor(t, cfg)
	ctx := context.Background()

	// Create agent with old heartbeat
	agent := createHealthyAgent(t, memStore, "agent-1")
	agent.LastHeartbeat = time.Now().Add(-time.Hour) // Old heartbeat
	memStore.UpdateAgent(ctx, agent)

	// Watch for events
	ch := make(chan HealthEvent, 10)
	monitor.Watch(ch)

	// Run check
	monitor.checkAgents(ctx)

	// Should receive health change event
	select {
	case event := <-ch:
		if event.AgentID != "agent-1" {
			t.Errorf("expected agent-1, got %s", event.AgentID)
		}
		if event.OldStatus != core.StatusHealthy {
			t.Errorf("expected old status healthy, got %s", event.OldStatus)
		}
		if event.NewStatus != core.StatusUnhealthy {
			t.Errorf("expected new status unhealthy, got %s", event.NewStatus)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected health event")
	}

	// Verify agent is now unhealthy
	updated, _ := memStore.GetAgent(ctx, "agent-1")
	if updated.Status != core.StatusUnhealthy {
		t.Errorf("agent should be unhealthy, got %s", updated.Status)
	}
}

func TestMonitor_CheckAgents_DoesNotAffectHealthy(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  time.Hour, // Long TTL
	}
	monitor, memStore := setupMonitor(t, cfg)
	ctx := context.Background()

	// Create agent with recent heartbeat
	createHealthyAgent(t, memStore, "agent-1")

	// Watch for events
	ch := make(chan HealthEvent, 10)
	monitor.Watch(ch)

	// Run check
	monitor.checkAgents(ctx)

	// Should NOT receive event
	select {
	case event := <-ch:
		t.Errorf("should not receive event for healthy agent, got %+v", event)
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Verify agent is still healthy
	agent, _ := memStore.GetAgent(ctx, "agent-1")
	if agent.Status != core.StatusHealthy {
		t.Errorf("agent should still be healthy, got %s", agent.Status)
	}
}

func TestMonitor_CheckAgents_DoesNotAffectAlreadyUnhealthy(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  50 * time.Millisecond,
	}
	monitor, memStore := setupMonitor(t, cfg)
	ctx := context.Background()

	// Create already unhealthy agent with old heartbeat
	agent := createHealthyAgent(t, memStore, "agent-1")
	agent.Status = core.StatusUnhealthy
	agent.LastHeartbeat = time.Now().Add(-time.Hour)
	memStore.UpdateAgent(ctx, agent)

	// Watch for events
	ch := make(chan HealthEvent, 10)
	monitor.Watch(ch)

	// Run check
	monitor.checkAgents(ctx)

	// Should NOT receive event (already unhealthy)
	select {
	case event := <-ch:
		t.Errorf("should not receive event for already unhealthy agent, got %+v", event)
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

func TestMonitor_MultipleWatchers(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  50 * time.Millisecond,
	}
	monitor, memStore := setupMonitor(t, cfg)
	ctx := context.Background()

	// Create agent with old heartbeat
	agent := createHealthyAgent(t, memStore, "agent-1")
	agent.LastHeartbeat = time.Now().Add(-time.Hour)
	memStore.UpdateAgent(ctx, agent)

	// Create multiple watchers
	ch1 := make(chan HealthEvent, 10)
	ch2 := make(chan HealthEvent, 10)
	monitor.Watch(ch1)
	monitor.Watch(ch2)

	// Run check
	monitor.checkAgents(ctx)

	// Both should receive events
	received := 0
	timeout := time.After(100 * time.Millisecond)

loop:
	for received < 2 {
		select {
		case <-ch1:
			received++
		case <-ch2:
			received++
		case <-timeout:
			break loop
		}
	}

	if received != 2 {
		t.Errorf("expected 2 watchers to receive events, got %d", received)
	}
}

func TestMonitor_WatcherChannelFull(t *testing.T) {
	cfg := Config{
		CheckInterval: 50 * time.Millisecond,
		HeartbeatTTL:  50 * time.Millisecond,
	}
	monitor, memStore := setupMonitor(t, cfg)
	ctx := context.Background()

	// Create watcher with small buffer
	ch := make(chan HealthEvent, 1)
	monitor.Watch(ch)

	// Fill the channel
	ch <- HealthEvent{}

	// Create agent with old heartbeat
	agent := createHealthyAgent(t, memStore, "agent-1")
	agent.LastHeartbeat = time.Now().Add(-time.Hour)
	memStore.UpdateAgent(ctx, agent)

	// Run check - should not block even though channel is full
	done := make(chan bool)
	go func() {
		monitor.checkAgents(ctx)
		done <- true
	}()

	select {
	case <-done:
		// Good - didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("checkAgents blocked on full channel")
	}
}

func TestMonitor_IntegrationTest(t *testing.T) {
	cfg := Config{
		CheckInterval: 30 * time.Millisecond,
		HeartbeatTTL:  20 * time.Millisecond,
	}
	monitor, memStore := setupMonitor(t, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create agent
	agent := createHealthyAgent(t, memStore, "agent-1")

	// Watch for events
	ch := make(chan HealthEvent, 10)
	monitor.Watch(ch)

	// Start monitor
	go monitor.Start(ctx)

	// Simulate old heartbeat
	agent.LastHeartbeat = time.Now().Add(-time.Hour)
	memStore.UpdateAgent(ctx, agent)

	// Wait for health check to detect expired agent
	select {
	case event := <-ch:
		if event.AgentID != "agent-1" {
			t.Errorf("expected agent-1, got %s", event.AgentID)
		}
		if event.NewStatus != core.StatusUnhealthy {
			t.Errorf("expected unhealthy status, got %s", event.NewStatus)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("expected health event")
	}

	monitor.Stop()
}

func TestHealthEvent_Fields(t *testing.T) {
	now := time.Now()
	event := HealthEvent{
		AgentID:   "agent-1",
		OldStatus: core.StatusHealthy,
		NewStatus: core.StatusUnhealthy,
		Timestamp: now,
	}

	if event.AgentID != "agent-1" {
		t.Error("AgentID mismatch")
	}
	if event.OldStatus != core.StatusHealthy {
		t.Error("OldStatus mismatch")
	}
	if event.NewStatus != core.StatusUnhealthy {
		t.Error("NewStatus mismatch")
	}
	if !event.Timestamp.Equal(now) {
		t.Error("Timestamp mismatch")
	}
}

func TestConfig_Fields(t *testing.T) {
	cfg := Config{
		CheckInterval: 5 * time.Second,
		HeartbeatTTL:  15 * time.Second,
	}

	if cfg.CheckInterval != 5*time.Second {
		t.Error("CheckInterval mismatch")
	}
	if cfg.HeartbeatTTL != 15*time.Second {
		t.Error("HeartbeatTTL mismatch")
	}
}
