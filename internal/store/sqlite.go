package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Sithumli/Beacon/internal/core"
	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store interface with SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT,
		endpoint_host TEXT NOT NULL,
		endpoint_port INTEGER NOT NULL,
		endpoint_protocol TEXT DEFAULT 'grpc',
		capabilities TEXT NOT NULL,
		metadata TEXT,
		status TEXT DEFAULT 'healthy',
		registered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_heartbeat DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		from_agent TEXT NOT NULL,
		to_agent TEXT NOT NULL,
		capability TEXT NOT NULL,
		payload TEXT NOT NULL,
		status TEXT DEFAULT 'pending',
		result TEXT,
		error TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_to_agent ON tasks(to_agent);
	CREATE INDEX IF NOT EXISTS idx_tasks_from_agent ON tasks(from_agent);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// CreateAgent stores a new agent
func (s *SQLiteStore) CreateAgent(ctx context.Context, agent *core.Agent) error {
	caps, err := json.Marshal(agent.Capabilities)
	if err != nil {
		return err
	}

	meta, err := json.Marshal(agent.Metadata)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO agents (id, name, version, description, endpoint_host, endpoint_port,
			endpoint_protocol, capabilities, metadata, status, registered_at, last_heartbeat)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, agent.ID, agent.Name, agent.Version, agent.Description,
		agent.Endpoint.Host, agent.Endpoint.Port, agent.Endpoint.Protocol,
		string(caps), string(meta), string(agent.Status),
		agent.RegisteredAt, agent.LastHeartbeat)

	if err != nil {
		// Check for uniqueness/constraint violation (SQLITE_CONSTRAINT = 19)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to create agent: %w", err)
	}
	return nil
}

// GetAgent retrieves an agent by ID
func (s *SQLiteStore) GetAgent(ctx context.Context, id string) (*core.Agent, error) {
	var agent core.Agent
	var caps, meta string
	var status string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, version, description, endpoint_host, endpoint_port,
			endpoint_protocol, capabilities, metadata, status, registered_at, last_heartbeat
		FROM agents WHERE id = ?
	`, id).Scan(
		&agent.ID, &agent.Name, &agent.Version, &agent.Description,
		&agent.Endpoint.Host, &agent.Endpoint.Port, &agent.Endpoint.Protocol,
		&caps, &meta, &status, &agent.RegisteredAt, &agent.LastHeartbeat)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(caps), &agent.Capabilities); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(meta), &agent.Metadata); err != nil {
		return nil, err
	}
	agent.Status = core.AgentStatus(status)

	return &agent, nil
}

// UpdateAgent updates an existing agent
func (s *SQLiteStore) UpdateAgent(ctx context.Context, agent *core.Agent) error {
	caps, err := json.Marshal(agent.Capabilities)
	if err != nil {
		return err
	}

	meta, err := json.Marshal(agent.Metadata)
	if err != nil {
		return err
	}

	result, err := s.db.ExecContext(ctx, `
		UPDATE agents SET name = ?, version = ?, description = ?,
			endpoint_host = ?, endpoint_port = ?, endpoint_protocol = ?,
			capabilities = ?, metadata = ?, status = ?, last_heartbeat = ?
		WHERE id = ?
	`, agent.Name, agent.Version, agent.Description,
		agent.Endpoint.Host, agent.Endpoint.Port, agent.Endpoint.Protocol,
		string(caps), string(meta), string(agent.Status), agent.LastHeartbeat,
		agent.ID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteAgent removes an agent by ID
func (s *SQLiteStore) DeleteAgent(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListAgents returns all agents matching the filter
func (s *SQLiteStore) ListAgents(ctx context.Context, filter *AgentFilter) ([]*core.Agent, error) {
	query := "SELECT id, name, version, description, endpoint_host, endpoint_port, endpoint_protocol, capabilities, metadata, status, registered_at, last_heartbeat FROM agents WHERE 1=1"
	args := make([]interface{}, 0)

	if filter != nil {
		if filter.Status != nil {
			query += " AND status = ?"
			args = append(args, string(*filter.Status))
		}
		// Filter by tags using JSON extraction
		for _, tag := range filter.Tags {
			query += " AND metadata LIKE ?"
			args = append(args, "%"+tag+"%")
		}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := make([]*core.Agent, 0)
	for rows.Next() {
		var agent core.Agent
		var caps, meta string
		var status string

		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Version, &agent.Description,
			&agent.Endpoint.Host, &agent.Endpoint.Port, &agent.Endpoint.Protocol,
			&caps, &meta, &status, &agent.RegisteredAt, &agent.LastHeartbeat)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(caps), &agent.Capabilities); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(meta), &agent.Metadata); err != nil {
			return nil, err
		}
		agent.Status = core.AgentStatus(status)

		// Apply capability filter
		if filter != nil && filter.Capability != nil {
			if !agent.HasCapability(*filter.Capability) {
				continue
			}
		}

		agents = append(agents, &agent)
	}

	return agents, rows.Err()
}

// FindAgentsByCapability returns agents with the given capability
func (s *SQLiteStore) FindAgentsByCapability(ctx context.Context, capability string) ([]*core.Agent, error) {
	filter := &AgentFilter{Capability: &capability}
	return s.ListAgents(ctx, filter)
}

// CreateTask stores a new task
func (s *SQLiteStore) CreateTask(ctx context.Context, task *core.Task) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks (id, from_agent, to_agent, capability, payload, status,
			result, error, created_at, updated_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, task.ID, task.FromAgent, task.ToAgent, task.Capability,
		string(task.Payload), string(task.Status),
		nullableJSON(task.Result), nullableString(task.Error),
		task.CreatedAt, task.UpdatedAt, nullableTime(task.CompletedAt))

	if err != nil {
		// Check for uniqueness/constraint violation
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// GetTask retrieves a task by ID
func (s *SQLiteStore) GetTask(ctx context.Context, id string) (*core.Task, error) {
	var task core.Task
	var payload, status string
	var result, errMsg sql.NullString
	var completedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, from_agent, to_agent, capability, payload, status,
			result, error, created_at, updated_at, completed_at
		FROM tasks WHERE id = ?
	`, id).Scan(
		&task.ID, &task.FromAgent, &task.ToAgent, &task.Capability,
		&payload, &status, &result, &errMsg,
		&task.CreatedAt, &task.UpdatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	task.Payload = json.RawMessage(payload)
	task.Status = core.TaskStatus(status)
	if result.Valid {
		task.Result = json.RawMessage(result.String)
	}
	if errMsg.Valid {
		task.Error = errMsg.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// UpdateTask updates an existing task
func (s *SQLiteStore) UpdateTask(ctx context.Context, task *core.Task) error {
	task.UpdatedAt = time.Now().UTC()

	result, err := s.db.ExecContext(ctx, `
		UPDATE tasks SET status = ?, result = ?, error = ?, updated_at = ?, completed_at = ?
		WHERE id = ?
	`, string(task.Status), nullableJSON(task.Result), nullableString(task.Error),
		task.UpdatedAt, nullableTime(task.CompletedAt), task.ID)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteTask removes a task by ID
func (s *SQLiteStore) DeleteTask(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListTasks returns all tasks matching the filter
func (s *SQLiteStore) ListTasks(ctx context.Context, filter *TaskFilter) ([]*core.Task, error) {
	query := "SELECT id, from_agent, to_agent, capability, payload, status, result, error, created_at, updated_at, completed_at FROM tasks WHERE 1=1"
	args := make([]interface{}, 0)

	if filter != nil {
		if filter.Status != nil {
			query += " AND status = ?"
			args = append(args, string(*filter.Status))
		}
		if filter.FromAgent != nil {
			query += " AND from_agent = ?"
			args = append(args, *filter.FromAgent)
		}
		if filter.ToAgent != nil {
			query += " AND to_agent = ?"
			args = append(args, *filter.ToAgent)
		}
		if filter.Capability != nil {
			query += " AND capability = ?"
			args = append(args, *filter.Capability)
		}
	}

	query += " ORDER BY created_at DESC"

	if filter != nil {
		if filter.Limit != nil && *filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", *filter.Limit)
		}
		if filter.Offset != nil && *filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", *filter.Offset)
		}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*core.Task, 0)
	for rows.Next() {
		var task core.Task
		var payload, status string
		var result, errMsg sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.FromAgent, &task.ToAgent, &task.Capability,
			&payload, &status, &result, &errMsg,
			&task.CreatedAt, &task.UpdatedAt, &completedAt)
		if err != nil {
			return nil, err
		}

		task.Payload = json.RawMessage(payload)
		task.Status = core.TaskStatus(status)
		if result.Valid {
			task.Result = json.RawMessage(result.String)
		}
		if errMsg.Valid {
			task.Error = errMsg.String
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

// GetPendingTasksForAgent returns pending tasks for a specific agent
func (s *SQLiteStore) GetPendingTasksForAgent(ctx context.Context, agentID string) ([]*core.Task, error) {
	status := core.TaskPending
	filter := &TaskFilter{
		ToAgent: &agentID,
		Status:  &status,
	}
	return s.ListTasks(ctx, filter)
}

// Helper functions

func nullableJSON(data json.RawMessage) interface{} {
	if len(data) == 0 {
		return nil
	}
	return string(data)
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

// Ensure SQLiteStore implements Store
var _ Store = (*SQLiteStore)(nil)
