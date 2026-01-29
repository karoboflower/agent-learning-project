package scheduler

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"   // 等待中
	TaskStatusAssigned  TaskStatus = "ASSIGNED"  // 已分配
	TaskStatusRunning   TaskStatus = "RUNNING"   // 执行中
	TaskStatusCompleted TaskStatus = "COMPLETED" // 已完成
	TaskStatusFailed    TaskStatus = "FAILED"    // 失败
	TaskStatusCancelled TaskStatus = "CANCELLED" // 已取消
)

// TaskQueueItem 任务队列项
type TaskQueueItem struct {
	Task      *Task
	EnqueueAt time.Time
	Priority  int
}

// TaskQueue 任务队列
type TaskQueue struct {
	items    *list.List
	itemMap  map[string]*list.Element // taskID -> element
	mu       sync.RWMutex
	maxSize  int
	notEmpty *sync.Cond
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(maxSize int) *TaskQueue {
	tq := &TaskQueue{
		items:   list.New(),
		itemMap: make(map[string]*list.Element),
		maxSize: maxSize,
	}
	tq.notEmpty = sync.NewCond(&tq.mu)
	return tq
}

// Enqueue 入队
func (q *TaskQueue) Enqueue(task *Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.maxSize > 0 && q.items.Len() >= q.maxSize {
		return fmt.Errorf("queue is full")
	}

	if _, exists := q.itemMap[task.ID]; exists {
		return fmt.Errorf("task %s already in queue", task.ID)
	}

	item := &TaskQueueItem{
		Task:      task,
		EnqueueAt: time.Now(),
		Priority:  task.Priority,
	}

	// 按优先级插入
	inserted := false
	for e := q.items.Front(); e != nil; e = e.Next() {
		existing := e.Value.(*TaskQueueItem)
		if item.Priority > existing.Priority {
			elem := q.items.InsertBefore(item, e)
			q.itemMap[task.ID] = elem
			inserted = true
			break
		}
	}

	if !inserted {
		elem := q.items.PushBack(item)
		q.itemMap[task.ID] = elem
	}

	// 通知等待的goroutine
	q.notEmpty.Signal()

	return nil
}

// Dequeue 出队
func (q *TaskQueue) Dequeue() (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.items.Len() == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	elem := q.items.Front()
	item := elem.Value.(*TaskQueueItem)
	q.items.Remove(elem)
	delete(q.itemMap, item.Task.ID)

	return item.Task, nil
}

// DequeueWait 等待出队（阻塞）
func (q *TaskQueue) DequeueWait(timeout time.Duration) (*Task, error) {
	if timeout > 0 {
		// 使用轮询实现超时
		deadline := time.Now().Add(timeout)
		for {
			q.mu.Lock()
			if q.items.Len() > 0 {
				elem := q.items.Front()
				item := elem.Value.(*TaskQueueItem)
				q.items.Remove(elem)
				delete(q.itemMap, item.Task.ID)
				q.mu.Unlock()
				return item.Task, nil
			}
			q.mu.Unlock()

			if time.Now().After(deadline) {
				return nil, fmt.Errorf("dequeue timeout")
			}

			time.Sleep(10 * time.Millisecond)
		}
	}

	// 无限等待
	q.mu.Lock()
	for q.items.Len() == 0 {
		q.notEmpty.Wait()
	}

	elem := q.items.Front()
	if elem == nil {
		q.mu.Unlock()
		return nil, fmt.Errorf("queue is empty")
	}

	item := elem.Value.(*TaskQueueItem)
	q.items.Remove(elem)
	delete(q.itemMap, item.Task.ID)
	q.mu.Unlock()

	return item.Task, nil
}

// Remove 移除指定任务
func (q *TaskQueue) Remove(taskID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	elem, exists := q.itemMap[taskID]
	if !exists {
		return fmt.Errorf("task %s not found in queue", taskID)
	}

	q.items.Remove(elem)
	delete(q.itemMap, taskID)

	return nil
}

// Peek 查看队首任务（不移除）
func (q *TaskQueue) Peek() (*Task, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.items.Len() == 0 {
		return nil, fmt.Errorf("queue is empty")
	}

	elem := q.items.Front()
	item := elem.Value.(*TaskQueueItem)

	return item.Task, nil
}

// Contains 检查任务是否在队列中
func (q *TaskQueue) Contains(taskID string) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	_, exists := q.itemMap[taskID]
	return exists
}

// Size 获取队列大小
func (q *TaskQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.items.Len()
}

// IsEmpty 检查队列是否为空
func (q *TaskQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.items.Len() == 0
}

// IsFull 检查队列是否已满
func (q *TaskQueue) IsFull() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.maxSize > 0 && q.items.Len() >= q.maxSize
}

// Clear 清空队列
func (q *TaskQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items.Init()
	q.itemMap = make(map[string]*list.Element)
}

// List 列出所有任务
func (q *TaskQueue) List() []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*Task, 0, q.items.Len())
	for e := q.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*TaskQueueItem)
		tasks = append(tasks, item.Task)
	}

	return tasks
}

// GetTasksByPriority 按优先级获取任务
func (q *TaskQueue) GetTasksByPriority(minPriority int) []*Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tasks := make([]*Task, 0)
	for e := q.items.Front(); e != nil; e = e.Next() {
		item := e.Value.(*TaskQueueItem)
		if item.Priority >= minPriority {
			tasks = append(tasks, item.Task)
		}
	}

	return tasks
}

// TaskManager 任务管理器
type TaskManager struct {
	queue       *TaskQueue
	allocator   *TaskAllocator
	tasks       map[string]*Task // taskID -> Task
	assignments map[string]string // taskID -> agentID
	mu          sync.RWMutex
}

// NewTaskManager 创建任务管理器
func NewTaskManager(queue *TaskQueue, allocator *TaskAllocator) *TaskManager {
	return &TaskManager{
		queue:       queue,
		allocator:   allocator,
		tasks:       make(map[string]*Task),
		assignments: make(map[string]string),
	}
}

// SubmitTask 提交任务
func (m *TaskManager) SubmitTask(task *Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	if _, exists := m.tasks[task.ID]; exists {
		return fmt.Errorf("task %s already exists", task.ID)
	}

	task.Status = string(TaskStatusPending)
	m.tasks[task.ID] = task

	// 入队
	return m.queue.Enqueue(task)
}

// AssignTask 分配任务
func (m *TaskManager) AssignTask(taskID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return "", fmt.Errorf("task %s not found", taskID)
	}

	if task.Status != string(TaskStatusPending) {
		return "", fmt.Errorf("task %s is not pending (status: %s)", taskID, task.Status)
	}

	// 分配Agent
	agentID, err := m.allocator.Allocate(task)
	if err != nil {
		return "", fmt.Errorf("failed to allocate agent: %w", err)
	}

	// 更新任务状态
	task.Status = string(TaskStatusAssigned)
	task.AssignedAgentID = agentID
	m.assignments[taskID] = agentID

	// 从队列中移除
	m.queue.Remove(taskID)

	// 增加Agent任务计数
	if err := m.allocator.registry.IncrementTaskCount(agentID); err != nil {
		return "", fmt.Errorf("failed to increment task count: %w", err)
	}

	return agentID, nil
}

// CompleteTask 完成任务
func (m *TaskManager) CompleteTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	agentID, exists := m.assignments[taskID]
	if !exists {
		return fmt.Errorf("task %s has no assignment", taskID)
	}

	// 更新任务状态
	task.Status = string(TaskStatusCompleted)

	// 减少Agent任务计数
	if err := m.allocator.registry.DecrementTaskCount(agentID); err != nil {
		return fmt.Errorf("failed to decrement task count: %w", err)
	}

	return nil
}

// FailTask 标记任务失败
func (m *TaskManager) FailTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	agentID, exists := m.assignments[taskID]
	if exists {
		// 减少Agent任务计数
		m.allocator.registry.DecrementTaskCount(agentID)
	}

	// 更新任务状态
	task.Status = string(TaskStatusFailed)

	return nil
}

// CancelTask 取消任务
func (m *TaskManager) CancelTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	// 如果任务在队列中，移除
	if m.queue.Contains(taskID) {
		m.queue.Remove(taskID)
	}

	// 如果任务已分配，减少Agent任务计数
	if agentID, exists := m.assignments[taskID]; exists {
		m.allocator.registry.DecrementTaskCount(agentID)
		delete(m.assignments, taskID)
	}

	// 更新任务状态
	task.Status = string(TaskStatusCancelled)

	return nil
}

// GetTask 获取任务
func (m *TaskManager) GetTask(taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// ListTasks 列出所有任务
func (m *TaskManager) ListTasks() []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// ListTasksByStatus 按状态列出任务
func (m *TaskManager) ListTasksByStatus(status TaskStatus) []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0)
	for _, task := range m.tasks {
		if task.Status == string(status) {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// GetAssignment 获取任务分配信息
func (m *TaskManager) GetAssignment(taskID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agentID, exists := m.assignments[taskID]
	if !exists {
		return "", fmt.Errorf("task %s has no assignment", taskID)
	}

	return agentID, nil
}

// GetAgentTasks 获取Agent的所有任务
func (m *TaskManager) GetAgentTasks(agentID string) []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*Task, 0)
	for taskID, assignedAgentID := range m.assignments {
		if assignedAgentID == agentID {
			if task, exists := m.tasks[taskID]; exists {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks
}

// GetQueueSize 获取队列大小
func (m *TaskManager) GetQueueSize() int {
	return m.queue.Size()
}

// GetTaskCount 获取任务总数
func (m *TaskManager) GetTaskCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.tasks)
}

// GetTaskCountByStatus 按状态统计任务数量
func (m *TaskManager) GetTaskCountByStatus() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[string]int)
	for _, task := range m.tasks {
		counts[task.Status]++
	}

	return counts
}
