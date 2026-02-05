package interceptor

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/agent-learning/enterprise-platform/services/auth/internal/model"
	"github.com/agent-learning/enterprise-platform/services/auth/internal/service"
)

// AuthInterceptor 权限拦截器
type AuthInterceptor struct {
	permissionService *service.PermissionService
	jwtService        *service.JWTService
	publicMethods     map[string]bool // 公开方法（不需要认证）
}

// NewAuthInterceptor 创建权限拦截器
func NewAuthInterceptor(permService *service.PermissionService, jwtService *service.JWTService) *AuthInterceptor {
	return &AuthInterceptor{
		permissionService: permService,
		jwtService:        jwtService,
		publicMethods: map[string]bool{
			"/user.UserService/Login":    true,
			"/user.UserService/Register": true,
			"/health.HealthService/Check": true,
		},
	}
}

// UnaryInterceptor gRPC一元拦截器
func (i *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		// 1. 检查是否是公开方法
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 2. 提取访问上下文
		actx, err := i.extractAccessContext(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// 3. 注入上下文
		ctx = i.injectAccessContext(ctx, actx)

		// 4. 检查权限
		err = i.checkMethodPermission(ctx, actx, info.FullMethod)
		if err != nil {
			// 记录拒绝审计
			duration := time.Since(startTime).Milliseconds()
			i.permissionService.AuditAccess(ctx, actx, info.FullMethod, "", "denied", err.Error(), duration)
			return nil, status.Errorf(codes.PermissionDenied, "permission denied: %v", err)
		}

		// 5. 执行方法
		resp, err := handler(ctx, req)

		// 6. 记录审计
		duration := time.Since(startTime).Milliseconds()
		result := "success"
		details := ""
		if err != nil {
			result = "failure"
			details = err.Error()
		}
		i.permissionService.AuditAccess(ctx, actx, info.FullMethod, "", result, details, duration)

		return resp, err
	}
}

// StreamInterceptor gRPC流拦截器
func (i *AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		ctx := ss.Context()

		// 1. 检查是否是公开方法
		if i.publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		// 2. 提取访问上下文
		actx, err := i.extractAccessContext(ctx)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		// 3. 注入上下文
		ctx = i.injectAccessContext(ctx, actx)

		// 4. 检查权限
		err = i.checkMethodPermission(ctx, actx, info.FullMethod)
		if err != nil {
			duration := time.Since(startTime).Milliseconds()
			i.permissionService.AuditAccess(ctx, actx, info.FullMethod, "", "denied", err.Error(), duration)
			return status.Errorf(codes.PermissionDenied, "permission denied: %v", err)
		}

		// 5. 包装ServerStream
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// 6. 执行流处理
		err = handler(srv, wrappedStream)

		// 7. 记录审计
		duration := time.Since(startTime).Milliseconds()
		result := "success"
		details := ""
		if err != nil {
			result = "failure"
			details = err.Error()
		}
		i.permissionService.AuditAccess(ctx, actx, info.FullMethod, "", result, details, duration)

		return err
	}
}

// extractAccessContext 从gRPC metadata提取访问上下文
func (i *AuthInterceptor) extractAccessContext(ctx context.Context) (*model.AccessContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	// 提取JWT Token
	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, fmt.Errorf("missing authorization token")
	}

	token := tokens[0]
	// 移除 "Bearer " 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 验证JWT
	claims, err := i.jwtService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// 提取IP地址
	ipAddress := ""
	if ips := md.Get("x-forwarded-for"); len(ips) > 0 {
		ipAddress = ips[0]
	} else if ips := md.Get("x-real-ip"); len(ips) > 0 {
		ipAddress = ips[0]
	}

	// 提取User Agent
	userAgent := ""
	if uas := md.Get("user-agent"); len(uas) > 0 {
		userAgent = uas[0]
	}

	return &model.AccessContext{
		TenantID:  claims.TenantID,
		UserID:    claims.UserID,
		Username:  claims.Username,
		Roles:     claims.Roles,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}, nil
}

// injectAccessContext 将访问上下文注入到Context
func (i *AuthInterceptor) injectAccessContext(ctx context.Context, actx *model.AccessContext) context.Context {
	return context.WithValue(ctx, model.AccessContextKey, actx)
}

// checkMethodPermission 检查方法权限
func (i *AuthInterceptor) checkMethodPermission(ctx context.Context, actx *model.AccessContext, method string) error {
	// 方法 -> 权限映射
	methodPermissions := map[string]model.Permission{
		// Agent服务
		"/agent.AgentService/CreateAgent":  model.PermissionAgentCreate,
		"/agent.AgentService/ExecuteTask":  model.PermissionAgentExecute,
		"/agent.AgentService/GetAgent":     model.PermissionAgentView,
		"/agent.AgentService/DeleteAgent":  model.PermissionAgentDelete,

		// Task服务
		"/task.TaskService/CreateTask":     model.PermissionTaskCreate,
		"/task.TaskService/GetTask":        model.PermissionTaskView,
		"/task.TaskService/CancelTask":     model.PermissionTaskCancel,
		"/task.TaskService/RetryTask":      model.PermissionTaskRetry,

		// Tool服务
		"/tool.ToolService/ExecuteTool":    model.PermissionToolExecute,
		"/tool.ToolService/RegisterTool":   model.PermissionToolRegister,
		"/tool.ToolService/ListTools":      model.PermissionToolList,

		// Tenant服务
		"/tenant.TenantService/GetTenant":        model.PermissionTenantView,
		"/tenant.TenantService/UpdateTenantQuota": model.PermissionQuotaManage,

		// User服务
		"/user.UserService/CreateUser":     model.PermissionUserManage,
		"/user.UserService/UpdateUser":     model.PermissionUserManage,
		"/user.UserService/DeleteUser":     model.PermissionUserManage,
	}

	// 获取所需权限
	requiredPermission, ok := methodPermissions[method]
	if !ok {
		// 未配置的方法，默认允许（或者可以设置为拒绝）
		return nil
	}

	// 检查权限
	hasPermission, err := i.permissionService.CheckPermission(ctx, actx, requiredPermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("missing required permission: %s", requiredPermission)
	}

	return nil
}

// wrappedServerStream 包装ServerStream以注入新的Context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetAccessContext 从Context获取访问上下文
func GetAccessContext(ctx context.Context) (*model.AccessContext, error) {
	actx, ok := ctx.Value(model.AccessContextKey).(*model.AccessContext)
	if !ok {
		return nil, fmt.Errorf("access context not found")
	}
	return actx, nil
}

// 定义Context键
const (
	AccessContextKey = "access_context"
)
