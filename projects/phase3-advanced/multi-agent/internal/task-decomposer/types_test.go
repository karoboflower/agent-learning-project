package decomposer

import (
	"testing"
)

func TestNewTask(t *testing.T) {
	task := NewTask("task-001", "code_review", "Review code")

	if task.ID != "task-001" {
		t.Errorf("Expected ID task-001, got %s", task.ID)
	}

	if task.Type != "code_review" {
		t.Errorf("Expected Type code_review, got %s", task.Type)
	}

	if task.Description != "Review code" {
		t.Errorf("Expected Description 'Review code', got %s", task.Description)
	}

	if task.Priority != 5 {
		t.Errorf("Expected default Priority 5, got %d", task.Priority)
	}

	if task.Dependencies == nil {
		t.Error("Expected Dependencies to be initialized")
	}

	if task.Requirements == nil {
		t.Error("Expected Requirements to be initialized")
	}

	if task.Capabilities == nil {
		t.Error("Expected Capabilities to be initialized")
	}

	if task.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}

	if task.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestTask_AddDependency(t *testing.T) {
	task := NewTask("task-001", "test", "Test task")

	task.AddDependency("dep-1")
	task.AddDependency("dep-2")

	if len(task.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(task.Dependencies))
	}

	if task.Dependencies[0] != "dep-1" {
		t.Errorf("Expected first dependency dep-1, got %s", task.Dependencies[0])
	}

	if task.Dependencies[1] != "dep-2" {
		t.Errorf("Expected second dependency dep-2, got %s", task.Dependencies[1])
	}
}

func TestTask_AddCapability(t *testing.T) {
	task := NewTask("task-001", "test", "Test task")

	task.AddCapability("syntax_check")
	task.AddCapability("quality_check")

	if len(task.Capabilities) != 2 {
		t.Errorf("Expected 2 capabilities, got %d", len(task.Capabilities))
	}

	if task.Capabilities[0] != "syntax_check" {
		t.Errorf("Expected first capability syntax_check, got %s", task.Capabilities[0])
	}

	if task.Capabilities[1] != "quality_check" {
		t.Errorf("Expected second capability quality_check, got %s", task.Capabilities[1])
	}
}

func TestTask_SetRequirement(t *testing.T) {
	task := NewTask("task-001", "test", "Test task")

	task.SetRequirement("quality", "high")
	task.SetRequirement("timeout", 300)

	if len(task.Requirements) != 2 {
		t.Errorf("Expected 2 requirements, got %d", len(task.Requirements))
	}

	if task.Requirements["quality"] != "high" {
		t.Errorf("Expected quality requirement 'high', got %v", task.Requirements["quality"])
	}

	if task.Requirements["timeout"] != 300 {
		t.Errorf("Expected timeout requirement 300, got %v", task.Requirements["timeout"])
	}
}

func TestTask_GetComplexity(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Task)
		expected TaskComplexity
	}{
		{
			name: "Simple task",
			setup: func(task *Task) {
				// No modifications
			},
			expected: ComplexitySimple,
		},
		{
			name: "Moderate task - few dependencies",
			setup: func(task *Task) {
				task.AddDependency("dep-1")
				task.AddDependency("dep-2")
			},
			expected: ComplexityModerate,
		},
		{
			name: "Moderate task - few capabilities",
			setup: func(task *Task) {
				task.AddCapability("cap-1")
				task.AddCapability("cap-2")
			},
			expected: ComplexityModerate,
		},
		{
			name: "Complex task",
			setup: func(task *Task) {
				for i := 0; i < 3; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
				task.AddCapability("cap-1")
				task.AddCapability("cap-2")
			},
			expected: ComplexityComplex,
		},
		{
			name: "Very complex task",
			setup: func(task *Task) {
				for i := 0; i < 6; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
				for i := 0; i < 4; i++ {
					task.AddCapability("cap-" + string(rune('1'+i)))
				}
				for i := 0; i < 6; i++ {
					task.SetRequirement("req-"+string(rune('1'+i)), "value")
				}
			},
			expected: ComplexityVeryComplex,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("test-001", "test", "Test task")
			tt.setup(task)

			complexity := task.GetComplexity()
			if complexity != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, complexity)
			}
		})
	}
}

func TestTask_IsDecomposable(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Task)
		expected bool
	}{
		{
			name: "Simple task - not decomposable",
			setup: func(task *Task) {
				// No modifications
			},
			expected: false,
		},
		{
			name: "Moderate task - decomposable",
			setup: func(task *Task) {
				task.AddDependency("dep-1")
				task.AddCapability("cap-1")
			},
			expected: true,
		},
		{
			name: "Complex task - decomposable",
			setup: func(task *Task) {
				for i := 0; i < 3; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("test-001", "test", "Test task")
			tt.setup(task)

			decomposable := task.IsDecomposable()
			if decomposable != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, decomposable)
			}
		})
	}
}

func TestNewDependencyGraph(t *testing.T) {
	graph := NewDependencyGraph()

	if graph == nil {
		t.Fatal("NewDependencyGraph returned nil")
	}

	if graph.Nodes == nil {
		t.Error("Expected Nodes to be initialized")
	}

	if graph.Edges == nil {
		t.Error("Expected Edges to be initialized")
	}

	if len(graph.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(graph.Edges))
	}
}

func TestDependencyGraph_AddNode(t *testing.T) {
	graph := NewDependencyGraph()

	graph.AddNode("task-1")
	graph.AddNode("task-2")

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}

	if _, exists := graph.Nodes["task-1"]; !exists {
		t.Error("Node task-1 not found")
	}

	if _, exists := graph.Nodes["task-2"]; !exists {
		t.Error("Node task-2 not found")
	}

	// Adding duplicate should not create new node
	graph.AddNode("task-1")
	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes after duplicate add, got %d", len(graph.Nodes))
	}
}

func TestDependencyGraph_AddEdge(t *testing.T) {
	graph := NewDependencyGraph()

	graph.AddEdge("task-1", "task-2", 1)

	// Check nodes were created
	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}

	// Check edge was added
	if len(graph.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(graph.Edges))
	}

	edge := graph.Edges[0]
	if edge.From != "task-1" {
		t.Errorf("Expected edge from task-1, got %s", edge.From)
	}
	if edge.To != "task-2" {
		t.Errorf("Expected edge to task-2, got %s", edge.To)
	}
	if edge.Weight != 1 {
		t.Errorf("Expected edge weight 1, got %d", edge.Weight)
	}

	// Check node relationships
	node1 := graph.Nodes["task-1"]
	if len(node1.Dependents) != 1 || node1.Dependents[0] != "task-2" {
		t.Error("task-1 should have task-2 as dependent")
	}

	node2 := graph.Nodes["task-2"]
	if len(node2.Dependencies) != 1 || node2.Dependencies[0] != "task-1" {
		t.Error("task-2 should have task-1 as dependency")
	}
}

func TestDependencyGraph_HasCycle_NoCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create linear graph: 1 -> 2 -> 3
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-2", "task-3", 1)

	if graph.HasCycle() {
		t.Error("Expected no cycle in linear graph")
	}
}

func TestDependencyGraph_HasCycle_WithCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create cycle: 1 -> 2 -> 3 -> 1
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-2", "task-3", 1)
	graph.AddEdge("task-3", "task-1", 1)

	if !graph.HasCycle() {
		t.Error("Expected cycle to be detected")
	}
}

func TestDependencyGraph_HasCycle_SelfLoop(t *testing.T) {
	graph := NewDependencyGraph()

	// Create self-loop: 1 -> 1
	graph.AddEdge("task-1", "task-1", 1)

	if !graph.HasCycle() {
		t.Error("Expected self-loop cycle to be detected")
	}
}

func TestDependencyGraph_TopologicalSort_LinearGraph(t *testing.T) {
	graph := NewDependencyGraph()

	// Create linear graph: 1 -> 2 -> 3
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-2", "task-3", 1)

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 nodes in sorted order, got %d", len(sorted))
	}

	// Check order
	expected := []string{"task-1", "task-2", "task-3"}
	for i, taskID := range sorted {
		if taskID != expected[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expected[i], taskID)
		}
	}
}

func TestDependencyGraph_TopologicalSort_DAG(t *testing.T) {
	graph := NewDependencyGraph()

	// Create DAG:
	//     1
	//    / \
	//   2   3
	//    \ /
	//     4
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-1", "task-3", 1)
	graph.AddEdge("task-2", "task-4", 1)
	graph.AddEdge("task-3", "task-4", 1)

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 4 {
		t.Errorf("Expected 4 nodes in sorted order, got %d", len(sorted))
	}

	// Check task-1 comes first
	if sorted[0] != "task-1" {
		t.Errorf("Expected task-1 first, got %s", sorted[0])
	}

	// Check task-4 comes last
	if sorted[3] != "task-4" {
		t.Errorf("Expected task-4 last, got %s", sorted[3])
	}
}

func TestDependencyGraph_TopologicalSort_WithCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create cycle
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-2", "task-3", 1)
	graph.AddEdge("task-3", "task-1", 1)

	_, err := graph.TopologicalSort()
	if err == nil {
		t.Error("Expected error for graph with cycle")
	}
}

func TestDependencyGraph_GetLevel(t *testing.T) {
	graph := NewDependencyGraph()

	graph.AddNode("task-1")
	graph.Nodes["task-1"].Level = 0

	graph.AddNode("task-2")
	graph.Nodes["task-2"].Level = 1

	if level := graph.GetLevel("task-1"); level != 0 {
		t.Errorf("Expected level 0 for task-1, got %d", level)
	}

	if level := graph.GetLevel("task-2"); level != 1 {
		t.Errorf("Expected level 1 for task-2, got %d", level)
	}

	if level := graph.GetLevel("non-existent"); level != 0 {
		t.Errorf("Expected level 0 for non-existent node, got %d", level)
	}
}

func TestDependencyGraph_CalculateLevels(t *testing.T) {
	graph := NewDependencyGraph()

	// Create graph:
	//     1
	//    / \
	//   2   3
	//    \ /
	//     4
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-1", "task-3", 1)
	graph.AddEdge("task-2", "task-4", 1)
	graph.AddEdge("task-3", "task-4", 1)

	err := graph.CalculateLevels()
	if err != nil {
		t.Fatalf("CalculateLevels failed: %v", err)
	}

	// Check levels
	if graph.Nodes["task-1"].Level != 0 {
		t.Errorf("Expected task-1 at level 0, got %d", graph.Nodes["task-1"].Level)
	}

	if graph.Nodes["task-2"].Level != 1 {
		t.Errorf("Expected task-2 at level 1, got %d", graph.Nodes["task-2"].Level)
	}

	if graph.Nodes["task-3"].Level != 1 {
		t.Errorf("Expected task-3 at level 1, got %d", graph.Nodes["task-3"].Level)
	}

	if graph.Nodes["task-4"].Level != 2 {
		t.Errorf("Expected task-4 at level 2, got %d", graph.Nodes["task-4"].Level)
	}
}

func TestDependencyGraph_CalculateLevels_WithCycle(t *testing.T) {
	graph := NewDependencyGraph()

	// Create cycle
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-2", "task-1", 1)

	err := graph.CalculateLevels()
	if err == nil {
		t.Error("Expected error for graph with cycle")
	}
}

func TestDependencyGraph_GetParallelTasks(t *testing.T) {
	graph := NewDependencyGraph()

	// Create graph with multiple levels
	//     1
	//    /|\
	//   2 3 4  (level 1 - can run in parallel)
	//    \|/
	//     5   (level 2)
	graph.AddEdge("task-1", "task-2", 1)
	graph.AddEdge("task-1", "task-3", 1)
	graph.AddEdge("task-1", "task-4", 1)
	graph.AddEdge("task-2", "task-5", 1)
	graph.AddEdge("task-3", "task-5", 1)
	graph.AddEdge("task-4", "task-5", 1)

	err := graph.CalculateLevels()
	if err != nil {
		t.Fatalf("CalculateLevels failed: %v", err)
	}

	parallelGroups := graph.GetParallelTasks()

	if len(parallelGroups) != 3 {
		t.Errorf("Expected 3 levels, got %d", len(parallelGroups))
	}

	// Level 0: task-1
	if len(parallelGroups[0]) != 1 {
		t.Errorf("Level 0 should have 1 task, got %d", len(parallelGroups[0]))
	}

	// Level 1: task-2, task-3, task-4
	if len(parallelGroups[1]) != 3 {
		t.Errorf("Level 1 should have 3 tasks, got %d", len(parallelGroups[1]))
	}

	// Level 2: task-5
	if len(parallelGroups[2]) != 1 {
		t.Errorf("Level 2 should have 1 task, got %d", len(parallelGroups[2]))
	}
}

func BenchmarkAddNode(b *testing.B) {
	graph := NewDependencyGraph()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		graph.AddNode("task-" + string(rune(i)))
	}
}

func BenchmarkAddEdge(b *testing.B) {
	graph := NewDependencyGraph()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		from := "task-" + string(rune(i))
		to := "task-" + string(rune(i+1))
		graph.AddEdge(from, to, 1)
	}
}

func BenchmarkHasCycle(b *testing.B) {
	graph := NewDependencyGraph()

	// Create linear graph with 100 nodes
	for i := 0; i < 100; i++ {
		from := "task-" + string(rune(i))
		to := "task-" + string(rune(i+1))
		graph.AddEdge(from, to, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		graph.HasCycle()
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	graph := NewDependencyGraph()

	// Create linear graph with 100 nodes
	for i := 0; i < 100; i++ {
		from := "task-" + string(rune(i))
		to := "task-" + string(rune(i+1))
		graph.AddEdge(from, to, 1)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		graph.TopologicalSort()
	}
}
