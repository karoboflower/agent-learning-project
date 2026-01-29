package auth

import (
	"fmt"
	"testing"
)

func TestNewRoleManager(t *testing.T) {
	rm := NewRoleManager()

	if rm == nil {
		t.Fatal("NewRoleManager returned nil")
	}

	// 检查默认角色
	defaultRoles := []string{"admin", "developer", "viewer", "guest"}
	for _, roleID := range defaultRoles {
		role, err := rm.GetRole(roleID)
		if err != nil {
			t.Errorf("Default role %s not found", roleID)
		}
		if role.ID != roleID {
			t.Errorf("Expected role ID %s, got %s", roleID, role.ID)
		}
	}
}

func TestRoleManager_CreateRole(t *testing.T) {
	rm := NewRoleManager()

	role := &Role{
		ID:          "test-role",
		Name:        "Test Role",
		Description: "Test role for testing",
		Permissions: []Permission{PermissionToolExecute, PermissionToolList},
	}

	err := rm.CreateRole(role)
	if err != nil {
		t.Fatalf("CreateRole failed: %v", err)
	}

	// 验证创建
	retrieved, err := rm.GetRole("test-role")
	if err != nil {
		t.Fatalf("GetRole failed: %v", err)
	}

	if retrieved.Name != "Test Role" {
		t.Errorf("Expected name 'Test Role', got '%s'", retrieved.Name)
	}

	if len(retrieved.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(retrieved.Permissions))
	}
}

func TestRoleManager_HasPermission(t *testing.T) {
	rm := NewRoleManager()

	// Admin应该有所有权限
	if !rm.HasPermission("admin", PermissionToolExecute) {
		t.Error("Admin should have tool:execute permission")
	}

	// Developer应该有工具执行权限
	if !rm.HasPermission("developer", PermissionToolExecute) {
		t.Error("Developer should have tool:execute permission")
	}

	// Viewer不应该有工具执行权限
	if rm.HasPermission("viewer", PermissionToolExecute) {
		t.Error("Viewer should not have tool:execute permission")
	}
}

func TestRoleManager_AddRemovePermission(t *testing.T) {
	rm := NewRoleManager()

	// 为viewer添加执行权限
	err := rm.AddPermission("viewer", PermissionToolExecute)
	if err != nil {
		t.Fatalf("AddPermission failed: %v", err)
	}

	// 验证权限已添加
	if !rm.HasPermission("viewer", PermissionToolExecute) {
		t.Error("Permission should be added")
	}

	// 移除权限
	err = rm.RemovePermission("viewer", PermissionToolExecute)
	if err != nil {
		t.Fatalf("RemovePermission failed: %v", err)
	}

	// 验证权限已移除
	if rm.HasPermission("viewer", PermissionToolExecute) {
		t.Error("Permission should be removed")
	}
}

func TestUserManager_CreateUser(t *testing.T) {
	um := NewUserManager()

	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"developer"},
	}

	err := um.CreateUser(user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// 验证创建
	retrieved, err := um.GetUser("user-001")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrieved.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrieved.Username)
	}
}

func TestUserManager_AssignRevokeRole(t *testing.T) {
	um := NewUserManager()

	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{},
	}

	um.CreateUser(user)

	// 分配角色
	err := um.AssignRole("user-001", "developer")
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	// 验证角色已分配
	roles, err := um.GetUserRoles("user-001")
	if err != nil {
		t.Fatalf("GetUserRoles failed: %v", err)
	}

	if len(roles) != 1 || roles[0] != "developer" {
		t.Error("Role should be assigned")
	}

	// 撤销角色
	err = um.RevokeRole("user-001", "developer")
	if err != nil {
		t.Fatalf("RevokeRole failed: %v", err)
	}

	// 验证角色已撤销
	roles, err = um.GetUserRoles("user-001")
	if err != nil {
		t.Fatalf("GetUserRoles failed: %v", err)
	}

	if len(roles) != 0 {
		t.Error("Role should be revoked")
	}
}

func TestPermissionChecker_CheckPermission(t *testing.T) {
	rm := NewRoleManager()
	um := NewUserManager()
	pc := NewPermissionChecker(rm, um)

	// 创建用户并分配角色
	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Roles:    []string{"developer"},
	}
	um.CreateUser(user)

	// 检查权限
	hasPermission, err := pc.CheckPermission("user-001", PermissionToolExecute)
	if err != nil {
		t.Fatalf("CheckPermission failed: %v", err)
	}

	if !hasPermission {
		t.Error("User should have tool:execute permission")
	}

	// 检查没有的权限
	hasPermission, err = pc.CheckPermission("user-001", PermissionUserManage)
	if err != nil {
		t.Fatalf("CheckPermission failed: %v", err)
	}

	if hasPermission {
		t.Error("User should not have user:manage permission")
	}
}

func TestPermissionChecker_CheckResourceAccess(t *testing.T) {
	rm := NewRoleManager()
	um := NewUserManager()
	pc := NewPermissionChecker(rm, um)

	// 创建用户
	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Roles:    []string{"developer"},
	}
	um.CreateUser(user)

	// 创建资源
	resource := &Resource{
		ID:    "resource-001",
		Type:  ResourceTypeFile,
		Path:  "/test/file.txt",
		Owner: "user-002", // 不是当前用户
	}

	// 检查读权限（developer有）
	err := pc.CheckResourceAccess("user-001", resource, AccessLevelRead)
	if err != nil {
		t.Error("User should have read access")
	}

	// 检查管理员权限（developer没有）
	err = pc.CheckResourceAccess("user-001", resource, AccessLevelAdmin)
	if err == nil {
		t.Error("User should not have admin access")
	}

	// 检查拥有者权限
	resource.Owner = "user-001"
	err = pc.CheckResourceAccess("user-001", resource, AccessLevelAdmin)
	if err != nil {
		t.Error("Owner should have full access")
	}
}

func TestAuditLogger_Log(t *testing.T) {
	al := NewAuditLogger(100)

	log := &AuditLog{
		UserID:   "user-001",
		Username: "testuser",
		Action:   AuditActionToolExecute,
		Resource: "tool-001",
		Result:   AuditResultSuccess,
		Details:  "Test execution",
	}

	err := al.Log(log)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	// 验证日志已记录
	logs := al.GetLogs()
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	if logs[0].UserID != "user-001" {
		t.Errorf("Expected user ID 'user-001', got '%s'", logs[0].UserID)
	}
}

func TestAuditLogger_GetLogsByUser(t *testing.T) {
	al := NewAuditLogger(100)

	// 记录多条日志
	users := []string{"user-001", "user-002", "user-001", "user-003"}
	for i, userID := range users {
		al.Log(&AuditLog{
			UserID:   userID,
			Username: userID,
			Action:   AuditActionToolExecute,
			Resource: fmt.Sprintf("tool-%d", i),
			Result:   AuditResultSuccess,
		})
	}

	// 查询user-001的日志
	logs := al.GetLogsByUser("user-001")
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs for user-001, got %d", len(logs))
	}
}

func TestAuditLogger_GetLogsByAction(t *testing.T) {
	al := NewAuditLogger(100)

	// 记录不同动作的日志
	actions := []AuditAction{
		AuditActionToolExecute,
		AuditActionResourceRead,
		AuditActionToolExecute,
		AuditActionUserCreate,
	}

	for i, action := range actions {
		al.Log(&AuditLog{
			UserID:   "user-001",
			Username: "user-001",
			Action:   action,
			Resource: fmt.Sprintf("resource-%d", i),
			Result:   AuditResultSuccess,
		})
	}

	// 查询工具执行日志
	logs := al.GetLogsByAction(AuditActionToolExecute)
	if len(logs) != 2 {
		t.Errorf("Expected 2 tool execute logs, got %d", len(logs))
	}
}

func TestAuditLogger_GetStatistics(t *testing.T) {
	al := NewAuditLogger(100)

	// 记录多条日志
	results := []AuditResult{
		AuditResultSuccess,
		AuditResultSuccess,
		AuditResultFailure,
		AuditResultDenied,
	}

	for i, result := range results {
		al.Log(&AuditLog{
			UserID:   "user-001",
			Username: "user-001",
			Action:   AuditActionToolExecute,
			Resource: fmt.Sprintf("tool-%d", i),
			Result:   result,
		})
	}

	// 获取统计
	stats := al.GetStatistics()

	if stats.TotalLogs != 4 {
		t.Errorf("Expected 4 total logs, got %d", stats.TotalLogs)
	}

	if stats.SuccessCount != 2 {
		t.Errorf("Expected 2 success logs, got %d", stats.SuccessCount)
	}

	if stats.FailureCount != 1 {
		t.Errorf("Expected 1 failure log, got %d", stats.FailureCount)
	}

	if stats.DeniedCount != 1 {
		t.Errorf("Expected 1 denied log, got %d", stats.DeniedCount)
	}
}

func TestResourceManager_RegisterResource(t *testing.T) {
	rm := NewResourceManager()

	resource := &Resource{
		ID:    "resource-001",
		Type:  ResourceTypeFile,
		Path:  "/test/file.txt",
		Owner: "user-001",
	}

	err := rm.RegisterResource(resource)
	if err != nil {
		t.Fatalf("RegisterResource failed: %v", err)
	}

	// 验证注册
	retrieved, err := rm.GetResource("resource-001")
	if err != nil {
		t.Fatalf("GetResource failed: %v", err)
	}

	if retrieved.Path != "/test/file.txt" {
		t.Errorf("Expected path '/test/file.txt', got '%s'", retrieved.Path)
	}
}

func TestResourceManager_ListResourcesByType(t *testing.T) {
	rm := NewResourceManager()

	// 注册不同类型的资源
	resources := []*Resource{
		{ID: "res-001", Type: ResourceTypeFile, Path: "/file1", Owner: "user-001"},
		{ID: "res-002", Type: ResourceTypeAPI, Path: "/api/v1", Owner: "user-001"},
		{ID: "res-003", Type: ResourceTypeFile, Path: "/file2", Owner: "user-002"},
	}

	for _, res := range resources {
		rm.RegisterResource(res)
	}

	// 查询文件类型资源
	fileResources := rm.ListResourcesByType(ResourceTypeFile)
	if len(fileResources) != 2 {
		t.Errorf("Expected 2 file resources, got %d", len(fileResources))
	}
}

func TestAuthorizationManager_Integration(t *testing.T) {
	am := NewAuthorizationManager()

	// 先创建管理员用户
	admin := &User{
		ID:       "admin",
		Username: "Admin",
		Email:    "admin@example.com",
		Roles:    []string{"admin"},
	}
	am.GetUserManager().CreateUser(admin)

	// 创建用户
	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Email:    "test@example.com",
	}

	err := am.CreateUserWithRole("admin", "Admin", user, "developer")
	if err != nil {
		t.Fatalf("CreateUserWithRole failed: %v", err)
	}

	// 验证用户已创建
	retrieved, err := am.GetUserManager().GetUser("user-001")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if retrieved.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrieved.Username)
	}

	// 验证角色已分配
	roles, err := am.GetUserManager().GetUserRoles("user-001")
	if err != nil {
		t.Fatalf("GetUserRoles failed: %v", err)
	}

	if len(roles) != 1 || roles[0] != "developer" {
		t.Error("Role should be assigned")
	}

	// 检查审计日志
	logs := am.GetAuditLogger().GetLogs()
	if len(logs) == 0 {
		t.Error("Expected audit logs to be recorded")
	}
}

func TestAuthorizationManager_AuthorizeToolExecution(t *testing.T) {
	am := NewAuthorizationManager()

	// 创建用户
	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Roles:    []string{"developer"},
	}
	am.GetUserManager().CreateUser(user)

	// 授权工具执行
	err := am.AuthorizeToolExecution("user-001", "testuser", "tool-001")
	if err != nil {
		t.Error("User should be authorized to execute tool")
	}

	// 检查审计日志
	logs := am.GetAuditLogger().GetLogsByUser("user-001")
	if len(logs) == 0 {
		t.Error("Expected audit log to be recorded")
	}

	if logs[0].Action != AuditActionToolExecute {
		t.Error("Expected tool execute action")
	}
}

func BenchmarkPermissionChecker_CheckPermission(b *testing.B) {
	rm := NewRoleManager()
	um := NewUserManager()
	pc := NewPermissionChecker(rm, um)

	user := &User{
		ID:       "user-001",
		Username: "testuser",
		Roles:    []string{"developer"},
	}
	um.CreateUser(user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.CheckPermission("user-001", PermissionToolExecute)
	}
}

func BenchmarkAuditLogger_Log(b *testing.B) {
	al := NewAuditLogger(10000)

	log := &AuditLog{
		UserID:   "user-001",
		Username: "testuser",
		Action:   AuditActionToolExecute,
		Resource: "tool-001",
		Result:   AuditResultSuccess,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		al.Log(log)
	}
}
