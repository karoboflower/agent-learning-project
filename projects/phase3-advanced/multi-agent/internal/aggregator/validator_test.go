package aggregator

import (
	"testing"
)

func TestNewResultValidator(t *testing.T) {
	validator := NewResultValidator()

	if validator == nil {
		t.Fatal("NewResultValidator returned nil")
	}

	if validator.rules == nil {
		t.Error("rules slice not initialized")
	}
}

func TestResultValidator_Validate_BasicFields(t *testing.T) {
	validator := NewResultValidator()

	tests := []struct {
		name    string
		result  *TaskResult
		wantErr bool
	}{
		{
			name: "valid result",
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data:    make(map[string]interface{}),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			result: &TaskResult{
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data:    make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "missing TaskID",
			result: &TaskResult{
				ID:      "result-001",
				AgentID: "agent-001",
				Data:    make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "missing AgentID",
			result: &TaskResult{
				ID:     "result-001",
				TaskID: "task-001",
				Data:   make(map[string]interface{}),
			},
			wantErr: true,
		},
		{
			name: "nil Data",
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data:    nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequiredFieldsRule_Validate(t *testing.T) {
	rule := &RequiredFieldsRule{
		Fields: []string{"field1", "field2"},
	}

	tests := []struct {
		name    string
		result  *TaskResult
		wantErr bool
	}{
		{
			name: "all fields present",
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data: map[string]interface{}{
					"field1": "value1",
					"field2": "value2",
				},
			},
			wantErr: false,
		},
		{
			name: "missing field1",
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data: map[string]interface{}{
					"field2": "value2",
				},
			},
			wantErr: true,
		},
		{
			name: "missing field2",
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data: map[string]interface{}{
					"field1": "value1",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataTypeRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rule    *DataTypeRule
		result  *TaskResult
		wantErr bool
	}{
		{
			name: "string type valid",
			rule: &DataTypeRule{
				Field:        "name",
				ExpectedType: "string",
			},
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data: map[string]interface{}{
					"name": "test",
				},
			},
			wantErr: false,
		},
		{
			name: "number type valid",
			rule: &DataTypeRule{
				Field:        "count",
				ExpectedType: "number",
			},
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data: map[string]interface{}{
					"count": 42,
				},
			},
			wantErr: false,
		},
		{
			name: "type mismatch",
			rule: &DataTypeRule{
				Field:        "count",
				ExpectedType: "string",
			},
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data: map[string]interface{}{
					"count": 42,
				},
			},
			wantErr: true,
		},
		{
			name: "field not exists",
			rule: &DataTypeRule{
				Field:        "missing",
				ExpectedType: "string",
			},
			result: &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data:    make(map[string]interface{}),
			},
			wantErr: false, // Skip validation if field doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScoreRangeRule_Validate(t *testing.T) {
	rule := &ScoreRangeRule{
		MinScore: 0,
		MaxScore: 100,
	}

	tests := []struct {
		name    string
		score   float64
		wantErr bool
	}{
		{"valid score 50", 50, false},
		{"valid score 0", 0, false},
		{"valid score 100", 100, false},
		{"invalid score -10", -10, true},
		{"invalid score 150", 150, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &TaskResult{
				ID:      "result-001",
				TaskID:  "task-001",
				AgentID: "agent-001",
				Data:    make(map[string]interface{}),
				Score:   tt.score,
			}

			err := rule.Validate(result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResultValidator_AddRule(t *testing.T) {
	validator := NewResultValidator()

	rule := &RequiredFieldsRule{
		Fields: []string{"field1"},
	}

	validator.AddRule(rule)

	if len(validator.rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(validator.rules))
	}
}

func TestResultValidator_ValidateAndMark(t *testing.T) {
	validator := NewResultValidator()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data:    make(map[string]interface{}),
	}

	err := validator.ValidateAndMark(result)
	if err != nil {
		t.Fatalf("ValidateAndMark failed: %v", err)
	}

	if result.Status != ResultStatusValidated {
		t.Errorf("Expected status VALIDATED, got %s", result.Status)
	}

	if result.ValidatedAt == nil {
		t.Error("ValidatedAt should be set")
	}
}

func TestResultValidator_ValidateAndMark_Failed(t *testing.T) {
	validator := NewResultValidator()

	// Add a rule that will fail
	validator.AddRule(&RequiredFieldsRule{
		Fields: []string{"required_field"},
	})

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Status:  ResultStatusPending,
		Data:    make(map[string]interface{}),
	}

	err := validator.ValidateAndMark(result)
	if err == nil {
		t.Error("Expected validation error")
	}

	if result.Status != ResultStatusRejected {
		t.Errorf("Expected status REJECTED, got %s", result.Status)
	}

	if result.Error == "" {
		t.Error("Error message should be set")
	}
}

func TestResultValidator_ValidateMultiple(t *testing.T) {
	validator := NewResultValidator()

	results := []*TaskResult{
		{
			ID:      "result-001",
			TaskID:  "task-001",
			AgentID: "agent-001",
			Data:    make(map[string]interface{}),
		},
		{
			ID:      "result-002",
			TaskID:  "task-001",
			AgentID: "agent-002",
			Data:    make(map[string]interface{}),
		},
		{
			ID:     "result-003", // Missing TaskID - will fail
			Data:   make(map[string]interface{}),
		},
	}

	errors := validator.ValidateMultiple(results)

	// First two should pass
	if results[0].Status != ResultStatusValidated {
		t.Error("Result 1 should be validated")
	}
	if results[1].Status != ResultStatusValidated {
		t.Error("Result 2 should be validated")
	}

	// Third should fail
	if results[2].Status != ResultStatusRejected {
		t.Error("Result 3 should be rejected")
	}

	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	if _, exists := errors["result-003"]; !exists {
		t.Error("Expected error for result-003")
	}
}

func TestResultValidator_WithComplexRules(t *testing.T) {
	validator := NewResultValidator()

	// Add multiple rules
	validator.AddRule(&RequiredFieldsRule{
		Fields: []string{"result", "confidence"},
	})

	validator.AddRule(&DataTypeRule{
		Field:        "result",
		ExpectedType: "string",
	})

	validator.AddRule(&DataTypeRule{
		Field:        "confidence",
		ExpectedType: "number",
	})

	validator.AddRule(&ScoreRangeRule{
		MinScore: 0,
		MaxScore: 100,
	})

	validResult := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Data: map[string]interface{}{
			"result":     "success",
			"confidence": 95.5,
		},
		Score: 85,
	}

	err := validator.ValidateAndMark(validResult)
	if err != nil {
		t.Errorf("Valid result should pass: %v", err)
	}

	invalidResult := &TaskResult{
		ID:      "result-002",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Data: map[string]interface{}{
			"result":     "success",
			"confidence": "high", // Wrong type
		},
		Score: 85,
	}

	err = validator.ValidateAndMark(invalidResult)
	if err == nil {
		t.Error("Invalid result should fail validation")
	}
}

func BenchmarkResultValidator_Validate(b *testing.B) {
	validator := NewResultValidator()

	result := &TaskResult{
		ID:      "result-001",
		TaskID:  "task-001",
		AgentID: "agent-001",
		Data:    make(map[string]interface{}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(result)
	}
}
