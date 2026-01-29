package auth

import (
	"fmt"
)

// AuthorizationManager 授权管理器（集成所有组件）
type AuthorizationManager struct {
	roleManager       *RoleManager
	userManager       *UserManager
	permissionChecker *PermissionChecker
	resourceManager   *ResourceManager
	auditLogger       *AuditLogger
}

// NewAuthorizationManager 创建授权管理器
func NewAuthorizationManager() *AuthorizationManager {
	roleManager := NewRoleManager()
	userManager := NewUserManager()
	permissionChecker := NewPermissionChecker(roleManager, userManager)
	resourceManager := NewResourceManager()
	auditLogger := NewAuditLogger(10000)

	// 添加控制台审计处理器
	auditLogger.AddHandler(&ConsoleAuditHandler{})

	return &AuthorizationManager{
		roleManager:       roleManager,
		userManager:       userManager,
		permissionChecker: permissionChecker,
		resourceManager:   resourceManager,
		auditLogger:       auditLogger,
	}
}

// GetRoleManager 获取角色管理器
func (am *AuthorizationManager) GetRoleManager() *RoleManager {
	return am.roleManager
}

// GetUserManager 获取用户管理器
func (am *AuthorizationManager) GetUserManager() *UserManager {
	return am.userManager
}

// GetPermissionChecker 获取权限检查器
func (am *AuthorizationManager) GetPermissionChecker() *PermissionChecker {
	return am.permissionChecker
}

// GetResourceManager 获取资源管理器
func (am *AuthorizationManager) GetResourceManager() *ResourceManager {
	return am.resourceManager
}

// GetAuditLogger 获取审计日志记录器
func (am *AuthorizationManager) GetAuditLogger() *AuditLogger {
	return am.auditLogger
}

// AuthorizeToolExecution 授权工具执行
func (am *AuthorizationManager) AuthorizeToolExecution(userID, username, toolID string) error {
	// 检查权限
	err := am.permissionChecker.CheckToolExecute(userID, toolID)

	// 记录审计日志
	result := AuditResultSuccess
	details := fmt.Sprintf("Tool execution authorized: %s", toolID)
	if err != nil {
		result = AuditResultDenied
		details = fmt.Sprintf("Tool execution denied: %s - %v", toolID, err)
	}

	am.auditLogger.LogToolExecution(userID, username, toolID, result, details, 0)

	return err
}

// AuthorizeResourceAccess 授权资源访问
func (am *AuthorizationManager) AuthorizeResourceAccess(userID, username, resourceID string, accessLevel AccessLevel) error {
	// 获取资源
	resource, err := am.resourceManager.GetResource(resourceID)
	if err != nil {
		am.auditLogger.LogResourceAccess(userID, username, resourceID, AuditActionResourceRead, AuditResultFailure, fmt.Sprintf("Resource not found: %v", err))
		return err
	}

	// 检查权限
	err = am.permissionChecker.CheckResourceAccess(userID, resource, accessLevel)

	// 记录审计日志
	var action AuditAction
	switch accessLevel {
	case AccessLevelRead:
		action = AuditActionResourceRead
	case AccessLevelWrite:
		action = AuditActionResourceWrite
	default:
		action = AuditActionResourceRead
	}

	result := AuditResultSuccess
	details := fmt.Sprintf("Resource access authorized: %s (level: %s)", resourceID, accessLevel)
	if err != nil {
		result = AuditResultDenied
		details = fmt.Sprintf("Resource access denied: %s - %v", resourceID, err)
	}

	am.auditLogger.LogResourceAccess(userID, username, resourceID, action, result, details)

	return err
}

// CreateUserWithRole 创建用户并分配角色
func (am *AuthorizationManager) CreateUserWithRole(operatorID, operatorName string, user *User, roleID string) error {
	// 检查操作者权限
	if err := am.permissionChecker.CheckUserManagement(operatorID); err != nil {
		am.auditLogger.LogUserAction(operatorID, operatorName, AuditActionUserCreate, user.ID, AuditResultDenied, fmt.Sprintf("Permission denied: %v", err))
		return err
	}

	// 创建用户
	if err := am.userManager.CreateUser(user); err != nil {
		am.auditLogger.LogUserAction(operatorID, operatorName, AuditActionUserCreate, user.ID, AuditResultFailure, fmt.Sprintf("Failed to create user: %v", err))
		return err
	}

	// 分配角色
	if roleID != "" {
		if err := am.userManager.AssignRole(user.ID, roleID); err != nil {
			am.auditLogger.LogRoleAction(operatorID, operatorName, roleID, AuditActionRoleAssign, AuditResultFailure, fmt.Sprintf("Failed to assign role: %v", err))
			return err
		}
	}

	am.auditLogger.LogUserAction(operatorID, operatorName, AuditActionUserCreate, user.ID, AuditResultSuccess, fmt.Sprintf("User created with role: %s", roleID))

	return nil
}

// AssignRoleToUser 为用户分配角色
func (am *AuthorizationManager) AssignRoleToUser(operatorID, operatorName, userID, roleID string) error {
	// 检查操作者权限
	if err := am.permissionChecker.CheckRoleManagement(operatorID); err != nil {
		am.auditLogger.LogRoleAction(operatorID, operatorName, roleID, AuditActionRoleAssign, AuditResultDenied, fmt.Sprintf("Permission denied: %v", err))
		return err
	}

	// 分配角色
	if err := am.userManager.AssignRole(userID, roleID); err != nil {
		am.auditLogger.LogRoleAction(operatorID, operatorName, roleID, AuditActionRoleAssign, AuditResultFailure, fmt.Sprintf("Failed to assign role: %v", err))
		return err
	}

	am.auditLogger.LogRoleAction(operatorID, operatorName, roleID, AuditActionRoleAssign, AuditResultSuccess, fmt.Sprintf("Role assigned to user: %s", userID))

	return nil
}

// RegisterResourceWithOwner 注册资源并设置拥有者
func (am *AuthorizationManager) RegisterResourceWithOwner(userID, username string, resource *Resource) error {
	// 检查权限
	if err := am.permissionChecker.CheckResourceCreate(userID, resource.Type); err != nil {
		am.auditLogger.LogResourceAccess(userID, username, resource.ID, AuditActionResourceCreate, AuditResultDenied, fmt.Sprintf("Permission denied: %v", err))
		return err
	}

	// 设置拥有者
	resource.Owner = userID

	// 注册资源
	if err := am.resourceManager.RegisterResource(resource); err != nil {
		am.auditLogger.LogResourceAccess(userID, username, resource.ID, AuditActionResourceCreate, AuditResultFailure, fmt.Sprintf("Failed to register resource: %v", err))
		return err
	}

	am.auditLogger.LogResourceAccess(userID, username, resource.ID, AuditActionResourceCreate, AuditResultSuccess, fmt.Sprintf("Resource registered: %s (type: %s)", resource.ID, resource.Type))

	return nil
}
