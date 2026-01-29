package decomposer

import (
	"testing"
)

func TestNewDecomposer(t *testing.T) {
	config := DefaultConfig()
	d := NewDecomposer(config)

	if d == nil {
		t.Fatal("NewDecomposer returned nil")
	}

	if d.config.Strategy != StrategyHybrid {
		t.Errorf("Expected strategy %s, got %s", StrategyHybrid, d.config.Strategy)
	}

	if len(d.rules) == 0 {
		t.Error("Expected default rules to be registered")
	}
}

func TestDecompose_SimpleTask(t *testing.T) {
	config := DefaultConfig()
	config.ComplexityAnalysis = true
	d := NewDecomposer(config)

	task := NewTask("simple-001", "simple_query", "What is 2+2?")

	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	if len(result.SubTasks) != 1 {
		t.Errorf("Simple task should not be decomposed, got %d sub-tasks", len(result.SubTasks))
	}
}

func TestDecompose_ByDependency(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyDependency,
		MaxDepth: 3,
	}
	d := NewDecomposer(config)

	task := NewTask("dep-001", "complex_task", "Task with dependencies")
	task.AddDependency("dep-1")
	task.AddDependency("dep-2")

	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Should create: 2 dependency tasks + 1 main task = 3 sub-tasks
	if len(result.SubTasks) != 3 {
		t.Errorf("Expected 3 sub-tasks, got %d", len(result.SubTasks))
	}

	// Check main task dependencies
	mainTask := result.SubTasks[2]
	if len(mainTask.Dependencies) != 2 {
		t.Errorf("Main task should have 2 dependencies, got %d", len(mainTask.Dependencies))
	}
}

func TestDecompose_ByPriority(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyPriority,
		MaxDepth: 3,
	}
	d := NewDecomposer(config)

	task := NewTask("priority-001", "task", "Priority-based task")
	task.Priority = 5

	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Should create 3 phases: preparation, execution, verification
	if len(result.SubTasks) != 3 {
		t.Errorf("Expected 3 sub-tasks, got %d", len(result.SubTasks))
	}

	// Check phase order
	phases := []string{"preparation", "execution", "verification"}
	for i, subTask := range result.SubTasks {
		if subTask.Type != phases[i] {
			t.Errorf("Phase %d should be %s, got %s", i, phases[i], subTask.Type)
		}
	}

	// Check dependencies
	if len(result.SubTasks[0].Dependencies) != 0 {
		t.Error("First phase should have no dependencies")
	}
	if len(result.SubTasks[1].Dependencies) != 1 {
		t.Error("Second phase should have 1 dependency")
	}
	if len(result.SubTasks[2].Dependencies) != 1 {
		t.Error("Third phase should have 1 dependency")
	}
}

func TestDecompose_ByCapability(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyCapability,
		MaxDepth: 3,
	}
	d := NewDecomposer(config)

	task := NewTask("cap-001", "analysis", "Capability-based task")
	task.AddCapability("syntax_check")
	task.AddCapability("quality_check")
	task.AddCapability("security_check")

	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Should create: 3 capability tasks + 1 aggregate task = 4 sub-tasks
	if len(result.SubTasks) != 4 {
		t.Errorf("Expected 4 sub-tasks, got %d", len(result.SubTasks))
	}

	// Check aggregate task
	aggregateTask := result.SubTasks[3]
	if aggregateTask.Type != "aggregate" {
		t.Errorf("Last task should be aggregate, got %s", aggregateTask.Type)
	}
	if len(aggregateTask.Dependencies) != 3 {
		t.Errorf("Aggregate task should have 3 dependencies, got %d", len(aggregateTask.Dependencies))
	}
}

func TestDecompose_Hybrid_CodeReview(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyHybrid,
		MaxDepth: 3,
	}
	d := NewDecomposer(config)

	task := NewTask("review-001", "code_review", "Review code")

	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Should apply code_review rule: syntax, quality, security checks
	if len(result.SubTasks) != 3 {
		t.Errorf("Expected 3 sub-tasks for code review, got %d", len(result.SubTasks))
	}

	// Check task types
	expectedTypes := map[string]bool{
		"syntax_check":   true,
		"quality_check":  true,
		"security_check": true,
	}

	for _, subTask := range result.SubTasks {
		if !expectedTypes[subTask.Type] {
			t.Errorf("Unexpected sub-task type: %s", subTask.Type)
		}
	}
}

func TestDecompose_Hybrid_DocumentProcessing(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyHybrid,
		MaxDepth: 3,
	}
	d := NewDecomposer(config)

	task := NewTask("doc-001", "document_processing", "Process document")

	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Should apply document_processing rule: parse, analyze, summarize
	if len(result.SubTasks) != 3 {
		t.Errorf("Expected 3 sub-tasks for document processing, got %d", len(result.SubTasks))
	}

	// Check task types
	expectedTypes := []string{"parse", "analyze", "summarize"}
	for i, subTask := range result.SubTasks {
		if subTask.Type != expectedTypes[i] {
			t.Errorf("Sub-task %d should be %s, got %s", i, expectedTypes[i], subTask.Type)
		}
	}

	// Check levels
	for i, subTask := range result.SubTasks {
		if subTask.Level != i {
			t.Errorf("Sub-task %d should be at level %d, got %d", i, i, subTask.Level)
		}
	}
}

func TestDecompose_InvalidTask(t *testing.T) {
	d := NewDecomposer(nil)

	tests := []struct {
		name string
		task *Task
	}{
		{"nil task", nil},
		{"empty ID", &Task{Type: "test"}},
		{"empty Type", &Task{ID: "test-001"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.Decompose(tt.task)
			if err == nil {
				t.Error("Expected error for invalid task")
			}
		})
	}
}

func TestDecompose_DependencyGraph(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyPriority,
	}
	d := NewDecomposer(config)

	task := NewTask("graph-001", "task", "Task with graph")
	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Check graph exists
	if result.Graph == nil {
		t.Fatal("Expected dependency graph to be created")
	}

	// Check all sub-tasks are in graph
	for _, subTask := range result.SubTasks {
		if _, exists := result.Graph.Nodes[subTask.ID]; !exists {
			t.Errorf("Sub-task %s not in graph", subTask.ID)
		}
	}

	// Check levels are calculated
	for _, subTask := range result.SubTasks {
		level := result.Graph.GetLevel(subTask.ID)
		if level != subTask.Level {
			t.Errorf("Level mismatch for %s: graph=%d, task=%d", subTask.ID, level, subTask.Level)
		}
	}

	// Check no cycles
	if result.Graph.HasCycle() {
		t.Error("Graph should not have cycles")
	}
}

func TestRegisterRule(t *testing.T) {
	d := NewDecomposer(nil)
	initialRuleCount := len(d.rules)

	rule := &DecompositionRule{
		Name:     "test_rule",
		TaskType: "test",
		Condition: func(t *Task) bool {
			return t.Type == "test"
		},
		Decompose: func(t *Task) ([]*SubTask, error) {
			return []*SubTask{{ID: "test-1"}}, nil
		},
		Priority: 5,
	}

	d.RegisterRule(rule)

	if len(d.rules) != initialRuleCount+1 {
		t.Errorf("Expected %d rules, got %d", initialRuleCount+1, len(d.rules))
	}
}

func TestDecompose_CustomRule(t *testing.T) {
	config := &DecomposerConfig{
		Strategy: StrategyHybrid,
	}
	d := NewDecomposer(config)

	// Register custom rule
	rule := &DecompositionRule{
		Name:     "custom_task",
		TaskType: "custom",
		Condition: func(t *Task) bool {
			return t.Type == "custom"
		},
		Decompose: func(t *Task) ([]*SubTask, error) {
			return []*SubTask{
				{
					ID:          t.ID + "-sub-1",
					ParentID:    t.ID,
					Type:        "custom_step_1",
					Description: "Step 1",
					Level:       0,
				},
				{
					ID:          t.ID + "-sub-2",
					ParentID:    t.ID,
					Type:        "custom_step_2",
					Description: "Step 2",
					Level:       1,
				},
			}, nil
		},
		Priority: 10,
	}
	d.RegisterRule(rule)

	// Test custom task
	task := NewTask("custom-001", "custom", "Custom task")
	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	if len(result.SubTasks) != 2 {
		t.Errorf("Expected 2 sub-tasks, got %d", len(result.SubTasks))
	}

	if result.SubTasks[0].Type != "custom_step_1" {
		t.Errorf("First sub-task should be custom_step_1, got %s", result.SubTasks[0].Type)
	}
}

func TestDecompose_Metadata(t *testing.T) {
	d := NewDecomposer(nil)

	task := NewTask("meta-001", "task", "Task with metadata")
	result, err := d.Decompose(task)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Check metadata exists
	if result.Metadata == nil {
		t.Fatal("Expected metadata to be created")
	}

	// Check sub_task_count
	subTaskCount, ok := result.Metadata["sub_task_count"].(int)
	if !ok {
		t.Error("Expected sub_task_count in metadata")
	}
	if subTaskCount != len(result.SubTasks) {
		t.Errorf("Metadata sub_task_count=%d, actual=%d", subTaskCount, len(result.SubTasks))
	}

	// Check max_level
	if _, ok := result.Metadata["max_level"]; !ok {
		t.Error("Expected max_level in metadata")
	}
}

func BenchmarkDecompose_ByDependency(b *testing.B) {
	config := &DecomposerConfig{
		Strategy: StrategyDependency,
	}
	d := NewDecomposer(config)

	task := NewTask("bench-001", "task", "Benchmark task")
	task.AddDependency("dep-1")
	task.AddDependency("dep-2")
	task.AddDependency("dep-3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decompose(task)
		if err != nil {
			b.Fatalf("Decompose failed: %v", err)
		}
	}
}

func BenchmarkDecompose_ByCapability(b *testing.B) {
	config := &DecomposerConfig{
		Strategy: StrategyCapability,
	}
	d := NewDecomposer(config)

	task := NewTask("bench-002", "task", "Benchmark task")
	task.AddCapability("cap1")
	task.AddCapability("cap2")
	task.AddCapability("cap3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.Decompose(task)
		if err != nil {
			b.Fatalf("Decompose failed: %v", err)
		}
	}
}
