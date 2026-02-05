package model

import "time"

// StreamEvent 流式事件
type StreamEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // start, chunk, end, error
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// AsyncTask 异步任务
type AsyncTask struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	UserID      string                 `json:"user_id"`
	Type        string                 `json:"type"` // agent_execution, batch_process, etc.
	Priority    int                    `json:"priority"` // 1-10, higher = more important
	Payload     map[string]interface{} `json:"payload"`
	Status      string                 `json:"status"` // pending, processing, completed, failed
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Progress    int                    `json:"progress"` // 0-100
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	ScheduledAt time.Time              `json:"scheduled_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// BatchRequest 批量请求
type BatchRequest struct {
	ID        string                   `json:"id"`
	TenantID  string                   `json:"tenant_id"`
	UserID    string                   `json:"user_id"`
	Requests  []map[string]interface{} `json:"requests"`
	Status    string                   `json:"status"` // pending, processing, completed
	Results   []map[string]interface{} `json:"results,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// PromptTemplate Prompt模板
type PromptTemplate struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Template         string            `json:"template"`
	Variables        []string          `json:"variables"`
	CompressedLength int               `json:"compressed_length"`
	OriginalLength   int               `json:"original_length"`
	CompressionRatio float64           `json:"compression_ratio"`
	Metadata         map[string]string `json:"metadata"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// ContextWindow 上下文窗口
type ContextWindow struct {
	MaxTokens       int     `json:"max_tokens"`
	CurrentTokens   int     `json:"current_tokens"`
	UsagePercent    float64 `json:"usage_percent"`
	Messages        []Message `json:"messages"`
	PruneStrategy   string  `json:"prune_strategy"` // oldest, least_important, summarize
}

// Message 消息
type Message struct {
	Role      string    `json:"role"` // system, user, assistant
	Content   string    `json:"content"`
	Tokens    int       `json:"tokens"`
	Timestamp time.Time `json:"timestamp"`
	Important bool      `json:"important"` // 是否重要，不能被剪枝
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Key        string                 `json:"key"`
	Value      map[string]interface{} `json:"value"`
	TTL        time.Duration          `json:"ttl"`
	Hits       int64                  `json:"hits"`
	CreatedAt  time.Time              `json:"created_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
	AccessedAt time.Time              `json:"accessed_at"`
}

// CacheStats 缓存统计
type CacheStats struct {
	TotalKeys     int64   `json:"total_keys"`
	TotalHits     int64   `json:"total_hits"`
	TotalMisses   int64   `json:"total_misses"`
	HitRate       float64 `json:"hit_rate"`
	AvgAccessTime float64 `json:"avg_access_time_ms"`
	MemoryUsage   int64   `json:"memory_usage_bytes"`
	EvictionCount int64   `json:"eviction_count"`
}

// ConnectionPool 连接池
type ConnectionPool struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"` // database, http, grpc
	MinSize     int            `json:"min_size"`
	MaxSize     int            `json:"max_size"`
	CurrentSize int            `json:"current_size"`
	ActiveConns int            `json:"active_conns"`
	IdleConns   int            `json:"idle_conns"`
	WaitCount   int64          `json:"wait_count"`
	WaitTime    time.Duration  `json:"wait_time"`
	MaxLifetime time.Duration  `json:"max_lifetime"`
	MaxIdleTime time.Duration  `json:"max_idle_time"`
	Stats       ConnectionPoolStats `json:"stats"`
}

// ConnectionPoolStats 连接池统计
type ConnectionPoolStats struct {
	TotalConns    int64         `json:"total_conns"`
	AcquiredConns int64         `json:"acquired_conns"`
	ReleasedConns int64         `json:"released_conns"`
	AvgWaitTime   time.Duration `json:"avg_wait_time"`
	MaxWaitTime   time.Duration `json:"max_wait_time"`
}

// AutoScalingConfig 自动扩缩容配置
type AutoScalingConfig struct {
	Enabled            bool    `json:"enabled"`
	MinReplicas        int     `json:"min_replicas"`
	MaxReplicas        int     `json:"max_replicas"`
	TargetCPU          float64 `json:"target_cpu_percent"`
	TargetMemory       float64 `json:"target_memory_percent"`
	TargetQPS          int64   `json:"target_qps"`
	ScaleUpThreshold   float64 `json:"scale_up_threshold"`
	ScaleDownThreshold float64 `json:"scale_down_threshold"`
	CooldownPeriod     int     `json:"cooldown_period_seconds"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TenantID           string        `json:"tenant_id"`
	ServiceName        string        `json:"service_name"`
	RequestCount       int64         `json:"request_count"`
	AvgResponseTime    time.Duration `json:"avg_response_time"`
	P50ResponseTime    time.Duration `json:"p50_response_time"`
	P95ResponseTime    time.Duration `json:"p95_response_time"`
	P99ResponseTime    time.Duration `json:"p99_response_time"`
	ErrorRate          float64       `json:"error_rate"`
	Throughput         float64       `json:"throughput_rps"`
	CPUUsage           float64       `json:"cpu_usage_percent"`
	MemoryUsage        float64       `json:"memory_usage_percent"`
	ActiveConnections  int           `json:"active_connections"`
	QueuedTasks        int           `json:"queued_tasks"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
	TokensPerSecond    float64       `json:"tokens_per_second"`
	Timestamp          time.Time     `json:"timestamp"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Category    string                 `json:"category"` // latency, throughput, cost, resource
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"` // high, medium, low
	Effort      string                 `json:"effort"` // high, medium, low
	Priority    int                    `json:"priority"` // 1-10
	Metrics     map[string]interface{} `json:"metrics"`
	CreatedAt   time.Time              `json:"created_at"`
}
