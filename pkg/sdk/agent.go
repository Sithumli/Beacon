package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AgentStatus represents agent health status
type AgentStatus string

const (
	StatusHealthy   AgentStatus = "healthy"
	StatusUnhealthy AgentStatus = "unhealthy"
	StatusUnknown   AgentStatus = "unknown"
)

// Schema represents a JSON schema
type Schema struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
}

// CapabilityConfig defines a capability for registration
type CapabilityConfig struct {
	Name         string
	Description  string
	InputSchema  Schema
	OutputSchema Schema
	Handler      TaskHandler
}

// AgentConfig holds configuration for agent registration
type AgentConfig struct {
	Name         string
	Version      string
	Description  string
	Host         string
	Port         int
	Protocol     string
	Author       string
	Tags         []string
	Capabilities []CapabilityConfig
}

// AgentInfo contains information about a registered agent
type AgentInfo struct {
	ID            string
	Name          string
	Version       string
	Description   string
	Host          string
	Port          int
	Status        AgentStatus
	Capabilities  []CapabilityInfo
	RegisteredAt  time.Time
	LastHeartbeat time.Time
}

// CapabilityInfo contains information about a capability
type CapabilityInfo struct {
	Name        string
	Description string
}

// TaskHandler is a function that handles a task
type TaskHandler func(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)

// RegisteredAgent represents an agent registered with the platform
type RegisteredAgent struct {
	ID      string
	Name    string
	Version string
	client  *Client
}

// Agent is a builder for creating and running agents
type Agent struct {
	config       AgentConfig
	handlers     map[string]TaskHandler
	client       *Client
	serverAddr   string
	registered   *RegisteredAgent
	httpServer   *http.Server
	mu           sync.Mutex
	running      bool
	stopCh       chan struct{}
	heartbeatTTL time.Duration
}

// AgentBuilder helps build an agent with a fluent API
type AgentBuilder struct {
	config   AgentConfig
	handlers map[string]TaskHandler
}

// NewAgent creates a new agent builder
func NewAgent(name string) *AgentBuilder {
	return &AgentBuilder{
		config: AgentConfig{
			Name:     name,
			Version:  "1.0.0",
			Protocol: "http",
		},
		handlers: make(map[string]TaskHandler),
	}
}

// WithVersion sets the agent version
func (b *AgentBuilder) WithVersion(version string) *AgentBuilder {
	b.config.Version = version
	return b
}

// WithDescription sets the agent description
func (b *AgentBuilder) WithDescription(description string) *AgentBuilder {
	b.config.Description = description
	return b
}

// WithEndpoint sets the agent endpoint
func (b *AgentBuilder) WithEndpoint(host string, port int) *AgentBuilder {
	b.config.Host = host
	b.config.Port = port
	return b
}

// WithAuthor sets the agent author
func (b *AgentBuilder) WithAuthor(author string) *AgentBuilder {
	b.config.Author = author
	return b
}

// WithTags sets the agent tags
func (b *AgentBuilder) WithTags(tags ...string) *AgentBuilder {
	b.config.Tags = tags
	return b
}

// WithCapability adds a capability with a handler
func (b *AgentBuilder) WithCapability(name, description string, handler TaskHandler) *AgentBuilder {
	cap := CapabilityConfig{
		Name:        name,
		Description: description,
		Handler:     handler,
	}
	b.config.Capabilities = append(b.config.Capabilities, cap)
	b.handlers[name] = handler
	return b
}

// WithCapabilitySchema adds a capability with schema and handler
func (b *AgentBuilder) WithCapabilitySchema(name, description string, input, output Schema, handler TaskHandler) *AgentBuilder {
	cap := CapabilityConfig{
		Name:         name,
		Description:  description,
		InputSchema:  input,
		OutputSchema: output,
		Handler:      handler,
	}
	b.config.Capabilities = append(b.config.Capabilities, cap)
	b.handlers[name] = handler
	return b
}

// Build creates the agent
func (b *AgentBuilder) Build() *Agent {
	return &Agent{
		config:       b.config,
		handlers:     b.handlers,
		stopCh:       make(chan struct{}),
		heartbeatTTL: 10 * time.Second,
	}
}

// Register registers the agent with the platform
func (a *Agent) Register(serverAddr string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	client, err := NewClient(serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	a.client = client
	a.serverAddr = serverAddr

	ctx := context.Background()
	registered, err := client.RegisterAgent(ctx, &a.config)
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to register: %w", err)
	}

	a.registered = registered
	return nil
}

// Start starts the agent and begins processing tasks
func (a *Agent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent already running")
	}
	a.running = true
	a.mu.Unlock()

	// Start HTTP server for receiving tasks
	if err := a.startHTTPServer(); err != nil {
		return err
	}

	// Start heartbeat goroutine
	go a.heartbeatLoop(ctx)

	// Start task polling (since we're using HTTP, we poll for tasks)
	go a.pollTasks(ctx)

	// Wait for context cancellation or stop signal
	select {
	case <-ctx.Done():
	case <-a.stopCh:
	}

	return a.cleanup()
}

// Stop stops the agent gracefully
func (a *Agent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return
	}

	close(a.stopCh)
}

func (a *Agent) startHTTPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/execute", a.handleExecute)
	mux.HandleFunc("/health", a.handleHealth)

	addr := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
	a.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error but don't crash
		}
	}()

	return nil
}

func (a *Agent) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TaskID     string          `json:"task_id"`
		Capability string          `json:"capability"`
		Payload    json.RawMessage `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	handler, ok := a.handlers[req.Capability]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "unknown capability: " + req.Capability,
		})
		return
	}

	result, err := handler(r.Context(), req.Payload)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  json.RawMessage(result),
	})
}

func (a *Agent) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"healthy": true})
}

func (a *Agent) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(a.heartbeatTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			if a.client != nil && a.registered != nil {
				a.client.Heartbeat(ctx, a.registered.ID)
			}
		}
	}
}

func (a *Agent) pollTasks(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.processPendingTasks(ctx)
		}
	}
}

func (a *Agent) processPendingTasks(ctx context.Context) {
	if a.client == nil || a.registered == nil {
		return
	}

	// Get pending tasks for this agent
	// Note: In a production system, you'd want a dedicated endpoint for this
	// For now, we rely on the server pushing tasks or the task being fetched individually
}

func (a *Agent) cleanup() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.running = false

	if a.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.httpServer.Shutdown(ctx)
	}

	if a.client != nil && a.registered != nil {
		ctx := context.Background()
		a.client.DeregisterAgent(ctx, a.registered.ID)
		a.client.Close()
	}

	return nil
}

// GetID returns the registered agent ID
func (a *Agent) GetID() string {
	if a.registered != nil {
		return a.registered.ID
	}
	return ""
}
