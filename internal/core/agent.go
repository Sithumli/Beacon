package core

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// AgentStatus represents the current health status of an agent
type AgentStatus string

const (
	StatusHealthy   AgentStatus = "healthy"
	StatusUnhealthy AgentStatus = "unhealthy"
	StatusUnknown   AgentStatus = "unknown"
)

// Endpoint represents how to connect to an agent
type Endpoint struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // grpc, http
}

// Validate checks if the endpoint is valid
func (e *Endpoint) Validate() error {
	if e.Host == "" {
		return errors.New("endpoint host is required")
	}
	if e.Port <= 0 || e.Port > 65535 {
		return errors.New("endpoint port must be between 1 and 65535")
	}
	if e.Protocol == "" {
		e.Protocol = "grpc"
	}
	if e.Protocol != "grpc" && e.Protocol != "http" {
		return errors.New("endpoint protocol must be 'grpc' or 'http'")
	}
	return nil
}

// Address returns the full address string
func (e *Endpoint) Address() string {
	return e.Host + ":" + strconv.Itoa(e.Port)
}

// Metadata contains optional agent metadata
type Metadata struct {
	Author string   `json:"author,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

// Agent represents a registered agent in the platform
type Agent struct {
	ID            string       `json:"agent_id"`
	Name          string       `json:"name"`
	Version       string       `json:"version"`
	Description   string       `json:"description"`
	Endpoint      Endpoint     `json:"endpoint"`
	Capabilities  []Capability `json:"capabilities"`
	Metadata      Metadata     `json:"metadata"`
	Status        AgentStatus  `json:"status"`
	RegisteredAt  time.Time    `json:"registered_at"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
}

// NewAgent creates a new agent with a generated ID
func NewAgent(name, version, description string) *Agent {
	now := time.Now().UTC()
	return &Agent{
		ID:            uuid.New().String(),
		Name:          name,
		Version:       version,
		Description:   description,
		Capabilities:  make([]Capability, 0),
		Status:        StatusHealthy,
		RegisteredAt:  now,
		LastHeartbeat: now,
	}
}

// Validate checks if the agent configuration is valid
func (a *Agent) Validate() error {
	if a.Name == "" {
		return errors.New("agent name is required")
	}
	if a.Version == "" {
		return errors.New("agent version is required")
	}
	if err := a.Endpoint.Validate(); err != nil {
		return err
	}
	for _, cap := range a.Capabilities {
		if err := cap.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// AddCapability adds a capability to the agent
func (a *Agent) AddCapability(cap Capability) error {
	if err := cap.Validate(); err != nil {
		return err
	}
	// Check for duplicates
	for _, existing := range a.Capabilities {
		if existing.Name == cap.Name {
			return errors.New("capability already exists: " + cap.Name)
		}
	}
	a.Capabilities = append(a.Capabilities, cap)
	return nil
}

// HasCapability checks if the agent has a specific capability
func (a *Agent) HasCapability(name string) bool {
	for _, cap := range a.Capabilities {
		if cap.Name == name {
			return true
		}
	}
	return false
}

// GetCapability returns a capability by name
func (a *Agent) GetCapability(name string) *Capability {
	for i := range a.Capabilities {
		if a.Capabilities[i].Name == name {
			return &a.Capabilities[i]
		}
	}
	return nil
}

// UpdateHeartbeat updates the last heartbeat time and status
func (a *Agent) UpdateHeartbeat() {
	a.LastHeartbeat = time.Now().UTC()
	a.Status = StatusHealthy
}

// IsExpired checks if the agent has exceeded the heartbeat TTL
func (a *Agent) IsExpired(ttl time.Duration) bool {
	return time.Since(a.LastHeartbeat) > ttl
}

// MarkUnhealthy marks the agent as unhealthy
func (a *Agent) MarkUnhealthy() {
	a.Status = StatusUnhealthy
}
