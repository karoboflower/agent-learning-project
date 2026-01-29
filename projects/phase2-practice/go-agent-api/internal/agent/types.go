package agent

import (
	"time"
)

// AgentType defines the type of agent
type AgentType string

const (
	AgentTypeGeneral     AgentType = "general"
	AgentTypeCodeReview  AgentType = "code_review"
	AgentTypeDocQA       AgentType = "doc_qa"
	AgentTypeAPIHandler  AgentType = "api_handler"
)

// AgentStatus defines the current status of an agent
type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "idle"
	AgentStatusBusy       AgentStatus = "busy"
	AgentStatusError      AgentStatus = "error"
	AgentStatusTerminated AgentStatus = "terminated"
)

// AgentConfig holds configuration for an agent
type AgentConfig struct {
	Model       string                 `json:"model"`
	Temperature float32                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Tools       []string               `json:"tools"`
	Extra       map[string]interface{} `json:"extra"`
}

// Agent represents an agent instance
type Agent struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      AgentType    `json:"type"`
	Status    AgentStatus  `json:"status"`
	Config    AgentConfig  `json:"config"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Error     string       `json:"error,omitempty"`
}

// CreateAgentRequest represents a request to create an agent
type CreateAgentRequest struct {
	Name   string      `json:"name" binding:"required"`
	Type   AgentType   `json:"type" binding:"required"`
	Config AgentConfig `json:"config"`
}

// TaskType defines the type of task
type TaskType string

const (
	TaskTypeQuery       TaskType = "query"
	TaskTypeCodeReview  TaskType = "code_review"
	TaskTypeSearch      TaskType = "search"
	TaskTypeFileOps     TaskType = "file_ops"
	TaskTypeCustom      TaskType = "custom"
)

// TaskStatus defines the current status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task represents a task to be executed by an agent
type Task struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	Type      TaskType               `json:"type"`
	Input     string                 `json:"input"`
	Output    string                 `json:"output,omitempty"`
	Status    TaskStatus             `json:"status"`
	Priority  int                    `json:"priority"`
	Tools     []string               `json:"tools"`
	Metadata  map[string]interface{} `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	StartedAt *time.Time             `json:"started_at,omitempty"`
	EndedAt   *time.Time             `json:"ended_at,omitempty"`
}

// CreateTaskRequest represents a request to create a task
type CreateTaskRequest struct {
	AgentID  string                 `json:"agent_id" binding:"required"`
	Type     TaskType               `json:"type" binding:"required"`
	Input    string                 `json:"input" binding:"required"`
	Priority int                    `json:"priority"`
	Tools    []string               `json:"tools"`
	Metadata map[string]interface{} `json:"metadata"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	Status    TaskStatus             `json:"status"`
	Output    string                 `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Duration  int64                  `json:"duration_ms"`
	CreatedAt time.Time              `json:"created_at"`
	EndedAt   time.Time              `json:"ended_at"`
}
