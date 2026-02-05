package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/agent-learning/enterprise-platform/monitoring/tracing/internal/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HTTPTracingMiddleware HTTP追踪中间件
func HTTPTracingMiddleware(tr *tracer.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从请求头提取追踪上下文
			ctx := tr.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// 开始span
			ctx, span := tr.StartHTTPServerSpan(ctx, r.Method, r.URL.Path)
			defer span.End()

			// 设置请求属性
			span.SetAttributes(
				attribute.String("http.host", r.Host),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.target", r.URL.Path),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
				attribute.String("net.peer.ip", r.RemoteAddr),
			)

			// 提取常用头部
			if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
				span.SetAttributes(attribute.String("request.id", requestID))
			}
			if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
				span.SetAttributes(attribute.String("tenant.id", tenantID))
			}
			if userID := r.Header.Get("X-User-ID"); userID != "" {
				span.SetAttributes(attribute.String("user.id", userID))
			}

			// 包装ResponseWriter以捕获状态码
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 记录开始时间
			startTime := time.Now()

			// 调用下一个处理器
			next.ServeHTTP(rw, r.WithContext(ctx))

			// 记录响应信息
			duration := time.Since(startTime)
			span.SetAttributes(
				attribute.Int("http.status_code", rw.statusCode),
				attribute.Int64("http.response.body.size", rw.bytesWritten),
				attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
			)

			// 设置状态
			if rw.statusCode >= 500 {
				span.SetStatus(trace.StatusCode(trace.StatusError), fmt.Sprintf("HTTP %d", rw.statusCode))
			} else if rw.statusCode >= 400 {
				span.SetStatus(trace.StatusCode(trace.StatusError), fmt.Sprintf("HTTP %d", rw.statusCode))
			}

			// 添加完成事件
			span.AddEvent("http.request.completed", trace.WithAttributes(
				attribute.Int("status_code", rw.statusCode),
				attribute.Int64("duration_ms", duration.Milliseconds()),
			))
		})
	}
}

// responseWriter 包装ResponseWriter以捕获状态码和字节数
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// GRPCUnaryTracingInterceptor gRPC一元调用追踪拦截器
func GRPCUnaryTracingInterceptor(tr *tracer.Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 从metadata提取追踪上下文
		md, _ := metadata.FromIncomingContext(ctx)
		ctx = tr.Extract(ctx, &metadataCarrier{md: md})

		// 开始span
		ctx, span := tr.StartGRPCServerSpan(ctx, info.FullMethod)
		defer span.End()

		// 设置属性
		span.SetAttributes(
			attribute.String("rpc.service", info.FullMethod),
			attribute.String("rpc.system", "grpc"),
		)

		// 提取常用metadata
		if requestID := getMetadataValue(md, "x-request-id"); requestID != "" {
			span.SetAttributes(attribute.String("request.id", requestID))
		}
		if tenantID := getMetadataValue(md, "x-tenant-id"); tenantID != "" {
			span.SetAttributes(attribute.String("tenant.id", tenantID))
		}

		// 记录开始时间
		startTime := time.Now()

		// 调用处理器
		resp, err := handler(ctx, req)

		// 记录持续时间
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Float64("rpc.duration_ms", float64(duration.Milliseconds())),
		)

		// 设置状态
		if err != nil {
			st := status.Convert(err)
			span.SetAttributes(
				attribute.Int("rpc.grpc.status_code", int(st.Code())),
				attribute.String("rpc.grpc.message", st.Message()),
			)
			span.SetStatus(trace.StatusCode(trace.StatusError), st.Message())
			span.RecordError(err)
		} else {
			span.SetAttributes(attribute.Int("rpc.grpc.status_code", 0))
		}

		// 添加完成事件
		span.AddEvent("rpc.request.completed", trace.WithAttributes(
			attribute.Int64("duration_ms", duration.Milliseconds()),
		))

		return resp, err
	}
}

// GRPCStreamTracingInterceptor gRPC流式调用追踪拦截器
func GRPCStreamTracingInterceptor(tr *tracer.Tracer) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// 从metadata提取追踪上下文
		ctx := ss.Context()
		md, _ := metadata.FromIncomingContext(ctx)
		ctx = tr.Extract(ctx, &metadataCarrier{md: md})

		// 开始span
		ctx, span := tr.StartGRPCServerSpan(ctx, info.FullMethod)
		defer span.End()

		// 设置属性
		span.SetAttributes(
			attribute.String("rpc.service", info.FullMethod),
			attribute.String("rpc.system", "grpc"),
			attribute.Bool("rpc.stream", true),
			attribute.Bool("rpc.stream.client", info.IsClientStream),
			attribute.Bool("rpc.stream.server", info.IsServerStream),
		)

		// 记录开始时间
		startTime := time.Now()

		// 包装ServerStream
		wrappedStream := &tracedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// 调用处理器
		err := handler(srv, wrappedStream)

		// 记录持续时间
		duration := time.Since(startTime)
		span.SetAttributes(
			attribute.Float64("rpc.duration_ms", float64(duration.Milliseconds())),
		)

		// 设置状态
		if err != nil {
			st := status.Convert(err)
			span.SetAttributes(
				attribute.Int("rpc.grpc.status_code", int(st.Code())),
				attribute.String("rpc.grpc.message", st.Message()),
			)
			span.SetStatus(trace.StatusCode(trace.StatusError), st.Message())
			span.RecordError(err)
		}

		return err
	}
}

// tracedServerStream 带追踪的ServerStream
type tracedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (ss *tracedServerStream) Context() context.Context {
	return ss.ctx
}

// metadataCarrier metadata载体，实现TextMapCarrier接口
type metadataCarrier struct {
	md metadata.MD
}

func (mc *metadataCarrier) Get(key string) string {
	values := mc.md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (mc *metadataCarrier) Set(key, value string) {
	mc.md.Set(key, value)
}

func (mc *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.md))
	for key := range mc.md {
		keys = append(keys, key)
	}
	return keys
}

// getMetadataValue 获取metadata值
func getMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// DatabaseTracingWrapper 数据库追踪包装器
type DatabaseTracingWrapper struct {
	tracer *tracer.Tracer
	dbType string
}

// NewDatabaseTracingWrapper 创建数据库追踪包装器
func NewDatabaseTracingWrapper(tr *tracer.Tracer, dbType string) *DatabaseTracingWrapper {
	return &DatabaseTracingWrapper{
		tracer: tr,
		dbType: dbType,
	}
}

// WrapQuery 包装查询
func (dtw *DatabaseTracingWrapper) WrapQuery(ctx context.Context, operation, table, query string, fn func() error) error {
	ctx, span := dtw.tracer.StartDatabaseSpan(ctx, dtw.dbType, operation, table)
	defer span.End()

	span.SetAttributes(
		attribute.String("db.statement", query),
	)

	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("db.duration_ms", float64(duration.Milliseconds())),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(trace.StatusCode(trace.StatusError), err.Error())
	}

	return err
}

// CacheTracingWrapper 缓存追踪包装器
type CacheTracingWrapper struct {
	tracer *tracer.Tracer
}

// NewCacheTracingWrapper 创建缓存追踪包装器
func NewCacheTracingWrapper(tr *tracer.Tracer) *CacheTracingWrapper {
	return &CacheTracingWrapper{
		tracer: tr,
	}
}

// WrapOperation 包装操作
func (ctw *CacheTracingWrapper) WrapOperation(ctx context.Context, operation, key string, fn func() (bool, error)) (bool, error) {
	ctx, span := ctw.tracer.StartCacheSpan(ctx, operation, key)
	defer span.End()

	startTime := time.Now()
	hit, err := fn()
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Bool("cache.hit", hit),
		attribute.Float64("cache.duration_ms", float64(duration.Microseconds())/1000.0),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(trace.StatusCode(trace.StatusError), err.Error())
	}

	return hit, err
}

// MessageQueueTracingWrapper 消息队列追踪包装器
type MessageQueueTracingWrapper struct {
	tracer *tracer.Tracer
}

// NewMessageQueueTracingWrapper 创建消息队列追踪包装器
func NewMessageQueueTracingWrapper(tr *tracer.Tracer) *MessageQueueTracingWrapper {
	return &MessageQueueTracingWrapper{
		tracer: tr,
	}
}

// WrapPublish 包装发布操作
func (mqtw *MessageQueueTracingWrapper) WrapPublish(ctx context.Context, queue, message string, fn func(context.Context) error) error {
	ctx, span := mqtw.tracer.StartMessageQueueSpan(ctx, "publish", queue)
	defer span.End()

	span.SetAttributes(
		attribute.Int("messaging.message.payload_size", len(message)),
	)

	// 注入追踪上下文到消息
	carrier := propagation.MapCarrier{}
	mqtw.tracer.Inject(ctx, carrier)

	startTime := time.Now()
	err := fn(ctx)
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("messaging.duration_ms", float64(duration.Milliseconds())),
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(trace.StatusCode(trace.StatusError), err.Error())
	}

	return err
}

// HTTPClientTracingWrapper HTTP客户端追踪包装器
type HTTPClientTracingWrapper struct {
	tracer *tracer.Tracer
}

// NewHTTPClientTracingWrapper 创建HTTP客户端追踪包装器
func NewHTTPClientTracingWrapper(tr *tracer.Tracer) *HTTPClientTracingWrapper {
	return &HTTPClientTracingWrapper{
		tracer: tr,
	}
}

// WrapRequest 包装HTTP请求
func (hctw *HTTPClientTracingWrapper) WrapRequest(ctx context.Context, method, url string, fn func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	ctx, span := hctw.tracer.StartHTTPClientSpan(ctx, method, url)
	defer span.End()

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// 注入追踪上下文
	hctw.tracer.Inject(ctx, propagation.HeaderCarrier(req.Header))

	startTime := time.Now()
	resp, err := fn(req)
	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Float64("http.client.duration_ms", float64(duration.Milliseconds())),
	)

	if resp != nil {
		span.SetAttributes(
			attribute.Int("http.status_code", resp.StatusCode),
			attribute.Int64("http.response.body.size", resp.ContentLength),
		)
		if resp.StatusCode >= 400 {
			span.SetStatus(trace.StatusCode(trace.StatusError), fmt.Sprintf("HTTP %d", resp.StatusCode))
		}
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(trace.StatusCode(trace.StatusError), err.Error())
	}

	return resp, err
}
