package quota

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agent-learning/enterprise-platform/services/tenant/internal/model"
	"github.com/agent-learning/enterprise-platform/services/tenant/internal/repository"
)

// QuotaManager ��额管理器
type QuotaManager struct {
	repo  *repository.TenantRepository
	cache sync.Map // 缓存配额信息
}

// NewQuotaManager 创建配额管理器
func NewQuotaManager(repo *repository.TenantRepository) *QuotaManager {
	qm := &QuotaManager{
		repo: repo,
	}

	// 启动后台任务：定时清理API调用计数
	go qm.resetAPICallsCounter()

	// 启动后台任务：月度使用量重置
	go qm.resetMonthlyUsage()

	return qm
}

// CheckQuota 检查配额
func (qm *QuotaManager) CheckQuota(ctx context.Context, tenantID, quotaType string, requestAmount int64) (bool, int64, error) {
	// 获取配额
	quota, err := qm.getQuota(ctx, tenantID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get quota: %w", err)
	}

	// 获取使用情况
	usage, err := qm.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get usage: %w", err)
	}

	// 检查是否超限
	exceeded, reason := usage.IsQuotaExceeded(*quota, quotaType, requestAmount)
	if exceeded {
		return false, 0, fmt.Errorf("quota exceeded: %s", reason)
	}

	// 计算剩余配额
	remaining := qm.calculateRemaining(usage, quota, quotaType)

	return true, remaining, nil
}

// ConsumeQuota 消费配额
func (qm *QuotaManager) ConsumeQuota(ctx context.Context, tenantID, quotaType string, amount int64) error {
	// 先检查配额
	allowed, _, err := qm.CheckQuota(ctx, tenantID, quotaType, amount)
	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("quota check failed")
	}

	// 增加使用量
	err = qm.repo.IncrementUsage(ctx, tenantID, quotaType, amount)
	if err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}

	// 清除缓存
	qm.cache.Delete(tenantID)

	return nil
}

// ReleaseQuota 释放配额（例如任务完成后）
func (qm *QuotaManager) ReleaseQuota(ctx context.Context, tenantID, quotaType string, amount int64) error {
	err := qm.repo.IncrementUsage(ctx, tenantID, quotaType, -amount)
	if err != nil {
		return fmt.Errorf("failed to decrement usage: %w", err)
	}

	// 清除缓存
	qm.cache.Delete(tenantID)

	return nil
}

// GetUsagePercentage 获取使用率
func (qm *QuotaManager) GetUsagePercentage(ctx context.Context, tenantID string) (map[string]float64, error) {
	quota, err := qm.getQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	usage, err := qm.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return usage.GetUsagePercentage(*quota), nil
}

// UpdateQuota 更新配额
func (qm *QuotaManager) UpdateQuota(ctx context.Context, tenantID string, quota *model.TenantQuota) error {
	quota.TenantID = tenantID

	err := qm.repo.UpdateQuota(ctx, quota)
	if err != nil {
		return err
	}

	// 清除缓存
	qm.cache.Delete(tenantID)

	return nil
}

// getQuota 获取配额（带缓存）
func (qm *QuotaManager) getQuota(ctx context.Context, tenantID string) (*model.TenantQuota, error) {
	// 先查缓存
	if cached, ok := qm.cache.Load(tenantID); ok {
		if quota, ok := cached.(*model.TenantQuota); ok {
			return quota, nil
		}
	}

	// 从数据库获取
	quota, err := qm.repo.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	qm.cache.Store(tenantID, quota)

	return quota, nil
}

// calculateRemaining 计算剩余配额
func (qm *QuotaManager) calculateRemaining(usage *model.TenantUsage, quota *model.TenantQuota, quotaType string) int64 {
	switch quotaType {
	case "users":
		if quota.MaxUsers == -1 {
			return -1 // 无限制
		}
		return int64(quota.MaxUsers - usage.CurrentUsers)
	case "agents":
		if quota.MaxAgents == -1 {
			return -1
		}
		return int64(quota.MaxAgents - usage.CurrentAgents)
	case "tokens":
		if quota.MaxTokensPerMonth == -1 {
			return -1
		}
		return quota.MaxTokensPerMonth - usage.TokensUsedThisMonth
	case "storage":
		if quota.MaxStorageBytes == -1 {
			return -1
		}
		return quota.MaxStorageBytes - usage.StorageUsedBytes
	case "tasks":
		if quota.MaxConcurrentTasks == -1 {
			return -1
		}
		return int64(quota.MaxConcurrentTasks - usage.ActiveTasks)
	case "api_calls":
		if quota.MaxAPICallsPerMinute == -1 {
			return -1
		}
		return int64(quota.MaxAPICallsPerMinute - usage.APICallsThisMinute)
	default:
		return 0
	}
}

// resetAPICallsCounter 每分钟重置API调用计数
func (qm *QuotaManager) resetAPICallsCounter() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// TODO: 重置所有租户的API调用计数
		// 这里简化处理，实际应该遍历所有租户
		qm.cache.Range(func(key, value interface{}) bool {
			if tenantID, ok := key.(string); ok {
				ctx := context.Background()
				usage, err := qm.repo.GetUsage(ctx, tenantID)
				if err == nil {
					usage.APICallsThisMinute = 0
					qm.repo.UpdateUsage(ctx, usage)
				}
			}
			return true
		})
	}
}

// resetMonthlyUsage 每月1号重置月度使用量
func (qm *QuotaManager) resetMonthlyUsage() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		// 检查是否是每月1号凌晨
		if now.Day() == 1 && now.Hour() == 0 {
			// TODO: 重置所有租户的月度使用量
			// 这里简化处理，实际应该遍历所有租户
			qm.cache.Range(func(key, value interface{}) bool {
				if tenantID, ok := key.(string); ok {
					ctx := context.Background()
					qm.repo.ResetMonthlyUsage(ctx, tenantID)
				}
				return true
			})
		}
	}
}

// Alert 配额告警
type Alert struct {
	TenantID   string
	QuotaType  string
	Percentage float64
	Timestamp  time.Time
}

// CheckAlerts 检查配额告警
func (qm *QuotaManager) CheckAlerts(ctx context.Context, tenantID string) ([]Alert, error) {
	percentages, err := qm.GetUsagePercentage(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var alerts []Alert

	// 检查超过80%的配额项
	for quotaType, percentage := range percentages {
		if percentage >= 80.0 {
			alerts = append(alerts, Alert{
				TenantID:   tenantID,
				QuotaType:  quotaType,
				Percentage: percentage,
				Timestamp:  time.Now(),
			})
		}
	}

	return alerts, nil
}
