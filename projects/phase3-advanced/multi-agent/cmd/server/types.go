package main

import (
	"time"

	"github.com/agent-learning/multi-agent/internal/scheduler"
)

// WebTask Web界面使用的任务结构
type WebTask struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Priority    int       `json:"priority"`
	Description string    `json:"description"`
	Capabilities []string `json:"capabilities"`
	Status      string    `json:"status"`
	AssignedTo  string    `json:"assigned_to"`
	Progress    int       `json:"progress"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToSchedulerTask 转换为scheduler.Task
func (wt *WebTask) ToSchedulerTask() *scheduler.Task {
	return &scheduler.Task{
		ID:                  wt.ID,
		Type:                wt.Type,
		Priority:            wt.Priority,
		RequiredCapabilities: wt.Capabilities,
		AssignedAgentID:     wt.AssignedTo,
		Status:              wt.Status,
		Metadata: map[string]interface{}{
			"description": wt.Description,
			"progress":    wt.Progress,
			"created_at":  wt.CreatedAt.Format(time.RFC3339),
		},
	}
}

// FromSchedulerTask 从scheduler.Task转换
func FromSchedulerTask(st *scheduler.Task) *WebTask {
	wt := &WebTask{
		ID:           st.ID,
		Type:         st.Type,
		Priority:     st.Priority,
		Capabilities: st.RequiredCapabilities,
		Status:       st.Status,
		AssignedTo:   st.AssignedAgentID,
	}

	if desc, ok := st.Metadata["description"].(string); ok {
		wt.Description = desc
	}
	if prog, ok := st.Metadata["progress"].(int); ok {
		wt.Progress = prog
	}
	if createdStr, ok := st.Metadata["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdStr); err == nil {
			wt.CreatedAt = t
		}
	}

	return wt
}
