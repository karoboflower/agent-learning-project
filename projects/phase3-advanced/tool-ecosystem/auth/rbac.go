package auth

import (
	"fmt"
	"sync"
	"time"
)

// Permission 权限
type Permission string

const (
	// 工具权限
	PermissionToolExecute    Permission = "tool:execute"     // 执行工具
	PermissionToolRegister   Permission = "tool:register"    // 注册工具
	PermissionToolUnregister Permission = "tool:unregister"  // 注销工具
	PermissionToolList       Permission = "tool:list"        // 列出工具
	PermissionToolView       Permission = "tool:view"        // 查看工具详情

	// 资源权限
	PermissionResourceRead   Permission = "resource:read"    // 读取资源
	PermissionResourceWrite  Permission = "resource:write"   // 写入资源
	PermissionResourceDelete Permission = "resource:delete"  // 删除资源
	PermissionResourceCreate Permission = "resource:create"  // 创建资源

	// 管理权限
	PermissionUserManage      Permission = "user:manage"      // 管理用户
	PermissionRoleManage      Permission = "role:manage"      // 管理角色
	PermissionPermissionManage Permission = "permission:manage" // 管理权限
	PermissionAuditView       Permission = "audit:view"       // 查看审计日志
)

// Role 角色
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// User 用户
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"` // Role IDs
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeFile     ResourceType = "file"
	ResourceTypeAPI      ResourceType = "api"
	ResourceTypeDatabase ResourceType = "database"
	ResourceTypeTool     ResourceType = "tool"
	ResourceTypeAgent    ResourceType = "agent"
)

// Resource 资源
type Resource struct {
	ID       string       `json:"id"`
	Type     ResourceType `json:"type"`
	Path     string       `json:"path"` // 资源路径或标识
	Owner    string       `json:"owner"` // 拥有者用户ID
	Metadata map[string]interface{} `json:"metadata"`
}

// AccessLevel 访问级别
type AccessLevel string

const (
	AccessLevelNone   AccessLevel = "none"   // 无权限
	AccessLevelRead   AccessLevel = "read"   // 只读
	AccessLevelWrite  AccessLevel = "write"  // 读写
	AccessLevelAdmin  AccessLevel = "admin"  // 管理员
)

// RoleManager 角色管理器
type RoleManager struct {
	roles map[string]*Role
	mu    sync.RWMutex
}

// NewRoleManager 创建角色管理器
func NewRoleManager() *RoleManager {
	rm := &RoleManager{
		roles: make(map[string]*Role),
	}

	// 初始化默认角色
	rm.initDefaultRoles()

	return rm
}

// initDefaultRoles 初始化默认角色
func (rm *RoleManager) initDefaultRoles() {
	// 管理员角色
	adminRole := &Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full system access",
		Permissions: []Permission{
			PermissionToolExecute,
			PermissionToolRegister,
			PermissionToolUnregister,
			PermissionToolList,
			PermissionToolView,
			PermissionResourceRead,
			PermissionResourceWrite,
			PermissionResourceDelete,
			PermissionResourceCreate,
			PermissionUserManage,
			PermissionRoleManage,
			PermissionPermissionManage,
			PermissionAuditView,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 开发者角色
	developerRole := &Role{
		ID:          "developer",
		Name:        "Developer",
		Description: "Can execute tools and manage resources",
		Permissions: []Permission{
			PermissionToolExecute,
			PermissionToolList,
			PermissionToolView,
			PermissionResourceRead,
			PermissionResourceWrite,
			PermissionResourceCreate,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 只读角色
	viewerRole := &Role{
		ID:          "viewer",
		Name:        "Viewer",
		Description: "Read-only access",
		Permissions: []Permission{
			PermissionToolList,
			PermissionToolView,
			PermissionResourceRead,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Guest角色
	guestRole := &Role{
		ID:          "guest",
		Name:        "Guest",
		Description: "Limited access",
		Permissions: []Permission{
			PermissionToolList,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rm.roles["admin"] = adminRole
	rm.roles["developer"] = developerRole
	rm.roles["viewer"] = viewerRole
	rm.roles["guest"] = guestRole
}

// CreateRole 创建角色
func (rm *RoleManager) CreateRole(role *Role) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if role.ID == "" {
		return fmt.Errorf("role ID cannot be empty")
	}

	if role.Name == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if _, exists := rm.roles[role.ID]; exists {
		return fmt.Errorf("role %s already exists", role.ID)
	}

	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	rm.roles[role.ID] = role
	return nil
}

// UpdateRole 更新角色
func (rm *RoleManager) UpdateRole(role *Role) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.roles[role.ID]; !exists {
		return fmt.Errorf("role %s not found", role.ID)
	}

	role.UpdatedAt = time.Now()
	rm.roles[role.ID] = role
	return nil
}

// DeleteRole 删除角色
func (rm *RoleManager) DeleteRole(roleID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 不允许删除默认角色
	if roleID == "admin" || roleID == "developer" || roleID == "viewer" || roleID == "guest" {
		return fmt.Errorf("cannot delete default role: %s", roleID)
	}

	if _, exists := rm.roles[roleID]; !exists {
		return fmt.Errorf("role %s not found", roleID)
	}

	delete(rm.roles, roleID)
	return nil
}

// GetRole 获取角色
func (rm *RoleManager) GetRole(roleID string) (*Role, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	role, exists := rm.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role %s not found", roleID)
	}

	return role, nil
}

// ListRoles 列出所有角色
func (rm *RoleManager) ListRoles() []*Role {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	roles := make([]*Role, 0, len(rm.roles))
	for _, role := range rm.roles {
		roles = append(roles, role)
	}

	return roles
}

// HasPermission 检查角色是否有指定权限
func (rm *RoleManager) HasPermission(roleID string, permission Permission) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	role, exists := rm.roles[roleID]
	if !exists {
		return false
	}

	for _, p := range role.Permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// AddPermission 为角色添加权限
func (rm *RoleManager) AddPermission(roleID string, permission Permission) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	role, exists := rm.roles[roleID]
	if !exists {
		return fmt.Errorf("role %s not found", roleID)
	}

	// 检查是否已存在
	for _, p := range role.Permissions {
		if p == permission {
			return nil // 已存在，不需要添加
		}
	}

	role.Permissions = append(role.Permissions, permission)
	role.UpdatedAt = time.Now()

	return nil
}

// RemovePermission 移除角色的权限
func (rm *RoleManager) RemovePermission(roleID string, permission Permission) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	role, exists := rm.roles[roleID]
	if !exists {
		return fmt.Errorf("role %s not found", roleID)
	}

	// 查找并移除
	for i, p := range role.Permissions {
		if p == permission {
			role.Permissions = append(role.Permissions[:i], role.Permissions[i+1:]...)
			role.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("permission %s not found in role %s", permission, roleID)
}

// UserManager 用户管理器
type UserManager struct {
	users map[string]*User
	mu    sync.RWMutex
}

// NewUserManager 创建用户管理器
func NewUserManager() *UserManager {
	return &UserManager{
		users: make(map[string]*User),
	}
}

// CreateUser 创建用户
func (um *UserManager) CreateUser(user *User) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if user.ID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if _, exists := um.users[user.ID]; exists {
		return fmt.Errorf("user %s already exists", user.ID)
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	um.users[user.ID] = user
	return nil
}

// UpdateUser 更新用户
func (um *UserManager) UpdateUser(user *User) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.users[user.ID]; !exists {
		return fmt.Errorf("user %s not found", user.ID)
	}

	user.UpdatedAt = time.Now()
	um.users[user.ID] = user
	return nil
}

// DeleteUser 删除用户
func (um *UserManager) DeleteUser(userID string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.users[userID]; !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	delete(um.users, userID)
	return nil
}

// GetUser 获取用户
func (um *UserManager) GetUser(userID string) (*User, error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	user, exists := um.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return user, nil
}

// ListUsers 列出所有用户
func (um *UserManager) ListUsers() []*User {
	um.mu.RLock()
	defer um.mu.RUnlock()

	users := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}

	return users
}

// AssignRole 为用户分配角色
func (um *UserManager) AssignRole(userID, roleID string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	// 检查是否已有该角色
	for _, r := range user.Roles {
		if r == roleID {
			return nil // 已存在
		}
	}

	user.Roles = append(user.Roles, roleID)
	user.UpdatedAt = time.Now()

	return nil
}

// RevokeRole 撤销用户的角色
func (um *UserManager) RevokeRole(userID, roleID string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	user, exists := um.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	// 查找并移除
	for i, r := range user.Roles {
		if r == roleID {
			user.Roles = append(user.Roles[:i], user.Roles[i+1:]...)
			user.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("role %s not found for user %s", roleID, userID)
}

// GetUserRoles 获取用户的所有角色ID
func (um *UserManager) GetUserRoles(userID string) ([]string, error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	user, exists := um.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return user.Roles, nil
}
