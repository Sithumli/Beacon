package registry

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Sithumli/Beacon/internal/core"
)

// HTTPHandler provides HTTP endpoints for the registry service
type HTTPHandler struct {
	service *Service
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{service: svc}
}

// RegisterRoutes registers HTTP routes on the given mux
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/agents", h.handleAgents)
	mux.HandleFunc("/api/v1/agents/", h.handleAgent)
	mux.HandleFunc("/api/v1/discover", h.handleDiscover)
	mux.HandleFunc("/api/v1/heartbeat", h.handleHeartbeat)
}

func (h *HTTPHandler) handleAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		// List agents
		agents, err := h.service.ListAgents(ctx, nil)
		if err != nil {
			jsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse(w, agents)

	case http.MethodPost:
		// Register agent
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, "invalid request body", http.StatusBadRequest)
			return
		}

		agent := &core.Agent{
			Name:        req.Name,
			Version:     req.Version,
			Description: req.Description,
			Endpoint: core.Endpoint{
				Host:     req.Endpoint.Host,
				Port:     req.Endpoint.Port,
				Protocol: req.Endpoint.Protocol,
			},
			Capabilities: make([]core.Capability, len(req.Capabilities)),
			Metadata: core.Metadata{
				Author: req.Metadata.Author,
				Tags:   req.Metadata.Tags,
			},
		}

		for i, cap := range req.Capabilities {
			agent.Capabilities[i] = core.Capability{
				Name:        cap.Name,
				Description: cap.Description,
			}
		}

		registered, err := h.service.Register(ctx, agent)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		jsonResponse(w, RegisterResponse{
			AgentID: registered.ID,
			Agent:   agentToResponse(registered),
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) handleAgent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := strings.TrimPrefix(r.URL.Path, "/api/v1/agents/")

	if agentID == "" {
		jsonError(w, "agent ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		agent, err := h.service.GetAgent(ctx, agentID)
		if err != nil {
			jsonError(w, "agent not found", http.StatusNotFound)
			return
		}
		jsonResponse(w, agentToResponse(agent))

	case http.MethodDelete:
		if err := h.service.Deregister(ctx, agentID); err != nil {
			jsonError(w, "agent not found", http.StatusNotFound)
			return
		}
		jsonResponse(w, map[string]bool{"success": true})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *HTTPHandler) handleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	capability := r.URL.Query().Get("capability")
	if capability == "" {
		jsonError(w, "capability parameter required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	agents, err := h.service.Discover(ctx, capability)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]*AgentResponse, len(agents))
	for i, agent := range agents {
		responses[i] = agentToResponse(agent)
	}
	jsonResponse(w, responses)
}

func (h *HTTPHandler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.service.Heartbeat(ctx, req.AgentID); err != nil {
		jsonError(w, "agent not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, HeartbeatResponse{
		Success: true,
		Status:  "healthy",
	})
}

// Request/Response types

type RegisterRequest struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Description  string              `json:"description"`
	Endpoint     EndpointRequest     `json:"endpoint"`
	Capabilities []CapabilityRequest `json:"capabilities"`
	Metadata     MetadataRequest     `json:"metadata"`
}

type EndpointRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type CapabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MetadataRequest struct {
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

type RegisterResponse struct {
	AgentID string         `json:"agent_id"`
	Agent   *AgentResponse `json:"agent"`
}

type AgentResponse struct {
	ID            string               `json:"agent_id"`
	Name          string               `json:"name"`
	Version       string               `json:"version"`
	Description   string               `json:"description"`
	Endpoint      EndpointResponse     `json:"endpoint"`
	Capabilities  []CapabilityResponse `json:"capabilities"`
	Metadata      MetadataResponse     `json:"metadata"`
	Status        string               `json:"status"`
	RegisteredAt  string               `json:"registered_at"`
	LastHeartbeat string               `json:"last_heartbeat"`
}

type EndpointResponse struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type CapabilityResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MetadataResponse struct {
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

type HeartbeatRequest struct {
	AgentID string `json:"agent_id"`
}

type HeartbeatResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

func agentToResponse(a *core.Agent) *AgentResponse {
	caps := make([]CapabilityResponse, len(a.Capabilities))
	for i, cap := range a.Capabilities {
		caps[i] = CapabilityResponse{
			Name:        cap.Name,
			Description: cap.Description,
		}
	}

	return &AgentResponse{
		ID:          a.ID,
		Name:        a.Name,
		Version:     a.Version,
		Description: a.Description,
		Endpoint: EndpointResponse{
			Host:     a.Endpoint.Host,
			Port:     a.Endpoint.Port,
			Protocol: a.Endpoint.Protocol,
		},
		Capabilities: caps,
		Metadata: MetadataResponse{
			Author: a.Metadata.Author,
			Tags:   a.Metadata.Tags,
		},
		Status:        string(a.Status),
		RegisteredAt:  a.RegisteredAt.Format("2006-01-02T15:04:05Z07:00"),
		LastHeartbeat: a.LastHeartbeat.Format("2006-01-02T15:04:05Z07:00"),
	}
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
