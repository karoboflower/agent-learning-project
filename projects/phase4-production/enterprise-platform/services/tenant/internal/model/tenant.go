package model

import (
	"time"
)

// TenantPlan 租户计划类型
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"       // 免费版
	TenantPlanStarter    TenantPlan = "starter"    // 入门版
	TenantPlanPro        TenantPlan = "pro"        // 专业版
	TenantPlanEnterprise TenantPlan = "enterprise" // 企业版
)

// TenantStatus 租户状态
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"    // 激活
	TenantStatusSuspended TenantStatus = "suspended" // 暂停
	TenantStatusCancelled TenantStatus = "cancelled" // 已取消
)

// Tenant 租户模型
type Tenant struct {
	ID        string       `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	Company   string       `json:"company" db:"company"`
	Email     string       `json:"email" db:"email"`
	Plan      TenantPlan   `json:"plan" db:"plan"`
	Status    TenantStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// TenantQuota 租户配额
type TenantQuota struct {
	ID                    string    `json:"id" db:"id"`
	TenantID              string    `json:"tenant_id" db:"tenant_id"`
	MaxUsers              int       `json:"max_users" db:"max_users"`
	MaxAgents             int       `json:"max_agents" db:"max_agents"`
	MaxTokensPerMonth     int64     `json:"max_tokens_per_month" db:"max_tokens_per_month"`
	MaxStorageBytes       int64     `json:"max_storage_bytes" db:"max_storage_bytes"`
	MaxConcurrentTasks    int       `json:"max_concurrent_tasks" db:"max_concurrent_tasks"`
	MaxAPICallsPerMinute  int       `json:"max_api_calls_per_minute" db:"max_api_calls_per_minute"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// TenantUsage 租户使用情况
type TenantUsage struct {
	ID                  string    `json:"id" db:"id"`
	TenantID            string    `json:"tenant_id" db:"tenant_id"`
	CurrentUsers        int       `json:"current_users" db:"current_users"`
	CurrentAgents       int       `json:"current_agents" db:"current_agents"`
	TokensUsedThisMonth int64     `json:"tokens_used_this_month" db:"tokens_used_this_month"`
	StorageUsedBytes    int64     `json:"storage_used_bytes" db:"storage_used_bytes"`
	ActiveTasks         int       `json:"active_tasks" db:"active_tasks"`
	APICallsThisMinute  int       `json:"api_calls_this_minute" db:"api_calls_this_minute"`
	LastUpdated         time.Time `json:"last_updated" db:"last_updated"`
}

// TenantFeature 租户功能开关
type TenantFeature struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Feature   string    `json:"feature" db:"feature"`
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TenantConfig 租户配置
type TenantConfig struct {
	ID        string    `json:"id" db:"id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Key       string    `json:"key" db:"key"`
	Value     string    `json:"value" db:"value"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetDefaultQuota 获取默认配额
func GetDefaultQuota(plan TenantPlan) TenantQuota {
	switch plan {
	case TenantPlanFree:
		return TenantQuota{
			MaxUsers:              5,
			MaxAgents:             3,
			MaxTokensPerMonth:     100000,    // 10万tokens/月
			MaxStorageBytes:       1073741824, // 1GB
			MaxConcurrentTasks:    5,
			MaxAPICallsPerMinute:  60,
		}
	case TenantPlanStarter:
		return TenantQuota{
			MaxUsers:              20,
			MaxAgents:             10,
			MaxTokensPerMonth:     1000000,     // 100万tokens/月
			MaxStorageBytes:       10737418240,  // 10GB
			MaxConcurrentTasks:    20,
			MaxAPICallsPerMinute:  300,
		}
	case TenantPlanPro:
		return TenantQuota{
			MaxUsers:              100,
			MaxAgents:             50,
			MaxTokensPerMonth:     10000000,    // 1000万tokens/月
			MaxStorageBytes:       107374182400, // 100GB
			MaxConcurrentTasks:    100,
			MaxAPICallsPerMinute:  1000,
		}
	case TenantPlanEnterprise:
		return TenantQuota{
			MaxUsers:              -1,  // 无限制
			MaxAgents:             -1,  // 无限制
			MaxTokensPerMonth:     -1,  // 无限制
			MaxStorageBytes:       -1,  // 无限制
			MaxConcurrentTasks:    -1,  // 无限制
			MaxAPICallsPerMinute:  -1,  // 无限制
		}
	default:
		return GetDefaultQuota(TenantPlanFree)
	}
}

// IsQuotaExceeded 检查配额是否超限
func (u *TenantUsage) IsQuotaExceeded(quota TenantQuota, quotaType string, requestAmount int64) (bool, string) {
	switch quotaType {
	case "users":
		if quota.MaxUsers == -1 {
			return false, ""
		}
		if u.CurrentUsers >= quota.MaxUsers {
			return true, "用户数量已达上限"
		}
	case "agents":
		if quota.MaxAgents == -1 {
			return false, ""
		}
		if u.CurrentAgents >= quota.MaxAgents {
			return true, "Agent数量已达上限"
		}
	case "tokens":
		if quota.MaxTokensPerMonth == -1 {
			return false, ""
		}
		if u.TokensUsedThisMonth+requestAmount > quota.MaxTokensPerMonth {
			return true, "本月Token配额已用完"
		}
	case "storage":
		if quota.MaxStorageBytes == -1 {
			return false, ""
		}
		if u.StorageUsedBytes+requestAmount > quota.MaxStorageBytes {
			return true, "存储空间已满"
		}
	case "tasks":
		if quota.MaxConcurrentTasks == -1 {
			return false, ""
		}
		if u.ActiveTasks >= quota.MaxConcurrentTasks {
			return true, "并发任务数已达上限"
		}
	case "api_calls":
		if quota.MaxAPICallsPerMinute == -1 {
			return false, ""
		}
		if u.APICallsThisMinute >= quota.MaxAPICallsPerMinute {
			return true, "API调用频率超限"
		}
	}
	return false, ""
}

// GetUsagePercentage 获取使用率百分比
func (u *TenantUsage) GetUsagePercentage(quota TenantQuota) map[string]float64 {
	percentages := make(map[string]float64)

	if quota.MaxUsers > 0 {
		percentages["users"] = float64(u.CurrentUsers) / float64(quota.MaxUsers) * 100
	}
	if quota.MaxAgents > 0 {
		percentages["agents"] = float64(u.CurrentAgents) / float64(quota.MaxAgents) * 100
	}
	if quota.MaxTokensPerMonth > 0 {
		percentages["tokens"] = float64(u.TokensUsedThisMonth) / float64(quota.MaxTokensPerMonth) * 100
	}
	if quota.MaxStorageBytes > 0 {
		percentages["storage"] = float64(u.StorageUsedBytes) / float64(quota.MaxStorageBytes) * 100
	}
	if quota.MaxConcurrentTasks > 0 {
		percentages["tasks"] = float64(u.ActiveTasks) / float64(quota.MaxConcurrentTasks) * 100
	}
	if quota.MaxAPICallsPerMinute > 0 {
		percentages["api_calls"] = float64(u.APICallsThisMinute) / float64(quota.MaxAPICallsPerMinute) * 100
	}

	return percentages
}
