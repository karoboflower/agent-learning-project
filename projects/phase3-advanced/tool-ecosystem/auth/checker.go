package auth

import (
	"fmt"
	"sync"
)

// PermissionChecker 权限检查器
type PermissionChecker struct {
	roleManager *RoleManager
	userManager *UserManager
	mu          sync.RWMutex
}

// NewPermissionChecker 创建权限检查器
func NewPermissionChecker(roleManager *RoleManager, userManager *UserManager) *PermissionChecker {
	return &PermissionChecker{
		roleManager: roleManager,
		userManager: userManager,
	}
}

// CheckPermission 检查用户是否有指定权限
func (pc *PermissionChecker) CheckPermission(userID string, permission Permission) (bool, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	// 获取用户
	user, err := pc.userManager.GetUser(userID)
	if err != nil {
		return false, err
	}

	// 检查用户的所有角色
	for _, roleID := range user.Roles {
		if pc.roleManager.HasPermission(roleID, permission) {
			return true, nil
		}
	}

	return false, nil
}

// CheckToolExecute 检查工具执行权限
func (pc *PermissionChecker) CheckToolExecute(userID, toolID string) error {
	hasPermission, err := pc.CheckPermission(userID, PermissionToolExecute)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to execute tool %s", userID, toolID)
	}

	return nil
}

// CheckToolRegister 检查工具注册权限
func (pc *PermissionChecker) CheckToolRegister(userID string) error {
	hasPermission, err := pc.CheckPermission(userID, PermissionToolRegister)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to register tools", userID)
	}

	return nil
}

// CheckResourceAccess 检查资源访问权限
func (pc *PermissionChecker) CheckResourceAccess(userID string, resource *Resource, accessLevel AccessLevel) error {
	// 检查是否是资源拥有者
	if resource.Owner == userID {
		return nil // 拥有者有全部权限
	}

	// 根据访问级别检查权限
	switch accessLevel {
	case AccessLevelRead:
		hasPermission, err := pc.CheckPermission(userID, PermissionResourceRead)
		if err != nil {
			return err
		}
		if !hasPermission {
			return fmt.Errorf("user %s does not have read permission for resource %s", userID, resource.ID)
		}

	case AccessLevelWrite:
		hasPermission, err := pc.CheckPermission(userID, PermissionResourceWrite)
		if err != nil {
			return err
		}
		if !hasPermission {
			return fmt.Errorf("user %s does not have write permission for resource %s", userID, resource.ID)
		}

	case AccessLevelAdmin:
		// 需要管理员权限或资源拥有者
		hasPermission, err := pc.CheckPermission(userID, PermissionUserManage)
		if err != nil {
			return err
		}
		if !hasPermission {
			return fmt.Errorf("user %s does not have admin permission for resource %s", userID, resource.ID)
		}

	default:
		return fmt.Errorf("invalid access level: %s", accessLevel)
	}

	return nil
}

// CheckResourceCreate 检查资源创建权限
func (pc *PermissionChecker) CheckResourceCreate(userID string, resourceType ResourceType) error {
	hasPermission, err := pc.CheckPermission(userID, PermissionResourceCreate)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to create %s resources", userID, resourceType)
	}

	return nil
}

// CheckResourceDelete 检查资源删除权限
func (pc *PermissionChecker) CheckResourceDelete(userID string, resource *Resource) error {
	// 只有拥有者或管理员可以删除
	if resource.Owner == userID {
		return nil
	}

	hasPermission, err := pc.CheckPermission(userID, PermissionResourceDelete)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to delete resource %s", userID, resource.ID)
	}

	return nil
}

// CheckUserManagement 检查用户管理权限
func (pc *PermissionChecker) CheckUserManagement(userID string) error {
	hasPermission, err := pc.CheckPermission(userID, PermissionUserManage)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to manage users", userID)
	}

	return nil
}

// CheckRoleManagement 检查角色管理权限
func (pc *PermissionChecker) CheckRoleManagement(userID string) error {
	hasPermission, err := pc.CheckPermission(userID, PermissionRoleManage)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to manage roles", userID)
	}

	return nil
}

// CheckAuditView 检查审计日志查看权限
func (pc *PermissionChecker) CheckAuditView(userID string) error {
	hasPermission, err := pc.CheckPermission(userID, PermissionAuditView)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("user %s does not have permission to view audit logs", userID)
	}

	return nil
}

// GetUserPermissions 获取用户的所有权限
func (pc *PermissionChecker) GetUserPermissions(userID string) ([]Permission, error) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	user, err := pc.userManager.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// 收集所有角色的权限（去重）
	permissionSet := make(map[Permission]bool)

	for _, roleID := range user.Roles {
		role, err := pc.roleManager.GetRole(roleID)
		if err != nil {
			continue
		}

		for _, permission := range role.Permissions {
			permissionSet[permission] = true
		}
	}

	// 转换为切片
	permissions := make([]Permission, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// HasAnyPermission 检查用户是否有任意一个权限
func (pc *PermissionChecker) HasAnyPermission(userID string, permissions []Permission) (bool, error) {
	for _, permission := range permissions {
		hasPermission, err := pc.CheckPermission(userID, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// HasAllPermissions 检查用户是否有所有权限
func (pc *PermissionChecker) HasAllPermissions(userID string, permissions []Permission) (bool, error) {
	for _, permission := range permissions {
		hasPermission, err := pc.CheckPermission(userID, permission)
		if err != nil {
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}

	return true, nil
}

// ResourceManager 资源管理器
type ResourceManager struct {
	resources map[string]*Resource
	mu        sync.RWMutex
}

// NewResourceManager 创建资源管理器
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make(map[string]*Resource),
	}
}

// RegisterResource 注册资源
func (rm *ResourceManager) RegisterResource(resource *Resource) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if resource.ID == "" {
		return fmt.Errorf("resource ID cannot be empty")
	}

	if _, exists := rm.resources[resource.ID]; exists {
		return fmt.Errorf("resource %s already exists", resource.ID)
	}

	if resource.Metadata == nil {
		resource.Metadata = make(map[string]interface{})
	}

	rm.resources[resource.ID] = resource
	return nil
}

// UnregisterResource 注销资源
func (rm *ResourceManager) UnregisterResource(resourceID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.resources[resourceID]; !exists {
		return fmt.Errorf("resource %s not found", resourceID)
	}

	delete(rm.resources, resourceID)
	return nil
}

// GetResource 获取资源
func (rm *ResourceManager) GetResource(resourceID string) (*Resource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resource, exists := rm.resources[resourceID]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", resourceID)
	}

	return resource, nil
}

// ListResources 列出所有资源
func (rm *ResourceManager) ListResources() []*Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resources := make([]*Resource, 0, len(rm.resources))
	for _, resource := range rm.resources {
		resources = append(resources, resource)
	}

	return resources
}

// ListResourcesByType 按类型列出资源
func (rm *ResourceManager) ListResourcesByType(resourceType ResourceType) []*Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resources := make([]*Resource, 0)
	for _, resource := range rm.resources {
		if resource.Type == resourceType {
			resources = append(resources, resource)
		}
	}

	return resources
}

// ListResourcesByOwner 按拥有者列出资源
func (rm *ResourceManager) ListResourcesByOwner(ownerID string) []*Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resources := make([]*Resource, 0)
	for _, resource := range rm.resources {
		if resource.Owner == ownerID {
			resources = append(resources, resource)
		}
	}

	return resources
}
