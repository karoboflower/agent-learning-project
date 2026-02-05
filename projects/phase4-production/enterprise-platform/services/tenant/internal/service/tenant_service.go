package service

import (
	"context"
	"fmt"

	"github.com/agent-learning/enterprise-platform/services/tenant/internal/isolation"
	"github.com/agent-learning/enterprise-platform/services/tenant/internal/model"
	"github.com/agent-learning/enterprise-platform/services/tenant/internal/quota"
	"github.com/agent-learning/enterprise-platform/services/tenant/internal/repository"
)

// TenantService 租户服务
type TenantService struct {
	repo            *repository.TenantRepository
	quotaManager    *quota.QuotaManager
	isolationMgr    *isolation.IsolationManager
	cacheIsolation  *isolation.CacheIsolation
	storageIsolation *isolation.StorageIsolation
}

// NewTenantService 创建租户服务
func NewTenantService(
	repo *repository.TenantRepository,
	quotaManager *quota.QuotaManager,
	isolationMgr *isolation.IsolationManager,
) *TenantService {
	return &TenantService{
		repo:             repo,
		quotaManager:     quotaManager,
		isolationMgr:     isolationMgr,
		cacheIsolation:   isolation.NewCacheIsolation(),
		storageIsolation: isolation.NewStorageIsolation("/var/data/tenants"),
	}
}

// CreateTenant 创建租户
func (s *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*CreateTenantResponse, error) {
	// 检查邮箱是否已存在
	existing, _ := s.repo.GetTenantByEmail(ctx, req.Email)
	if existing != nil {
		return nil, fmt.Errorf("tenant with email %s already exists", req.Email)
	}

	// 创建租户
	tenant := &model.Tenant{
		Name:    req.Name,
		Company: req.Company,
		Email:   req.Email,
		Plan:    model.TenantPlan(req.Plan),
		Status:  model.TenantStatusActive,
	}

	err := s.repo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 创建默认配额
	defaultQuota := model.GetDefaultQuota(tenant.Plan)
	defaultQuota.TenantID = tenant.ID

	err = s.repo.CreateQuota(ctx, &defaultQuota)
	if err != nil {
		return nil, fmt.Errorf("failed to create quota: %w", err)
	}

	// 初始化使用情况
	usage := &model.TenantUsage{
		TenantID:            tenant.ID,
		CurrentUsers:        0,
		CurrentAgents:       0,
		TokensUsedThisMonth: 0,
		StorageUsedBytes:    0,
		ActiveTasks:         0,
		APICallsThisMinute:  0,
	}

	err = s.repo.CreateUsage(ctx, usage)
	if err != nil {
		return nil, fmt.Errorf("failed to create usage: %w", err)
	}

	// 设置默认功能开关
	defaultFeatures := s.getDefaultFeatures(tenant.Plan)
	for feature, enabled := range defaultFeatures {
		s.repo.SetFeature(ctx, tenant.ID, feature, enabled)
	}

	return &CreateTenantResponse{
		TenantID:  tenant.ID,
		Plan:      string(tenant.Plan),
		Quota:     &defaultQuota,
		CreatedAt: tenant.CreatedAt,
	}, nil
}

// GetTenant 获取租户
func (s *TenantService) GetTenant(ctx context.Context, tenantID string) (*GetTenantResponse, error) {
	tenant, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	quota, err := s.repo.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	usage, err := s.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	features, err := s.repo.GetFeatures(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return &GetTenantResponse{
		Tenant:   tenant,
		Quota:    quota,
		Usage:    usage,
		Features: features,
	}, nil
}

// UpdateTenantQuota 更新租户配额
func (s *TenantService) UpdateTenantQuota(ctx context.Context, tenantID string, quota *model.TenantQuota) error {
	// 验证租户存在
	_, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	// 更新配额
	return s.quotaManager.UpdateQuota(ctx, tenantID, quota)
}

// GetTenantUsage 获取租户使用情况
func (s *TenantService) GetTenantUsage(ctx context.Context, tenantID string) (*GetTenantUsageResponse, error) {
	usage, err := s.repo.GetUsage(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	quota, err := s.repo.GetQuota(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	percentages := usage.GetUsagePercentage(*quota)

	// 计算总体使用率
	var totalPercentage float64
	count := 0
	for _, p := range percentages {
		totalPercentage += p
		count++
	}
	avgPercentage := totalPercentage / float64(count)

	return &GetTenantUsageResponse{
		Usage:            usage,
		Quota:            quota,
		UsagePercentages: percentages,
		AvgPercentage:    avgPercentage,
	}, nil
}

// CheckQuota 检查配额
func (s *TenantService) CheckQuota(ctx context.Context, tenantID, quotaType string, requestAmount int64) (*CheckQuotaResponse, error) {
	allowed, remaining, err := s.quotaManager.CheckQuota(ctx, tenantID, quotaType, requestAmount)

	var reason string
	if err != nil {
		reason = err.Error()
	}

	return &CheckQuotaResponse{
		Allowed:   allowed,
		Remaining: remaining,
		Reason:    reason,
	}, err
}

// ConsumeQuota 消费配额
func (s *TenantService) ConsumeQuota(ctx context.Context, tenantID, quotaType string, amount int64) error {
	return s.quotaManager.ConsumeQuota(ctx, tenantID, quotaType, amount)
}

// UpdateTenantFeatures 更新租户功能开关
func (s *TenantService) UpdateTenantFeatures(ctx context.Context, tenantID string, features map[string]bool) error {
	// 验证租户存在
	_, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	// 更新每个功能开关
	for feature, enabled := range features {
		err = s.repo.SetFeature(ctx, tenantID, feature, enabled)
		if err != nil {
			return fmt.Errorf("failed to set feature %s: %w", feature, err)
		}
	}

	return nil
}

// UpgradeTenant 升级租户计划
func (s *TenantService) UpgradeTenant(ctx context.Context, tenantID string, newPlan model.TenantPlan) error {
	tenant, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	// 更新租户计划
	tenant.Plan = newPlan
	err = s.repo.UpdateTenant(ctx, tenant)
	if err != nil {
		return err
	}

	// 更新配额
	newQuota := model.GetDefaultQuota(newPlan)
	newQuota.TenantID = tenantID
	err = s.repo.UpdateQuota(ctx, &newQuota)
	if err != nil {
		return err
	}

	// 更新功能开关
	features := s.getDefaultFeatures(newPlan)
	return s.UpdateTenantFeatures(ctx, tenantID, features)
}

// SuspendTenant 暂停租户
func (s *TenantService) SuspendTenant(ctx context.Context, tenantID, reason string) error {
	tenant, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	tenant.Status = model.TenantStatusSuspended
	return s.repo.UpdateTenant(ctx, tenant)
}

// ReactivateTenant 重新激活租户
func (s *TenantService) ReactivateTenant(ctx context.Context, tenantID string) error {
	tenant, err := s.repo.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	tenant.Status = model.TenantStatusActive
	return s.repo.UpdateTenant(ctx, tenant)
}

// GetStoragePath 获取租户存储路径
func (s *TenantService) GetStoragePath(tenantID string) string {
	return s.storageIsolation.GetStoragePath(tenantID)
}

// GetCacheKey 获取租户缓存键
func (s *TenantService) GetCacheKey(tenantID, key string) string {
	return s.cacheIsolation.GetCacheKey(tenantID, key)
}

// getDefaultFeatures 获取计划的默认功能开关
func (s *TenantService) getDefaultFeatures(plan model.TenantPlan) map[string]bool {
	features := map[string]bool{
		"agent_execution":     true,
		"tool_integration":    true,
		"api_access":          true,
		"webhooks":            false,
		"custom_models":       false,
		"advanced_analytics":  false,
		"priority_support":    false,
		"sso":                 false,
		"audit_logs":          false,
		"data_export":         false,
	}

	switch plan {
	case model.TenantPlanFree:
		// 免费版：基础功能
		features["webhooks"] = false
		features["custom_models"] = false
		features["advanced_analytics"] = false

	case model.TenantPlanStarter:
		// 入门版：增加webhooks
		features["webhooks"] = true

	case model.TenantPlanPro:
		// 专业版：增加高级功能
		features["webhooks"] = true
		features["custom_models"] = true
		features["advanced_analytics"] = true
		features["audit_logs"] = true

	case model.TenantPlanEnterprise:
		// 企业版：所有功能
		for k := range features {
			features[k] = true
		}
	}

	return features
}

// DTO类型定义

type CreateTenantRequest struct {
	Name    string `json:"name"`
	Company string `json:"company"`
	Email   string `json:"email"`
	Plan    string `json:"plan"`
}

type CreateTenantResponse struct {
	TenantID  string              `json:"tenant_id"`
	Plan      string              `json:"plan"`
	Quota     *model.TenantQuota  `json:"quota"`
	CreatedAt interface{}         `json:"created_at"`
}

type GetTenantResponse struct {
	Tenant   *model.Tenant       `json:"tenant"`
	Quota    *model.TenantQuota  `json:"quota"`
	Usage    *model.TenantUsage  `json:"usage"`
	Features map[string]bool     `json:"features"`
}

type GetTenantUsageResponse struct {
	Usage            *model.TenantUsage `json:"usage"`
	Quota            *model.TenantQuota `json:"quota"`
	UsagePercentages map[string]float64 `json:"usage_percentages"`
	AvgPercentage    float64            `json:"avg_percentage"`
}

type CheckQuotaResponse struct {
	Allowed   bool   `json:"allowed"`
	Remaining int64  `json:"remaining"`
	Reason    string `json:"reason"`
}
