package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector Prometheus指标收集器
type MetricsCollector struct {
	// HTTP请求指标
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestSize      *prometheus.HistogramVec
	HTTPResponseSize     *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// gRPC请求指标
	GRPCRequestsTotal   *prometheus.CounterVec
	GRPCRequestDuration *prometheus.HistogramVec
	GRPCStreamMessages  *prometheus.CounterVec

	// Agent执行指标
	AgentExecutionsTotal    *prometheus.CounterVec
	AgentExecutionDuration  *prometheus.HistogramVec
	AgentExecutionErrors    *prometheus.CounterVec
	AgentActiveExecutions   prometheus.Gauge
	AgentQueueLength        prometheus.Gauge
	AgentTokensConsumed     *prometheus.CounterVec
	AgentCostTotal          *prometheus.CounterVec

	// Task处理指标
	TasksCreatedTotal     *prometheus.CounterVec
	TasksCompletedTotal   *prometheus.CounterVec
	TasksFailedTotal      *prometheus.CounterVec
	TaskDuration          *prometheus.HistogramVec
	TaskQueueLength       *prometheus.GaugeVec
	TaskRetries           *prometheus.CounterVec
	TaskActiveTasks       *prometheus.GaugeVec

	// 缓存指标
	CacheHits             *prometheus.CounterVec
	CacheMisses           *prometheus.CounterVec
	CacheSize             *prometheus.GaugeVec
	CacheEvictions        *prometheus.CounterVec
	CacheOperationLatency *prometheus.HistogramVec

	// 数据库指标
	DBConnectionsActive   *prometheus.GaugeVec
	DBConnectionsIdle     *prometheus.GaugeVec
	DBConnectionsWait     *prometheus.CounterVec
	DBQueryDuration       *prometheus.HistogramVec
	DBQueriesTotal        *prometheus.CounterVec
	DBTransactionsTotal   *prometheus.CounterVec

	// 消息队列指标
	MQMessagesPublished   *prometheus.CounterVec
	MQMessagesConsumed    *prometheus.CounterVec
	MQMessagesFailed      *prometheus.CounterVec
	MQMessageLatency      *prometheus.HistogramVec
	MQQueueDepth          *prometheus.GaugeVec

	// 资源使用指标
	CPUUsage              *prometheus.GaugeVec
	MemoryUsage           *prometheus.GaugeVec
	GoroutinesActive      prometheus.Gauge
	HeapAllocBytes        prometheus.Gauge
	HeapInuseBytes        prometheus.Gauge
	StackInuseBytes       prometheus.Gauge

	// 成本指标
	CostPerRequest        *prometheus.HistogramVec
	CostBudgetUtilization *prometheus.GaugeVec
	CostAlerts            *prometheus.CounterVec

	// 租户指标
	TenantRequestsTotal   *prometheus.CounterVec
	TenantQuotaUsage      *prometheus.GaugeVec
	TenantActiveUsers     *prometheus.GaugeVec

	// 业务指标
	UsersActiveTotal      *prometheus.GaugeVec
	UsersRegisteredTotal  *prometheus.CounterVec
	ToolInvocationsTotal  *prometheus.CounterVec
	ToolExecutionDuration *prometheus.HistogramVec
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(namespace string) *MetricsCollector {
	return &MetricsCollector{
		// HTTP请求指标
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status", "tenant_id"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "tenant_id"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being served",
			},
		),

		// gRPC请求指标
		GRPCRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "grpc_requests_total",
				Help:      "Total number of gRPC requests",
			},
			[]string{"service", "method", "status", "tenant_id"},
		),
		GRPCRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "grpc_request_duration_seconds",
				Help:      "gRPC request duration in seconds",
				Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"service", "method", "tenant_id"},
		),
		GRPCStreamMessages: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "grpc_stream_messages_total",
				Help:      "Total number of gRPC stream messages",
			},
			[]string{"service", "method", "direction"},
		),

		// Agent执行指标
		AgentExecutionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "agent_executions_total",
				Help:      "Total number of agent executions",
			},
			[]string{"agent_id", "status", "tenant_id"},
		),
		AgentExecutionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "agent_execution_duration_seconds",
				Help:      "Agent execution duration in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600},
			},
			[]string{"agent_id", "tenant_id"},
		),
		AgentExecutionErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "agent_execution_errors_total",
				Help:      "Total number of agent execution errors",
			},
			[]string{"agent_id", "error_type", "tenant_id"},
		),
		AgentActiveExecutions: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agent_active_executions",
				Help:      "Current number of active agent executions",
			},
		),
		AgentQueueLength: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "agent_queue_length",
				Help:      "Current length of agent execution queue",
			},
		),
		AgentTokensConsumed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "agent_tokens_consumed_total",
				Help:      "Total number of tokens consumed by agents",
			},
			[]string{"agent_id", "model", "token_type", "tenant_id"},
		),
		AgentCostTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "agent_cost_total",
				Help:      "Total cost of agent executions in USD",
			},
			[]string{"agent_id", "model", "tenant_id"},
		),

		// Task处理指标
		TasksCreatedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tasks_created_total",
				Help:      "Total number of tasks created",
			},
			[]string{"task_type", "priority", "tenant_id"},
		),
		TasksCompletedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tasks_completed_total",
				Help:      "Total number of tasks completed",
			},
			[]string{"task_type", "tenant_id"},
		),
		TasksFailedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tasks_failed_total",
				Help:      "Total number of tasks failed",
			},
			[]string{"task_type", "error_type", "tenant_id"},
		),
		TaskDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "task_duration_seconds",
				Help:      "Task execution duration in seconds",
				Buckets:   []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800},
			},
			[]string{"task_type", "tenant_id"},
		),
		TaskQueueLength: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "task_queue_length",
				Help:      "Current length of task queue",
			},
			[]string{"priority", "tenant_id"},
		),
		TaskRetries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "task_retries_total",
				Help:      "Total number of task retries",
			},
			[]string{"task_type", "tenant_id"},
		),
		TaskActiveTasks: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "task_active_tasks",
				Help:      "Current number of active tasks",
			},
			[]string{"task_type", "tenant_id"},
		),

		// 缓存指标
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache_name", "tenant_id"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache_name", "tenant_id"},
		),
		CacheSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cache_size_bytes",
				Help:      "Current cache size in bytes",
			},
			[]string{"cache_name"},
		),
		CacheEvictions: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_evictions_total",
				Help:      "Total number of cache evictions",
			},
			[]string{"cache_name", "reason"},
		),
		CacheOperationLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "cache_operation_duration_seconds",
				Help:      "Cache operation duration in seconds",
				Buckets:   []float64{.0001, .0005, .001, .005, .01, .05, .1},
			},
			[]string{"cache_name", "operation"},
		),

		// 数据库指标
		DBConnectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_active",
				Help:      "Current number of active database connections",
			},
			[]string{"database"},
		),
		DBConnectionsIdle: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Current number of idle database connections",
			},
			[]string{"database"},
		),
		DBConnectionsWait: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_connections_wait_total",
				Help:      "Total number of times waited for database connection",
			},
			[]string{"database"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .05, .1, .5, 1, 5},
			},
			[]string{"database", "query_type"},
		),
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"database", "query_type", "status"},
		),
		DBTransactionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_transactions_total",
				Help:      "Total number of database transactions",
			},
			[]string{"database", "status"},
		),

		// 消息队列指标
		MQMessagesPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mq_messages_published_total",
				Help:      "Total number of messages published to queue",
			},
			[]string{"queue", "tenant_id"},
		),
		MQMessagesConsumed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mq_messages_consumed_total",
				Help:      "Total number of messages consumed from queue",
			},
			[]string{"queue", "tenant_id"},
		),
		MQMessagesFailed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "mq_messages_failed_total",
				Help:      "Total number of failed message processing",
			},
			[]string{"queue", "error_type", "tenant_id"},
		),
		MQMessageLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "mq_message_latency_seconds",
				Help:      "Message processing latency in seconds",
				Buckets:   []float64{.1, .5, 1, 5, 10, 30, 60},
			},
			[]string{"queue", "tenant_id"},
		),
		MQQueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "mq_queue_depth",
				Help:      "Current depth of message queue",
			},
			[]string{"queue"},
		),

		// 资源使用指标
		CPUUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cpu_usage_percent",
				Help:      "CPU usage percentage",
			},
			[]string{"service"},
		),
		MemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memory_usage_bytes",
				Help:      "Memory usage in bytes",
			},
			[]string{"service", "type"},
		),
		GoroutinesActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "goroutines_active",
				Help:      "Current number of active goroutines",
			},
		),
		HeapAllocBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "heap_alloc_bytes",
				Help:      "Current heap allocation in bytes",
			},
		),
		HeapInuseBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "heap_inuse_bytes",
				Help:      "Current heap in-use in bytes",
			},
		),
		StackInuseBytes: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "stack_inuse_bytes",
				Help:      "Current stack in-use in bytes",
			},
		),

		// 成本指标
		CostPerRequest: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "cost_per_request_usd",
				Help:      "Cost per request in USD",
				Buckets:   []float64{.0001, .001, .01, .1, 1, 10},
			},
			[]string{"service", "tenant_id"},
		),
		CostBudgetUtilization: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cost_budget_utilization_percent",
				Help:      "Cost budget utilization percentage",
			},
			[]string{"tenant_id", "period"},
		),
		CostAlerts: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cost_alerts_total",
				Help:      "Total number of cost alerts",
			},
			[]string{"tenant_id", "alert_type", "severity"},
		),

		// 租户指标
		TenantRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tenant_requests_total",
				Help:      "Total number of requests per tenant",
			},
			[]string{"tenant_id", "service"},
		),
		TenantQuotaUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "tenant_quota_usage_percent",
				Help:      "Tenant quota usage percentage",
			},
			[]string{"tenant_id", "quota_type"},
		),
		TenantActiveUsers: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "tenant_active_users",
				Help:      "Current number of active users per tenant",
			},
			[]string{"tenant_id"},
		),

		// 业务指标
		UsersActiveTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "users_active_total",
				Help:      "Total number of active users",
			},
			[]string{"tenant_id"},
		),
		UsersRegisteredTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_registered_total",
				Help:      "Total number of registered users",
			},
			[]string{"tenant_id"},
		),
		ToolInvocationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tool_invocations_total",
				Help:      "Total number of tool invocations",
			},
			[]string{"tool_name", "status", "tenant_id"},
		),
		ToolExecutionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "tool_execution_duration_seconds",
				Help:      "Tool execution duration in seconds",
				Buckets:   []float64{.1, .5, 1, 5, 10, 30, 60},
			},
			[]string{"tool_name", "tenant_id"},
		),
	}
}

// RecordHTTPRequest 记录HTTP请求
func (mc *MetricsCollector) RecordHTTPRequest(method, path, status, tenantID string, duration time.Duration, requestSize, responseSize int) {
	mc.HTTPRequestsTotal.WithLabelValues(method, path, status, tenantID).Inc()
	mc.HTTPRequestDuration.WithLabelValues(method, path, tenantID).Observe(duration.Seconds())
	mc.HTTPRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
	mc.HTTPResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// RecordGRPCRequest 记录gRPC请求
func (mc *MetricsCollector) RecordGRPCRequest(service, method, status, tenantID string, duration time.Duration) {
	mc.GRPCRequestsTotal.WithLabelValues(service, method, status, tenantID).Inc()
	mc.GRPCRequestDuration.WithLabelValues(service, method, tenantID).Observe(duration.Seconds())
}

// RecordAgentExecution 记录Agent执行
func (mc *MetricsCollector) RecordAgentExecution(agentID, status, tenantID string, duration time.Duration, inputTokens, outputTokens int64, cost float64, model string) {
	mc.AgentExecutionsTotal.WithLabelValues(agentID, status, tenantID).Inc()
	mc.AgentExecutionDuration.WithLabelValues(agentID, tenantID).Observe(duration.Seconds())
	mc.AgentTokensConsumed.WithLabelValues(agentID, model, "input", tenantID).Add(float64(inputTokens))
	mc.AgentTokensConsumed.WithLabelValues(agentID, model, "output", tenantID).Add(float64(outputTokens))
	mc.AgentCostTotal.WithLabelValues(agentID, model, tenantID).Add(cost)
}

// RecordCacheOperation 记录缓存操作
func (mc *MetricsCollector) RecordCacheOperation(cacheName, operation string, hit bool, duration time.Duration, tenantID string) {
	if hit {
		mc.CacheHits.WithLabelValues(cacheName, tenantID).Inc()
	} else {
		mc.CacheMisses.WithLabelValues(cacheName, tenantID).Inc()
	}
	mc.CacheOperationLatency.WithLabelValues(cacheName, operation).Observe(duration.Seconds())
}
