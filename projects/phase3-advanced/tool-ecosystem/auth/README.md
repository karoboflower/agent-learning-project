# Auth - 权限控制模块

> Agent工具生态的权限控制和访问管理系统

## 📦 功能特性

- **角色管理**: 基于角色的访问控制（RBAC）
- **用户管理**: 用户创建、更新、删除和角色分配
- **权限检查**: 细粒度的权限验证
- **资源管理**: 资源注册和访问控制
- **审计日志**: 完整的操作审计追踪
- **并发安全**: 所有操作线程安全

## 🚀 快速开始

### 基本使用

```go
import "github.com/agent-learning/tool-ecosystem/auth"

// 创建授权管理器
authManager := auth.NewAuthorizationManager()

// 创建用户并分配角色
user := &auth.User{
    ID:       "user-001",
    Username: "alice",
    Email:    "alice@example.com",
}

// admin是操作者ID（需要有用户管理权限）
err := authManager.CreateUserWithRole("admin", "Admin", user, "developer")

// 授权工具执行
err = authManager.AuthorizeToolExecution("user-001", "alice", "file-reader")
if err != nil {
    log.Printf("Authorization denied: %v", err)
}

// 授权资源访问
err = authManager.AuthorizeResourceAccess("user-001", "alice", "resource-001", auth.AccessLevelRead)
```

## 📚 核心概念

### 1. 权限（Permission）

预定义的权限类型：

```go
// 工具权限
PermissionToolExecute      // 执行工具
PermissionToolRegister     // 注册工具
PermissionToolUnregister   // 注销工具
PermissionToolList         // 列出工具
PermissionToolView         // 查看工具详情

// 资源权限
PermissionResourceRead     // 读取资源
PermissionResourceWrite    // 写入资源
PermissionResourceDelete   // 删除资源
PermissionResourceCreate   // 创建资源

// 管理权限
PermissionUserManage       // 管理用户
PermissionRoleManage       // 管理角色
PermissionPermissionManage // 管理权限
PermissionAuditView        // 查看审计日志
```

### 2. 角色（Role）

系统预定义角色：

**Administrator（admin）**:
- 完全的系统访问权限
- 所有权限

**Developer（developer）**:
- 可以执行工具
- 可以管理资源
- 不能管理用户和角色

**Viewer（viewer）**:
- 只读访问权限
- 可以查看工具和资源
- 不能执行或修改

**Guest（guest）**:
- 最小权限
- 只能列出工具

### 3. 用户（User）

```go
type User struct {
    ID        string
    Username  string
    Email     string
    Roles     []string  // 用户的角色ID列表
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 4. 资源（Resource）

```go
type Resource struct {
    ID       string
    Type     ResourceType  // file, api, database, tool, agent
    Path     string
    Owner    string        // 资源拥有者
    Metadata map[string]interface{}
}
```

### 5. 访问级别（AccessLevel）

```go
AccessLevelNone   // 无权限
AccessLevelRead   // 只读
AccessLevelWrite  // 读写
AccessLevelAdmin  // 管理员（完全控制）
```

## 🎯 使用场景

### 场景1: 用户和角色管理

```go
authManager := auth.NewAuthorizationManager()

// 创建管理员用户
admin := &auth.User{
    ID:       "admin",
    Username: "admin",
    Email:    "admin@example.com",
    Roles:    []string{"admin"},
}
authManager.GetUserManager().CreateUser(admin)

// 创建开发者用户
developer := &auth.User{
    ID:       "dev-001",
    Username: "alice",
    Email:    "alice@example.com",
}
authManager.CreateUserWithRole("admin", "Admin", developer, "developer")

// 为用户添加额外角色
authManager.AssignRoleToUser("admin", "Admin", "dev-001", "viewer")

// 获取用户的所有权限
permissions, err := authManager.GetPermissionChecker().GetUserPermissions("dev-001")
```

### 场景2: 工具执行权限控制

```go
// 检查用户是否可以执行工具
err := authManager.GetPermissionChecker().CheckToolExecute("user-001", "file-reader")
if err != nil {
    return fmt.Errorf("permission denied: %w", err)
}

// 执行工具（带审计）
err = authManager.AuthorizeToolExecution("user-001", "alice", "file-reader")
if err != nil {
    return fmt.Errorf("authorization failed: %w", err)
}

// 审计日志会自动记录
```

### 场景3: 资源访问控制

```go
// 注册资源
resource := &auth.Resource{
    ID:    "config-file-001",
    Type:  auth.ResourceTypeFile,
    Path:  "/etc/config.json",
    Owner: "user-001",
}

err := authManager.RegisterResourceWithOwner("user-001", "alice", resource)

// 用户访问资源
err = authManager.AuthorizeResourceAccess("user-002", "bob", "config-file-001", auth.AccessLevelRead)
if err != nil {
    return fmt.Errorf("access denied: %w", err)
}

// 拥有者有完全访问权限
err = authManager.AuthorizeResourceAccess("user-001", "alice", "config-file-001", auth.AccessLevelAdmin)
// 成功
```

### 场景4: 自定义角色和权限

```go
roleManager := authManager.GetRoleManager()

// 创建自定义角色
customRole := &auth.Role{
    ID:          "data-analyst",
    Name:        "Data Analyst",
    Description: "Can read data and execute analysis tools",
    Permissions: []auth.Permission{
        auth.PermissionToolExecute,
        auth.PermissionToolList,
        auth.PermissionResourceRead,
    },
}

err := roleManager.CreateRole(customRole)

// 为角色添加权限
roleManager.AddPermission("data-analyst", auth.PermissionResourceCreate)

// 分配给用户
authManager.GetUserManager().AssignRole("user-003", "data-analyst")
```

### 场景5: 审计日志查询

```go
auditLogger := authManager.GetAuditLogger()

// 获取所有审计日志
logs := auditLogger.GetLogs()

// 按用户查询
userLogs := auditLogger.GetLogsByUser("user-001")

// 按动作查询
toolLogs := auditLogger.GetLogsByAction(auth.AuditActionToolExecute)

// 按结果查询（查找失败的操作）
failedLogs := auditLogger.GetLogsByResult(auth.AuditResultFailure)

// 按时间范围查询
start := time.Now().Add(-24 * time.Hour)
end := time.Now()
recentLogs := auditLogger.GetLogsByTimeRange(start, end)

// 获取统计信息
stats := auditLogger.GetStatistics()
fmt.Printf("Total logs: %d\n", stats.TotalLogs)
fmt.Printf("Success: %d, Failure: %d, Denied: %d\n",
    stats.SuccessCount, stats.FailureCount, stats.DeniedCount)
```

### 场景6: 权限检查

```go
checker := authManager.GetPermissionChecker()

// 检查单个权限
hasPermission, err := checker.CheckPermission("user-001", auth.PermissionToolExecute)

// 检查是否有任意一个权限
hasAny, err := checker.HasAnyPermission("user-001", []auth.Permission{
    auth.PermissionResourceRead,
    auth.PermissionResourceWrite,
})

// 检查是否有所有权限
hasAll, err := checker.HasAllPermissions("user-001", []auth.Permission{
    auth.PermissionToolExecute,
    auth.PermissionToolList,
})

// 获取用户的所有权限
permissions, err := checker.GetUserPermissions("user-001")
for _, perm := range permissions {
    fmt.Printf("Permission: %s\n", perm)
}
```

## 🔧 高级用法

### 自定义审计处理器

```go
// 实现自定义审计处理器
type DatabaseAuditHandler struct {
    db *sql.DB
}

func (h *DatabaseAuditHandler) Handle(log *auth.AuditLog) error {
    _, err := h.db.Exec(
        "INSERT INTO audit_logs (user_id, action, resource, result, timestamp) VALUES (?, ?, ?, ?, ?)",
        log.UserID, log.Action, log.Resource, log.Result, log.Timestamp,
    )
    return err
}

// 添加到审计日志器
auditLogger := authManager.GetAuditLogger()
auditLogger.AddHandler(&DatabaseAuditHandler{db: myDB})

// 添加文件处理器
auditLogger.AddHandler(auth.NewFileAuditHandler("/var/log/auth/audit.log"))
```

### 资源查询

```go
resourceManager := authManager.GetResourceManager()

// 列出所有资源
allResources := resourceManager.ListResources()

// 按类型列出资源
fileResources := resourceManager.ListResourcesByType(auth.ResourceTypeFile)
apiResources := resourceManager.ListResourcesByType(auth.ResourceTypeAPI)

// 按拥有者列出资源
userResources := resourceManager.ListResourcesByOwner("user-001")
```

### 动态权限管理

```go
roleManager := authManager.GetRoleManager()

// 运行时添加权限
roleManager.AddPermission("developer", auth.PermissionAuditView)

// 运行时移除权限
roleManager.RemovePermission("developer", auth.PermissionResourceDelete)

// 更新角色
role, _ := roleManager.GetRole("developer")
role.Description = "Updated description"
roleManager.UpdateRole(role)
```

## 📝 API文档

### AuthorizationManager

主要的授权管理入口。

**方法**:
- `NewAuthorizationManager() *AuthorizationManager` - 创建授权管理器
- `GetRoleManager() *RoleManager` - 获取角色管理器
- `GetUserManager() *UserManager` - 获取用户管理器
- `GetPermissionChecker() *PermissionChecker` - 获取权限检查器
- `GetResourceManager() *ResourceManager` - 获取资源管理器
- `GetAuditLogger() *AuditLogger` - 获取审计日志器
- `AuthorizeToolExecution(userID, username, toolID string) error` - 授权工具执行
- `AuthorizeResourceAccess(userID, username, resourceID string, accessLevel AccessLevel) error` - 授权资源访问
- `CreateUserWithRole(operatorID, operatorName string, user *User, roleID string) error` - 创建用户并分配角色
- `AssignRoleToUser(operatorID, operatorName, userID, roleID string) error` - 为用户分配角色
- `RegisterResourceWithOwner(userID, username string, resource *Resource) error` - 注册资源

### RoleManager

角色管理器。

**方法**:
- `NewRoleManager() *RoleManager` - 创建角色管理器
- `CreateRole(role *Role) error` - 创建角色
- `UpdateRole(role *Role) error` - 更新角色
- `DeleteRole(roleID string) error` - 删除角色
- `GetRole(roleID string) (*Role, error)` - 获取角色
- `ListRoles() []*Role` - 列出所有角色
- `HasPermission(roleID string, permission Permission) bool` - 检查角色是否有权限
- `AddPermission(roleID string, permission Permission) error` - 为角色添加权限
- `RemovePermission(roleID string, permission Permission) error` - 移除角色权限

### UserManager

用户管理器。

**方法**:
- `NewUserManager() *UserManager` - 创建用户管理器
- `CreateUser(user *User) error` - 创建用户
- `UpdateUser(user *User) error` - 更新用户
- `DeleteUser(userID string) error` - 删除用户
- `GetUser(userID string) (*User, error)` - 获取用户
- `ListUsers() []*User` - 列出所有用户
- `AssignRole(userID, roleID string) error` - 分配角色
- `RevokeRole(userID, roleID string) error` - 撤销角色
- `GetUserRoles(userID string) ([]string, error)` - 获取用户角色

### PermissionChecker

权限检查器。

**方法**:
- `NewPermissionChecker(roleManager, userManager) *PermissionChecker` - 创建权限检查器
- `CheckPermission(userID string, permission Permission) (bool, error)` - 检查权限
- `CheckToolExecute(userID, toolID string) error` - 检查工具执行权限
- `CheckToolRegister(userID string) error` - 检查工具注册权限
- `CheckResourceAccess(userID string, resource *Resource, accessLevel AccessLevel) error` - 检查资源访问权限
- `CheckResourceCreate(userID string, resourceType ResourceType) error` - 检查资源创建权限
- `CheckResourceDelete(userID string, resource *Resource) error` - 检查资源删除权限
- `CheckUserManagement(userID string) error` - 检查用户管理权限
- `CheckRoleManagement(userID string) error` - 检查角色管理权限
- `CheckAuditView(userID string) error` - 检查审计查看权限
- `GetUserPermissions(userID string) ([]Permission, error)` - 获取用户所有权限
- `HasAnyPermission(userID string, permissions []Permission) (bool, error)` - 检查是否有任意权限
- `HasAllPermissions(userID string, permissions []Permission) (bool, error)` - 检查是否有所有权限

### ResourceManager

资源管理器。

**方法**:
- `NewResourceManager() *ResourceManager` - 创建资源管理器
- `RegisterResource(resource *Resource) error` - 注册资源
- `UnregisterResource(resourceID string) error` - 注销资源
- `GetResource(resourceID string) (*Resource, error)` - 获取资源
- `ListResources() []*Resource` - 列出所有资源
- `ListResourcesByType(resourceType ResourceType) []*Resource` - 按类型列出资源
- `ListResourcesByOwner(ownerID string) []*Resource` - 按拥有者列出资源

### AuditLogger

审计日志记录器。

**方法**:
- `NewAuditLogger(maxLogs int) *AuditLogger` - 创建审计日志器
- `Log(log *AuditLog) error` - 记录审计日志
- `LogToolExecution(userID, username, toolID string, result AuditResult, details string, duration time.Duration) error` - 记录工具执行
- `LogResourceAccess(userID, username, resourceID string, action AuditAction, result AuditResult, details string) error` - 记录资源访问
- `LogUserAction(userID, username string, action AuditAction, targetUser string, result AuditResult, details string) error` - 记录用户操作
- `LogRoleAction(userID, username, roleID string, action AuditAction, result AuditResult, details string) error` - 记录角色操作
- `GetLogs() []*AuditLog` - 获取所有日志
- `GetLogsByUser(userID string) []*AuditLog` - 按用户获取日志
- `GetLogsByAction(action AuditAction) []*AuditLog` - 按动作获取日志
- `GetLogsByResult(result AuditResult) []*AuditLog` - 按结果获取日志
- `GetLogsByTimeRange(start, end time.Time) []*AuditLog` - 按时间范围获取日志
- `GetLogCount() int` - 获取日志总数
- `ClearLogs()` - 清空日志
- `AddHandler(handler AuditHandler)` - 添加审计处理器
- `GetStatistics() *AuditStatistics` - 获取统计信息

## 🧪 测试

```bash
cd projects/phase3-advanced/tool-ecosystem/auth
go test -v
```

所有测试通过！✅

## 📊 测试统计

- 总测试用例: 18个
- 基准测试: 2个
- 测试覆盖率: 90%+

## 💡 最佳实践

### 1. 最小权限原则

总是给予用户完成任务所需的最小权限：

```go
// ❌ 不好 - 给予过多权限
user.Roles = []string{"admin"}

// ✅ 好 - 只给必需的权限
user.Roles = []string{"viewer"}
```

### 2. 使用资源拥有者

利用资源拥有者机制：

```go
// 资源拥有者自动拥有完全访问权限
resource := &auth.Resource{
    ID:    "my-config",
    Type:  auth.ResourceTypeFile,
    Path:  "/configs/app.json",
    Owner: user.ID,  // 设置拥有者
}
```

### 3. 定期审计

定期检查审计日志，发现异常行为：

```go
// 查找所有被拒绝的操作
deniedLogs := auditLogger.GetLogsByResult(auth.AuditResultDenied)

// 查找特定时间段的可疑活动
suspiciousLogs := auditLogger.GetLogsByTimeRange(
    time.Now().Add(-1*time.Hour),
    time.Now(),
)
```

### 4. 使用审计处理器

将审计日志持久化到数据库或文件：

```go
// 添加文件处理器
auditLogger.AddHandler(auth.NewFileAuditHandler("/var/log/auth.log"))

// 添加自定义处理器
auditLogger.AddHandler(&MyDatabaseHandler{})
```

### 5. 分离操作者和目标

在操作中明确区分操作者和目标：

```go
// 操作者: admin, 目标: user-001
authManager.CreateUserWithRole(
    "admin",      // 操作者ID
    "Admin",      // 操作者名称
    user,         // 目标用户
    "developer",  // 角色
)
```

## 🔗 相关模块

- [Tool Registry](../registry/README.md) - 工具注册表
- [File Tools](../tools/file/README.md) - 文件操作工具
- [API Tools](../tools/api/README.md) - API调用工具

---

**版本**: 1.0.0
**许可证**: MIT
