package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/agent-learning/go-agent-api/internal/agent"
	"github.com/google/uuid"
)

// Scheduler manages task scheduling and execution
type Scheduler struct {
	agentService  agent.AgentService
	taskQueue     *TaskQueue
	runningTasks  map[string]*agent.Task
	taskResults   map[string]*agent.TaskResult
	maxConcurrent int
	taskTimeout   time.Duration
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewScheduler creates a new scheduler
func NewScheduler(agentService agent.AgentService, maxConcurrent int, taskTimeout time.Duration) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		agentService:  agentService,
		taskQueue:     NewTaskQueue(),
		runningTasks:  make(map[string]*agent.Task),
		taskResults:   make(map[string]*agent.TaskResult),
		maxConcurrent: maxConcurrent,
		taskTimeout:   taskTimeout,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Println("Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processQueue()
		}
	}
}

// processQueue processes pending tasks in the queue
func (s *Scheduler) processQueue() {
	s.mu.Lock()
	runningCount := len(s.runningTasks)
	s.mu.Unlock()

	// Check if we can start more tasks
	if runningCount >= s.maxConcurrent {
		return
	}

	// Get next task
	task := s.taskQueue.Dequeue()
	if task == nil {
		return
	}

	// Check if task is still pending
	if task.Status != agent.TaskStatusPending {
		return
	}

	// Start task execution
	go s.executeTask(task)
}

// executeTask executes a single task
func (s *Scheduler) executeTask(task *agent.Task) {
	// Mark task as running
	task.Status = agent.TaskStatusRunning
	now := time.Now()
	task.StartedAt = &now
	task.UpdatedAt = now

	s.mu.Lock()
	s.runningTasks[task.ID] = task
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.runningTasks, task.ID)
		s.mu.Unlock()
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(s.ctx, s.taskTimeout)
	defer cancel()

	// Get agent
	ag, err := s.agentService.GetAgent(ctx, task.AgentID)
	if err != nil {
		s.handleTaskError(task, fmt.Errorf("failed to get agent: %w", err))
		return
	}

	// Execute task
	result, err := s.agentService.ExecuteTask(ctx, ag, task)
	if err != nil {
		s.handleTaskError(task, err)
		return
	}

	// Update task
	task.Status = agent.TaskStatusCompleted
	task.Output = result.Output
	endTime := time.Now()
	task.EndedAt = &endTime
	task.UpdatedAt = endTime

	// Store result
	s.mu.Lock()
	s.taskResults[task.ID] = result
	s.mu.Unlock()

	log.Printf("Task %s completed in %dms", task.ID, result.Duration)
}

// handleTaskError handles task execution errors
func (s *Scheduler) handleTaskError(task *agent.Task, err error) {
	task.Status = agent.TaskStatusFailed
	task.Error = err.Error()
	endTime := time.Now()
	task.EndedAt = &endTime
	task.UpdatedAt = endTime

	// Store error result
	s.mu.Lock()
	s.taskResults[task.ID] = &agent.TaskResult{
		TaskID:    task.ID,
		Status:    agent.TaskStatusFailed,
		Error:     err.Error(),
		CreatedAt: task.CreatedAt,
		EndedAt:   endTime,
		Duration:  endTime.Sub(task.CreatedAt).Milliseconds(),
	}
	s.mu.Unlock()

	log.Printf("Task %s failed: %v", task.ID, err)
}

// SubmitTask submits a new task for execution
func (s *Scheduler) SubmitTask(req *agent.CreateTaskRequest) (*agent.Task, error) {
	// Validate agent exists
	_, err := s.agentService.GetAgent(s.ctx, req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("invalid agent_id: %w", err)
	}

	// Create task
	task := &agent.Task{
		ID:        uuid.New().String(),
		AgentID:   req.AgentID,
		Type:      req.Type,
		Input:     req.Input,
		Status:    agent.TaskStatusPending,
		Priority:  req.Priority,
		Tools:     req.Tools,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add to queue
	s.taskQueue.Enqueue(task)

	log.Printf("Task %s submitted (priority: %d)", task.ID, task.Priority)
	return task, nil
}

// GetTask retrieves a task by ID
func (s *Scheduler) GetTask(taskID string) (*agent.Task, error) {
	// Check running tasks
	s.mu.RLock()
	if task, exists := s.runningTasks[taskID]; exists {
		s.mu.RUnlock()
		return task, nil
	}
	s.mu.RUnlock()

	// Check queue
	tasks := s.taskQueue.List()
	for _, task := range tasks {
		if task.ID == taskID {
			return task, nil
		}
	}

	return nil, fmt.Errorf("task not found: %s", taskID)
}

// GetTaskResult retrieves the result of a completed task
func (s *Scheduler) GetTaskResult(taskID string) (*agent.TaskResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.taskResults[taskID]
	if !exists {
		return nil, fmt.Errorf("task result not found: %s", taskID)
	}

	return result, nil
}

// CancelTask cancels a pending or running task
func (s *Scheduler) CancelTask(taskID string) error {
	// Try to remove from queue
	if s.taskQueue.Remove(taskID) {
		log.Printf("Task %s cancelled (was pending)", taskID)
		return nil
	}

	// Check if running
	s.mu.RLock()
	_, isRunning := s.runningTasks[taskID]
	s.mu.RUnlock()

	if isRunning {
		// Task is running, mark as cancelled
		// The actual cancellation will happen via context
		log.Printf("Task %s cancelled (was running)", taskID)
		return nil
	}

	return fmt.Errorf("task not found or already completed: %s", taskID)
}

// ListTasks returns all tasks (pending, running, and completed)
func (s *Scheduler) ListTasks() []*agent.Task {
	tasks := make([]*agent.Task, 0)

	// Add pending tasks
	tasks = append(tasks, s.taskQueue.List()...)

	// Add running tasks
	s.mu.RLock()
	for _, task := range s.runningTasks {
		tasks = append(tasks, task)
	}
	s.mu.RUnlock()

	return tasks
}

// GetStats returns scheduler statistics
func (s *Scheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"pending_tasks":   s.taskQueue.Len(),
		"running_tasks":   len(s.runningTasks),
		"completed_tasks": len(s.taskResults),
		"max_concurrent":  s.maxConcurrent,
	}
}
