package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	MaxQueueSize       int                // 最大队列大小
	AllocationStrategy AllocationStrategy // 分配策略
	HeartbeatInterval  time.Duration      // 心跳检查间隔
	HeartbeatTimeout   time.Duration      // 心跳超时时间
	WorkerCount        int                // 工作协程数
}

// DefaultSchedulerConfig 默认配置
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		MaxQueueSize:       1000,
		AllocationStrategy: StrategyLoadBalance,
		HeartbeatInterval:  30 * time.Second,
		HeartbeatTimeout:   90 * time.Second,
		WorkerCount:        5,
	}
}

// Scheduler 任务调度器
type Scheduler struct {
	config    *SchedulerConfig
	registry  *AgentRegistry
	allocator *TaskAllocator
	queue     *TaskQueue
	manager   *TaskManager

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// NewScheduler 创建调度器
func NewScheduler(config *SchedulerConfig) *Scheduler {
	if config == nil {
		config = DefaultSchedulerConfig()
	}

	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, config.AllocationStrategy)
	queue := NewTaskQueue(config.MaxQueueSize)
	manager := NewTaskManager(queue, allocator)

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		config:    config,
		registry:  registry,
		allocator: allocator,
		queue:     queue,
		manager:   manager,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 启动心跳检查
	s.wg.Add(1)
	go s.heartbeatChecker()

	// 启动任务分配工作协程
	for i := 0; i < s.config.WorkerCount; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	return nil
}

// Stop 停止调度器
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancel()
	s.wg.Wait()

	return nil
}

// RegisterAgent 注册Agent
func (s *Scheduler) RegisterAgent(agent *Agent) error {
	return s.registry.Register(agent)
}

// UnregisterAgent 注销Agent
func (s *Scheduler) UnregisterAgent(agentID string) error {
	return s.registry.Unregister(agentID)
}

// GetAgent 获取Agent信息
func (s *Scheduler) GetAgent(agentID string) (*Agent, error) {
	return s.registry.GetAgent(agentID)
}

// ListAgents 列出所有Agent
func (s *Scheduler) ListAgents() []*Agent {
	return s.registry.ListAgents()
}

// UpdateAgentStatus 更新Agent状态
func (s *Scheduler) UpdateAgentStatus(agentID string, status AgentStatus) error {
	return s.registry.UpdateAgentStatus(agentID, status)
}

// UpdateAgentHeartbeat 更新Agent心跳
func (s *Scheduler) UpdateAgentHeartbeat(agentID string) error {
	return s.registry.UpdateHeartbeat(agentID)
}

// SubmitTask 提交任务
func (s *Scheduler) SubmitTask(task *Task) error {
	return s.manager.SubmitTask(task)
}

// AssignTask 手动分配任务
func (s *Scheduler) AssignTask(taskID string) (string, error) {
	return s.manager.AssignTask(taskID)
}

// CompleteTask 完成任务
func (s *Scheduler) CompleteTask(taskID string) error {
	return s.manager.CompleteTask(taskID)
}

// FailTask 标记任务失败
func (s *Scheduler) FailTask(taskID string) error {
	return s.manager.FailTask(taskID)
}

// CancelTask 取消任务
func (s *Scheduler) CancelTask(taskID string) error {
	return s.manager.CancelTask(taskID)
}

// GetTask 获取任务
func (s *Scheduler) GetTask(taskID string) (*Task, error) {
	return s.manager.GetTask(taskID)
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks() []*Task {
	return s.manager.ListTasks()
}

// ListTasksByStatus 按状态列出任务
func (s *Scheduler) ListTasksByStatus(status TaskStatus) []*Task {
	return s.manager.ListTasksByStatus(status)
}

// GetAgentTasks 获取Agent的所有任务
func (s *Scheduler) GetAgentTasks(agentID string) []*Task {
	return s.manager.GetAgentTasks(agentID)
}

// GetStatistics 获取统计信息
func (s *Scheduler) GetStatistics() *Statistics {
	return &Statistics{
		TotalAgents:        s.registry.GetAgentCount(),
		AgentsByStatus:     s.registry.GetAgentCountByStatus(),
		TotalTasks:         s.manager.GetTaskCount(),
		TasksByStatus:      s.manager.GetTaskCountByStatus(),
		QueueSize:          s.manager.GetQueueSize(),
		AllocationStrategy: string(s.allocator.GetStrategy()),
	}
}

// SetAllocationStrategy 设置分配策略
func (s *Scheduler) SetAllocationStrategy(strategy AllocationStrategy) {
	s.allocator.SetStrategy(strategy)
}

// GetAllocationStrategy 获取分配策略
func (s *Scheduler) GetAllocationStrategy() AllocationStrategy {
	return s.allocator.GetStrategy()
}

// worker 工作协程
func (s *Scheduler) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// 从队列中取出任务
			task, err := s.queue.DequeueWait(5 * time.Second)
			if err != nil {
				// 超时或队列为空，继续等待
				continue
			}

			// 分配任务
			agentID, err := s.manager.AssignTask(task.ID)
			if err != nil {
				// 分配失败，任务状态已经回到pending，会在下次尝试
				fmt.Printf("Worker %d: failed to assign task %s: %v\n", id, task.ID, err)

				// 重新入队
				if err := s.queue.Enqueue(task); err != nil {
					fmt.Printf("Worker %d: failed to re-enqueue task %s: %v\n", id, task.ID, err)
				}

				// 等待一段时间再重试
				time.Sleep(1 * time.Second)
				continue
			}

			fmt.Printf("Worker %d: assigned task %s to agent %s\n", id, task.ID, agentID)
		}
	}
}

// heartbeatChecker 心跳检查器
func (s *Scheduler) heartbeatChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			timeoutAgents := s.registry.CheckHeartbeat(s.config.HeartbeatTimeout)
			if len(timeoutAgents) > 0 {
				fmt.Printf("Heartbeat checker: %d agents timed out: %v\n", len(timeoutAgents), timeoutAgents)
			}
		}
	}
}

// Statistics 统计信息
type Statistics struct {
	TotalAgents        int                  `json:"total_agents"`
	AgentsByStatus     map[AgentStatus]int  `json:"agents_by_status"`
	TotalTasks         int                  `json:"total_tasks"`
	TasksByStatus      map[string]int       `json:"tasks_by_status"`
	QueueSize          int                  `json:"queue_size"`
	AllocationStrategy string               `json:"allocation_strategy"`
}
