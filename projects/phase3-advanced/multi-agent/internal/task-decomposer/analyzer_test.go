package decomposer

import (
	"testing"
)

func TestNewComplexityAnalyzer(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	if analyzer == nil {
		t.Fatal("NewComplexityAnalyzer returned nil")
	}

	// Check default weights
	if analyzer.weights.DependencyWeight != 0.3 {
		t.Errorf("Expected DependencyWeight 0.3, got %.2f", analyzer.weights.DependencyWeight)
	}
	if analyzer.weights.CapabilityWeight != 0.3 {
		t.Errorf("Expected CapabilityWeight 0.3, got %.2f", analyzer.weights.CapabilityWeight)
	}
	if analyzer.weights.RequirementWeight != 0.2 {
		t.Errorf("Expected RequirementWeight 0.2, got %.2f", analyzer.weights.RequirementWeight)
	}
	if analyzer.weights.TypeWeight != 0.2 {
		t.Errorf("Expected TypeWeight 0.2, got %.2f", analyzer.weights.TypeWeight)
	}
}

func TestAnalyze_SimpleTask(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("simple-001", "simple_query", "Simple task")

	complexity := analyzer.Analyze(task)

	if complexity != ComplexitySimple {
		t.Errorf("Expected ComplexitySimple, got %v", complexity)
	}
}

func TestAnalyze_ModerateTask(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("moderate-001", "calculation", "Moderate task")
	task.AddDependency("dep-1")
	task.AddCapability("math")

	complexity := analyzer.Analyze(task)

	if complexity != ComplexityModerate {
		t.Errorf("Expected ComplexityModerate, got %v", complexity)
	}
}

func TestAnalyze_ComplexTask(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("complex-001", "refactoring", "Complex task")
	task.AddDependency("dep-1")
	task.AddDependency("dep-2")
	task.AddDependency("dep-3")
	task.AddCapability("code_analysis")
	task.AddCapability("refactoring")

	complexity := analyzer.Analyze(task)

	if complexity != ComplexityComplex {
		t.Errorf("Expected ComplexityComplex, got %v", complexity)
	}
}

func TestAnalyze_VeryComplexTask(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("very-complex-001", "system_design", "Very complex task")
	for i := 0; i < 6; i++ {
		task.AddDependency("dep-" + string(rune('1'+i)))
	}
	for i := 0; i < 4; i++ {
		task.AddCapability("cap-" + string(rune('1'+i)))
	}
	for i := 0; i < 6; i++ {
		task.SetRequirement("req-"+string(rune('1'+i)), "value")
	}

	complexity := analyzer.Analyze(task)

	if complexity != ComplexityVeryComplex {
		t.Errorf("Expected ComplexityVeryComplex, got %v", complexity)
	}
}

func TestGetTypeComplexity(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		taskType string
		expected float64
	}{
		{"simple_query", 1.0},
		{"calculation", 1.5},
		{"document_processing", 2.5},
		{"code_review", 3.0},
		{"data_analysis", 3.5},
		{"refactoring", 4.0},
		{"system_design", 5.0},
		{"unknown_type", 2.0}, // default
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			complexity := analyzer.getTypeComplexity(tt.taskType)
			if complexity != tt.expected {
				t.Errorf("Type %s: expected %.2f, got %.2f", tt.taskType, tt.expected, complexity)
			}
		})
	}
}

func TestGetRecommendedStrategy(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name     string
		setup    func(*Task)
		expected DecompositionStrategy
	}{
		{
			name: "VeryComplex -> Hybrid",
			setup: func(task *Task) {
				task.Type = "system_design"
				for i := 0; i < 6; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
			},
			expected: StrategyHybrid,
		},
		{
			name: "Complex with dependencies -> Dependency",
			setup: func(task *Task) {
				task.Type = "refactoring"
				task.AddDependency("dep-1")
				task.AddDependency("dep-2")
				task.AddCapability("cap-1")
			},
			expected: StrategyDependency,
		},
		{
			name: "Complex without dependencies -> Capability",
			setup: func(task *Task) {
				task.Type = "refactoring"
				task.AddCapability("cap-1")
				task.AddCapability("cap-2")
			},
			expected: StrategyCapability,
		},
		{
			name: "Moderate with capabilities -> Capability",
			setup: func(task *Task) {
				task.Type = "calculation"
				task.AddCapability("math")
			},
			expected: StrategyCapability,
		},
		{
			name: "Moderate without capabilities -> Priority",
			setup: func(task *Task) {
				task.Type = "calculation"
			},
			expected: StrategyPriority,
		},
		{
			name: "Simple -> Priority",
			setup: func(task *Task) {
				task.Type = "simple_query"
			},
			expected: StrategyPriority,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("test-001", "test", "Test task")
			tt.setup(task)

			strategy := analyzer.GetRecommendedStrategy(task)
			if strategy != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, strategy)
			}
		})
	}
}

func TestEstimateSubTaskCount(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name     string
		setup    func(*Task)
		minCount int
		maxCount int
	}{
		{
			name: "Simple task",
			setup: func(task *Task) {
				task.Type = "simple_query"
			},
			minCount: 1,
			maxCount: 1,
		},
		{
			name: "Moderate task",
			setup: func(task *Task) {
				task.Type = "calculation"
				task.AddDependency("dep-1")
			},
			minCount: 3,
			maxCount: 5,
		},
		{
			name: "Complex task",
			setup: func(task *Task) {
				task.Type = "refactoring"
				task.AddDependency("dep-1")
				task.AddDependency("dep-2")
				task.AddCapability("cap-1")
				task.AddCapability("cap-2")
			},
			minCount: 5,
			maxCount: 10,
		},
		{
			name: "Very complex task",
			setup: func(task *Task) {
				task.Type = "system_design"
				for i := 0; i < 6; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
				for i := 0; i < 4; i++ {
					task.AddCapability("cap-" + string(rune('1'+i)))
				}
			},
			minCount: 8,
			maxCount: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("test-001", "test", "Test task")
			tt.setup(task)

			count := analyzer.EstimateSubTaskCount(task)
			if count < tt.minCount || count > tt.maxCount {
				t.Errorf("Expected count between %d and %d, got %d", tt.minCount, tt.maxCount, count)
			}
		})
	}
}

func TestGenerateReport(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("report-001", "refactoring", "Test task")
	task.AddDependency("dep-1")
	task.AddDependency("dep-2")
	task.AddCapability("code_analysis")
	task.AddCapability("refactoring")
	task.SetRequirement("quality", "high")

	report := analyzer.GenerateReport(task)

	// Check basic fields
	if report.TaskID != task.ID {
		t.Errorf("Expected TaskID %s, got %s", task.ID, report.TaskID)
	}

	if report.Complexity == 0 {
		t.Error("Expected non-zero complexity")
	}

	if report.Score == 0 {
		t.Error("Expected non-zero score")
	}

	if report.RecommendedStrategy == "" {
		t.Error("Expected recommended strategy")
	}

	if report.EstimatedSubTasks == 0 {
		t.Error("Expected non-zero estimated sub-tasks")
	}

	// Check factors
	if report.Factors == nil {
		t.Fatal("Expected factors to be set")
	}

	expectedFactors := []string{"dependencies", "capabilities", "requirements", "type_complexity"}
	for _, factor := range expectedFactors {
		if _, exists := report.Factors[factor]; !exists {
			t.Errorf("Expected factor %s to exist", factor)
		}
	}

	// Check recommendations
	if report.Recommendations == nil {
		t.Fatal("Expected recommendations to be set")
	}
}

func TestGenerateReport_Recommendations(t *testing.T) {
	analyzer := NewComplexityAnalyzer()

	tests := []struct {
		name              string
		setup             func(*Task)
		expectedRecCount  int
		containsKeyword   string
	}{
		{
			name: "Many dependencies",
			setup: func(task *Task) {
				for i := 0; i < 6; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
			},
			expectedRecCount: 1,
			containsKeyword:  "dependencies",
		},
		{
			name: "Many capabilities",
			setup: func(task *Task) {
				for i := 0; i < 4; i++ {
					task.AddCapability("cap-" + string(rune('1'+i)))
				}
			},
			expectedRecCount: 1,
			containsKeyword:  "capabilities",
		},
		{
			name: "Very complex",
			setup: func(task *Task) {
				task.Type = "system_design"
				for i := 0; i < 6; i++ {
					task.AddDependency("dep-" + string(rune('1'+i)))
				}
				for i := 0; i < 4; i++ {
					task.AddCapability("cap-" + string(rune('1'+i)))
				}
			},
			expectedRecCount: 3,
			containsKeyword:  "complex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("test-001", "test", "Test task")
			tt.setup(task)

			report := analyzer.GenerateReport(task)

			if len(report.Recommendations) < tt.expectedRecCount {
				t.Errorf("Expected at least %d recommendations, got %d", tt.expectedRecCount, len(report.Recommendations))
			}
		})
	}
}

func TestNewSubTaskGenerator(t *testing.T) {
	generator := NewSubTaskGenerator()

	if generator == nil {
		t.Fatal("NewSubTaskGenerator returned nil")
	}

	if generator.idCounter != 0 {
		t.Errorf("Expected idCounter 0, got %d", generator.idCounter)
	}
}

func TestGenerate(t *testing.T) {
	generator := NewSubTaskGenerator()

	task := NewTask("parent-001", "test", "Parent task")
	task.Priority = 5

	count := 3
	subTasks, err := generator.Generate(task, count)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(subTasks) != count {
		t.Errorf("Expected %d sub-tasks, got %d", count, len(subTasks))
	}

	// Check first task has no dependencies
	if len(subTasks[0].Dependencies) != 0 {
		t.Error("First sub-task should have no dependencies")
	}

	// Check sequential dependencies
	for i := 1; i < len(subTasks); i++ {
		if len(subTasks[i].Dependencies) != 1 {
			t.Errorf("Sub-task %d should have 1 dependency, got %d", i, len(subTasks[i].Dependencies))
		}
		if subTasks[i].Dependencies[0] != subTasks[i-1].ID {
			t.Errorf("Sub-task %d should depend on task %d", i, i-1)
		}
	}

	// Check levels
	for i, subTask := range subTasks {
		if subTask.Level != i {
			t.Errorf("Sub-task %d should be at level %d, got %d", i, i, subTask.Level)
		}
	}
}

func TestGenerate_InvalidCount(t *testing.T) {
	generator := NewSubTaskGenerator()
	task := NewTask("parent-001", "test", "Parent task")

	_, err := generator.Generate(task, 0)
	if err == nil {
		t.Error("Expected error for count 0")
	}

	_, err = generator.Generate(task, -1)
	if err == nil {
		t.Error("Expected error for negative count")
	}
}

func TestGenerateWithPattern_Sequential(t *testing.T) {
	generator := NewSubTaskGenerator()
	task := NewTask("parent-001", "test", "Parent task")

	subTasks, err := generator.GenerateWithPattern(task, "sequential")
	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	// Should generate: prepare, execute, verify
	if len(subTasks) != 3 {
		t.Errorf("Expected 3 sub-tasks, got %d", len(subTasks))
	}

	expectedPhases := []string{"prepare", "execute", "verify"}
	for i, subTask := range subTasks {
		if subTask.Type != expectedPhases[i] {
			t.Errorf("Phase %d should be %s, got %s", i, expectedPhases[i], subTask.Type)
		}
	}

	// Check dependencies
	for i := 1; i < len(subTasks); i++ {
		if len(subTasks[i].Dependencies) != 1 {
			t.Errorf("Sub-task %d should have 1 dependency", i)
		}
	}
}

func TestGenerateWithPattern_Parallel(t *testing.T) {
	generator := NewSubTaskGenerator()
	task := NewTask("parent-001", "test", "Parent task")
	task.AddCapability("cap1")
	task.AddCapability("cap2")

	subTasks, err := generator.GenerateWithPattern(task, "parallel")
	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	// Should generate tasks based on capabilities
	if len(subTasks) != 2 {
		t.Errorf("Expected 2 sub-tasks, got %d", len(subTasks))
	}

	// Check all tasks are at level 0 (parallel)
	for i, subTask := range subTasks {
		if subTask.Level != 0 {
			t.Errorf("Parallel task %d should be at level 0, got %d", i, subTask.Level)
		}
		if len(subTask.Dependencies) != 0 {
			t.Errorf("Parallel task %d should have no dependencies", i)
		}
	}
}

func TestGenerateWithPattern_Pipeline(t *testing.T) {
	generator := NewSubTaskGenerator()
	task := NewTask("parent-001", "test", "Parent task")

	subTasks, err := generator.GenerateWithPattern(task, "pipeline")
	if err != nil {
		t.Fatalf("GenerateWithPattern failed: %v", err)
	}

	// Should generate: input, process, output
	if len(subTasks) != 3 {
		t.Errorf("Expected 3 sub-tasks, got %d", len(subTasks))
	}

	expectedStages := []string{"input", "process", "output"}
	for i, subTask := range subTasks {
		if subTask.Type != expectedStages[i] {
			t.Errorf("Stage %d should be %s, got %s", i, expectedStages[i], subTask.Type)
		}
	}

	// Check pipeline dependencies
	for i := 1; i < len(subTasks); i++ {
		if len(subTasks[i].Dependencies) != 1 {
			t.Errorf("Pipeline stage %d should have 1 dependency", i)
		}
	}
}

func TestGenerateWithPattern_UnknownPattern(t *testing.T) {
	generator := NewSubTaskGenerator()
	task := NewTask("parent-001", "test", "Parent task")

	_, err := generator.GenerateWithPattern(task, "unknown_pattern")
	if err == nil {
		t.Error("Expected error for unknown pattern")
	}
}

func BenchmarkAnalyze(b *testing.B) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("bench-001", "refactoring", "Benchmark task")
	task.AddDependency("dep-1")
	task.AddDependency("dep-2")
	task.AddCapability("cap-1")
	task.AddCapability("cap-2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze(task)
	}
}

func BenchmarkGenerateReport(b *testing.B) {
	analyzer := NewComplexityAnalyzer()

	task := NewTask("bench-001", "refactoring", "Benchmark task")
	task.AddDependency("dep-1")
	task.AddDependency("dep-2")
	task.AddCapability("cap-1")
	task.AddCapability("cap-2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.GenerateReport(task)
	}
}

func BenchmarkGenerate(b *testing.B) {
	generator := NewSubTaskGenerator()
	task := NewTask("bench-001", "test", "Benchmark task")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generator.Generate(task, 5)
	}
}
