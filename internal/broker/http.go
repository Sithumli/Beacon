package broker

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Sithumli/Beacon/internal/core"
)

// HTTPHandler provides HTTP endpoints for the broker service
type HTTPHandler struct {
	service *Service
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{service: svc}
}

// RegisterRoutes registers HTTP routes on the given mux
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/tasks", h.handleTasks)
	mux.HandleFunc("/api/v1/tasks/", h.handleTask)
	mux.HandleFunc("/api/v1/route", h.handleRoute)
	mux.HandleFunc("/api/v1/pending", h.handlePendingTasks)
}

func (h *HTTPHandler) handleTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		// List tasks
		tasks, err := h.service.ListTasks(ctx, nil)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responses := make([]*TaskResponse, len(tasks))
		for i, task := range tasks {
			responses[i] = taskToResponse(task)
		}
		jsonResponse(w, responses)

	case http.MethodPost:
		// Send task
		var req SendTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		payload, err := json.Marshal(req.Payload)
		if err != nil {
			jsonError(w, "invalid payload", http.StatusBadRequest)
			return
		}

		task, err := h.service.SendTask(ctx, req.FromAgent, req.ToAgent, req.Capability, payload)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		jsonResponse(w, SendTaskResponse{
			TaskID: task.ID,
			Task:   taskToResponse(task),
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) handleTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	parts := strings.Split(path, "/")
	taskID := parts[0]

	if taskID == "" {
		jsonError(w, "task ID required", http.StatusBadRequest)
		return
	}

	// Check for sub-routes like /tasks/{id}/cancel
	if len(parts) > 1 && parts[1] == "cancel" && r.Method == http.MethodPost {
		task, err := h.service.CancelTask(ctx, taskID)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, taskToResponse(task))
		return
	}

	switch r.Method {
	case http.MethodGet:
		task, err := h.service.GetTask(ctx, taskID)
		if err != nil {
			jsonError(w, "task not found", http.StatusNotFound)
			return
		}
		jsonResponse(w, taskToResponse(task))

	case http.MethodPatch:
		var req UpdateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var result json.RawMessage
		if req.Result != nil {
			marshaledResult, err := json.Marshal(req.Result)
			if err != nil {
				jsonError(w, "invalid result payload", http.StatusBadRequest)
				return
			}
			result = marshaledResult
		}

		task, err := h.service.UpdateTask(ctx, taskID, core.TaskStatus(req.Status), result, req.Error)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonResponse(w, taskToResponse(task))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) handlePendingTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		jsonError(w, "agent_id query parameter required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	tasks, err := h.service.GetPendingTasks(ctx, agentID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = taskToResponse(task)
	}
	jsonResponse(w, responses)
}

func (h *HTTPHandler) handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req RouteTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	payload, err := json.Marshal(req.Payload)
	if err != nil {
		jsonError(w, "invalid payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	task, err := h.service.RouteTask(ctx, req.FromAgent, req.Capability, payload)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonResponse(w, RouteTaskResponse{
		TaskID:  task.ID,
		ToAgent: task.ToAgent,
		Task:    taskToResponse(task),
	})
}

// Request/Response types

type SendTaskRequest struct {
	FromAgent  string      `json:"from_agent"`
	ToAgent    string      `json:"to_agent"`
	Capability string      `json:"capability"`
	Payload    interface{} `json:"payload"`
}

type SendTaskResponse struct {
	TaskID string        `json:"task_id"`
	Task   *TaskResponse `json:"task"`
}

type UpdateTaskRequest struct {
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type RouteTaskRequest struct {
	FromAgent  string      `json:"from_agent"`
	Capability string      `json:"capability"`
	Payload    interface{} `json:"payload"`
}

type RouteTaskResponse struct {
	TaskID  string        `json:"task_id"`
	ToAgent string        `json:"to_agent"`
	Task    *TaskResponse `json:"task"`
}

type TaskResponse struct {
	ID          string      `json:"task_id"`
	FromAgent   string      `json:"from_agent"`
	ToAgent     string      `json:"to_agent"`
	Capability  string      `json:"capability"`
	Payload     interface{} `json:"payload"`
	Status      string      `json:"status"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	CompletedAt *string     `json:"completed_at,omitempty"`
}

func taskToResponse(t *core.Task) *TaskResponse {
	resp := &TaskResponse{
		ID:         t.ID,
		FromAgent:  t.FromAgent,
		ToAgent:    t.ToAgent,
		Capability: t.Capability,
		Status:     string(t.Status),
		Error:      t.Error,
		CreatedAt:  t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Parse payload JSON
	if len(t.Payload) > 0 {
		var payload interface{}
		if json.Unmarshal(t.Payload, &payload) == nil {
			resp.Payload = payload
		}
	}

	// Parse result JSON
	if len(t.Result) > 0 {
		var result interface{}
		if json.Unmarshal(t.Result, &result) == nil {
			resp.Result = result
		}
	}

	if t.CompletedAt != nil {
		completedAt := t.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &completedAt
	}

	return resp
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
