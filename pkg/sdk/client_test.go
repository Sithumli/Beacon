package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantURL string
	}{
		{
			name:    "with scheme",
			addr:    "http://localhost:8080",
			wantURL: "http://localhost:8080",
		},
		{
			name:    "without scheme",
			addr:    "localhost:8080",
			wantURL: "http://localhost:8080",
		},
		{
			name:    "https scheme",
			addr:    "https://api.example.com",
			wantURL: "https://api.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.addr)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if client.baseURL != tt.wantURL {
				t.Errorf("baseURL = %s, want %s", client.baseURL, tt.wantURL)
			}
			if client.httpClient == nil {
				t.Error("httpClient should not be nil")
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	client, _ := NewClient("localhost:8080")

	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestClient_RegisterAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/agents" {
			t.Errorf("expected /api/v1/agents, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(registerResponse{
			AgentID: "agent-123",
			Agent: &agentResponse{
				ID:      "agent-123",
				Name:    "Test Agent",
				Version: "1.0.0",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	config := &AgentConfig{
		Name:    "Test Agent",
		Version: "1.0.0",
		Host:    "localhost",
		Port:    50051,
	}

	registered, err := client.RegisterAgent(ctx, config)
	if err != nil {
		t.Fatalf("RegisterAgent() error = %v", err)
	}

	if registered.ID != "agent-123" {
		t.Errorf("ID = %s, want agent-123", registered.ID)
	}
}

func TestClient_DeregisterAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/agents/agent-123" {
			t.Errorf("expected /api/v1/agents/agent-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	err := client.DeregisterAgent(ctx, "agent-123")
	if err != nil {
		t.Fatalf("DeregisterAgent() error = %v", err)
	}
}

func TestClient_Heartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/heartbeat" {
			t.Errorf("expected /api/v1/heartbeat, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(heartbeatResponse{Success: true, Status: "healthy"})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	err := client.Heartbeat(ctx, "agent-123")
	if err != nil {
		t.Fatalf("Heartbeat() error = %v", err)
	}
}

func TestClient_DiscoverAgents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("capability") != "echo" {
			t.Errorf("expected capability=echo, got %s", r.URL.Query().Get("capability"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*agentResponse{
			{ID: "agent-1", Name: "Agent 1"},
			{ID: "agent-2", Name: "Agent 2"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	agents, err := client.DiscoverAgents(ctx, "echo")
	if err != nil {
		t.Fatalf("DiscoverAgents() error = %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestClient_ListAgents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/agents" {
			t.Errorf("expected /api/v1/agents, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*agentResponse{
			{ID: "agent-1"},
			{ID: "agent-2"},
			{ID: "agent-3"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	agents, err := client.ListAgents(ctx)
	if err != nil {
		t.Fatalf("ListAgents() error = %v", err)
	}

	if len(agents) != 3 {
		t.Errorf("expected 3 agents, got %d", len(agents))
	}
}

func TestClient_GetAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/agents/agent-123" {
			t.Errorf("expected /api/v1/agents/agent-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(agentResponse{
			ID:      "agent-123",
			Name:    "Test Agent",
			Version: "1.0.0",
			Status:  "healthy",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	agent, err := client.GetAgent(ctx, "agent-123")
	if err != nil {
		t.Fatalf("GetAgent() error = %v", err)
	}

	if agent.ID != "agent-123" {
		t.Errorf("ID = %s, want agent-123", agent.ID)
	}
}

func TestClient_SendTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/tasks" {
			t.Errorf("expected /api/v1/tasks, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(sendTaskResponse{
			TaskID: "task-456",
			Task: &taskResponse{
				ID:         "task-456",
				FromAgent:  "sender",
				ToAgent:    "receiver",
				Capability: "echo",
				Status:     "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	task, err := client.SendTask(ctx, "sender", "receiver", "echo", map[string]string{"msg": "hello"})
	if err != nil {
		t.Fatalf("SendTask() error = %v", err)
	}

	if task.ID != "task-456" {
		t.Errorf("ID = %s, want task-456", task.ID)
	}
}

func TestClient_RouteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/route" {
			t.Errorf("expected /api/v1/route, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(routeTaskResponse{
			TaskID:  "task-789",
			ToAgent: "selected-agent",
			Task: &taskResponse{
				ID:         "task-789",
				FromAgent:  "sender",
				ToAgent:    "selected-agent",
				Capability: "echo",
				Status:     "pending",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	task, err := client.RouteTask(ctx, "sender", "echo", map[string]string{"msg": "hello"})
	if err != nil {
		t.Fatalf("RouteTask() error = %v", err)
	}

	if task.ID != "task-789" {
		t.Errorf("ID = %s, want task-789", task.ID)
	}
}

func TestClient_GetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/tasks/task-123" {
			t.Errorf("expected /api/v1/tasks/task-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(taskResponse{
			ID:     "task-123",
			Status: "completed",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	task, err := client.GetTask(ctx, "task-123")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if task.ID != "task-123" {
		t.Errorf("ID = %s, want task-123", task.ID)
	}
}

func TestClient_GetPendingTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Query().Get("agent_id") != "agent-123" {
			t.Errorf("expected agent_id=agent-123, got %s", r.URL.Query().Get("agent_id"))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*taskResponse{
			{ID: "task-1", Status: "pending"},
			{ID: "task-2", Status: "pending"},
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	tasks, err := client.GetPendingTasks(ctx, "agent-123")
	if err != nil {
		t.Fatalf("GetPendingTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestClient_UpdateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/tasks/task-123" {
			t.Errorf("expected /api/v1/tasks/task-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(taskResponse{
			ID:     "task-123",
			Status: "completed",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	task, err := client.UpdateTask(ctx, "task-123", TaskCompleted, map[string]string{"result": "done"}, "")
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if task.Status != TaskCompleted {
		t.Errorf("Status = %s, want completed", task.Status)
	}
}

func TestClient_WaitForTask(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		status := "pending"
		if callCount >= 3 {
			status = "completed"
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(taskResponse{
			ID:     "task-123",
			Status: status,
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	task, err := client.WaitForTask(ctx, "task-123", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForTask() error = %v", err)
	}

	if task.Status != TaskCompleted {
		t.Errorf("Status = %s, want completed", task.Status)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 calls, got %d", callCount)
	}
}

func TestClient_WaitForTask_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(taskResponse{
			ID:     "task-123",
			Status: "pending",
		})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.WaitForTask(ctx, "task-123", 10*time.Millisecond)
	if err == nil {
		t.Error("expected context deadline error")
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "something went wrong"})
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	_, err := client.GetAgent(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for bad request")
	}
	if err.Error() != "something went wrong" {
		t.Errorf("error = %v, want 'something went wrong'", err)
	}
}

func TestClient_ErrorResponse_NoMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	_, err := client.GetAgent(ctx, "agent-123")
	if err == nil {
		t.Error("expected error for server error")
	}
}

func TestResponseToAgentInfo(t *testing.T) {
	resp := &agentResponse{
		ID:          "agent-123",
		Name:        "Test Agent",
		Version:     "1.0.0",
		Description: "A test agent",
		Endpoint: endpointResponse{
			Host:     "localhost",
			Port:     50051,
			Protocol: "grpc",
		},
		Capabilities: []capabilityResponse{
			{Name: "echo", Description: "Echo"},
		},
		Status:        "healthy",
		RegisteredAt:  "2024-01-01T00:00:00Z",
		LastHeartbeat: "2024-01-01T01:00:00Z",
	}

	info := responseToAgentInfo(resp)

	if info.ID != "agent-123" {
		t.Errorf("ID = %s, want agent-123", info.ID)
	}
	if info.Name != "Test Agent" {
		t.Errorf("Name = %s, want Test Agent", info.Name)
	}
	if info.Host != "localhost" {
		t.Errorf("Host = %s, want localhost", info.Host)
	}
	if info.Port != 50051 {
		t.Errorf("Port = %d, want 50051", info.Port)
	}
	if len(info.Capabilities) != 1 {
		t.Errorf("Capabilities count = %d, want 1", len(info.Capabilities))
	}
	if info.Status != StatusHealthy {
		t.Errorf("Status = %s, want healthy", info.Status)
	}
}

func TestResponseToTaskInfo(t *testing.T) {
	completedAt := "2024-01-01T01:00:00Z"
	resp := &taskResponse{
		ID:          "task-123",
		FromAgent:   "sender",
		ToAgent:     "receiver",
		Capability:  "echo",
		Payload:     map[string]string{"input": "test"},
		Status:      "completed",
		Result:      map[string]string{"output": "result"},
		Error:       "",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:30:00Z",
		CompletedAt: &completedAt,
	}

	info := responseToTaskInfo(resp)

	if info.ID != "task-123" {
		t.Errorf("ID = %s, want task-123", info.ID)
	}
	if info.FromAgent != "sender" {
		t.Errorf("FromAgent = %s, want sender", info.FromAgent)
	}
	if info.ToAgent != "receiver" {
		t.Errorf("ToAgent = %s, want receiver", info.ToAgent)
	}
	if info.Status != TaskCompleted {
		t.Errorf("Status = %s, want completed", info.Status)
	}
	if info.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

func TestResponseToTaskInfo_NilCompletedAt(t *testing.T) {
	resp := &taskResponse{
		ID:          "task-123",
		Status:      "pending",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-01T00:00:00Z",
		CompletedAt: nil,
	}

	info := responseToTaskInfo(resp)

	if info.CompletedAt != nil {
		t.Error("CompletedAt should be nil")
	}
}
