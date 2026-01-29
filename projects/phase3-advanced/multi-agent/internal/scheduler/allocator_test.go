package scheduler

import (
	"fmt"
	"testing"
)

func TestNewTaskAllocator(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)

	if allocator == nil {
		t.Fatal("NewTaskAllocator returned nil")
	}

	if allocator.GetStrategy() != StrategyLoadBalance {
		t.Errorf("Expected strategy LOAD_BALANCE, got %s", allocator.GetStrategy())
	}
}

func TestTaskAllocator_SetStrategy(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)

	allocator.SetStrategy(StrategyCapability)

	if allocator.GetStrategy() != StrategyCapability {
		t.Errorf("Expected strategy CAPABILITY, got %s", allocator.GetStrategy())
	}
}

func TestTaskAllocator_AllocateByCapability(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyCapability)

	// Register agents
	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"code_review"}, Status: AgentStatusIdle, MaxTasks: 5},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"testing"}, Status: AgentStatusIdle, MaxTasks: 5},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"code_review", "testing"}, Status: AgentStatusIdle, MaxTasks: 5},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	// Test task with code_review capability
	task := &Task{
		ID:                   "task-001",
		Type:                 "review",
		Priority:             5,
		RequiredCapabilities: []string{"code_review"},
	}

	agentID, err := allocator.Allocate(task)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	// Should be agent-001 or agent-003
	if agentID != "agent-001" && agentID != "agent-003" {
		t.Errorf("Expected agent-001 or agent-003, got %s", agentID)
	}

	// Test task with both capabilities
	task2 := &Task{
		ID:                   "task-002",
		Type:                 "review",
		Priority:             5,
		RequiredCapabilities: []string{"code_review", "testing"},
	}

	agentID, err = allocator.Allocate(task2)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	// Should be agent-003 (only one with both capabilities)
	if agentID != "agent-003" {
		t.Errorf("Expected agent-003, got %s", agentID)
	}
}

func TestTaskAllocator_AllocateByLoadBalance(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)

	// Register agents with different loads
	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 3, Load: 0.6},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 1, Load: 0.2},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 2, Load: 0.4},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	agentID, err := allocator.Allocate(task)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	// Should be agent-002 (lowest load)
	if agentID != "agent-002" {
		t.Errorf("Expected agent-002 (lowest load), got %s", agentID)
	}
}

func TestTaskAllocator_AllocateByPriority(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyPriority)

	// Register agents
	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 3, Load: 0.6},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 1, Load: 0.2},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	// High priority task
	highPriorityTask := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 9,
	}

	agentID, err := allocator.Allocate(highPriorityTask)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	// High priority should go to lowest load agent
	if agentID != "agent-002" {
		t.Errorf("High priority task should go to agent-002, got %s", agentID)
	}
}

func TestTaskAllocator_AllocateByRoundRobin(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyRoundRobin)

	// Register agents
	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	// Allocate 3 tasks
	allocations := make([]string, 3)
	for i := 0; i < 3; i++ {
		task := &Task{
			ID:       fmt.Sprintf("task-%03d", i),
			Type:     "test",
			Priority: 5,
		}

		agentID, err := allocator.Allocate(task)
		if err != nil {
			t.Fatalf("Allocate failed: %v", err)
		}

		allocations[i] = agentID
	}

	// Each agent should be used once
	agentCount := make(map[string]int)
	for _, agentID := range allocations {
		agentCount[agentID]++
	}

	if len(agentCount) != 3 {
		t.Errorf("Expected 3 different agents, got %d", len(agentCount))
	}

	for agentID, count := range agentCount {
		if count != 1 {
			t.Errorf("Agent %s used %d times, expected 1", agentID, count)
		}
	}
}

func TestTaskAllocator_NoAvailableAgents(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	_, err := allocator.Allocate(task)
	if err == nil {
		t.Error("Expected error when no agents available")
	}
}

func TestTaskAllocator_NoMatchingCapabilities(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyCapability)

	// Register agent with different capability
	agent := &Agent{
		ID:           "agent-001",
		Name:         "Agent 1",
		Capabilities: []string{"testing"},
		Status:       AgentStatusIdle,
		MaxTasks:     5,
	}
	registry.Register(agent)

	// Task requires different capability
	task := &Task{
		ID:                   "task-001",
		Type:                 "review",
		Priority:             5,
		RequiredCapabilities: []string{"code_review"},
	}

	_, err := allocator.Allocate(task)
	if err == nil {
		t.Error("Expected error when no matching capabilities")
	}
}

func TestTaskAllocator_BatchAllocate(t *testing.T) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)

	// Register agents
	for i := 0; i < 3; i++ {
		agent := &Agent{
			ID:           fmt.Sprintf("agent-%03d", i),
			Name:         fmt.Sprintf("Agent %d", i),
			Capabilities: []string{"test"},
			Status:       AgentStatusIdle,
			MaxTasks:     5,
		}
		registry.Register(agent)
	}

	// Create tasks
	tasks := make([]*Task, 5)
	for i := 0; i < 5; i++ {
		tasks[i] = &Task{
			ID:       fmt.Sprintf("task-%03d", i),
			Type:     "test",
			Priority: i, // Different priorities
		}
	}

	allocations, errors := allocator.BatchAllocate(tasks)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d errors", len(errors))
	}

	if len(allocations) != 5 {
		t.Errorf("Expected 5 allocations, got %d", len(allocations))
	}

	// Check all tasks were allocated
	for _, task := range tasks {
		if _, ok := allocations[task.ID]; !ok {
			t.Errorf("Task %s was not allocated", task.ID)
		}
	}
}

func TestHasAllCapabilities(t *testing.T) {
	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"code_review", "testing", "refactoring"},
	}

	tests := []struct {
		name         string
		capabilities []string
		expected     bool
	}{
		{"single matching", []string{"code_review"}, true},
		{"multiple matching", []string{"code_review", "testing"}, true},
		{"all matching", []string{"code_review", "testing", "refactoring"}, true},
		{"single non-matching", []string{"deployment"}, false},
		{"partial matching", []string{"code_review", "deployment"}, false},
		{"empty", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasAllCapabilities(agent, tt.capabilities)
			if result != tt.expected {
				t.Errorf("hasAllCapabilities(%v) = %v, expected %v", tt.capabilities, result, tt.expected)
			}
		})
	}
}

func TestCalculateLoad(t *testing.T) {
	tests := []struct {
		name         string
		agent        *Agent
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name: "no load",
			agent: &Agent{
				Load:         0.0,
				CurrentTasks: 0,
				MaxTasks:     10,
			},
			expectedMin: 0.0,
			expectedMax: 0.0,
		},
		{
			name: "half load",
			agent: &Agent{
				Load:         0.5,
				CurrentTasks: 5,
				MaxTasks:     10,
			},
			expectedMin: 0.4,
			expectedMax: 0.6,
		},
		{
			name: "full load",
			agent: &Agent{
				Load:         1.0,
				CurrentTasks: 10,
				MaxTasks:     10,
			},
			expectedMin: 0.9,
			expectedMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			load := calculateLoad(tt.agent)
			if load < tt.expectedMin || load > tt.expectedMax {
				t.Errorf("calculateLoad() = %f, expected between %f and %f", load, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func BenchmarkTaskAllocator_Allocate(b *testing.B) {
	registry := NewAgentRegistry()
	allocator := NewTaskAllocator(registry, StrategyLoadBalance)

	// Setup
	for i := 0; i < 100; i++ {
		agent := &Agent{
			ID:           fmt.Sprintf("agent-%d", i),
			Name:         fmt.Sprintf("Agent %d", i),
			Capabilities: []string{"test"},
			Status:       AgentStatusIdle,
			MaxTasks:     10,
		}
		registry.Register(agent)
	}

	task := &Task{
		ID:       "task-001",
		Type:     "test",
		Priority: 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allocator.Allocate(task)
	}
}
