package aggregator

import (
	"testing"
)

func TestNewResultMerger(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	if merger == nil {
		t.Fatal("NewResultMerger returned nil")
	}

	if merger.strategy != MergeStrategyVoting {
		t.Errorf("Expected strategy VOTING, got %s", merger.strategy)
	}

	if merger.conflictStrategy != ConflictResolutionVoting {
		t.Errorf("Expected conflict strategy VOTING, got %s", merger.conflictStrategy)
	}
}

func TestResultMerger_SetMinResults(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	merger.SetMinResults(3)

	if merger.minResults != 3 {
		t.Errorf("Expected minResults 3, got %d", merger.minResults)
	}
}

func TestResultMerger_Merge_InsufficientResults(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)
	merger.SetMinResults(3)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data:    make(map[string]interface{}),
		},
	}

	_, err := merger.Merge("task-001", results)
	if err == nil {
		t.Error("Expected error for insufficient results")
	}
}

func TestResultMerger_Merge_NoValidatedResults(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusPending, // Not validated
			Data:    make(map[string]interface{}),
		},
	}

	_, err := merger.Merge("task-001", results)
	if err == nil {
		t.Error("Expected error for no validated results")
	}
}

func TestResultMerger_MergeByVoting(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 85,
		},
		{
			ID:      "result-003",
			TaskID:  "task-001",
			AgentID: "agent-003",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "B",
			},
			Score: 75,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if aggregated == nil {
		t.Fatal("Aggregated result is nil")
	}

	// Answer "A" should win (2 votes vs 1 vote)
	if aggregated.MergedData["answer"] != "A" {
		t.Errorf("Expected answer A, got %v", aggregated.MergedData["answer"])
	}
}

func TestResultMerger_MergeByAveraging(t *testing.T) {
	merger := NewResultMerger(MergeStrategyAveraging, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"value": 10,
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"value": 20,
			},
			Score: 85,
		},
		{
			ID:      "result-003",
			TaskID:  "task-001",
			AgentID: "agent-003",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"value": 30,
			},
			Score: 75,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Average should be (10 + 20 + 30) / 3 = 20
	avgValue := aggregated.MergedData["value"].(float64)
	if avgValue != 20.0 {
		t.Errorf("Expected average 20.0, got %.2f", avgValue)
	}
}

func TestResultMerger_MergeByWeighted(t *testing.T) {
	merger := NewResultMerger(MergeStrategyWeighted, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"value": 10.0,
			},
			Score: 50, // Weight 50
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"value": 20.0,
			},
			Score: 100, // Weight 100
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Weighted average = (10*50 + 20*100) / (50+100) = 2500 / 150 = 16.67
	avgValue := aggregated.MergedData["value"].(float64)
	expected := (10.0*50 + 20.0*100) / (50 + 100)
	if avgValue != expected {
		t.Errorf("Expected weighted average %.2f, got %.2f", expected, avgValue)
	}
}

func TestResultMerger_MergeByConsensus(t *testing.T) {
	merger := NewResultMerger(MergeStrategyConsensus, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"agreed":    "yes",
				"disagreed": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"agreed":    "yes",
				"disagreed": "B",
			},
			Score: 85,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Only "agreed" should be in merged data (both agents agree)
	if aggregated.MergedData["agreed"] != "yes" {
		t.Error("Expected 'agreed' field in merged data")
	}

	// "disagreed" should not be in merged data
	if _, exists := aggregated.MergedData["disagreed"]; exists {
		t.Error("'disagreed' field should not be in merged data")
	}
}

func TestResultMerger_MergeByPriority(t *testing.T) {
	merger := NewResultMerger(MergeStrategyPriority, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "B",
			},
			Score: 95, // Highest score
		},
		{
			ID:      "result-003",
			TaskID:  "task-001",
			AgentID: "agent-003",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "C",
			},
			Score: 70,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should use result with highest score (95)
	if aggregated.MergedData["answer"] != "B" {
		t.Errorf("Expected answer B (highest score), got %v", aggregated.MergedData["answer"])
	}
}

func TestResultMerger_DetectConflicts(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "B",
			},
			Score: 85,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should detect conflict in "answer" field
	if len(aggregated.Conflicts) == 0 {
		t.Error("Expected conflicts to be detected")
	}

	if aggregated.Conflicts[0].Field != "answer" {
		t.Errorf("Expected conflict in 'answer' field, got %s", aggregated.Conflicts[0].Field)
	}
}

func TestResultMerger_ResolveConflictByVoting(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 85,
		},
		{
			ID:      "result-003",
			TaskID:  "task-001",
			AgentID: "agent-003",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "B",
			},
			Score: 90,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Conflict should be resolved
	if len(aggregated.Conflicts) > 0 {
		conflict := aggregated.Conflicts[0]
		if conflict.Resolution == "" {
			t.Error("Conflict should have resolution")
		}
		if conflict.ResolvedAt == nil {
			t.Error("Conflict should have resolved time")
		}
	}
}

func TestResultMerger_ResolveConflictByHighScore(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionHighScore)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "B",
			},
			Score: 95, // Highest score
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should have conflict
	if len(aggregated.Conflicts) == 0 {
		t.Fatal("Expected conflict")
	}

	// Resolution should mention agent-002 (highest score)
	if aggregated.Conflicts[0].Resolution == "" {
		t.Error("Conflict should be resolved")
	}
}

func TestResultMerger_CalculateConfidence(t *testing.T) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 90,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 85,
		},
	}

	aggregated, err := merger.Merge("task-001", results)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Confidence should be calculated
	if aggregated.Confidence < 0 || aggregated.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %.2f", aggregated.Confidence)
	}

	// High scores and no conflicts should give high confidence
	if aggregated.Confidence < 0.5 {
		t.Errorf("Expected high confidence for good results, got %.2f", aggregated.Confidence)
	}
}

func TestNewResultAggregator(t *testing.T) {
	aggregator := NewResultAggregator(MergeStrategyVoting, ConflictResolutionVoting)

	if aggregator == nil {
		t.Fatal("NewResultAggregator returned nil")
	}

	if aggregator.store == nil {
		t.Error("store not initialized")
	}

	if aggregator.validator == nil {
		t.Error("validator not initialized")
	}

	if aggregator.merger == nil {
		t.Error("merger not initialized")
	}
}

func TestResultAggregator_AddResult(t *testing.T) {
	aggregator := NewResultAggregator(MergeStrategyVoting, ConflictResolutionVoting)

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Data:    make(map[string]interface{}),
		Score:   85,
	}

	err := aggregator.AddResult(result)
	if err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	// Result should be validated
	if result.Status != ResultStatusValidated {
		t.Errorf("Expected status VALIDATED, got %s", result.Status)
	}

	// Should be in store
	retrieved, err := aggregator.GetResult("result-001")
	if err != nil {
		t.Fatalf("GetResult failed: %v", err)
	}

	if retrieved.ID != "result-001" {
		t.Error("Result not stored correctly")
	}
}

func TestResultAggregator_AggregateTask(t *testing.T) {
	aggregator := NewResultAggregator(MergeStrategyVoting, ConflictResolutionVoting)

	// Add multiple results
	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 85,
		},
	}

	for _, result := range results {
		aggregator.AddResult(result)
	}

	// Aggregate
	aggregated, err := aggregator.AggregateTask("task-001")
	if err != nil {
		t.Fatalf("AggregateTask failed: %v", err)
	}

	if aggregated.TaskID != "task-001" {
		t.Error("Wrong task ID in aggregated result")
	}

	if len(aggregated.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(aggregated.Results))
	}

	if aggregated.MergedData["answer"] != "A" {
		t.Error("Merged data incorrect")
	}
}

func BenchmarkResultMerger_MergeByVoting(b *testing.B) {
	merger := NewResultMerger(MergeStrategyVoting, ConflictResolutionVoting)

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 80,
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Status:  ResultStatusValidated,
			Data: map[string]interface{}{
				"answer": "A",
			},
			Score: 85,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger.Merge("task-001", results)
	}
}
