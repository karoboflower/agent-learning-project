package scheduler

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTaskQueue(t *testing.T) {
	queue := NewTaskQueue(100)

	if queue == nil {
		t.Fatal("NewTaskQueue returned nil")
	}

	if queue.Size() != 0 {
		t.Error("New queue should be empty")
	}
}

func TestTaskQueue_Enqueue(t *testing.T) {
	queue := NewTaskQueue(10)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	err := queue.Enqueue(task)
	if err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	if queue.Size() != 1 {
		t.Errorf("Expected size 1, got %d", queue.Size())
	}
}

func TestTaskQueue_EnqueuePriority(t *testing.T) {
	queue := NewTaskQueue(10)

	tasks := []*Task{
		{ID: "task-001", Type: "test", Priority: 5},
		{ID: "task-002", Type: "test", Priority: 10},
		{ID: "task-003", Type: "test", Priority: 3},
		{ID: "task-004", Type: "test", Priority: 8},
	}

	for _, task := range tasks {
		queue.Enqueue(task)
	}

	// Check order (highest priority first)
	expectedOrder := []string{"task-002", "task-004", "task-001", "task-003"}

	for i, expectedID := range expectedOrder {
		task, err := queue.Dequeue()
		if err != nil {
			t.Fatalf("Dequeue failed: %v", err)
		}

		if task.ID != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, task.ID)
		}
	}
}

func TestTaskQueue_Dequeue(t *testing.T) {
	queue := NewTaskQueue(10)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	queue.Enqueue(task)

	dequeued, err := queue.Dequeue()
	if err != nil {
		t.Fatalf("Dequeue failed: %v", err)
	}

	if dequeued.ID != task.ID {
		t.Errorf("Expected task %s, got %s", task.ID, dequeued.ID)
	}

	if queue.Size() != 0 {
		t.Errorf("Expected size 0, got %d", queue.Size())
	}
}

func TestTaskQueue_DequeueEmpty(t *testing.T) {
	queue := NewTaskQueue(10)

	_, err := queue.Dequeue()
	if err == nil {
		t.Error("Expected error when dequeuing from empty queue")
	}
}

func TestTaskQueue_DequeueWait(t *testing.T) {
	queue := NewTaskQueue(10)

	// Start goroutine to enqueue after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		task := &Task{
			ID:       "task-001",
			Type:     "test",
			Priority: 5,
		}
		queue.Enqueue(task)
	}()

	// Wait for task
	task, err := queue.DequeueWait(1 * time.Second)
	if err != nil {
		t.Fatalf("DequeueWait failed: %v", err)
	}

	if task.ID != "task-001" {
		t.Errorf("Expected task-001, got %s", task.ID)
	}
}

func TestTaskQueue_DequeueWaitTimeout(t *testing.T) {
	queue := NewTaskQueue(10)

	_, err := queue.DequeueWait(100 * time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestTaskQueue_Remove(t *testing.T) {
	queue := NewTaskQueue(10)

	tasks := []*Task{
		{ID: "task-001", Type: "test", Priority: 5},
		{ID: "task-002", Type: "test", Priority: 10},
		{ID: "task-003", Type: "test", Priority: 3},
	}

	for _, task := range tasks {
		queue.Enqueue(task)
	}

	err := queue.Remove("task-002")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if queue.Size() != 2 {
		t.Errorf("Expected size 2, got %d", queue.Size())
	}

	if queue.Contains("task-002") {
		t.Error("task-002 should be removed")
	}
}

func TestTaskQueue_Peek(t *testing.T) {
	queue := NewTaskQueue(10)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	queue.Enqueue(task)

	peeked, err := queue.Peek()
	if err != nil {
		t.Fatalf("Peek failed: %v", err)
	}

	if peeked.ID != task.ID {
		t.Errorf("Expected task %s, got %s", task.ID, peeked.ID)
	}

	// Queue should still have the task
	if queue.Size() != 1 {
		t.Errorf("Expected size 1, got %d", queue.Size())
	}
}

func TestTaskQueue_Contains(t *testing.T) {
	queue := NewTaskQueue(10)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	queue.Enqueue(task)

	if !queue.Contains("task-001") {
		t.Error("Queue should contain task-001")
	}

	if queue.Contains("task-999") {
		t.Error("Queue should not contain task-999")
	}
}

func TestTaskQueue_FullQueue(t *testing.T) {
	queue := NewTaskQueue(2)

	queue.Enqueue(&Task{ID: "task-001", Priority: 5})
	queue.Enqueue(&Task{ID: "task-002", Priority: 5})

	err := queue.Enqueue(&Task{ID: "task-003", Priority: 5})
	if err == nil {
		t.Error("Expected error when queue is full")
	}

	if !queue.IsFull() {
		t.Error("Queue should be full")
	}
}

func TestTaskQueue_Clear(t *testing.T) {
	queue := NewTaskQueue(10)

	for i := 0; i < 5; i++ {
		task := &Task{
			ID:       fmt.Sprintf("task-%03d", i),
			Priority: 5,
		}
		queue.Enqueue(task)
	}

	queue.Clear()

	if queue.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", queue.Size())
	}

	if !queue.IsEmpty() {
		t.Error("Queue should be empty after clear")
	}
}

func TestTaskQueue_List(t *testing.T) {
	queue := NewTaskQueue(10)

	tasks := []*Task{
		{ID: "task-001", Priority: 5},
		{ID: "task-002", Priority: 10},
		{ID: "task-003", Priority: 3},
	}

	for _, task := range tasks {
		queue.Enqueue(task)
	}

	list := queue.List()

	if len(list) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(list))
	}

	// Should be in priority order
	if list[0].ID != "task-002" {
		t.Errorf("First task should be task-002, got %s", list[0].ID)
	}
}

func TestTaskQueue_GetTasksByPriority(t *testing.T) {
	queue := NewTaskQueue(10)

	tasks := []*Task{
		{ID: "task-001", Priority: 5},
		{ID: "task-002", Priority: 10},
		{ID: "task-003", Priority: 3},
		{ID: "task-004", Priority: 8},
	}

	for _, task := range tasks {
		queue.Enqueue(task)
	}

	highPriorityTasks := queue.GetTasksByPriority(8)

	if len(highPriorityTasks) != 2 {
		t.Errorf("Expected 2 high priority tasks, got %d", len(highPriorityTasks))
	}
}

func TestNewTaskManager(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	if manager == nil {
		t.Fatal("NewTaskManager returned nil")
	}
}

func TestTaskManager_SubmitTask(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	err := manager.SubmitTask(task)
	if err != nil {
		t.Fatalf("SubmitTask failed: %v", err)
	}

	if task.Status != string(TaskStatusPending) {
		t.Errorf("Expected status PENDING, got %s", task.Status)
	}

	if manager.GetQueueSize() != 1 {
		t.Errorf("Expected queue size 1, got %d", manager.GetQueueSize())
	}
}

func TestTaskManager_AssignTask(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	// Register agent
	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
		Status:       AgentStatusIdle,
		MaxTasks:     5,
	}
	registry.Register(agent)

	// Submit task
	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}
	manager.SubmitTask(task)

	// Assign task
	agentID, err := manager.AssignTask("task-001")
	if err != nil {
		t.Fatalf("AssignTask failed: %v", err)
	}

	if agentID != "agent-001" {
		t.Errorf("Expected agent-001, got %s", agentID)
	}

	// Check task status
	assigned, _ := manager.GetTask("task-001")
	if assigned.Status != string(TaskStatusAssigned) {
		t.Errorf("Expected status ASSIGNED, got %s", assigned.Status)
	}

	if assigned.AssignedAgentID != "agent-001" {
		t.Errorf("Expected assigned agent agent-001, got %s", assigned.AssignedAgentID)
	}
}

func TestTaskManager_CompleteTask(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	// Setup
	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
		Status:       AgentStatusIdle,
		MaxTasks:     5,
	}
	registry.Register(agent)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}
	manager.SubmitTask(task)
	manager.AssignTask("task-001")

	// Complete task
	err := manager.CompleteTask("task-001")
	if err != nil {
		t.Fatalf("CompleteTask failed: %v", err)
	}

	completed, _ := manager.GetTask("task-001")
	if completed.Status != string(TaskStatusCompleted) {
		t.Errorf("Expected status COMPLETED, got %s", completed.Status)
	}
}

func TestTaskManager_FailTask(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	// Setup
	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
		Status:       AgentStatusIdle,
		MaxTasks:     5,
	}
	registry.Register(agent)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}
	manager.SubmitTask(task)
	manager.AssignTask("task-001")

	// Fail task
	err := manager.FailTask("task-001")
	if err != nil {
		t.Fatalf("FailTask failed: %v", err)
	}

	failed, _ := manager.GetTask("task-001")
	if failed.Status != string(TaskStatusFailed) {
		t.Errorf("Expected status FAILED, got %s", failed.Status)
	}
}

func TestTaskManager_CancelTask(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}
	manager.SubmitTask(task)

	// Cancel task
	err := manager.CancelTask("task-001")
	if err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}

	cancelled, _ := manager.GetTask("task-001")
	if cancelled.Status != string(TaskStatusCancelled) {
		t.Errorf("Expected status CANCELLED, got %s", cancelled.Status)
	}

	// Should be removed from queue
	if queue.Contains("task-001") {
		t.Error("Task should be removed from queue")
	}
}

func TestTaskManager_ListTasksByStatus(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	// Submit tasks with different statuses
	for i := 0; i < 3; i++ {
		task := &Task{
			ID:       fmt.Sprintf("task-%03d", i),
			Type:     "test",
			Priority: 5,
		}
		manager.SubmitTask(task)
	}

	pendingTasks := manager.ListTasksByStatus(TaskStatusPending)
	if len(pendingTasks) != 3 {
		t.Errorf("Expected 3 pending tasks, got %d", len(pendingTasks))
	}
}

func TestTaskManager_GetAgentTasks(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)
	queue := NewTaskQueue(100)
	manager := NewTaskManager(queue, allocator)

	// Setup
	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
		Status:       AgentStatusIdle,
		MaxTasks:     5,
	}
	registry.Register(agent)

	// Submit and assign tasks
	for i := 0; i < 3; i++ {
		task := &Task{
			ID:       fmt.Sprintf("task-%03d", i),
			Type:     "test",
			Priority: 5,
		}
		manager.SubmitTask(task)
		manager.AssignTask(task.ID)
	}

	agentTasks := manager.GetAgentTasks("agent-001")
	if len(agentTasks) != 3 {
		t.Errorf("Expected 3 tasks for agent-001, got %d", len(agentTasks))
	}
}

func BenchmarkTaskQueue_Enqueue(b *testing.B) {
	queue := NewTaskQueue(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:       fmt.Sprintf("task-%d", i),
			Priority: i % 10,
		}
		queue.Enqueue(task)
	}
}

func BenchmarkTaskQueue_Dequeue(b *testing.B) {
	queue := NewTaskQueue(10000)

	// Setup
	for i := 0; i < b.N; i++ {
		task := &Task{
			ID:       fmt.Sprintf("task-%d", i),
			Priority: i % 10,
		}
		queue.Enqueue(task)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Dequeue()
	}
}
