package aggregator

import (
	"testing"
	"time"
)

func TestNewResultStore(t *testing.T) {
	store := NewResultStore()

	if store == nil {
		t.Fatal("NewResultStore returned nil")
	}

	if store.results == nil {
		t.Error("results map not initialized")
	}

	if store.byTask == nil {
		t.Error("byTask map not initialized")
	}
}

func TestResultStore_AddResult(t *testing.T) {
	store := NewResultStore()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data: map[string]interface{}{
			"key": "value",
		},
		CreatedAt: time.Now(),
	}

	err := store.AddResult(result)
	if err != nil {
		t.Fatalf("AddResult failed: %v", err)
	}

	// Check result was added
	retrieved, err := store.GetResult("result-001")
	if err != nil {
		t.Fatalf("GetResult failed: %v", err)
	}

	if retrieved.ID != result.ID {
		t.Errorf("Expected ID %s, got %s", result.ID, retrieved.ID)
	}

	// Check byTask index
	byTask := store.GetResultsByTask("task-001")
	if len(byTask) != 1 {
		t.Errorf("Expected 1 result for task, got %d", len(byTask))
	}
}

func TestResultStore_AddResult_Duplicate(t *testing.T) {
	store := NewResultStore()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data:    make(map[string]interface{}),
	}

	store.AddResult(result)

	err := store.AddResult(result)
	if err == nil {
		t.Error("Expected error when adding duplicate result")
	}
}

func TestResultStore_GetResult(t *testing.T) {
	store := NewResultStore()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data:    make(map[string]interface{}),
	}

	store.AddResult(result)

	retrieved, err := store.GetResult("result-001")
	if err != nil {
		t.Fatalf("GetResult failed: %v", err)
	}

	if retrieved.ID != "result-001" {
		t.Errorf("Expected ID result-001, got %s", retrieved.ID)
	}

	// Try to get non-existent result
	_, err = store.GetResult("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent result")
	}
}

func TestResultStore_GetResultsByTask(t *testing.T) {
	store := NewResultStore()

	// Add multiple results for same task
	for i := 0; i < 3; i++ {
		result := &TaskResult{
			ID:      string(rune('a' + i)),
			TaskID:  "task-001",
			AgentID: string(rune('1' + i)),
			Status:  ResultStatusPending,
			Data:    make(map[string]interface{}),
		}
		store.AddResult(result)
	}

	results := store.GetResultsByTask("task-001")
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestResultStore_UpdateResult(t *testing.T) {
	store := NewResultStore()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data:    make(map[string]interface{}),
	}

	store.AddResult(result)

	// Update result
	result.Status = ResultStatusValidated
	err := store.UpdateResult(result)
	if err != nil {
		t.Fatalf("UpdateResult failed: %v", err)
	}

	// Verify update
	retrieved, _ := store.GetResult("result-001")
	if retrieved.Status != ResultStatusValidated {
		t.Errorf("Expected status VALIDATED, got %s", retrieved.Status)
	}
}

func TestResultStore_DeleteResult(t *testing.T) {
	store := NewResultStore()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data:    make(map[string]interface{}),
	}

	store.AddResult(result)

	err := store.DeleteResult("result-001")
	if err != nil {
		t.Fatalf("DeleteResult failed: %v", err)
	}

	// Verify deletion
	_, err = store.GetResult("result-001")
	if err == nil {
		t.Error("Expected error after deletion")
	}

	// Check byTask index
	byTask := store.GetResultsByTask("task-001")
	if len(byTask) != 0 {
		t.Errorf("Expected 0 results after deletion, got %d", len(byTask))
	}
}

func TestResultStore_GetAllResults(t *testing.T) {
	store := NewResultStore()

	// Add multiple results
	for i := 0; i < 5; i++ {
		result := &TaskResult{
			ID:      string(rune('a' + i)),
			TaskID:  "task-001",
			AgentID: string(rune('1' + i)),
			Status:  ResultStatusPending,
			Data:    make(map[string]interface{}),
		}
		store.AddResult(result)
	}

	all := store.GetAllResults()
	if len(all) != 5 {
		t.Errorf("Expected 5 results, got %d", len(all))
	}
}

func TestResultStore_GetResultCount(t *testing.T) {
	store := NewResultStore()

	if store.GetResultCount() != 0 {
		t.Error("Expected 0 results initially")
	}

	// Add results
	for i := 0; i < 3; i++ {
		result := &TaskResult{
			ID:      string(rune('a' + i)),
			TaskID:  "task-001",
			AgentID: string(rune('1' + i)),
			Status:  ResultStatusPending,
			Data:    make(map[string]interface{}),
		}
		store.AddResult(result)
	}

	if store.GetResultCount() != 3 {
		t.Errorf("Expected 3 results, got %d", store.GetResultCount())
	}
}

func TestResultStore_GetResultsByStatus(t *testing.T) {
	store := NewResultStore()

	// Add results with different statuses
	statuses := []ResultStatus{
		ResultStatusPending,
		ResultStatusValidated,
		ResultStatusValidated,
		ResultStatusRejected,
	}

	for i, status := range statuses {
		result := &TaskResult{
			ID:      string(rune('a' + i)),
			TaskID:  "task-001",
			AgentID: string(rune('1' + i)),
			Status:  status,
			Data:    make(map[string]interface{}),
		}
		store.AddResult(result)
	}

	validated := store.GetResultsByStatus(ResultStatusValidated)
	if len(validated) != 2 {
		t.Errorf("Expected 2 validated results, got %d", len(validated))
	}

	pending := store.GetResultsByStatus(ResultStatusPending)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending result, got %d", len(pending))
	}
}

func TestSerializeResult(t *testing.T) {
	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusValidated,
		Data: map[string]interface{}{
			"key": "value",
		},
		CreatedAt: time.Now(),
		Score:     85.5,
	}

	data, err := SerializeResult(result)
	if err != nil {
		t.Fatalf("SerializeResult failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}
}

func TestDeserializeResult(t *testing.T) {
	jsonData := []byte(`{
		"id": "result-001",
		"task_id": "task-001",
		"agent_id": "agent-001",
		"status": "VALIDATED",
		"data": {"key": "value"},
		"metadata": {},
		"error": "",
		"created_at": "2026-01-29T10:00:00Z",
		"score": 85.5
	}`)

	result, err := DeserializeResult(jsonData)
	if err != nil {
		t.Fatalf("DeserializeResult failed: %v", err)
	}

	if result.ID != "result-001" {
		t.Errorf("Expected ID result-001, got %s", result.ID)
	}

	if result.Status != ResultStatusValidated {
		t.Errorf("Expected status VALIDATED, got %s", result.Status)
	}

	if result.Score != 85.5 {
		t.Errorf("Expected score 85.5, got %.2f", result.Score)
	}
}

func TestSerializeAggregatedResult(t *testing.T) {
	aggregated := &AggregatedResult{
		TaskID: "task-001",
		Results: []*TaskResult{
			{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Status:  ResultStatusValidated,
				Data:    make(map[string]interface{}),
			},
		},
		MergedData: map[string]interface{}{
			"key": "value",
		},
		Conflicts:  make([]*Conflict, 0),
		Strategy:   "VOTING",
		Confidence: 0.85,
		CreatedAt:  time.Now(),
	}

	data, err := SerializeAggregatedResult(aggregated)
	if err != nil {
		t.Fatalf("SerializeAggregatedResult failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Serialized data is empty")
	}
}

func BenchmarkResultStore_AddResult(b *testing.B) {
	store := NewResultStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &TaskResult{
			ID:      string(rune(i)),
			TaskID:  "task-001",
			AgentID: string(rune(i + 1000)),
			Status:  ResultStatusPending,
			Data:    make(map[string]interface{}),
		}
		store.AddResult(result)
	}
}

func BenchmarkResultStore_GetResult(b *testing.B) {
	store := NewResultStore()

	// Setup
	for i := 0; i < 100; i++ {
		result := &TaskResult{
			ID:      string(rune('a' + i)),
			TaskID:  "task-001",
			AgentID: string(rune('1' + i)),
			Status:  ResultStatusPending,
			Data:    make(map[string]interface{}),
		}
		store.AddResult(result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetResult(string(rune('a' + (i % 100))))
	}
}
