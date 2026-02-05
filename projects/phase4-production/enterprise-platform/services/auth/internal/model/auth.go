package model

import (
	"time"
)

// Permission 权限类型
type Permission string

// 工具权限
const (
	PermissionToolExecute    Permission = "tool:execute"
	PermissionToolRegister   Permission = "tool:register"
	PermissionToolUnregister Permission = "tool:unregister"
	PermissionToolList       Permission = "tool:list"
	PermissionToolView       Permission = "tool:view"
)

// 资源权限
const (
	PermissionResourceRead   Permission = "resource:read"
	PermissionResourceWrite  Permission = "resource:write"
	PermissionResourceDelete Permission = "resource:delete"
	PermissionResourceCreate Permission = "resource:create"
)

// 管理权限
const (
	PermissionUserManage       Permission = "user:manage"
	PermissionRoleManage       Permission = "role:manage"
	PermissionPermissionManage Permission = "permission:manage"
	PermissionAuditView        Permission = "audit:view"
)

// Agent权限（新增）
const (
	PermissionAgentCreate  Permission = "agent:create"
	PermissionAgentExecute Permission = "agent:execute"
	PermissionAgentView    Permission = "agent:view"
	PermissionAgentDelete  Permission = "agent:delete"
)

// 任务权限（新增）
const (
	PermissionTaskCreate Permission = "task:create"
	PermissionTaskView   Permission = "task:view"
	PermissionTaskCancel Permission = "task:cancel"
	PermissionTaskRetry  Permission = "task:retry"
)

// 租户权限（新增）
const (
	PermissionTenantManage Permission = "tenant:manage"
	PermissionTenantView   Permission = "tenant:view"
	PermissionQuotaManage  Permission = "quota:manage"
)

// API权限（新增）
const (
	PermissionAPIRead  Permission = "api:read"
	PermissionAPIWrite Permission = "api:write"
	PermissionAPIAdmin Permission = "api:admin"
)

// Role 角色
type Role struct {
	ID          string       `json:"id" db:"id"`
	TenantID    string       `json:"tenant_id" db:"tenant_id"` // 租户ID（支持租户级别角色）
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	IsSystem    bool         `json:"is_system" db:"is_system"` // 是否系统角色
	ParentID    string       `json:"parent_id" db:"parent_id"` // 父角色ID（支持继承）
	Permissions []Permission `json:"permissions" db:"-"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID         string     `json:"id" db:"id"`
	RoleID     string     `json:"role_id" db:"role_id"`
	Permission Permission `json:"permission" db:"permission"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// User 用户
type User struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // bcrypt hash
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole 用户角色关联
type UserRole struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	RoleID    string    `json:"role_id" db:"role_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Resource 资源
type Resource struct {
	ID        string                 `json:"id" db:"id"`
	TenantID  string                 `json:"tenant_id" db:"tenant_id"`
	Type      string                 `json:"type" db:"type"` // agent, task, tool, file, api
	Name      string                 `json:"name" db:"name"`
	Path      string                 `json:"path" db:"path"`
	Owner     string                 `json:"owner" db:"owner"` // 资源拥有者
	Metadata  map[string]interface{} `json:"metadata" db:"-"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// ResourcePermission 资源级别权限
type ResourcePermission struct {
	ID         string     `json:"id" db:"id"`
	ResourceID string     `json:"resource_id" db:"resource_id"`
	UserID     string     `json:"user_id" db:"user_id"`
	RoleID     string     `json:"role_id" db:"role_id"`
	Permission Permission `json:"permission" db:"permission"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// PolicyRule 策略规则（ABAC属性访问控制）
type PolicyRule struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Subject     string    `json:"subject" db:"subject"`     // 主体：user:*, role:admin
	Resource    string    `json:"resource" db:"resource"`   // 资源：agent:*, task:123
	Action      string    `json:"action" db:"action"`       // 动作：execute, read, write
	Effect      string    `json:"effect" db:"effect"`       // 效果：allow, deny
	Conditions  string    `json:"conditions" db:"conditions"` // 条件：JSON格式
	Priority    int       `json:"priority" db:"priority"`   // 优先级（数字越大优先级越高）
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	Result    string    `json:"result" db:"result"` // success, denied, failure
	Details   string    `json:"details" db:"details"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Duration  int64     `json:"duration" db:"duration"` // 毫秒
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetSystemRoles 获取系统预定义角色
func GetSystemRoles() []*Role {
	return []*Role{
		{
			ID:          "system-admin",
			Name:        "System Administrator",
			Description: "完全的系统访问权限",
			IsSystem:    true,
			Permissions: []Permission{
				// 所有权限
				PermissionToolExecute, PermissionToolRegister, PermissionToolUnregister, PermissionToolList, PermissionToolView,
				PermissionResourceRead, PermissionResourceWrite, PermissionResourceDelete, PermissionResourceCreate,
				PermissionUserManage, PermissionRoleManage, PermissionPermissionManage, PermissionAuditView,
				PermissionAgentCreate, PermissionAgentExecute, PermissionAgentView, PermissionAgentDelete,
				PermissionTaskCreate, PermissionTaskView, PermissionTaskCancel, PermissionTaskRetry,
				PermissionTenantManage, PermissionTenantView, PermissionQuotaManage,
				PermissionAPIRead, PermissionAPIWrite, PermissionAPIAdmin,
			},
		},
		{
			ID:          "tenant-admin",
			Name:        "Tenant Administrator",
			Description: "租户管理员",
			IsSystem:    true,
			Permissions: []Permission{
				PermissionToolExecute, PermissionToolList, PermissionToolView,
				PermissionResourceRead, PermissionResourceWrite, PermissionResourceCreate,
				PermissionUserManage, PermissionRoleManage, PermissionAuditView,
				PermissionAgentCreate, PermissionAgentExecute, PermissionAgentView, PermissionAgentDelete,
				PermissionTaskCreate, PermissionTaskView, PermissionTaskCancel, PermissionTaskRetry,
				PermissionTenantView,
				PermissionAPIRead, PermissionAPIWrite,
			},
		},
		{
			ID:          "developer",
			Name:        "Developer",
			Description: "开发者",
			IsSystem:    true,
			Permissions: []Permission{
				PermissionToolExecute, PermissionToolList, PermissionToolView,
				PermissionResourceRead, PermissionResourceWrite, PermissionResourceCreate,
				PermissionAgentCreate, PermissionAgentExecute, PermissionAgentView,
				PermissionTaskCreate, PermissionTaskView,
				PermissionAPIRead, PermissionAPIWrite,
			},
		},
		{
			ID:          "viewer",
			Name:        "Viewer",
			Description: "查看者",
			IsSystem:    true,
			Permissions: []Permission{
				PermissionToolList, PermissionToolView,
				PermissionResourceRead,
				PermissionAgentView,
				PermissionTaskView,
				PermissionAPIRead,
			},
		},
		{
			ID:          "guest",
			Name:        "Guest",
			Description: "访客",
			IsSystem:    true,
			Permissions: []Permission{
				PermissionToolList,
				PermissionAgentView,
				PermissionTaskView,
			},
		},
	}
}

// HasPermission 检查角色是否有指定权限（支持继承）
func (r *Role) HasPermission(perm Permission) bool {
	for _, p := range r.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// AccessContext 访问上下文
type AccessContext struct {
	TenantID  string
	UserID    string
	Username  string
	Roles     []string
	IPAddress string
	UserAgent string
}
