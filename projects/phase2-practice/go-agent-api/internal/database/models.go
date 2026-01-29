package database

import (
	"database/sql"
	"time"
)

// AgentRecord represents an agent record in the database
type AgentRecord struct {
	ID        string
	Name      string
	Type      string
	Status    string
	Config    string // JSON string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TaskRecord represents a task record in the database
type TaskRecord struct {
	ID        string
	AgentID   string
	Type      string
	Input     string
	Output    sql.NullString
	Status    string
	Priority  int
	Tools     string // JSON array
	Metadata  string // JSON object
	Error     sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
	StartedAt sql.NullTime
	EndedAt   sql.NullTime
}

// TaskResultRecord represents a task result record in the database
type TaskResultRecord struct {
	ID         int
	TaskID     string
	Status     string
	Output     sql.NullString
	Error      sql.NullString
	Metadata   string // JSON object
	DurationMs int64
	CreatedAt  time.Time
	EndedAt    time.Time
}
