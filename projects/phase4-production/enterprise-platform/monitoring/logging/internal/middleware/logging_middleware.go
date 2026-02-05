package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/agent-learning/enterprise-platform/monitoring/logging/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// HTTPLoggingMiddleware HTTP日志中间件
func HTTPLoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应记录器
			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				written:        0,
			}

			// 添加请求ID到上下文
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			ctx := context.WithValue(r.Context(), "request_id", requestID)
			r = r.WithContext(ctx)

			// 记录请求开始
			log.WithContext(ctx).Info("HTTP request started",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.String("remote_addr", r.RemoteAddr),
				logger.String("user_agent", r.UserAgent()),
			)

			// 处理请求
			next.ServeHTTP(rec, r)

			// 记录请求完成
			duration := time.Since(start)
			log.WithContext(ctx).Info("HTTP request completed",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", rec.statusCode),
				logger.Int("size", rec.written),
				logger.Duration("duration", duration),
			)
		})
	}
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.written += n
	return n, err
}

// GRPCUnaryLoggingInterceptor gRPC一元调用日志拦截器
func GRPCUnaryLoggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// 从metadata提取信息
		md, _ := metadata.FromIncomingContext(ctx)
		requestID := extractMetadata(md, "x-request-id")
		tenantID := extractMetadata(md, "x-tenant-id")

		// 添加到上下文
		ctx = context.WithValue(ctx, "request_id", requestID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		// 记录请求开始
		log.WithContext(ctx).Info("gRPC request started",
			logger.String("method", info.FullMethod),
		)

		// 处理请求
		resp, err := handler(ctx, req)

		// 记录请求完成
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.WithContext(ctx).Error("gRPC request failed",
				logger.String("method", info.FullMethod),
				logger.String("code", st.Code().String()),
				logger.String("error", st.Message()),
				logger.Duration("duration", duration),
			)
		} else {
			log.WithContext(ctx).Info("gRPC request completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
			)
		}

		return resp, err
	}
}

// GRPCStreamLoggingInterceptor gRPC流式调用日志拦截器
func GRPCStreamLoggingInterceptor(log *logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// 从metadata提取信息
		md, _ := metadata.FromIncomingContext(ss.Context())
		requestID := extractMetadata(md, "x-request-id")
		tenantID := extractMetadata(md, "x-tenant-id")

		// 添加到上下文
		ctx := context.WithValue(ss.Context(), "request_id", requestID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		// 记录流开始
		log.WithContext(ctx).Info("gRPC stream started",
			logger.String("method", info.FullMethod),
			logger.Bool("client_stream", info.IsClientStream),
			logger.Bool("server_stream", info.IsServerStream),
		)

		// 包装流
		wrappedStream := &loggingServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// 处理流
		err := handler(srv, wrappedStream)

		// 记录流完成
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.WithContext(ctx).Error("gRPC stream failed",
				logger.String("method", info.FullMethod),
				logger.String("code", st.Code().String()),
				logger.String("error", st.Message()),
				logger.Duration("duration", duration),
			)
		} else {
			log.WithContext(ctx).Info("gRPC stream completed",
				logger.String("method", info.FullMethod),
				logger.Duration("duration", duration),
			)
		}

		return err
	}
}

// loggingServerStream 日志流包装器
type loggingServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *loggingServerStream) Context() context.Context {
	return s.ctx
}

// extractMetadata 从metadata提取值
func extractMetadata(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// RecoveryMiddleware 恢复中间件（捕获panic）
func RecoveryMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.WithContext(r.Context()).Error("Panic recovered",
						logger.String("method", r.Method),
						logger.String("path", r.URL.Path),
						logger.Any("panic", err),
					)

					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// GRPCRecoveryInterceptor gRPC恢复拦截器
func GRPCRecoveryInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.WithContext(ctx).Error("Panic recovered",
					logger.String("method", info.FullMethod),
					logger.Any("panic", r),
				)
				err = status.Errorf(500, "Internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
