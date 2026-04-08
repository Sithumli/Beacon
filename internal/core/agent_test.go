package core

import (
	"testing"
	"time"
)

func TestNewAgent(t *testing.T) {
	agent := NewAgent("TestAgent", "1.0.0", "A test agent")

	if agent.Name != "TestAgent" {
		t.Errorf("expected name 'TestAgent', got '%s'", agent.Name)
	}
	if agent.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", agent.Version)
	}
	if agent.ID == "" {
		t.Error("expected ID to be generated")
	}
	if agent.Status != StatusHealthy {
		t.Errorf("expected status 'healthy', got '%s'", agent.Status)
	}
}

func TestAgentValidate(t *testing.T) {
	tests := []struct {
		name    string
		agent   *Agent
		wantErr bool
	}{
		{
			name: "valid agent",
			agent: &Agent{
				Name:    "TestAgent",
				Version: "1.0.0",
				Endpoint: Endpoint{
					Host:     "localhost",
					Port:     50051,
					Protocol: "grpc",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			agent: &Agent{
				Version: "1.0.0",
				Endpoint: Endpoint{
					Host: "localhost",
					Port: 50051,
				},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			agent: &Agent{
				Name: "TestAgent",
				Endpoint: Endpoint{
					Host: "localhost",
					Port: 50051,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			agent: &Agent{
				Name:    "TestAgent",
				Version: "1.0.0",
				Endpoint: Endpoint{
					Host: "localhost",
					Port: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentCapabilities(t *testing.T) {
	agent := NewAgent("TestAgent", "1.0.0", "A test agent")

	cap := Capability{
		Name:        "echo",
		Description: "Echo capability",
	}

	err := agent.AddCapability(cap)
	if err != nil {
		t.Errorf("failed to add capability: %v", err)
	}

	if !agent.HasCapability("echo") {
		t.Error("expected agent to have 'echo' capability")
	}

	if agent.HasCapability("nonexistent") {
		t.Error("expected agent to not have 'nonexistent' capability")
	}

	// Test duplicate
	err = agent.AddCapability(cap)
	if err == nil {
		t.Error("expected error when adding duplicate capability")
	}
}

func TestAgentHeartbeat(t *testing.T) {
	agent := NewAgent("TestAgent", "1.0.0", "A test agent")

	oldHeartbeat := agent.LastHeartbeat
	time.Sleep(10 * time.Millisecond)
	agent.UpdateHeartbeat()

	if !agent.LastHeartbeat.After(oldHeartbeat) {
		t.Error("expected heartbeat to be updated")
	}
}

func TestAgentIsExpired(t *testing.T) {
	agent := NewAgent("TestAgent", "1.0.0", "A test agent")

	// Should not be expired immediately
	if agent.IsExpired(time.Hour) {
		t.Error("agent should not be expired")
	}

	// Set old heartbeat
	agent.LastHeartbeat = time.Now().Add(-2 * time.Hour)

	if !agent.IsExpired(time.Hour) {
		t.Error("agent should be expired")
	}
}
