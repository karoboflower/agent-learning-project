package aggregator

import (
	"fmt"
	"time"
)

// ValidationRule 验证规则
type ValidationRule interface {
	Validate(result *TaskResult) error
	Name() string
}

// RequiredFieldsRule 必需字段验证规则
type RequiredFieldsRule struct {
	Fields []string
}

func (r *RequiredFieldsRule) Name() string {
	return "RequiredFields"
}

func (r *RequiredFieldsRule) Validate(result *TaskResult) error {
	for _, field := range r.Fields {
		if _, exists := result.Data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	return nil
}

// DataTypeRule 数据类型验证规则
type DataTypeRule struct {
	Field        string
	ExpectedType string // "string", "number", "boolean", "object", "array"
}

func (r *DataTypeRule) Name() string {
	return "DataType"
}

func (r *DataTypeRule) Validate(result *TaskResult) error {
	value, exists := result.Data[r.Field]
	if !exists {
		return nil // 如果字段不存在，跳过类型检查
	}

	var actualType string
	switch value.(type) {
	case string:
		actualType = "string"
	case int, int32, int64, float32, float64:
		actualType = "number"
	case bool:
		actualType = "boolean"
	case map[string]interface{}:
		actualType = "object"
	case []interface{}:
		actualType = "array"
	default:
		actualType = "unknown"
	}

	if actualType != r.ExpectedType {
		return fmt.Errorf("field '%s' expected type '%s', got '%s'", r.Field, r.ExpectedType, actualType)
	}

	return nil
}

// ScoreRangeRule 分数范围验证规则
type ScoreRangeRule struct {
	MinScore float64
	MaxScore float64
}

func (r *ScoreRangeRule) Name() string {
	return "ScoreRange"
}

func (r *ScoreRangeRule) Validate(result *TaskResult) error {
	if result.Score < r.MinScore || result.Score > r.MaxScore {
		return fmt.Errorf("score %.2f is out of range [%.2f, %.2f]", result.Score, r.MinScore, r.MaxScore)
	}
	return nil
}

// ResultValidator 结果验证器
type ResultValidator struct {
	rules []ValidationRule
}

// NewResultValidator 创建验证器
func NewResultValidator() *ResultValidator {
	return &ResultValidator{
		rules: make([]ValidationRule, 0),
	}
}

// AddRule 添加验证规则
func (v *ResultValidator) AddRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
}

// Validate 验证结果
func (v *ResultValidator) Validate(result *TaskResult) error {
	// 基本字段验证
	if result.ID == "" {
		return fmt.Errorf("result ID is empty")
	}
	if result.TaskID == "" {
		return fmt.Errorf("task ID is empty")
	}
	if result.AgentID == "" {
		return fmt.Errorf("agent ID is empty")
	}
	if result.Data == nil {
		return fmt.Errorf("result data is nil")
	}

	// 应用自定义规则
	for _, rule := range v.rules {
		if err := rule.Validate(result); err != nil {
			return fmt.Errorf("%s validation failed: %w", rule.Name(), err)
		}
	}

	return nil
}

// ValidateAndMark 验证结果并标记状态
func (v *ResultValidator) ValidateAndMark(result *TaskResult) error {
	if err := v.Validate(result); err != nil {
		result.Status = ResultStatusRejected
		result.Error = err.Error()
		return err
	}

	result.Status = ResultStatusValidated
	now := time.Now()
	result.ValidatedAt = &now

	return nil
}

// ValidateMultiple 批量验证
func (v *ResultValidator) ValidateMultiple(results []*TaskResult) map[string]error {
	errors := make(map[string]error)

	for _, result := range results {
		if err := v.ValidateAndMark(result); err != nil {
			errors[result.ID] = err
		}
	}

	return errors
}
