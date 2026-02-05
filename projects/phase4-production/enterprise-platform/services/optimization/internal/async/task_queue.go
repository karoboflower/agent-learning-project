package async

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

// TaskQueue 任务队列
type TaskQueue struct {
	tasks    chan *model.AsyncTask
	workers  int
	handlers map[string]TaskHandler
	mu       sync.RWMutex
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// TaskHandler 任务处理器
type TaskHandler func(context.Context, *model.AsyncTask) (map[string]interface{}, error)

// NewTaskQueue 创建任务队列
func NewTaskQueue(workers int, queueSize int) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())

	tq := &TaskQueue{
		tasks:    make(chan *model.AsyncTask, queueSize),
		workers:  workers,
		handlers: make(map[string]TaskHandler),
		ctx:      ctx,
		cancel:   cancel,
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		tq.wg.Add(1)
		go tq.worker(i)
	}

	return tq
}

// RegisterHandler 注册任务处理器
func (tq *TaskQueue) RegisterHandler(taskType string, handler TaskHandler) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	tq.handlers[taskType] = handler
}

// Enqueue 入队任务
func (tq *TaskQueue) Enqueue(task *model.AsyncTask) error {
	select {
	case tq.tasks <- task:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("queue is full, timeout enqueueing task")
	}
}

// worker 工作协程
func (tq *TaskQueue) worker(id int) {
	defer tq.wg.Done()

	for {
		select {
		case <-tq.ctx.Done():
			return

		case task, ok := <-tq.tasks:
			if !ok {
				return
			}

			tq.processTask(task)
		}
	}
}

// processTask 处理任务
func (tq *TaskQueue) processTask(task *model.AsyncTask) {
	// 更新任务状态为处理中
	now := time.Now()
	task.Status = "processing"
	task.StartedAt = &now

	// 获取处理器
	tq.mu.RLock()
	handler, ok := tq.handlers[task.Type]
	tq.mu.RUnlock()

	if !ok {
		task.Status = "failed"
		task.Error = fmt.Sprintf("no handler registered for task type: %s", task.Type)
		completed := time.Now()
		task.CompletedAt = &completed
		return
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(tq.ctx, 30*time.Minute)
	defer cancel()

	// 执行任务
	result, err := handler(ctx, task)

	completed := time.Now()
	task.CompletedAt = &completed

	if err != nil {
		// 检查是否需要重试
		if task.RetryCount < task.MaxRetries {
			task.RetryCount++
			task.Status = "pending"
			task.StartedAt = nil
			task.CompletedAt = nil

			// 重新入队（带延迟）
			go func() {
				time.Sleep(time.Duration(task.RetryCount*5) * time.Second)
				tq.Enqueue(task)
			}()
		} else {
			task.Status = "failed"
			task.Error = err.Error()
		}
	} else {
		task.Status = "completed"
		task.Result = result
		task.Progress = 100
	}
}

// Shutdown 关闭队列
func (tq *TaskQueue) Shutdown() {
	tq.cancel()
	close(tq.tasks)
	tq.wg.Wait()
}

// GetQueueSize 获取队列大小
func (tq *TaskQueue) GetQueueSize() int {
	return len(tq.tasks)
}

// PriorityTaskQueue 优先级任务队列
type PriorityTaskQueue struct {
	highQueue   *TaskQueue
	normalQueue *TaskQueue
	lowQueue    *TaskQueue
}

// NewPriorityTaskQueue 创建优先级任务队列
func NewPriorityTaskQueue(workers int, queueSize int) *PriorityTaskQueue {
	// 分配工作协程：高优先级50%，普通30%，低优先级20%
	highWorkers := workers * 5 / 10
	normalWorkers := workers * 3 / 10
	lowWorkers := workers - highWorkers - normalWorkers

	return &PriorityTaskQueue{
		highQueue:   NewTaskQueue(highWorkers, queueSize),
		normalQueue: NewTaskQueue(normalWorkers, queueSize),
		lowQueue:    NewTaskQueue(lowWorkers, queueSize),
	}
}

// RegisterHandler 注册任务处理器
func (ptq *PriorityTaskQueue) RegisterHandler(taskType string, handler TaskHandler) {
	ptq.highQueue.RegisterHandler(taskType, handler)
	ptq.normalQueue.RegisterHandler(taskType, handler)
	ptq.lowQueue.RegisterHandler(taskType, handler)
}

// Enqueue 入队任务
func (ptq *PriorityTaskQueue) Enqueue(task *model.AsyncTask) error {
	// 根据优先级选择队列
	if task.Priority >= 7 {
		return ptq.highQueue.Enqueue(task)
	} else if task.Priority >= 4 {
		return ptq.normalQueue.Enqueue(task)
	} else {
		return ptq.lowQueue.Enqueue(task)
	}
}

// Shutdown 关闭队列
func (ptq *PriorityTaskQueue) Shutdown() {
	ptq.highQueue.Shutdown()
	ptq.normalQueue.Shutdown()
	ptq.lowQueue.Shutdown()
}

// GetStats 获取统计信息
func (ptq *PriorityTaskQueue) GetStats() map[string]int {
	return map[string]int{
		"high_queue":   ptq.highQueue.GetQueueSize(),
		"normal_queue": ptq.normalQueue.GetQueueSize(),
		"low_queue":    ptq.lowQueue.GetQueueSize(),
	}
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	queue     *PriorityTaskQueue
	scheduler *time.Ticker
	tasks     map[string]*model.AsyncTask
	mu        sync.RWMutex
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(queue *PriorityTaskQueue) *TaskScheduler {
	ts := &TaskScheduler{
		queue:     queue,
		scheduler: time.NewTicker(1 * time.Second),
		tasks:     make(map[string]*model.AsyncTask),
	}

	go ts.run()

	return ts
}

// run 运行调度器
func (ts *TaskScheduler) run() {
	for range ts.scheduler.C {
		ts.checkScheduledTasks()
	}
}

// checkScheduledTasks 检查待调度任务
func (ts *TaskScheduler) checkScheduledTasks() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()

	for id, task := range ts.tasks {
		if task.Status == "pending" && task.ScheduledAt.Before(now) {
			// 入队任务
			if err := ts.queue.Enqueue(task); err == nil {
				delete(ts.tasks, id)
			}
		}
	}
}

// ScheduleTask 调度任务
func (ts *TaskScheduler) ScheduleTask(task *model.AsyncTask) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tasks[task.ID] = task
}

// CancelTask 取消任务
func (ts *TaskScheduler) CancelTask(taskID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task, ok := ts.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != "pending" {
		return fmt.Errorf("cannot cancel task in status: %s", task.Status)
	}

	delete(ts.tasks, taskID)
	return nil
}

// Stop 停止调度器
func (ts *TaskScheduler) Stop() {
	ts.scheduler.Stop()
}
