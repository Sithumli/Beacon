package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/registry"
	"github.com/Sithumli/Beacon/internal/store"
)

func setupHTTPHandler(t *testing.T) (*HTTPHandler, *Service, *registry.Service, *store.MemoryStore) {
	t.Helper()
	memStore := store.NewMemoryStore()
	regService := registry.NewService(memStore)
	brokerService := NewService(memStore, regService)
	handler := NewHTTPHandler(brokerService)
	return handler, brokerService, regService, memStore
}

func TestNewHTTPHandler(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	if handler == nil {
		t.Fatal("handler should not be nil")
	}
	if handler.service == nil {
		t.Fatal("service should not be nil")
	}
}

func TestHTTPHandler_RegisterRoutes(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)
	mux := http.NewServeMux()

	handler.RegisterRoutes(mux)

	// Routes are registered - we can't easily verify this without making requests
}

func TestHTTPHandler_HandleTasks_GET(t *testing.T) {
	handler, service, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent and create tasks
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	service.SendTask(ctx, "sender", "receiver-1", "echo", payload)

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()

	handler.handleTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var tasks []*TaskResponse
	json.Unmarshal(w.Body.Bytes(), &tasks)

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestHTTPHandler_HandleTasks_POST(t *testing.T) {
	handler, _, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	reqBody := SendTaskRequest{
		FromAgent:  "sender",
		ToAgent:    "receiver-1",
		Capability: "echo",
		Payload:    map[string]string{"message": "hello"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleTasks(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp SendTaskResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.TaskID == "" {
		t.Error("task ID should not be empty")
	}
}

func TestHTTPHandler_HandleTasks_POST_InvalidJSON(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleTasks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleTasks_POST_AgentNotFound(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	reqBody := SendTaskRequest{
		FromAgent:  "sender",
		ToAgent:    "nonexistent",
		Capability: "echo",
		Payload:    map[string]string{"message": "hello"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleTasks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleTasks_MethodNotAllowed(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("DELETE", "/api/v1/tasks", nil)
	w := httptest.NewRecorder()

	handler.handleTasks(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleTask_GET(t *testing.T) {
	handler, service, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent and create task
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender", "receiver-1", "echo", payload)

	req := httptest.NewRequest("GET", "/api/v1/tasks/"+task.ID, nil)
	w := httptest.NewRecorder()

	handler.handleTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp TaskResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.ID != task.ID {
		t.Errorf("task ID mismatch: got %s, want %s", resp.ID, task.ID)
	}
}

func TestHTTPHandler_HandleTask_GET_NotFound(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/tasks/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.handleTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleTask_PATCH(t *testing.T) {
	handler, service, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent and create task
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender", "receiver-1", "echo", payload)

	reqBody := UpdateTaskRequest{
		Status: "running",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PATCH", "/api/v1/tasks/"+task.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp TaskResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != "running" {
		t.Errorf("expected status running, got %s", resp.Status)
	}
}

func TestHTTPHandler_HandleTask_PATCH_WithResult(t *testing.T) {
	handler, service, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent and create task
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender", "receiver-1", "echo", payload)

	// Start the task first
	service.UpdateTask(ctx, task.ID, core.TaskRunning, nil, "")

	reqBody := UpdateTaskRequest{
		Status: "completed",
		Result: map[string]string{"output": "world"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("PATCH", "/api/v1/tasks/"+task.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHTTPHandler_HandleTask_Cancel(t *testing.T) {
	handler, service, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent and create task
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	task, _ := service.SendTask(ctx, "sender", "receiver-1", "echo", payload)

	req := httptest.NewRequest("POST", "/api/v1/tasks/"+task.ID+"/cancel", nil)
	w := httptest.NewRecorder()

	handler.handleTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp TaskResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Status != "cancelled" {
		t.Errorf("expected status cancelled, got %s", resp.Status)
	}
}

func TestHTTPHandler_HandleTask_EmptyID(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/tasks/", nil)
	w := httptest.NewRecorder()

	handler.handleTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandlePendingTasks(t *testing.T) {
	handler, service, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent and create task
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	payload := json.RawMessage(`{"message": "hello"}`)
	service.SendTask(ctx, "sender", "receiver-1", "echo", payload)

	req := httptest.NewRequest("GET", "/api/v1/pending?agent_id=receiver-1", nil)
	w := httptest.NewRecorder()

	handler.handlePendingTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var tasks []*TaskResponse
	json.Unmarshal(w.Body.Bytes(), &tasks)

	if len(tasks) != 1 {
		t.Errorf("expected 1 pending task, got %d", len(tasks))
	}
}

func TestHTTPHandler_HandlePendingTasks_MissingAgentID(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/pending", nil)
	w := httptest.NewRecorder()

	handler.handlePendingTasks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandlePendingTasks_MethodNotAllowed(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/pending?agent_id=receiver-1", nil)
	w := httptest.NewRecorder()

	handler.handlePendingTasks(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleRoute(t *testing.T) {
	handler, _, regService, _ := setupHTTPHandler(t)
	ctx := context.Background()

	// Register agent
	agent := &core.Agent{
		ID:           "receiver-1",
		Name:         "Test Agent",
		Version:      "1.0.0",
		Endpoint:     core.Endpoint{Host: "localhost", Port: 50051, Protocol: "grpc"},
		Capabilities: []core.Capability{{Name: "echo", Description: "Echo"}},
		Status:       core.StatusHealthy,
	}
	regService.Register(ctx, agent)

	reqBody := RouteTaskRequest{
		FromAgent:  "sender",
		Capability: "echo",
		Payload:    map[string]string{"message": "hello"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/route", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleRoute(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp RouteTaskResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.TaskID == "" {
		t.Error("task ID should not be empty")
	}
	if resp.ToAgent != "receiver-1" {
		t.Errorf("expected to_agent receiver-1, got %s", resp.ToAgent)
	}
}

func TestHTTPHandler_HandleRoute_NoAgents(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	reqBody := RouteTaskRequest{
		FromAgent:  "sender",
		Capability: "echo",
		Payload:    map[string]string{"message": "hello"},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/route", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleRoute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleRoute_InvalidJSON(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("POST", "/api/v1/route", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.handleRoute(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHTTPHandler_HandleRoute_MethodNotAllowed(t *testing.T) {
	handler, _, _, _ := setupHTTPHandler(t)

	req := httptest.NewRequest("GET", "/api/v1/route", nil)
	w := httptest.NewRecorder()

	handler.handleRoute(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestTaskToResponse(t *testing.T) {
	task := &core.Task{
		ID:         "task-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Capability: "echo",
		Payload:    json.RawMessage(`{"input": "test"}`),
		Status:     core.TaskCompleted,
		Result:     json.RawMessage(`{"output": "result"}`),
		Error:      "",
	}

	resp := taskToResponse(task)

	if resp.ID != task.ID {
		t.Errorf("ID mismatch")
	}
	if resp.FromAgent != task.FromAgent {
		t.Errorf("FromAgent mismatch")
	}
	if resp.Status != string(task.Status) {
		t.Errorf("Status mismatch")
	}
	if resp.Payload == nil {
		t.Error("Payload should not be nil")
	}
	if resp.Result == nil {
		t.Error("Result should not be nil")
	}
}

func TestTaskToResponse_EmptyPayload(t *testing.T) {
	task := &core.Task{
		ID:         "task-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Capability: "echo",
		Status:     core.TaskPending,
	}

	resp := taskToResponse(task)

	if resp.Payload != nil {
		t.Error("Payload should be nil for empty payload")
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
