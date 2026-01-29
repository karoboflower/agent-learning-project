package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// AgentStatus Agent状态
type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "IDLE"       // 空闲
	AgentStatusBusy       AgentStatus = "BUSY"       // 忙碌
	AgentStatusOffline    AgentStatus = "OFFLINE"    // 离线
	AgentStatusMaintenance AgentStatus = "MAINTENANCE" // 维护中
)

// Agent Agent信息
type Agent struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Capabilities []string               `json:"capabilities"` // Agent能力列表
	Status       AgentStatus            `json:"status"`
	Load         float64                `json:"load"`          // 负载 0-1
	MaxTasks     int                    `json:"max_tasks"`     // 最大任务数
	CurrentTasks int                    `json:"current_tasks"` // 当前任务数
	Metadata     map[string]interface{} `json:"metadata"`
	RegisteredAt time.Time              `json:"registered_at"`
	LastHeartbeat time.Time             `json:"last_heartbeat"`
}

// AgentRegistry Agent注册表
type AgentRegistry struct {
	agents map[string]*Agent
	mu     sync.RWMutex
}

// NewAgentRegistry 创建Agent注册表
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*Agent),
	}
}

// Register 注册Agent
func (r *AgentRegistry) Register(agent *Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent.ID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	if agent.Name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}

	if len(agent.Capabilities) == 0 {
		return fmt.Errorf("agent must have at least one capability")
	}

	// 设置默认值
	if agent.MaxTasks == 0 {
		agent.MaxTasks = 10
	}

	if agent.Status == "" {
		agent.Status = AgentStatusIdle
	}

	if agent.Metadata == nil {
		agent.Metadata = make(map[string]interface{})
	}

	agent.RegisteredAt = time.Now()
	agent.LastHeartbeat = time.Now()

	r.agents[agent.ID] = agent

	return nil
}

// Unregister 注销Agent
func (r *AgentRegistry) Unregister(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(r.agents, agentID)
	return nil
}

// GetAgent 获取Agent信息
func (r *AgentRegistry) GetAgent(agentID string) (*Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	return agent, nil
}

// ListAgents 列出所有Agent
func (r *AgentRegistry) ListAgents() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}

// FindAgentsByCapability 根据能力查找Agent
func (r *AgentRegistry) FindAgentsByCapability(capability string) []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0)
	for _, agent := range r.agents {
		if hasCapability(agent, capability) {
			agents = append(agents, agent)
		}
	}

	return agents
}

// FindAvailableAgents 查找可用的Agent
func (r *AgentRegistry) FindAvailableAgents() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0)
	for _, agent := range r.agents {
		if agent.Status == AgentStatusIdle && agent.CurrentTasks < agent.MaxTasks {
			agents = append(agents, agent)
		}
	}

	return agents
}

// UpdateAgentStatus 更新Agent状态
func (r *AgentRegistry) UpdateAgentStatus(agentID string, status AgentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.Status = status
	return nil
}

// UpdateAgentLoad 更新Agent负载
func (r *AgentRegistry) UpdateAgentLoad(agentID string, load float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if load < 0 || load > 1 {
		return fmt.Errorf("load must be between 0 and 1")
	}

	agent.Load = load
	return nil
}

// IncrementTaskCount 增加Agent任务计数
func (r *AgentRegistry) IncrementTaskCount(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if agent.CurrentTasks >= agent.MaxTasks {
		return fmt.Errorf("agent %s has reached max tasks", agentID)
	}

	agent.CurrentTasks++

	// 更新状态
	if agent.CurrentTasks >= agent.MaxTasks {
		agent.Status = AgentStatusBusy
	}

	return nil
}

// DecrementTaskCount 减少Agent任务计数
func (r *AgentRegistry) DecrementTaskCount(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	if agent.CurrentTasks > 0 {
		agent.CurrentTasks--
	}

	// 更新状态
	if agent.CurrentTasks < agent.MaxTasks && agent.Status == AgentStatusBusy {
		agent.Status = AgentStatusIdle
	}

	return nil
}

// UpdateHeartbeat 更新心跳时间
func (r *AgentRegistry) UpdateHeartbeat(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	agent.LastHeartbeat = time.Now()
	return nil
}

// CheckHeartbeat 检查心跳超时
func (r *AgentRegistry) CheckHeartbeat(timeout time.Duration) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	timeoutAgents := make([]string, 0)

	for _, agent := range r.agents {
		if now.Sub(agent.LastHeartbeat) > timeout {
			// 标记为离线
			agent.Status = AgentStatusOffline
			timeoutAgents = append(timeoutAgents, agent.ID)
		}
	}

	return timeoutAgents
}

// GetAgentCount 获取Agent数量
func (r *AgentRegistry) GetAgentCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.agents)
}

// GetAgentCountByStatus 按状态统计Agent数量
func (r *AgentRegistry) GetAgentCountByStatus() map[AgentStatus]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[AgentStatus]int)
	for _, agent := range r.agents {
		counts[agent.Status]++
	}

	return counts
}

// 辅助函数

func hasCapability(agent *Agent, capability string) bool {
	for _, cap := range agent.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}
