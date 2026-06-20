package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/store"
)

func setupHTTPHandler(t *testing.T) (*HTTPHandler, *Service, *store.MemoryStore) {
	t.Helper()
	memStore := store.NewMemoryStore()
	service := NewService(memStore)
	handler := NewHTTPHandler(service)
	return handler, service, memStore
}

func TestNewHTTPHandler(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	if handler == nil {
		t.Fatal("handler should not be nil")
	}
	if handler.service == nil {
		t.Fatal("service should not be nil")
	}
}

func TestHTTPHandler_RegisterRoutes(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)
	mux := http.NewServeMux()

	handler.RegisterRoutes(mux)

	// Routes should be registered without panic
}

func TestHTTPHandler_HandleAgents_GET(t *testing.T) {
	handler, service, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register some agents
	agent := &core.Agent{
		ID:       "agent-1",
		Name:     "Test Agent",
		Version:  "1.0.0",
		Endpoint: core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
	}
	service.Register(ctx, agent)

	req := httptest.NewRequest("GET", "/api/v1/agents", nil)
	w := httptest.NewRecorder()

	handler.handleAgents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var agents []*core.Agent
	json.Unmarshal(w.Body.Bytes(), &agents)

	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
}

func TestHTTPHandler_HandleAgents_POST(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	reqBody := RegisterRequest{
		Name:    "Test Agent",
		Version: "1.0.0",
		Endpoint: EndpointRequest{
			Host:     "localhost",
			Port:     50051,
			Protocol: "grpc",
		},
		Capabilities: []CapabilityRequest{
			{Name: "echo", Description: "Echo capability"},
		},
		Metadata: MetadataRequest{
			Author: "Test Author",
			Tags:   []string{"test"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/agents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleAgents(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp RegisterResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.AgentID == "" {
		t.Error("agent ID should not be empty")
	}
}

func TestHTTPHandler_HandleAgents_POST_InvalidJSON(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/agents", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleAgents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleAgents_POST_ValidationError(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	reqBody := RegisterRequest{
		Name: "Test Agent",
		// Missing version
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/agents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleAgents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleAgents_MethodNotAllowed(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("DELETE", "/api/v1/agents", nil)
	w := httptest.NewRecorder()

	handler.handleAgents(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleAgent_GET(t *testing.T) {
	handler, service, _ := setupHTTPHandler(t)
	ctx := context.Background()

	agent := &core.Agent{
		ID:       "agent-1",
		Name:     "Test Agent",
		Version:  "1.0.0",
		Endpoint: core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
	}
	service.Register(ctx, agent)

	req := httptest.NewRequest("GET", "/api/v1/agents/agent-1", nil)
	w := httptest.NewRecorder()

	handler.handleAgent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp AgentResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.ID != "agent-1" {
		t.Errorf("expected agent-1, got %s", resp.ID)
	}
}

func TestHTTPHandler_HandleAgent_GET_NotFound(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/agents/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.handleAgent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleAgent_DELETE(t *testing.T) {
	handler, service, _ := setupHTTPHandler(t)
	ctx := context.Background()

	agent := &core.Agent{
		ID:       "agent-1",
		Name:     "Test Agent",
		Version:  "1.0.0",
		Endpoint: core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
	}
	service.Register(ctx, agent)

	req := httptest.NewRequest("DELETE", "/api/v1/agents/agent-1", nil)
	w := httptest.NewRecorder()

	handler.handleAgent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]bool
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp["success"] {
		t.Error("expected success: true")
	}
}

func TestHTTPHandler_HandleAgent_DELETE_NotFound(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("DELETE", "/api/v1/agents/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.handleAgent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleAgent_EmptyID(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/agents/", nil)
	w := httptest.NewRecorder()

	handler.handleAgent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleAgent_MethodNotAllowed(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("PATCH", "/api/v1/agents/agent-1", nil)
	w := httptest.NewRecorder()

	handler.handleAgent(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleDiscover(t *testing.T) {
	handler, service, _ := setupHTTPHandler(t)
	ctx := context.Background()

	agent := &core.Agent{
		ID:       "agent-1",
		Name:     "Test Agent",
		Version:  "1.0.0",
		Endpoint: core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{
			{Name: "echo", Description: "Echo"},
		},
		Status: core.StatusHealthy,
	}
	service.Register(ctx, agent)

	req := httptest.NewRequest("GET", "/api/v1/discover?capability=echo", nil)
	w := httptest.NewRecorder()

	handler.handleDiscover(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var agents []*AgentResponse
	json.Unmarshal(w.Body.Bytes(), &agents)

	if len(agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(agents))
	}
}

func TestHTTPHandler_HandleDiscover_MissingCapability(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/discover", nil)
	w := httptest.NewRecorder()

	handler.handleDiscover(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleDiscover_MethodNotAllowed(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/discover?capability=echo", nil)
	w := httptest.NewRecorder()

	handler.handleDiscover(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleHeartbeat(t *testing.T) {
	handler, service, _ := setupHTTPHandler(t)
	ctx := context.Background()

	agent := &core.Agent{
		ID:       "agent-1",
		Name:     "Test Agent",
		Version:  "1.0.0",
		Endpoint: core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
	}
	service.Register(ctx, agent)

	reqBody := HeartbeatRequest{AgentID: "agent-1"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/heartbeat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleHeartbeat(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp HeartbeatResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success: true")
	}
	if resp.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", resp.Status)
	}
}

func TestHTTPHandler_HandleHeartbeat_InvalidJSON(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/heartbeat", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleHeartbeat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleHeartbeat_AgentNotFound(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	reqBody := HeartbeatRequest{AgentID: "nonexistent"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/heartbeat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleHeartbeat(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleHeartbeat_MethodNotAllowed(t *testing.T) {
	handler, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/heartbeat", nil)
	w := httptest.NewRecorder()

	handler.handleHeartbeat(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestAgentToResponse(t *testing.T) {
	agent := &core.Agent{
		ID:          "agent-1",
		Name:        "Test Agent",
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
			Author: "Test Author",
			Tags:   []string{"test", "demo"},
		},
		Status: core.StatusHealthy,
	}

	resp := agentToResponse(agent)

	if resp.ID != agent.ID {
		t.Errorf("ID mismatch")
	}
	if resp.Name != agent.Name {
		t.Errorf("Name mismatch")
	}
	if resp.Version != agent.Version {
		t.Errorf("Version mismatch")
	}
	if resp.Description != agent.Description {
		t.Errorf("Description mismatch")
	}
	if resp.Endpoint.Host != agent.Endpoint.Host {
		t.Errorf("Endpoint host mismatch")
	}
	if resp.Endpoint.Port != agent.Endpoint.Port {
		t.Errorf("Endpoint port mismatch")
	}
	if len(resp.Capabilities) != len(agent.Capabilities) {
		t.Errorf("Capabilities count mismatch")
	}
	if resp.Metadata.Author != agent.Metadata.Author {
		t.Errorf("Metadata author mismatch")
	}
	if len(resp.Metadata.Tags) != len(agent.Metadata.Tags) {
		t.Errorf("Metadata tags count mismatch")
	}
	if resp.Status != string(agent.Status) {
		t.Errorf("Status mismatch")
	}
}

func TestJsonResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	jsonResponse(w, data)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["key"] != "value" {
		t.Error("response data mismatch")
	}
}

func TestJsonError(t *testing.T) {
	w := httptest.NewRecorder()

	jsonError(w, "something went wrong", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["error"] != "something went wrong" {
		t.Errorf("error message mismatch: got %s", result["error"])
	}
}
