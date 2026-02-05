package model

import (
	"time"
)

// TokenUsage Token使用记录
type TokenUsage struct {
	ID           string    `json:"id" db:"id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	AgentID      string    `json:"agent_id" db:"agent_id"`
	TaskID       string    `json:"task_id" db:"task_id"`
	Model        string    `json:"model" db:"model"`           // gpt-4, claude-3, gemini-pro
	Provider     string    `json:"provider" db:"provider"`     // openai, anthropic, google
	InputTokens  int64     `json:"input_tokens" db:"input_tokens"`
	OutputTokens int64     `json:"output_tokens" db:"output_tokens"`
	TotalTokens  int64     `json:"total_tokens" db:"total_tokens"`
	InputCost    float64   `json:"input_cost" db:"input_cost"`   // USD
	OutputCost   float64   `json:"output_cost" db:"output_cost"` // USD
	TotalCost    float64   `json:"total_cost" db:"total_cost"`   // USD
	Duration     int64     `json:"duration" db:"duration"`       // 毫秒
	Cached       bool      `json:"cached" db:"cached"`           // 是否使用缓存
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
}

// CostStatistics 成本统计
type CostStatistics struct {
	ID             string    `json:"id" db:"id"`
	TenantID       string    `json:"tenant_id" db:"tenant_id"`
	Period         string    `json:"period" db:"period"`         // daily, weekly, monthly
	PeriodStart    time.Time `json:"period_start" db:"period_start"`
	PeriodEnd      time.Time `json:"period_end" db:"period_end"`
	TotalTokens    int64     `json:"total_tokens" db:"total_tokens"`
	TotalCost      float64   `json:"total_cost" db:"total_cost"`
	RequestCount   int64     `json:"request_count" db:"request_count"`
	AvgTokensPerRequest float64 `json:"avg_tokens_per_request" db:"avg_tokens_per_request"`
	AvgCostPerRequest   float64 `json:"avg_cost_per_request" db:"avg_cost_per_request"`
	CachedRequests int64     `json:"cached_requests" db:"cached_requests"`
	CacheHitRate   float64   `json:"cache_hit_rate" db:"cache_hit_rate"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ModelCostBreakdown 模型成本分解
type ModelCostBreakdown struct {
	Model        string  `json:"model"`
	Provider     string  `json:"provider"`
	Tokens       int64   `json:"tokens"`
	Cost         float64 `json:"cost"`
	RequestCount int64   `json:"request_count"`
	Percentage   float64 `json:"percentage"`
}

// UserCostBreakdown 用户成本分解
type UserCostBreakdown struct {
	UserID       string  `json:"user_id"`
	Username     string  `json:"username"`
	Tokens       int64   `json:"tokens"`
	Cost         float64 `json:"cost"`
	RequestCount int64   `json:"request_count"`
	Percentage   float64 `json:"percentage"`
}

// CostForecast 成本预测
type CostForecast struct {
	ID           string    `json:"id" db:"id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	ForecastDate time.Time `json:"forecast_date" db:"forecast_date"`
	PredictedTokens int64  `json:"predicted_tokens" db:"predicted_tokens"`
	PredictedCost   float64 `json:"predicted_cost" db:"predicted_cost"`
	Confidence      float64 `json:"confidence" db:"confidence"` // 0.0-1.0
	Method          string  `json:"method" db:"method"`         // linear, exponential, arima
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// CostAlert 成本告警
type CostAlert struct {
	ID           string    `json:"id" db:"id"`
	TenantID     string    `json:"tenant_id" db:"tenant_id"`
	AlertType    string    `json:"alert_type" db:"alert_type"` // threshold, forecast, anomaly
	Severity     string    `json:"severity" db:"severity"`     // info, warning, critical
	Title        string    `json:"title" db:"title"`
	Message      string    `json:"message" db:"message"`
	CurrentValue float64   `json:"current_value" db:"current_value"`
	ThresholdValue float64 `json:"threshold_value" db:"threshold_value"`
	Status       string    `json:"status" db:"status"`         // active, acknowledged, resolved
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at" db:"acknowledged_at"`
	ResolvedAt   *time.Time `json:"resolved_at" db:"resolved_at"`
}

// ModelPricing LLM模型定价
type ModelPricing struct {
	Model            string  `json:"model"`
	Provider         string  `json:"provider"`
	InputPricePerK   float64 `json:"input_price_per_k"`   // USD per 1K tokens
	OutputPricePerK  float64 `json:"output_price_per_k"`  // USD per 1K tokens
	ContextWindow    int     `json:"context_window"`      // 最大上下文窗口
}

// GetModelPricing 获取模型定价（2026年1月定价）
func GetModelPricing() map[string]ModelPricing {
	return map[string]ModelPricing{
		// OpenAI GPT-4
		"gpt-4": {
			Model:           "gpt-4",
			Provider:        "openai",
			InputPricePerK:  0.03,  // $0.03/1K tokens
			OutputPricePerK: 0.06,  // $0.06/1K tokens
			ContextWindow:   8192,
		},
		"gpt-4-32k": {
			Model:           "gpt-4-32k",
			Provider:        "openai",
			InputPricePerK:  0.06,
			OutputPricePerK: 0.12,
			ContextWindow:   32768,
		},
		"gpt-4-turbo": {
			Model:           "gpt-4-turbo",
			Provider:        "openai",
			InputPricePerK:  0.01,
			OutputPricePerK: 0.03,
			ContextWindow:   128000,
		},
		"gpt-3.5-turbo": {
			Model:           "gpt-3.5-turbo",
			Provider:        "openai",
			InputPricePerK:  0.0005,
			OutputPricePerK: 0.0015,
			ContextWindow:   16385,
		},

		// Anthropic Claude
		"claude-3-opus": {
			Model:           "claude-3-opus",
			Provider:        "anthropic",
			InputPricePerK:  0.015,
			OutputPricePerK: 0.075,
			ContextWindow:   200000,
		},
		"claude-3-sonnet": {
			Model:           "claude-3-sonnet",
			Provider:        "anthropic",
			InputPricePerK:  0.003,
			OutputPricePerK: 0.015,
			ContextWindow:   200000,
		},
		"claude-3-haiku": {
			Model:           "claude-3-haiku",
			Provider:        "anthropic",
			InputPricePerK:  0.00025,
			OutputPricePerK: 0.00125,
			ContextWindow:   200000,
		},

		// Google Gemini
		"gemini-pro": {
			Model:           "gemini-pro",
			Provider:        "google",
			InputPricePerK:  0.00025,
			OutputPricePerK: 0.0005,
			ContextWindow:   32760,
		},
		"gemini-pro-vision": {
			Model:           "gemini-pro-vision",
			Provider:        "google",
			InputPricePerK:  0.00025,
			OutputPricePerK: 0.0005,
			ContextWindow:   16384,
		},
		"gemini-ultra": {
			Model:           "gemini-ultra",
			Provider:        "google",
			InputPricePerK:  0.01,
			OutputPricePerK: 0.03,
			ContextWindow:   32760,
		},
	}
}

// CostBudget 成本预算
type CostBudget struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Period      string    `json:"period" db:"period"`      // monthly, quarterly, yearly
	Budget      float64   `json:"budget" db:"budget"`      // USD
	Spent       float64   `json:"spent" db:"spent"`        // USD
	Remaining   float64   `json:"remaining" db:"remaining"`
	AlertAt     float64   `json:"alert_at" db:"alert_at"`  // 告警阈值（百分比）
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CostOptimization 成本优化建议
type CostOptimization struct {
	Type        string  `json:"type"`        // model_switch, cache, batch, prompt_optimization
	Title       string  `json:"title"`
	Description string  `json:"description"`
	CurrentCost float64 `json:"current_cost"`
	EstimatedCost float64 `json:"estimated_cost"`
	Savings     float64 `json:"savings"`
	SavingsPercent float64 `json:"savings_percent"`
	Priority    string  `json:"priority"`    // high, medium, low
}

// CalculateCost 计算成本
func (t *TokenUsage) CalculateCost(pricing ModelPricing) {
	t.InputCost = float64(t.InputTokens) / 1000.0 * pricing.InputPricePerK
	t.OutputCost = float64(t.OutputTokens) / 1000.0 * pricing.OutputPricePerK
	t.TotalCost = t.InputCost + t.OutputCost
	t.TotalTokens = t.InputTokens + t.OutputTokens
}

// IsOverBudget 检查是否超预算
func (b *CostBudget) IsOverBudget() bool {
	return b.Spent >= b.Budget
}

// ShouldAlert 是否应该告警
func (b *CostBudget) ShouldAlert() bool {
	usagePercent := (b.Spent / b.Budget) * 100
	return usagePercent >= b.AlertAt
}

// GetUsagePercent 获取使用百分比
func (b *CostBudget) GetUsagePercent() float64 {
	if b.Budget == 0 {
		return 0
	}
	return (b.Spent / b.Budget) * 100
}
