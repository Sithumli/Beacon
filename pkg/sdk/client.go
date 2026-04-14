package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides a high-level interface to the A2A platform
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new client connected to the A2A platform
func NewClient(serverAddr string) (*Client, error) {
	// Ensure the address has a scheme
	baseURL := serverAddr
	if len(baseURL) > 0 && baseURL[0] != 'h' {
		baseURL = "http://" + baseURL
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Close closes the client (no-op for HTTP)
func (c *Client) Close() error {
	return nil
}

// RegisterAgent registers an agent with the platform
func (c *Client) RegisterAgent(ctx context.Context, config *AgentConfig) (*RegisteredAgent, error) {
	caps := make([]capabilityRequest, len(config.Capabilities))
	for i, cap := range config.Capabilities {
		caps[i] = capabilityRequest{
			Name:        cap.Name,
			Description: cap.Description,
		}
	}

	req := registerRequest{
		Name:        config.Name,
		Version:     config.Version,
		Description: config.Description,
		Endpoint: endpointRequest{
			Host:     config.Host,
			Port:     config.Port,
			Protocol: config.Protocol,
		},
		Capabilities: caps,
		Metadata: metadataRequest{
			Author: config.Author,
			Tags:   config.Tags,
		},
	}

	var resp registerResponse
	if err := c.post(ctx, "/api/v1/agents", req, &resp); err != nil {
		return nil, err
	}

	return &RegisteredAgent{
		ID:      resp.AgentID,
		Name:    config.Name,
		Version: config.Version,
		client:  c,
	}, nil
}

// DeregisterAgent removes an agent from the platform
func (c *Client) DeregisterAgent(ctx context.Context, agentID string) error {
	return c.delete(ctx, "/api/v1/agents/"+agentID)
}

// Heartbeat sends a heartbeat for an agent
func (c *Client) Heartbeat(ctx context.Context, agentID string) error {
	req := heartbeatRequest{AgentID: agentID}
	var resp heartbeatResponse
	return c.post(ctx, "/api/v1/heartbeat", req, &resp)
}

// DiscoverAgents finds agents with a specific capability
func (c *Client) DiscoverAgents(ctx context.Context, capability string) ([]*AgentInfo, error) {
	var agents []*agentResponse
	if err := c.get(ctx, "/api/v1/discover?capability="+capability, &agents); err != nil {
		return nil, err
	}

	result := make([]*AgentInfo, len(agents))
	for i, a := range agents {
		result[i] = responseToAgentInfo(a)
	}
	return result, nil
}

// ListAgents returns all registered agents
func (c *Client) ListAgents(ctx context.Context) ([]*AgentInfo, error) {
	var agents []*agentResponse
	if err := c.get(ctx, "/api/v1/agents", &agents); err != nil {
		return nil, err
	}

	result := make([]*AgentInfo, len(agents))
	for i, a := range agents {
		result[i] = responseToAgentInfo(a)
	}
	return result, nil
}

// GetAgent retrieves an agent by ID
func (c *Client) GetAgent(ctx context.Context, agentID string) (*AgentInfo, error) {
	var agent agentResponse
	if err := c.get(ctx, "/api/v1/agents/"+agentID, &agent); err != nil {
		return nil, err
	}
	return responseToAgentInfo(&agent), nil
}

// SendTask sends a task to a specific agent
func (c *Client) SendTask(ctx context.Context, fromAgent, toAgent, capability string, payload interface{}) (*TaskInfo, error) {
	req := sendTaskRequest{
		FromAgent:  fromAgent,
		ToAgent:    toAgent,
		Capability: capability,
		Payload:    payload,
	}

	var resp sendTaskResponse
	if err := c.post(ctx, "/api/v1/tasks", req, &resp); err != nil {
		return nil, err
	}

	return responseToTaskInfo(resp.Task), nil
}

// RouteTask routes a task to any available agent with the capability
func (c *Client) RouteTask(ctx context.Context, fromAgent, capability string, payload interface{}) (*TaskInfo, error) {
	req := routeTaskRequest{
		FromAgent:  fromAgent,
		Capability: capability,
		Payload:    payload,
	}

	var resp routeTaskResponse
	if err := c.post(ctx, "/api/v1/route", req, &resp); err != nil {
		return nil, err
	}

	return responseToTaskInfo(resp.Task), nil
}

// GetTask retrieves a task by ID
func (c *Client) GetTask(ctx context.Context, taskID string) (*TaskInfo, error) {
	var task taskResponse
	if err := c.get(ctx, "/api/v1/tasks/"+taskID, &task); err != nil {
		return nil, err
	}
	return responseToTaskInfo(&task), nil
}

// GetPendingTasks retrieves pending tasks for an agent
func (c *Client) GetPendingTasks(ctx context.Context, agentID string) ([]*TaskInfo, error) {
	var tasks []*taskResponse
	if err := c.get(ctx, "/api/v1/pending?agent_id="+agentID, &tasks); err != nil {
		return nil, err
	}

	result := make([]*TaskInfo, len(tasks))
	for i, t := range tasks {
		result[i] = responseToTaskInfo(t)
	}
	return result, nil
}

// UpdateTask updates a task's status
func (c *Client) UpdateTask(ctx context.Context, taskID string, status TaskStatus, result interface{}, errMsg string) (*TaskInfo, error) {
	req := updateTaskRequest{
		Status: string(status),
		Result: result,
		Error:  errMsg,
	}

	var task taskResponse
	if err := c.patch(ctx, "/api/v1/tasks/"+taskID, req, &task); err != nil {
		return nil, err
	}

	return responseToTaskInfo(&task), nil
}

// WaitForTask polls until a task is complete
func (c *Client) WaitForTask(ctx context.Context, taskID string, pollInterval time.Duration) (*TaskInfo, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			task, err := c.GetTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if task.IsFinal() {
				return task, nil
			}
		}
	}
}

// HTTP helper methods

func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, result)
}

func (c *Client) patch(ctx context.Context, path string, body, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doRequest(req, result)
}

func (c *Client) delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.baseURL+path, nil)
	if err != nil {
		return err
	}
	return c.doRequest(req, nil)
}

func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil && len(body) > 0 {
		return json.Unmarshal(body, result)
	}
	return nil
}

// Request/Response types for HTTP API

type registerRequest struct {
	Name         string              `json:"name"`
	Version      string              `json:"version"`
	Description  string              `json:"description"`
	Endpoint     endpointRequest     `json:"endpoint"`
	Capabilities []capabilityRequest `json:"capabilities"`
	Metadata     metadataRequest     `json:"metadata"`
}

type endpointRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type capabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type metadataRequest struct {
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

type registerResponse struct {
	AgentID string         `json:"agent_id"`
	Agent   *agentResponse `json:"agent"`
}

type agentResponse struct {
	ID            string               `json:"agent_id"`
	Name          string               `json:"name"`
	Version       string               `json:"version"`
	Description   string               `json:"description"`
	Endpoint      endpointResponse     `json:"endpoint"`
	Capabilities  []capabilityResponse `json:"capabilities"`
	Metadata      metadataResponse     `json:"metadata"`
	Status        string               `json:"status"`
	RegisteredAt  string               `json:"registered_at"`
	LastHeartbeat string               `json:"last_heartbeat"`
}

type endpointResponse struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

type capabilityResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type metadataResponse struct {
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

type heartbeatRequest struct {
	AgentID string `json:"agent_id"`
}

type heartbeatResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

type sendTaskRequest struct {
	FromAgent  string      `json:"from_agent"`
	ToAgent    string      `json:"to_agent"`
	Capability string      `json:"capability"`
	Payload    interface{} `json:"payload"`
}

type sendTaskResponse struct {
	TaskID string        `json:"task_id"`
	Task   *taskResponse `json:"task"`
}

type routeTaskRequest struct {
	FromAgent  string      `json:"from_agent"`
	Capability string      `json:"capability"`
	Payload    interface{} `json:"payload"`
}

type routeTaskResponse struct {
	TaskID  string        `json:"task_id"`
	ToAgent string        `json:"to_agent"`
	Task    *taskResponse `json:"task"`
}

type updateTaskRequest struct {
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type taskResponse struct {
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

func responseToAgentInfo(a *agentResponse) *AgentInfo {
	caps := make([]CapabilityInfo, len(a.Capabilities))
	for i, cap := range a.Capabilities {
		caps[i] = CapabilityInfo{
			Name:        cap.Name,
			Description: cap.Description,
		}
	}

	registeredAt, _ := time.Parse(time.RFC3339, a.RegisteredAt)
	lastHeartbeat, _ := time.Parse(time.RFC3339, a.LastHeartbeat)

	return &AgentInfo{
		ID:            a.ID,
		Name:          a.Name,
		Version:       a.Version,
		Description:   a.Description,
		Host:          a.Endpoint.Host,
		Port:          a.Endpoint.Port,
		Status:        AgentStatus(a.Status),
		Capabilities:  caps,
		RegisteredAt:  registeredAt,
		LastHeartbeat: lastHeartbeat,
	}
}

func responseToTaskInfo(t *taskResponse) *TaskInfo {
	createdAt, _ := time.Parse(time.RFC3339, t.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, t.UpdatedAt)

	info := &TaskInfo{
		ID:         t.ID,
		FromAgent:  t.FromAgent,
		ToAgent:    t.ToAgent,
		Capability: t.Capability,
		Status:     TaskStatus(t.Status),
		Error:      t.Error,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	if t.Payload != nil {
		info.Payload, _ = json.Marshal(t.Payload)
	}
	if t.Result != nil {
		info.Result, _ = json.Marshal(t.Result)
	}
	if t.CompletedAt != nil {
		completedAt, _ := time.Parse(time.RFC3339, *t.CompletedAt)
		info.CompletedAt = &completedAt
	}

	return info
}
