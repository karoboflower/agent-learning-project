package scheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/agent-learning/go-agent-api/internal/agent"
)

// TaskQueue is a priority queue for tasks
type TaskQueue struct {
	tasks []*agent.Task
	mu    sync.RWMutex
}

// NewTaskQueue creates a new task queue
func NewTaskQueue() *TaskQueue {
	tq := &TaskQueue{
		tasks: make([]*agent.Task, 0),
	}
	heap.Init(tq)
	return tq
}

// Len returns the number of tasks in the queue
// NOTE: This is part of heap.Interface and should NOT lock
// Locking is done in the public methods that call heap functions
func (tq *TaskQueue) Len() int {
	return len(tq.tasks)
}

// Less compares two tasks by priority (higher priority first)
// NOTE: This is part of heap.Interface and should NOT lock
func (tq *TaskQueue) Less(i, j int) bool {
	// Higher priority value = higher priority
	// If same priority, older task first (FIFO)
	if tq.tasks[i].Priority == tq.tasks[j].Priority {
		return tq.tasks[i].CreatedAt.Before(tq.tasks[j].CreatedAt)
	}
	return tq.tasks[i].Priority > tq.tasks[j].Priority
}

// Swap swaps two tasks in the queue
// NOTE: This is part of heap.Interface and should NOT lock
func (tq *TaskQueue) Swap(i, j int) {
	tq.tasks[i], tq.tasks[j] = tq.tasks[j], tq.tasks[i]
}

// Push adds a task to the queue
// NOTE: This is part of heap.Interface and should NOT lock
// This method is called by heap.Push, which is called from Enqueue
func (tq *TaskQueue) Push(x interface{}) {
	task := x.(*agent.Task)
	tq.tasks = append(tq.tasks, task)
}

// Pop removes and returns the highest priority task
// NOTE: This is part of heap.Interface and should NOT lock
// This method is called by heap.Pop, which is called from Dequeue
func (tq *TaskQueue) Pop() interface{} {
	old := tq.tasks
	n := len(old)
	task := old[n-1]
	old[n-1] = nil // avoid memory leak
	tq.tasks = old[0 : n-1]
	return task
}

// Enqueue adds a task to the queue with priority
func (tq *TaskQueue) Enqueue(task *agent.Task) {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	heap.Push(tq, task)
}

// Dequeue removes and returns the highest priority task
func (tq *TaskQueue) Dequeue() *agent.Task {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if len(tq.tasks) == 0 {
		return nil
	}
	return heap.Pop(tq).(*agent.Task)
}

// Peek returns the highest priority task without removing it
func (tq *TaskQueue) Peek() *agent.Task {
	tq.mu.RLock()
	defer tq.mu.RUnlock()
	if len(tq.tasks) == 0 {
		return nil
	}
	return tq.tasks[0]
}

// Size returns the number of tasks (thread-safe version of Len)
func (tq *TaskQueue) Size() int {
	tq.mu.RLock()
	defer tq.mu.RUnlock()
	return len(tq.tasks)
}

// Remove removes a specific task from the queue
func (tq *TaskQueue) Remove(taskID string) bool {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	for i, task := range tq.tasks {
		if task.ID == taskID {
			heap.Remove(tq, i)
			return true
		}
	}
	return false
}

// List returns all tasks in the queue
func (tq *TaskQueue) List() []*agent.Task {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	tasks := make([]*agent.Task, len(tq.tasks))
	copy(tasks, tq.tasks)
	return tasks
}

// Clear removes all tasks from the queue
func (tq *TaskQueue) Clear() {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.tasks = make([]*agent.Task, 0)
}

// GetByStatus returns tasks with specific status
func (tq *TaskQueue) GetByStatus(status agent.TaskStatus) []*agent.Task {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	tasks := make([]*agent.Task, 0)
	for _, task := range tq.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

// UpdateTaskStatus updates the status of a task in the queue
func (tq *TaskQueue) UpdateTaskStatus(taskID string, status agent.TaskStatus) bool {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	for _, task := range tq.tasks {
		if task.ID == taskID {
			task.Status = status
			task.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}
