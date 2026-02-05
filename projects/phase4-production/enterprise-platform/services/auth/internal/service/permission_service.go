package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/agent-learning/enterprise-platform/services/auth/internal/model"
	"github.com/agent-learning/enterprise-platform/services/auth/internal/repository"
)

// PermissionService 权限服务
type PermissionService struct {
	roleRepo     *repository.RoleRepository
	userRepo     *repository.UserRepository
	resourceRepo *repository.ResourceRepository
	policyRepo   *repository.PolicyRepository
	auditRepo    *repository.AuditRepository
}

// NewPermissionService 创建权限服务
func NewPermissionService(
	roleRepo *repository.RoleRepository,
	userRepo *repository.UserRepository,
	resourceRepo *repository.ResourceRepository,
	policyRepo *repository.PolicyRepository,
	auditRepo *repository.AuditRepository,
) *PermissionService {
	return &PermissionService{
		roleRepo:     roleRepo,
		userRepo:     userRepo,
		resourceRepo: resourceRepo,
		policyRepo:   policyRepo,
		auditRepo:    auditRepo,
	}
}

// CheckPermission 检查用户是否有指定权限
func (s *PermissionService) CheckPermission(ctx context.Context, actx *model.AccessContext, permission model.Permission) (bool, error) {
	// 1. 获取用户角色
	roles, err := s.userRepo.GetUserRoles(ctx, actx.UserID)
	if err != nil {
		return false, err
	}

	// 2. 检查角色权限（包括继承）
	for _, roleID := range roles {
		hasPermission, err := s.checkRolePermission(ctx, roleID, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// checkRolePermission 检查角色权限（支持继承）
func (s *PermissionService) checkRolePermission(ctx context.Context, roleID string, permission model.Permission) (bool, error) {
	role, err := s.roleRepo.GetRole(ctx, roleID)
	if err != nil {
		return false, err
	}

	// 检查当前角色
	permissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	// 检查父角色（继承）
	if role.ParentID != "" {
		return s.checkRolePermission(ctx, role.ParentID, permission)
	}

	return false, nil
}

// CheckResourceAccess 检查资源访问权限
func (s *PermissionService) CheckResourceAccess(ctx context.Context, actx *model.AccessContext, resourceID string, permission model.Permission) error {
	// 1. 获取资源
	resource, err := s.resourceRepo.GetResource(ctx, resourceID)
	if err != nil {
		return fmt.Errorf("resource not found: %w", err)
	}

	// 2. 验证租户
	if resource.TenantID != actx.TenantID {
		return fmt.Errorf("tenant mismatch")
	}

	// 3. 检查是否是资源拥有者
	if resource.Owner == actx.UserID {
		return nil // 拥有者有完全权限
	}

	// 4. 检查资源级别权限
	hasResourcePerm, err := s.resourceRepo.CheckResourcePermission(ctx, resourceID, actx.UserID, permission)
	if err == nil && hasResourcePerm {
		return nil
	}

	// 5. 检查角色权限
	hasRolePerm, err := s.CheckPermission(ctx, actx, permission)
	if err != nil {
		return err
	}

	if !hasRolePerm {
		return fmt.Errorf("permission denied: %s", permission)
	}

	return nil
}

// CheckAPIAccess 检查API访问权限
func (s *PermissionService) CheckAPIAccess(ctx context.Context, actx *model.AccessContext, method, path string) error {
	// 1. 根据HTTP方法确定所需权限
	var requiredPermission model.Permission

	switch method {
	case "GET", "HEAD":
		requiredPermission = model.PermissionAPIRead
	case "POST", "PUT", "PATCH", "DELETE":
		requiredPermission = model.PermissionAPIWrite
	default:
		return fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// 2. 检查权限
	hasPermission, err := s.CheckPermission(ctx, actx, requiredPermission)
	if err != nil {
		return err
	}

	if !hasPermission {
		return fmt.Errorf("insufficient API permissions for %s %s", method, path)
	}

	// 3. 检查策略规则（ABAC）
	allowed, err := s.checkPolicyRules(ctx, actx, method, path)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("access denied by policy")
	}

	return nil
}

// checkPolicyRules 检查策略规则（ABAC）
func (s *PermissionService) checkPolicyRules(ctx context.Context, actx *model.AccessContext, method, path string) (bool, error) {
	// 获取租户的所有策略
	policies, err := s.policyRepo.GetTenantPolicies(ctx, actx.TenantID)
	if err != nil {
		return false, err
	}

	if len(policies) == 0 {
		return true, nil // 没有策略，默认允许
	}

	// 按优先级排序，优先级高的先匹配
	// 这里简化处理，实际应该排序
	var matchedPolicy *model.PolicyRule
	maxPriority := -1

	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		// 匹配主体
		if !s.matchSubject(policy.Subject, actx) {
			continue
		}

		// 匹配资源
		if !s.matchResource(policy.Resource, path) {
			continue
		}

		// 匹配动作
		if !s.matchAction(policy.Action, method) {
			continue
		}

		// 找到优先级最高的匹配策略
		if policy.Priority > maxPriority {
			maxPriority = policy.Priority
			matchedPolicy = policy
		}
	}

	// 如果找到匹配的策略
	if matchedPolicy != nil {
		return matchedPolicy.Effect == "allow", nil
	}

	// 没有匹配的策略，默认允许
	return true, nil
}

// matchSubject 匹配主体
func (s *PermissionService) matchSubject(subject string, actx *model.AccessContext) bool {
	if subject == "*" {
		return true
	}

	// user:123 或 role:admin
	parts := strings.SplitN(subject, ":", 2)
	if len(parts) != 2 {
		return false
	}

	switch parts[0] {
	case "user":
		return parts[1] == actx.UserID || parts[1] == "*"
	case "role":
		for _, role := range actx.Roles {
			if role == parts[1] || parts[1] == "*" {
				return true
			}
		}
	}

	return false
}

// matchResource 匹配资源
func (s *PermissionService) matchResource(resource, path string) bool {
	if resource == "*" {
		return true
	}

	// agent:*, task:123, /api/v1/agents/*
	if strings.HasSuffix(resource, "*") {
		prefix := strings.TrimSuffix(resource, "*")
		return strings.HasPrefix(path, prefix)
	}

	return resource == path
}

// matchAction 匹配动作
func (s *PermissionService) matchAction(action, method string) bool {
	if action == "*" {
		return true
	}

	// 动作映射
	actionMap := map[string][]string{
		"read":   {"GET", "HEAD"},
		"write":  {"POST", "PUT", "PATCH"},
		"delete": {"DELETE"},
	}

	if methods, ok := actionMap[action]; ok {
		for _, m := range methods {
			if m == method {
				return true
			}
		}
	}

	return action == method
}

// AuditAccess 记录访问审计
func (s *PermissionService) AuditAccess(ctx context.Context, actx *model.AccessContext, action, resource, result, details string, duration int64) error {
	audit := &model.AuditLog{
		TenantID:  actx.TenantID,
		UserID:    actx.UserID,
		Username:  actx.Username,
		Action:    action,
		Resource:  resource,
		Result:    result,
		Details:   details,
		IPAddress: actx.IPAddress,
		UserAgent: actx.UserAgent,
		Duration:  duration,
	}

	return s.auditRepo.CreateAuditLog(ctx, audit)
}

// GetUserPermissions 获取用户的所有权限
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID string) ([]model.Permission, error) {
	roles, err := s.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permissionMap := make(map[model.Permission]bool)

	for _, roleID := range roles {
		permissions, err := s.getRoleAllPermissions(ctx, roleID)
		if err != nil {
			return nil, err
		}

		for _, perm := range permissions {
			permissionMap[perm] = true
		}
	}

	permissions := make([]model.Permission, 0, len(permissionMap))
	for perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// getRoleAllPermissions 获取角色的所有权限（包括继承）
func (s *PermissionService) getRoleAllPermissions(ctx context.Context, roleID string) ([]model.Permission, error) {
	role, err := s.roleRepo.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// 如果有父角色，递归获取父角色权限
	if role.ParentID != "" {
		parentPermissions, err := s.getRoleAllPermissions(ctx, role.ParentID)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, parentPermissions...)
	}

	// 去重
	permMap := make(map[model.Permission]bool)
	for _, p := range permissions {
		permMap[p] = true
	}

	result := make([]model.Permission, 0, len(permMap))
	for p := range permMap {
		result = append(result, p)
	}

	return result, nil
}

// GrantResourcePermission 授予资源权限
func (s *PermissionService) GrantResourcePermission(ctx context.Context, actx *model.AccessContext, resourceID, targetUserID string, permission model.Permission) error {
	// 1. 检查操作者权限
	err := s.CheckResourceAccess(ctx, actx, resourceID, model.PermissionResourceWrite)
	if err != nil {
		return fmt.Errorf("insufficient permissions to grant: %w", err)
	}

	// 2. 授予权限
	err = s.resourceRepo.GrantResourcePermission(ctx, resourceID, targetUserID, permission)
	if err != nil {
		return err
	}

	// 3. 记录审计
	return s.AuditAccess(ctx, actx, "permission.grant", resourceID, "success",
		fmt.Sprintf("Granted %s to user %s", permission, targetUserID), 0)
}

// RevokeResourcePermission 撤销资源权限
func (s *PermissionService) RevokeResourcePermission(ctx context.Context, actx *model.AccessContext, resourceID, targetUserID string, permission model.Permission) error {
	// 1. 检查操作者权限
	err := s.CheckResourceAccess(ctx, actx, resourceID, model.PermissionResourceWrite)
	if err != nil {
		return fmt.Errorf("insufficient permissions to revoke: %w", err)
	}

	// 2. 撤销权限
	err = s.resourceRepo.RevokeResourcePermission(ctx, resourceID, targetUserID, permission)
	if err != nil {
		return err
	}

	// 3. 记录审计
	return s.AuditAccess(ctx, actx, "permission.revoke", resourceID, "success",
		fmt.Sprintf("Revoked %s from user %s", permission, targetUserID), 0)
}
