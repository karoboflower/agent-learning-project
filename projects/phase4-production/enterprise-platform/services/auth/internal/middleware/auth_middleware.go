package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/agent-learning/enterprise-platform/services/auth/internal/model"
	"github.com/agent-learning/enterprise-platform/services/auth/internal/service"
)

// AuthMiddleware HTTP权限中间件
type AuthMiddleware struct {
	permissionService *service.PermissionService
	jwtService        *service.JWTService
	publicPaths       map[string]bool
}

// NewAuthMiddleware 创建HTTP权限中间件
func NewAuthMiddleware(permService *service.PermissionService, jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		permissionService: permService,
		jwtService:        jwtService,
		publicPaths: map[string]bool{
			"/api/v1/auth/login":    true,
			"/api/v1/auth/register": true,
			"/api/v1/health":        true,
			"/metrics":              true,
		},
	}
}

// Authenticate 认证中间件
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// 1. 检查是否是公开路径
		if m.publicPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// 2. 提取JWT Token
		token := m.extractToken(r)
		if token == "" {
			m.respondError(w, http.StatusUnauthorized, "missing authorization token")
			return
		}

		// 3. 验证Token
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			m.respondError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
			return
		}

		// 4. 构建访问上下文
		actx := &model.AccessContext{
			TenantID:  claims.TenantID,
			UserID:    claims.UserID,
			Username:  claims.Username,
			Roles:     claims.Roles,
			IPAddress: m.getIPAddress(r),
			UserAgent: r.UserAgent(),
		}

		// 5. 注入上下文
		ctx := context.WithValue(r.Context(), model.AccessContextKey, actx)

		// 6. 检查API权限
		err = m.permissionService.CheckAPIAccess(ctx, actx, r.Method, r.URL.Path)
		if err != nil {
			// 记录审计
			duration := time.Since(startTime).Milliseconds()
			m.permissionService.AuditAccess(ctx, actx, r.Method+" "+r.URL.Path, "", "denied", err.Error(), duration)

			m.respondError(w, http.StatusForbidden, "permission denied: "+err.Error())
			return
		}

		// 7. 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))

		// 8. 记录审计
		duration := time.Since(startTime).Milliseconds()
		m.permissionService.AuditAccess(ctx, actx, r.Method+" "+r.URL.Path, "", "success", "", duration)
	})
}

// TenantIsolation 租户隔离中间件
func (m *AuthMiddleware) TenantIsolation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从上下文获取访问上下文
		actx, ok := r.Context().Value(model.AccessContextKey).(*model.AccessContext)
		if !ok {
			m.respondError(w, http.StatusUnauthorized, "access context not found")
			return
		}

		// 从URL或请求体提取租户ID并验证
		// 这里简化处理，实际应该从路径参数或查询参数提取
		// 例如: /api/v1/tenants/{tenant_id}/agents

		// 验证租户ID匹配
		// if requestTenantID != actx.TenantID {
		//     m.respondError(w, http.StatusForbidden, "tenant mismatch")
		//     return
		// }

		next.ServeHTTP(w, r)
	})
}

// RateLimiting 速率限制中间件
func (m *AuthMiddleware) RateLimiting(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actx, ok := r.Context().Value(model.AccessContextKey).(*model.AccessContext)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		// TODO: 实现速率限制逻辑
		// 1. 检查租户的API调用配额
		// 2. 使用Redis记录API调用次数
		// 3. 如果超限，返回429错误

		_ = actx // 避免未使用变量警告

		next.ServeHTTP(w, r)
	})
}

// Logging 日志中间件
func (m *AuthMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// 创建响应包装器以捕获状态码
		wrapped := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// 记录请求日志
		duration := time.Since(startTime)
		// log.Printf("[%s] %s %s %d %v", r.Method, r.URL.Path, m.getIPAddress(r), wrapped.statusCode, duration)
		_ = duration // 避免未使用变量警告
	})
}

// CORS CORS中间件
func (m *AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractToken 从请求中提取JWT Token
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	// 1. 从Authorization header提取
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// 移除 "Bearer " 前缀
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
		return auth
	}

	// 2. 从Cookie提取
	cookie, err := r.Cookie("token")
	if err == nil {
		return cookie.Value
	}

	// 3. 从查询参数提取
	return r.URL.Query().Get("token")
}

// getIPAddress 获取客户端IP地址
func (m *AuthMiddleware) getIPAddress(r *http.Request) string {
	// 1. X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// 2. X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 3. RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// respondError 返回错误响应
func (m *AuthMiddleware) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + message + `"}`))
}

// responseWrapper 响应包装器
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetAccessContext 从HTTP请求上下文获取访问上下文
func GetAccessContext(r *http.Request) (*model.AccessContext, error) {
	actx, ok := r.Context().Value(model.AccessContextKey).(*model.AccessContext)
	if !ok {
		return nil, fmt.Errorf("access context not found")
	}
	return actx, nil
}

// RequirePermission 要求特定权限的中间件
func (m *AuthMiddleware) RequirePermission(permission model.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actx, err := GetAccessContext(r)
			if err != nil {
				m.respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hasPermission, err := m.permissionService.CheckPermission(r.Context(), actx, permission)
			if err != nil || !hasPermission {
				m.respondError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole 要求特定角色的中间件
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actx, err := GetAccessContext(r)
			if err != nil {
				m.respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hasRole := false
			for _, r := range actx.Roles {
				if r == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.respondError(w, http.StatusForbidden, "insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
