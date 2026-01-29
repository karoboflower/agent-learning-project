package scheduler

import (
	"fmt"
	"sort"
	"sync"
)

// AllocationStrategy 分配策略
type AllocationStrategy string

const (
	StrategyCapability   AllocationStrategy = "CAPABILITY"    // 基于能力
	StrategyLoadBalance  AllocationStrategy = "LOAD_BALANCE"  // 负载均衡
	StrategyPriority     AllocationStrategy = "PRIORITY"      // 基于优先级
	StrategyRoundRobin   AllocationStrategy = "ROUND_ROBIN"   // 轮询
)

// Task 任务定义
type Task struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Priority        int                    `json:"priority"`
	RequiredCapabilities []string          `json:"required_capabilities"`
	AssignedAgentID string                 `json:"assigned_agent_id,omitempty"`
	Status          string                 `json:"status"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TaskAllocator 任务分配器
type TaskAllocator struct {
	registry *AgentRegistry
	strategy AllocationStrategy
	mu       sync.RWMutex

	// 轮询计数器（用于round-robin）
	roundRobinIndex int
}

// NewTaskAllocator 创建任务分配器
func NewTaskAllocator(registry *AgentRegistry, strategy AllocationStrategy) *TaskAllocator {
	return &TaskAllocator{
		registry:        registry,
		strategy:        strategy,
		roundRobinIndex: 0,
	}
}

// SetStrategy 设置分配策略
func (a *TaskAllocator) SetStrategy(strategy AllocationStrategy) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.strategy = strategy
}

// GetStrategy 获取当前策略
func (a *TaskAllocator) GetStrategy() AllocationStrategy {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.strategy
}

// Allocate 分配任务
func (a *TaskAllocator) Allocate(task *Task) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if task == nil {
		return "", fmt.Errorf("task cannot be nil")
	}

	switch a.strategy {
	case StrategyCapability:
		return a.allocateByCapability(task)
	case StrategyLoadBalance:
		return a.allocateByLoadBalance(task)
	case StrategyPriority:
		return a.allocateByPriority(task)
	case StrategyRoundRobin:
		return a.allocateByRoundRobin(task)
	default:
		return "", fmt.Errorf("unknown allocation strategy: %s", a.strategy)
	}
}

// allocateByCapability 基于能力分配
func (a *TaskAllocator) allocateByCapability(task *Task) (string, error) {
	if len(task.RequiredCapabilities) == 0 {
		// 没有特定能力要求，选择任意可用Agent
		agents := a.registry.FindAvailableAgents()
		if len(agents) == 0 {
			return "", fmt.Errorf("no available agents")
		}
		return agents[0].ID, nil
	}

	// 查找具有所有必需能力的Agent
	candidateAgents := a.findAgentsWithCapabilities(task.RequiredCapabilities)
	if len(candidateAgents) == 0 {
		return "", fmt.Errorf("no agents with required capabilities: %v", task.RequiredCapabilities)
	}

	// 在候选Agent中选择负载最低的
	return a.selectAgentByLoad(candidateAgents), nil
}

// allocateByLoadBalance 负载均衡分配
func (a *TaskAllocator) allocateByLoadBalance(task *Task) (string, error) {
	agents := a.registry.FindAvailableAgents()
	if len(agents) == 0 {
		return "", fmt.Errorf("no available agents")
	}

	// 如果有能力要求，先过滤
	if len(task.RequiredCapabilities) > 0 {
		agents = a.filterByCapabilities(agents, task.RequiredCapabilities)
		if len(agents) == 0 {
			return "", fmt.Errorf("no agents with required capabilities")
		}
	}

	// 选择负载最低的Agent
	return a.selectAgentByLoad(agents), nil
}

// allocateByPriority 基于优先级分配
func (a *TaskAllocator) allocateByPriority(task *Task) (string, error) {
	agents := a.registry.FindAvailableAgents()
	if len(agents) == 0 {
		return "", fmt.Errorf("no available agents")
	}

	// 如果有能力要求，先过滤
	if len(task.RequiredCapabilities) > 0 {
		agents = a.filterByCapabilities(agents, task.RequiredCapabilities)
		if len(agents) == 0 {
			return "", fmt.Errorf("no agents with required capabilities")
		}
	}

	// 高优先级任务分配给负载最低的Agent
	// 低优先级任务可以分配给负载较高的Agent
	if task.Priority >= 8 {
		return a.selectAgentByLoad(agents), nil
	}

	// 中低优先级任务选择负载适中的Agent
	return a.selectAgentByBalancedLoad(agents), nil
}

// allocateByRoundRobin 轮询分配
func (a *TaskAllocator) allocateByRoundRobin(task *Task) (string, error) {
	agents := a.registry.FindAvailableAgents()
	if len(agents) == 0 {
		return "", fmt.Errorf("no available agents")
	}

	// 如果有能力要求，先过滤
	if len(task.RequiredCapabilities) > 0 {
		agents = a.filterByCapabilities(agents, task.RequiredCapabilities)
		if len(agents) == 0 {
			return "", fmt.Errorf("no agents with required capabilities")
		}
	}

	// 轮询选择
	agent := agents[a.roundRobinIndex%len(agents)]
	a.roundRobinIndex++

	return agent.ID, nil
}

// BatchAllocate 批量分配任务
func (a *TaskAllocator) BatchAllocate(tasks []*Task) (map[string]string, []error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	allocations := make(map[string]string)
	errors := make([]error, 0)

	// 按优先级排序
	sortedTasks := make([]*Task, len(tasks))
	copy(sortedTasks, tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].Priority > sortedTasks[j].Priority
	})

	for _, task := range sortedTasks {
		// 临时释放锁以避免死锁
		a.mu.Unlock()
		agentID, err := a.Allocate(task)
		a.mu.Lock()

		if err != nil {
			errors = append(errors, fmt.Errorf("failed to allocate task %s: %w", task.ID, err))
			continue
		}

		allocations[task.ID] = agentID
	}

	return allocations, errors
}

// 辅助方法

// findAgentsWithCapabilities 查找具有所有必需能力的Agent
func (a *TaskAllocator) findAgentsWithCapabilities(capabilities []string) []*Agent {
	agents := a.registry.FindAvailableAgents()
	result := make([]*Agent, 0)

	for _, agent := range agents {
		if hasAllCapabilities(agent, capabilities) {
			result = append(result, agent)
		}
	}

	return result
}

// filterByCapabilities 按能力过滤Agent
func (a *TaskAllocator) filterByCapabilities(agents []*Agent, capabilities []string) []*Agent {
	result := make([]*Agent, 0)

	for _, agent := range agents {
		if hasAllCapabilities(agent, capabilities) {
			result = append(result, agent)
		}
	}

	return result
}

// selectAgentByLoad 选择负载最低的Agent
func (a *TaskAllocator) selectAgentByLoad(agents []*Agent) string {
	if len(agents) == 0 {
		return ""
	}

	minLoadAgent := agents[0]
	minLoad := calculateLoad(agents[0])

	for _, agent := range agents[1:] {
		load := calculateLoad(agent)
		if load < minLoad {
			minLoad = load
			minLoadAgent = agent
		}
	}

	return minLoadAgent.ID
}

// selectAgentByBalancedLoad 选择负载适中的Agent（避免总是选择最空闲的）
func (a *TaskAllocator) selectAgentByBalancedLoad(agents []*Agent) string {
	if len(agents) == 0 {
		return ""
	}

	// 计算所有Agent的平均负载
	totalLoad := 0.0
	for _, agent := range agents {
		totalLoad += calculateLoad(agent)
	}
	avgLoad := totalLoad / float64(len(agents))

	// 选择负载最接近平均值且低于平均值的Agent
	var selectedAgent *Agent
	minDiff := 100.0

	for _, agent := range agents {
		load := calculateLoad(agent)
		if load <= avgLoad {
			diff := avgLoad - load
			if diff < minDiff {
				minDiff = diff
				selectedAgent = agent
			}
		}
	}

	if selectedAgent == nil {
		// 如果没有低于平均负载的，选择负载最���的
		return a.selectAgentByLoad(agents)
	}

	return selectedAgent.ID
}

// hasAllCapabilities 检查Agent是否具有所有必需能力
func hasAllCapabilities(agent *Agent, capabilities []string) bool {
	agentCapMap := make(map[string]bool)
	for _, cap := range agent.Capabilities {
		agentCapMap[cap] = true
	}

	for _, cap := range capabilities {
		if !agentCapMap[cap] {
			return false
		}
	}

	return true
}

// calculateLoad 计算Agent负载
func calculateLoad(agent *Agent) float64 {
	// 综合考虑显式负载和任务数量
	taskLoad := float64(agent.CurrentTasks) / float64(agent.MaxTasks)
	return (agent.Load + taskLoad) / 2.0
}
