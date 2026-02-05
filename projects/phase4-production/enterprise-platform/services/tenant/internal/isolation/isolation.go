package isolation

import (
	"context"
	"fmt"
)

// IsolationStrategy 数据隔离策略
type IsolationStrategy string

const (
	// DatabaseIsolation 数据库级隔离（每个租户独立数据库）
	DatabaseIsolation IsolationStrategy = "database"

	// SchemaIsolation Schema级隔离（共享数据库，独立Schema）
	SchemaIsolation IsolationStrategy = "schema"

	// RowIsolation 行级隔离（共享表，通过tenant_id区分）
	RowIsolation IsolationStrategy = "row"
)

// IsolationManager 数据隔离管理器
type IsolationManager struct {
	strategy IsolationStrategy
}

// NewIsolationManager 创建隔离管理器
func NewIsolationManager(strategy IsolationStrategy) *IsolationManager {
	return &IsolationManager{
		strategy: strategy,
	}
}

// GetConnectionString 获取租户的数据库连接字符串
func (im *IsolationManager) GetConnectionString(ctx context.Context, tenantID, baseConnStr string) (string, error) {
	switch im.strategy {
	case DatabaseIsolation:
		// 数据库级隔离：每个租户独立数据库
		// 例如: postgresql://user:pass@host:5432/tenant_abc123
		return im.getDatabaseIsolationConnStr(tenantID, baseConnStr), nil

	case SchemaIsolation:
		// Schema级隔离：共享数据库，独立Schema
		// 需要在查询时使用 SET search_path TO tenant_schema;
		return baseConnStr, nil

	case RowIsolation:
		// 行级隔离：共享表，通过WHERE tenant_id = 'xxx'过滤
		return baseConnStr, nil

	default:
		return "", fmt.Errorf("unknown isolation strategy: %s", im.strategy)
	}
}

// GetSchemaName 获取租户的Schema名称
func (im *IsolationManager) GetSchemaName(tenantID string) string {
	if im.strategy == SchemaIsolation {
		return fmt.Sprintf("tenant_%s", tenantID)
	}
	return "public"
}

// AddTenantFilter 为SQL查询添加租户过滤条件
func (im *IsolationManager) AddTenantFilter(ctx context.Context, tenantID, query string) (string, error) {
	if im.strategy == RowIsolation {
		// 为查询添加 WHERE tenant_id = 'xxx'
		// 这里简化处理，实际应该使用SQL解析器
		return fmt.Sprintf("%s WHERE tenant_id = '%s'", query, tenantID), nil
	}
	return query, nil
}

// getDatabaseIsolationConnStr 获取数据库隔离的连接字符串
func (im *IsolationManager) getDatabaseIsolationConnStr(tenantID, baseConnStr string) string {
	// 简化实现：将数据库名替换为租户专属数据库
	// 实际实现需要解析连接字符串并替换数据库名
	// postgresql://user:pass@host:5432/base_db -> postgresql://user:pass@host:5432/tenant_abc123
	return baseConnStr // 这里需要实际的URL解析和替换逻辑
}

// CacheIsolation 缓存隔离
type CacheIsolation struct{}

// NewCacheIsolation 创建缓存隔离管理器
func NewCacheIsolation() *CacheIsolation {
	return &CacheIsolation{}
}

// GetCacheKey 获取租户隔离的缓存键
func (ci *CacheIsolation) GetCacheKey(tenantID, key string) string {
	return fmt.Sprintf("tenant:%s:%s", tenantID, key)
}

// GetCacheNamespace 获取租户的缓存命名空间
func (ci *CacheIsolation) GetCacheNamespace(tenantID string) string {
	return fmt.Sprintf("tenant:%s", tenantID)
}

// StorageIsolation 存储隔离
type StorageIsolation struct {
	basePath string
}

// NewStorageIsolation 创建存储隔离管理器
func NewStorageIsolation(basePath string) *StorageIsolation {
	return &StorageIsolation{
		basePath: basePath,
	}
}

// GetStoragePath 获取租户的存储路径
func (si *StorageIsolation) GetStoragePath(tenantID string) string {
	// 每个租户独立的存储目录
	return fmt.Sprintf("%s/%s", si.basePath, tenantID)
}

// GetFilePath 获取租户文件的完整路径
func (si *StorageIsolation) GetFilePath(tenantID, filename string) string {
	return fmt.Sprintf("%s/%s/%s", si.basePath, tenantID, filename)
}

// NetworkIsolation 网络隔离
type NetworkIsolation struct{}

// NewNetworkIsolation 创建网络隔离管理器
func NewNetworkIsolation() *NetworkIsolation {
	return &NetworkIsolation{}
}

// GetSubdomain 获取租户的子域名
func (ni *NetworkIsolation) GetSubdomain(tenantID, baseDomain string) string {
	// 为每个租户分配子域名
	// 例如: tenant-abc123.agent-platform.com
	return fmt.Sprintf("%s.%s", tenantID, baseDomain)
}

// GetAPIEndpoint 获取租户的API端点
func (ni *NetworkIsolation) GetAPIEndpoint(tenantID, baseURL string) string {
	// 为每个租户提供独立的API端点
	return fmt.Sprintf("%s/tenants/%s", baseURL, tenantID)
}

// TenantContext 租户上下文
type TenantContext struct {
	TenantID string
	UserID   string
	Roles    []string
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// TenantContextKey 租户上下文键
	TenantContextKey ContextKey = "tenant_context"
)

// WithTenantContext 将租户上下文注入到Context
func WithTenantContext(ctx context.Context, tc *TenantContext) context.Context {
	return context.WithValue(ctx, TenantContextKey, tc)
}

// GetTenantContext 从Context获取租户上下文
func GetTenantContext(ctx context.Context) (*TenantContext, error) {
	tc, ok := ctx.Value(TenantContextKey).(*TenantContext)
	if !ok {
		return nil, fmt.Errorf("tenant context not found")
	}
	return tc, nil
}

// GetTenantID 从Context获取租户ID
func GetTenantID(ctx context.Context) (string, error) {
	tc, err := GetTenantContext(ctx)
	if err != nil {
		return "", err
	}
	return tc.TenantID, nil
}

// ValidateTenantAccess 验证租户访问权限
func ValidateTenantAccess(ctx context.Context, requiredTenantID string) error {
	tc, err := GetTenantContext(ctx)
	if err != nil {
		return fmt.Errorf("unauthorized: %w", err)
	}

	if tc.TenantID != requiredTenantID {
		return fmt.Errorf("unauthorized: tenant mismatch")
	}

	return nil
}

// ResourceIsolation 资源隔离配置
type ResourceIsolation struct {
	// CPU配额（毫核）
	CPUQuota int64

	// 内存配额（字节）
	MemoryQuota int64

	// 网络带宽配额（bps）
	NetworkQuota int64
}

// GetResourceIsolation 获取租户的资源隔离配置
func GetResourceIsolation(plan string) *ResourceIsolation {
	switch plan {
	case "free":
		return &ResourceIsolation{
			CPUQuota:     1000,           // 1 core
			MemoryQuota:  1073741824,     // 1GB
			NetworkQuota: 10485760,       // 10Mbps
		}
	case "starter":
		return &ResourceIsolation{
			CPUQuota:     2000,           // 2 cores
			MemoryQuota:  4294967296,     // 4GB
			NetworkQuota: 52428800,       // 50Mbps
		}
	case "pro":
		return &ResourceIsolation{
			CPUQuota:     4000,           // 4 cores
			MemoryQuota:  17179869184,    // 16GB
			NetworkQuota: 104857600,      // 100Mbps
		}
	case "enterprise":
		return &ResourceIsolation{
			CPUQuota:     -1,             // 无限制
			MemoryQuota:  -1,             // 无限制
			NetworkQuota: -1,             // 无限制
		}
	default:
		return GetResourceIsolation("free")
	}
}
