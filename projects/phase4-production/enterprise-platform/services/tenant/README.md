# Tenant Service - 租户服务

> 企业级Agent平台的多租户管理服务

## 📦 功能特性

- ✅ **租户管理** - 租户创建、更新、删除
- ✅ **配额管理** - 灵活的资源配额控制
- ✅ **使用监控** - 实时使用量追踪
- ✅ **功能开关** - 租户级别的功能控制
- ✅ **数据隔离** - 多种隔离策略（数据库/Schema/行级）
- ✅ **配额告警** - 超限自动告警

## 🏗️ 架构设计

### 租户计划

| 计划 | 用户数 | Agent数 | Token/月 | 存储 | 并发任务 | API调用/分钟 |
|------|--------|---------|----------|------|----------|--------------|
| **Free** | 5 | 3 | 10万 | 1GB | 5 | 60 |
| **Starter** | 20 | 10 | 100万 | 10GB | 20 | 300 |
| **Pro** | 100 | 50 | 1000万 | 100GB | 100 | 1000 |
| **Enterprise** | 无限 | 无限 | 无限 | 无限 | 无限 | 无限 |

### 数据隔离策略

#### 1. 数据库级隔离（Database Isolation）
```
tenant_abc123 (Database)
tenant_def456 (Database)
tenant_ghi789 (Database)
```

**优点**：
- 完全物理隔离
- 最高安全性
- 独立备份恢复

**缺点**：
- 运维成本高
- 资源利用率低

#### 2. Schema级隔离（Schema Isolation）
```
shared_db (Database)
├── tenant_abc123 (Schema)
├── tenant_def456 (Schema)
└── tenant_ghi789 (Schema)
```

**优���**：
- 逻辑隔离清晰
- 资源利用率高
- 易于管理

**缺点**：
- 需要在查询时切换Schema
- 备份粒度是数据库级别

#### 3. 行级隔离（Row Isolation）
```
shared_db.shared_table
├── row (tenant_id: abc123)
├── row (tenant_id: def456)
└── row (tenant_id: ghi789)
```

**优点**：
- 最高资源利用率
- 简单易实现
- 跨租户查询方便

**缺点**：
- 安全性较低
- 需要严格的WHERE子句

**本项目采用**：行级隔离（Row Isolation）+ 租户上下文验证

## 📊 数据模型

### 租户表（tenants）
```sql
id              VARCHAR(36)     PRIMARY KEY
name            VARCHAR(255)    租户��称
company         VARCHAR(255)    公司名称
email           VARCHAR(255)    联系邮箱
plan            VARCHAR(50)     订阅计划
status          VARCHAR(50)     状态（active/suspended/cancelled）
created_at      TIMESTAMP       创建时间
updated_at      TIMESTAMP       更新时间
```

### 配额表（tenant_quotas）
```sql
id                          VARCHAR(36)     PRIMARY KEY
tenant_id                   VARCHAR(36)     租户ID
max_users                   INTEGER         最大用户数
max_agents                  INTEGER         最大Agent数
max_tokens_per_month        BIGINT          每月最大Token数
max_storage_bytes           BIGINT          最大存储空间
max_concurrent_tasks        INTEGER         最大并发任务数
max_api_calls_per_minute    INTEGER         每分钟最大API调用数
created_at                  TIMESTAMP       创建时间
updated_at                  TIMESTAMP       更新时间
```

### 使用情况表（tenant_usage）
```sql
id                      VARCHAR(36)     PRIMARY KEY
tenant_id               VARCHAR(36)     租户ID
current_users           INTEGER         当前用户数
current_agents          INTEGER         当前Agent数
tokens_used_this_month  BIGINT          本月已用Token
storage_used_bytes      BIGINT          已用存储空间
active_tasks            INTEGER         活跃任务数
api_calls_this_minute   INTEGER         本分钟API调用数
last_updated            TIMESTAMP       最后更新时间
```

## 🚀 快速开始

### 启动服务

```bash
cd services/tenant

# 安装依赖
go mod download

# 运行数据库迁移
make migrate-up

# 启动服务
go run cmd/main.go
```

### API示例

#### 创建租户

```bash
curl -X POST http://localhost:8085/api/v1/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "company": "Acme Corporation",
    "email": "admin@acme.com",
    "plan": "pro"
  }'
```

响应：
```json
{
  "tenant_id": "tenant-abc123",
  "plan": "pro",
  "quota": {
    "max_users": 100,
    "max_agents": 50,
    "max_tokens_per_month": 10000000,
    "max_storage_bytes": 107374182400,
    "max_concurrent_tasks": 100,
    "max_api_calls_per_minute": 1000
  },
  "created_at": "2026-01-30T10:00:00Z"
}
```

#### 获取租户信息

```bash
curl http://localhost:8085/api/v1/tenants/tenant-abc123
```

响应：
```json
{
  "tenant": {
    "id": "tenant-abc123",
    "name": "Acme Corp",
    "company": "Acme Corporation",
    "email": "admin@acme.com",
    "plan": "pro",
    "status": "active"
  },
  "quota": {
    "max_users": 100,
    "max_agents": 50,
    "max_tokens_per_month": 10000000
  },
  "usage": {
    "current_users": 25,
    "current_agents": 15,
    "tokens_used_this_month": 2500000
  },
  "features": {
    "webhooks": true,
    "custom_models": true,
    "advanced_analytics": true,
    "sso": false
  }
}
```

#### 检查配额

```bash
curl -X POST http://localhost:8085/api/v1/tenants/tenant-abc123/check-quota \
  -H "Content-Type: application/json" \
  -d '{
    "quota_type": "tokens",
    "requested_amount": 50000
  }'
```

响应：
```json
{
  "allowed": true,
  "remaining": 7450000,
  "reason": ""
}
```

#### 获取使用情况

```bash
curl http://localhost:8085/api/v1/tenants/tenant-abc123/usage
```

响应：
```json
{
  "usage": {
    "current_users": 25,
    "current_agents": 15,
    "tokens_used_this_month": 2500000,
    "storage_used_bytes": 21474836480,
    "active_tasks": 8,
    "api_calls_this_minute": 45
  },
  "quota": {
    "max_users": 100,
    "max_agents": 50,
    "max_tokens_per_month": 10000000,
    "max_storage_bytes": 107374182400,
    "max_concurrent_tasks": 100,
    "max_api_calls_per_minute": 1000
  },
  "usage_percentages": {
    "users": 25.0,
    "agents": 30.0,
    "tokens": 25.0,
    "storage": 20.0,
    "tasks": 8.0,
    "api_calls": 4.5
  },
  "avg_percentage": 18.75
}
```

## 🛠️ 配额管理

### 消费配额

```go
import "github.com/agent-learning/enterprise-platform/services/tenant/internal/quota"

// 消费Token配额
err := quotaManager.ConsumeQuota(ctx, tenantID, "tokens", 1000)
if err != nil {
    // 配额不足
    return err
}
```

### 释放配额

```go
// 任务完成后释放并发任务配额
err := quotaManager.ReleaseQuota(ctx, tenantID, "tasks", 1)
```

### 检查配额

```go
allowed, remaining, err := quotaManager.CheckQuota(ctx, tenantID, "agents", 1)
if !allowed {
    return fmt.Errorf("cannot create agent: %w", err)
}
```

### 配额告警

```go
alerts, err := quotaManager.CheckAlerts(ctx, tenantID)
for _, alert := range alerts {
    if alert.Percentage >= 90.0 {
        // 发送紧急告警
    } else if alert.Percentage >= 80.0 {
        // 发送警告
    }
}
```

## 🔒 数据隔离

### 租户上下文

```go
import "github.com/agent-learning/enterprise-platform/services/tenant/internal/isolation"

// 在请求中注入租户上下文
ctx = isolation.WithTenantContext(ctx, &isolation.TenantContext{
    TenantID: "tenant-abc123",
    UserID:   "user-001",
    Roles:    []string{"admin"},
})

// 从上下文获取租户ID
tenantID, err := isolation.GetTenantID(ctx)

// 验证租户访问权限
err := isolation.ValidateTenantAccess(ctx, "tenant-abc123")
```

### 缓存隔离

```go
cacheIsolation := isolation.NewCacheIsolation()

// 获取租户隔离的缓存键
key := cacheIsolation.GetCacheKey(tenantID, "user:001")
// 结果: "tenant:tenant-abc123:user:001"
```

### 存储隔离

```go
storageIsolation := isolation.NewStorageIsolation("/var/data/tenants")

// 获取租户存储路径
path := storageIsolation.GetStoragePath(tenantID)
// 结果: "/var/data/tenants/tenant-abc123"

// 获取文件路径
filePath := storageIsolation.GetFilePath(tenantID, "report.pdf")
// 结果: "/var/data/tenants/tenant-abc123/report.pdf"
```

## 📊 监控指标

### Prometheus指标

```
# 租户总数
tenant_total{status="active"}

# 配额使用率
tenant_quota_usage_percentage{tenant_id="xxx", quota_type="tokens"}

# 配额超限次数
tenant_quota_exceeded_total{tenant_id="xxx", quota_type="tokens"}

# API调用频率
tenant_api_calls_per_minute{tenant_id="xxx"}
```

### Grafana面板

- 租户概览
- 配额使用趋势
- 超限告警统计
- 活跃租户排行

## 🧪 测试

```bash
# 单元测试
go test ./...

# 集成测试
go test -tags=integration ./...

# 性能测试
go test -bench=. ./internal/quota/...
```

## 📝 最佳实践

### 1. 总是验证租户上下文

```go
func (s *Service) DoSomething(ctx context.Context, resourceID string) error {
    // 验证租户访问权限
    tenantID, err := isolation.GetTenantID(ctx)
    if err != nil {
        return err
    }

    // 业务逻辑
    ...
}
```

### 2. 操作前检查配额

```go
// 创建Agent前检查配额
allowed, _, err := quotaManager.CheckQuota(ctx, tenantID, "agents", 1)
if !allowed {
    return fmt.Errorf("配额不足: %w", err)
}

// 执行操作
agent := createAgent()

// 消费配额
quotaManager.ConsumeQuota(ctx, tenantID, "agents", 1)
```

### 3. 操作完成后释放配额

```go
// 任务开始
quotaManager.ConsumeQuota(ctx, tenantID, "tasks", 1)

defer func() {
    // 任务结束，释放配额
    quotaManager.ReleaseQuota(ctx, tenantID, "tasks", 1)
}()
```

### 4. 定期检查告警

```go
// 定时任务：每小时检查配额告警
ticker := time.NewTicker(1 * time.Hour)
for range ticker.C {
    alerts, _ := quotaManager.CheckAlerts(ctx, tenantID)
    for _, alert := range alerts {
        sendAlert(alert)
    }
}
```

## 🔗 相关服务

- [User Service](../user/README.md) - 用户管理服务
- [Agent Service](../agent/README.md) - Agent服务
- [Cost Service](../cost/README.md) - 成本控制服务

---

**版本**: v1.0.0
**状态**: ✅ 实现完成
