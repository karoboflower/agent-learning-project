package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Config 追踪配置
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	SamplingRatio  float64
	MaxExportBatch int
	MaxQueueSize   int
}

// Tracer 追踪器
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	config   *Config
}

// NewTracer 创建追踪器
func NewTracer(config *Config) (*Tracer, error) {
	// 验证配置
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}
	if config.JaegerEndpoint == "" {
		config.JaegerEndpoint = "http://localhost:14268/api/traces"
	}
	if config.SamplingRatio == 0 {
		config.SamplingRatio = 1.0 // 默认100%采样
	}
	if config.MaxExportBatch == 0 {
		config.MaxExportBatch = 512
	}
	if config.MaxQueueSize == 0 {
		config.MaxQueueSize = 2048
	}

	// 创建Jaeger导出器
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	// 创建资源
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建追踪提供者
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(config.MaxExportBatch),
			sdktrace.WithMaxQueueSize(config.MaxQueueSize),
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SamplingRatio)),
	)

	// 设置全局追踪提供者
	otel.SetTracerProvider(tp)

	// 设置全局传播器
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// 创建追踪器
	tracer := tp.Tracer(config.ServiceName)

	return &Tracer{
		provider: tp,
		tracer:   tracer,
		config:   config,
	}, nil
}

// Start 开始一个span
func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// StartSpan 开始一个span（简化接口）
func (t *Tracer) StartSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	return t.Start(ctx, operation, trace.WithSpanKind(trace.SpanKindInternal))
}

// StartHTTPServerSpan 开始HTTP服务器span
func (t *Tracer) StartHTTPServerSpan(ctx context.Context, method, path string) (context.Context, trace.Span) {
	return t.Start(ctx, fmt.Sprintf("%s %s", method, path),
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPRoute(path),
		),
	)
}

// StartHTTPClientSpan 开始HTTP客户端span
func (t *Tracer) StartHTTPClientSpan(ctx context.Context, method, url string) (context.Context, trace.Span) {
	return t.Start(ctx, fmt.Sprintf("%s %s", method, url),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPURL(url),
		),
	)
}

// StartGRPCServerSpan 开始gRPC服务器span
func (t *Tracer) StartGRPCServerSpan(ctx context.Context, fullMethod string) (context.Context, trace.Span) {
	return t.Start(ctx, fullMethod,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			semconv.RPCSystem("grpc"),
			semconv.RPCMethod(fullMethod),
		),
	)
}

// StartGRPCClientSpan 开始gRPC客户端span
func (t *Tracer) StartGRPCClientSpan(ctx context.Context, fullMethod string) (context.Context, trace.Span) {
	return t.Start(ctx, fullMethod,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.RPCSystem("grpc"),
			semconv.RPCMethod(fullMethod),
		),
	)
}

// StartDatabaseSpan 开始数据库span
func (t *Tracer) StartDatabaseSpan(ctx context.Context, dbType, operation, table string) (context.Context, trace.Span) {
	return t.Start(ctx, fmt.Sprintf("%s.%s", dbType, operation),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystem(dbType),
			semconv.DBOperation(operation),
			semconv.DBSQLTable(table),
		),
	)
}

// StartCacheSpan 开始缓存span
func (t *Tracer) StartCacheSpan(ctx context.Context, operation, key string) (context.Context, trace.Span) {
	return t.Start(ctx, fmt.Sprintf("cache.%s", operation),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("cache.operation", operation),
			attribute.String("cache.key", key),
		),
	)
}

// StartMessageQueueSpan 开始消息队列span
func (t *Tracer) StartMessageQueueSpan(ctx context.Context, operation, queue string) (context.Context, trace.Span) {
	return t.Start(ctx, fmt.Sprintf("mq.%s", operation),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.operation", operation),
			attribute.String("messaging.destination", queue),
		),
	)
}

// RecordError 记录错误
func (t *Tracer) RecordError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanStatus 设置span状态
func (t *Tracer) SetSpanStatus(span trace.Span, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// AddEvent 添加事件
func (t *Tracer) AddEvent(span trace.Span, name string, attributes ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// SetAttributes 设置属性
func (t *Tracer) SetAttributes(span trace.Span, attributes ...attribute.KeyValue) {
	span.SetAttributes(attributes...)
}

// GetTraceID 获取追踪ID
func (t *Tracer) GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID 获取SpanID
func (t *Tracer) GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasSpanID() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// Shutdown 关闭追踪器
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t.provider != nil {
		return t.provider.Shutdown(ctx)
	}
	return nil
}

// Extract 从载体中提取追踪上下文
func (t *Tracer) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

// Inject 将追踪上下文注入载体
func (t *Tracer) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

// SpanHelper Span辅助方法
type SpanHelper struct {
	span trace.Span
}

// NewSpanHelper 创建Span辅助器
func NewSpanHelper(span trace.Span) *SpanHelper {
	return &SpanHelper{span: span}
}

// SetHTTPStatus 设置HTTP状态码
func (sh *SpanHelper) SetHTTPStatus(statusCode int) {
	sh.span.SetAttributes(semconv.HTTPStatusCode(statusCode))
	if statusCode >= 400 {
		sh.span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
	}
}

// SetGRPCStatus 设置gRPC状态
func (sh *SpanHelper) SetGRPCStatus(code int, message string) {
	sh.span.SetAttributes(
		attribute.Int("rpc.grpc.status_code", code),
	)
	if code != 0 {
		sh.span.SetStatus(codes.Error, message)
	}
}

// SetDatabaseQuery 设置数据库查询
func (sh *SpanHelper) SetDatabaseQuery(query string) {
	sh.span.SetAttributes(semconv.DBStatement(query))
}

// SetUserInfo 设置用户信息
func (sh *SpanHelper) SetUserInfo(userID, tenantID string) {
	sh.span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("tenant.id", tenantID),
	)
}

// SetAgentInfo 设置Agent信息
func (sh *SpanHelper) SetAgentInfo(agentID, agentType string) {
	sh.span.SetAttributes(
		attribute.String("agent.id", agentID),
		attribute.String("agent.type", agentType),
	)
}

// SetTaskInfo 设置任务信息
func (sh *SpanHelper) SetTaskInfo(taskID, taskType, status string) {
	sh.span.SetAttributes(
		attribute.String("task.id", taskID),
		attribute.String("task.type", taskType),
		attribute.String("task.status", status),
	)
}

// SetModelInfo 设置模型信息
func (sh *SpanHelper) SetModelInfo(provider, model string, tokens int) {
	sh.span.SetAttributes(
		attribute.String("llm.provider", provider),
		attribute.String("llm.model", model),
		attribute.Int("llm.tokens", tokens),
	)
}

// SetCostInfo 设置成本信息
func (sh *SpanHelper) SetCostInfo(cost float64, currency string) {
	sh.span.SetAttributes(
		attribute.Float64("cost.amount", cost),
		attribute.String("cost.currency", currency),
	)
}

// RecordMetric 记录指标
func (sh *SpanHelper) RecordMetric(name string, value float64, unit string) {
	sh.span.AddEvent(name, trace.WithAttributes(
		attribute.Float64("value", value),
		attribute.String("unit", unit),
	))
}

// GlobalTracer 全局追踪器实例
var globalTracer *Tracer

// SetGlobalTracer 设置全局追踪器
func SetGlobalTracer(t *Tracer) {
	globalTracer = t
}

// GetGlobalTracer 获取全局追踪器
func GetGlobalTracer() *Tracer {
	return globalTracer
}

// TracerFromContext 从上下文获取追踪器
func TracerFromContext(ctx context.Context) *Tracer {
	if globalTracer != nil {
		return globalTracer
	}
	return nil
}

// SpanFromContext 从上下文获取当前span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan 创建带span的上下文
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}
