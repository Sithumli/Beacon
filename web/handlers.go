package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/Sithumli/Beacon/internal/broker"
	"github.com/Sithumli/Beacon/internal/core"
	"github.com/Sithumli/Beacon/internal/registry"
	"github.com/Sithumli/Beacon/internal/store"
)

// Handler serves the web dashboard
type Handler struct {
	registry *registry.Service
	broker   *broker.Service
}

// NewHandler creates a new web handler
func NewHandler(reg *registry.Service, brk *broker.Service) *Handler {
	return &Handler{
		registry: reg,
		broker:   brk,
	}
}

// RegisterRoutes registers all web routes on the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Dashboard pages
	mux.HandleFunc("/", h.handleDashboard)
	mux.HandleFunc("/agents", h.handleAgents)
	mux.HandleFunc("/agents/", h.handleAgentDetail)
	mux.HandleFunc("/tasks", h.handleTasks)
	mux.HandleFunc("/discovery", h.handleDiscovery)
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/stats", h.handleStats)

	// API endpoints
	mux.HandleFunc("/api/agents", h.handleAPIAgents)
	mux.HandleFunc("/api/agents/", h.handleAPIAgentDetail)
	mux.HandleFunc("/api/tasks", h.handleAPITasks)
	mux.HandleFunc("/api/stats", h.handleAPIStats)
	mux.HandleFunc("/api/discover", h.handleAPIDiscover)

	// Static files
	mux.HandleFunc("/static/", h.handleStatic)

	// SSE for real-time updates
	mux.HandleFunc("/events", h.handleSSE)
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	agents, _ := h.registry.ListAgents(ctx, nil)
	tasks, _ := h.broker.ListTasks(ctx, nil)

	// Calculate stats
	activeCount := 0
	for _, a := range agents {
		if a.Status == core.StatusHealthy {
			activeCount++
		}
	}

	runningTasks := 0
	completedTasks := 0
	failedTasks := 0
	for _, t := range tasks {
		switch t.Status {
		case core.TaskPending, core.TaskRunning:
			runningTasks++
		case core.TaskCompleted:
			completedTasks++
		case core.TaskFailed:
			failedTasks++
		}
	}

	// Calculate health rate
	healthRate := 100.0
	if len(agents) > 0 {
		healthRate = float64(activeCount) / float64(len(agents)) * 100
	}

	// System load (mock for now)
	systemLoad := 42

	data := map[string]interface{}{
		"Page":           "dashboard",
		"TotalAgents":    len(agents),
		"ActiveAgents":   activeCount,
		"TasksDone":      completedTasks,
		"SystemLoad":     systemLoad,
		"HealthRate":     fmt.Sprintf("%.1f", healthRate),
		"RunningTasks":   runningTasks,
		"FailedTasks":    failedTasks,
		"CompletedTasks": completedTasks,
		"Agents":         limitAgents(agents, 5),
	}

	renderTemplate(w, "dashboard", data)
}

func (h *Handler) handleAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, err := h.registry.ListAgents(ctx, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Page":   "agents",
		"Agents": agents,
	}

	renderTemplate(w, "agents", data)
}

func (h *Handler) handleAgentDetail(w http.ResponseWriter, r *http.Request) {
	agentID := strings.TrimPrefix(r.URL.Path, "/agents/")
	if agentID == "" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	agent, err := h.registry.GetAgent(ctx, agentID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	filter := &store.TaskFilter{ToAgent: &agentID}
	tasks, _ := h.broker.ListTasks(ctx, filter)

	data := map[string]interface{}{
		"Page":   "agents",
		"Agent":  agent,
		"Tasks":  tasks,
		"Agents": []*core.Agent{agent},
	}

	renderTemplate(w, "agents", data)
}

func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.broker.ListTasks(ctx, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	runningTasks := 0
	completedTasks := 0
	failedTasks := 0
	for _, t := range tasks {
		switch t.Status {
		case core.TaskPending, core.TaskRunning:
			runningTasks++
		case core.TaskCompleted:
			completedTasks++
		case core.TaskFailed:
			failedTasks++
		}
	}

	data := map[string]interface{}{
		"Page":           "tasks",
		"Tasks":          tasks,
		"RunningTasks":   runningTasks,
		"FailedTasks":    failedTasks,
		"CompletedTasks": completedTasks,
	}

	renderTemplate(w, "tasks", data)
}

func (h *Handler) handleDiscovery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, _ := h.registry.ListAgents(ctx, nil)

	// Group by capabilities
	capabilityFilter := r.URL.Query().Get("capability")

	var filteredAgents []*core.Agent
	if capabilityFilter != "" {
		for _, a := range agents {
			for _, cap := range a.Capabilities {
				if strings.EqualFold(cap.Name, capabilityFilter) {
					filteredAgents = append(filteredAgents, a)
					break
				}
			}
		}
	} else {
		filteredAgents = agents
	}

	data := map[string]interface{}{
		"Page":   "discovery",
		"Agents": filteredAgents,
		"Filter": capabilityFilter,
	}

	renderTemplate(w, "discovery", data)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, _ := h.registry.ListAgents(ctx, nil)

	data := map[string]interface{}{
		"Page":   "health",
		"Agents": agents,
	}

	renderTemplate(w, "health", data)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, _ := h.registry.ListAgents(ctx, nil)
	tasks, _ := h.broker.ListTasks(ctx, nil)

	data := map[string]interface{}{
		"Page":   "stats",
		"Agents": agents,
		"Tasks":  tasks,
	}

	renderTemplate(w, "stats", data)
}

// API Handlers

func (h *Handler) handleAPIAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, err := h.registry.ListAgents(ctx, nil)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, agents)
}

func (h *Handler) handleAPIAgentDetail(w http.ResponseWriter, r *http.Request) {
	agentID := strings.TrimPrefix(r.URL.Path, "/api/agents/")
	if agentID == "" {
		jsonError(w, "agent ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	agent, err := h.registry.GetAgent(ctx, agentID)
	if err != nil {
		jsonError(w, "agent not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, agent)
}

func (h *Handler) handleAPITasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tasks, err := h.broker.ListTasks(ctx, nil)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, tasks)
}

func (h *Handler) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agents, _ := h.registry.ListAgents(ctx, nil)
	tasks, _ := h.broker.ListTasks(ctx, nil)

	healthyCount := 0
	for _, a := range agents {
		if a.Status == core.StatusHealthy {
			healthyCount++
		}
	}

	pendingTasks := 0
	completedTasks := 0
	failedTasks := 0
	for _, t := range tasks {
		switch t.Status {
		case core.TaskPending, core.TaskRunning:
			pendingTasks++
		case core.TaskCompleted:
			completedTasks++
		case core.TaskFailed:
			failedTasks++
		}
	}

	stats := map[string]interface{}{
		"total_agents":    len(agents),
		"healthy_agents":  healthyCount,
		"total_tasks":     len(tasks),
		"pending_tasks":   pendingTasks,
		"completed_tasks": completedTasks,
		"failed_tasks":    failedTasks,
	}

	jsonResponse(w, stats)
}

func (h *Handler) handleAPIDiscover(w http.ResponseWriter, r *http.Request) {
	capability := r.URL.Query().Get("capability")
	if capability == "" {
		jsonError(w, "capability required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	agents, err := h.registry.Discover(ctx, capability)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, agents)
}

func (h *Handler) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "style.css") {
		w.Header().Set("Content-Type", "text/css")
		w.Write([]byte(obsidianCSS))
		return
	}
	if strings.HasSuffix(r.URL.Path, "app.js") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(appJS))
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			agents, _ := h.registry.ListAgents(context.Background(), nil)
			tasks, _ := h.broker.ListTasks(context.Background(), nil)

			data := map[string]interface{}{
				"agents": len(agents),
				"tasks":  len(tasks),
			}
			jsonData, _ := json.Marshal(data)

			w.Write([]byte("data: "))
			w.Write(jsonData)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

// Helper functions

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func limitAgents(agents []*core.Agent, limit int) []*core.Agent {
	if len(agents) <= limit {
		return agents
	}
	return agents[:limit]
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl := template.New(name).Funcs(template.FuncMap{
		"truncateID": func(id string) string {
			if len(id) > 8 {
				return id[:8]
			}
			return id
		},
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"statusColor": func(status core.AgentStatus) string {
			if status == core.StatusHealthy {
				return "active"
			}
			return "inactive"
		},
		"taskStatusColor": func(status core.TaskStatus) string {
			switch status {
			case core.TaskRunning:
				return "running"
			case core.TaskCompleted:
				return "completed"
			case core.TaskFailed:
				return "failed"
			default:
				return "pending"
			}
		},
	})

	var err error
	tmpl, err = tmpl.Parse(getTemplate(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func getTemplate(name string) string {
	switch name {
	case "dashboard":
		return dashboardTemplate
	case "agents":
		return agentsTemplate
	case "tasks":
		return tasksTemplate
	case "discovery":
		return discoveryTemplate
	case "health":
		return healthTemplate
	case "stats":
		return statsTemplate
	default:
		return dashboardTemplate
	}
}
