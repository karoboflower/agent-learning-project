package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB wraps PostgreSQL database connection
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL connection
func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresDB{db: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// InitSchema initializes database schema
func (p *PostgresDB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		status VARCHAR(50) NOT NULL,
		config JSONB,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_agents_type ON agents(type);

	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(255) PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		input TEXT NOT NULL,
		output TEXT,
		status VARCHAR(50) NOT NULL,
		priority INTEGER DEFAULT 0,
		tools JSONB,
		metadata JSONB,
		error TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		started_at TIMESTAMP,
		ended_at TIMESTAMP,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_agent_id ON tasks(agent_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);

	CREATE TABLE IF NOT EXISTS task_results (
		id SERIAL PRIMARY KEY,
		task_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		output TEXT,
		error TEXT,
		metadata JSONB,
		duration_ms BIGINT,
		created_at TIMESTAMP NOT NULL,
		ended_at TIMESTAMP NOT NULL,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_task_results_task_id ON task_results(task_id);
	`

	_, err := p.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// SaveAgent saves an agent to the database
func (p *PostgresDB) SaveAgent(agent *AgentRecord) error {
	query := `
		INSERT INTO agents (id, name, type, status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			config = EXCLUDED.config,
			updated_at = EXCLUDED.updated_at
	`

	_, err := p.db.Exec(query,
		agent.ID, agent.Name, agent.Type, agent.Status,
		agent.Config, agent.CreatedAt, agent.UpdatedAt,
	)

	return err
}

// SaveTask saves a task to the database
func (p *PostgresDB) SaveTask(task *TaskRecord) error {
	query := `
		INSERT INTO tasks (id, agent_id, type, input, output, status, priority, tools, metadata, error, created_at, updated_at, started_at, ended_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			output = EXCLUDED.output,
			status = EXCLUDED.status,
			error = EXCLUDED.error,
			updated_at = EXCLUDED.updated_at,
			started_at = EXCLUDED.started_at,
			ended_at = EXCLUDED.ended_at
	`

	_, err := p.db.Exec(query,
		task.ID, task.AgentID, task.Type, task.Input, task.Output,
		task.Status, task.Priority, task.Tools, task.Metadata,
		task.Error, task.CreatedAt, task.UpdatedAt,
		task.StartedAt, task.EndedAt,
	)

	return err
}

// SaveTaskResult saves a task result to the database
func (p *PostgresDB) SaveTaskResult(result *TaskResultRecord) error {
	query := `
		INSERT INTO task_results (task_id, status, output, error, metadata, duration_ms, created_at, ended_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := p.db.Exec(query,
		result.TaskID, result.Status, result.Output, result.Error,
		result.Metadata, result.DurationMs, result.CreatedAt, result.EndedAt,
	)

	return err
}

// GetTaskHistory retrieves task history
func (p *PostgresDB) GetTaskHistory(limit int) ([]*TaskRecord, error) {
	query := `
		SELECT id, agent_id, type, input, output, status, priority, tools, metadata, error, created_at, updated_at, started_at, ended_at
		FROM tasks
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := p.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*TaskRecord, 0)
	for rows.Next() {
		var task TaskRecord
		err := rows.Scan(
			&task.ID, &task.AgentID, &task.Type, &task.Input, &task.Output,
			&task.Status, &task.Priority, &task.Tools, &task.Metadata,
			&task.Error, &task.CreatedAt, &task.UpdatedAt,
			&task.StartedAt, &task.EndedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}
