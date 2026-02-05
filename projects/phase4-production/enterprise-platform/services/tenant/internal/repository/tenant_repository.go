package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/agent-learning/enterprise-platform/services/tenant/internal/model"
)

// TenantRepository 租户数据访问层
type TenantRepository struct {
	db *sqlx.DB
}

// NewTenantRepository 创建租户仓库
func NewTenantRepository(db *sqlx.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// CreateTenant 创建租户
func (r *TenantRepository) CreateTenant(ctx context.Context, tenant *model.Tenant) error {
	tenant.ID = uuid.New().String()
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenants (id, name, company, email, plan, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Company,
		tenant.Email,
		tenant.Plan,
		tenant.Status,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	return err
}

// GetTenant 获取租户
func (r *TenantRepository) GetTenant(ctx context.Context, tenantID string) (*model.Tenant, error) {
	var tenant model.Tenant
	query := `SELECT * FROM tenants WHERE id = $1`

	err := r.db.GetContext(ctx, &tenant, query, tenantID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetTenantByEmail 根据邮箱获取租户
func (r *TenantRepository) GetTenantByEmail(ctx context.Context, email string) (*model.Tenant, error) {
	var tenant model.Tenant
	query := `SELECT * FROM tenants WHERE email = $1`

	err := r.db.GetContext(ctx, &tenant, query, email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", email)
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// UpdateTenant 更新租户
func (r *TenantRepository) UpdateTenant(ctx context.Context, tenant *model.Tenant) error {
	tenant.UpdatedAt = time.Now()

	query := `
		UPDATE tenants
		SET name = $2, company = $3, email = $4, plan = $5, status = $6, updated_at = $7
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Company,
		tenant.Email,
		tenant.Plan,
		tenant.Status,
		tenant.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}

	return nil
}

// DeleteTenant 删除租户
func (r *TenantRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenants WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, tenantID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}

	return nil
}

// ListTenants 列出租户
func (r *TenantRepository) ListTenants(ctx context.Context, limit, offset int) ([]*model.Tenant, error) {
	var tenants []*model.Tenant
	query := `SELECT * FROM tenants ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	err := r.db.SelectContext(ctx, &tenants, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return tenants, nil
}

// CreateQuota 创建配额
func (r *TenantRepository) CreateQuota(ctx context.Context, quota *model.TenantQuota) error {
	quota.ID = uuid.New().String()
	quota.CreatedAt = time.Now()
	quota.UpdatedAt = time.Now()

	query := `
		INSERT INTO tenant_quotas (
			id, tenant_id, max_users, max_agents, max_tokens_per_month,
			max_storage_bytes, max_concurrent_tasks, max_api_calls_per_minute,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		quota.ID,
		quota.TenantID,
		quota.MaxUsers,
		quota.MaxAgents,
		quota.MaxTokensPerMonth,
		quota.MaxStorageBytes,
		quota.MaxConcurrentTasks,
		quota.MaxAPICallsPerMinute,
		quota.CreatedAt,
		quota.UpdatedAt,
	)

	return err
}

// GetQuota 获取配额
func (r *TenantRepository) GetQuota(ctx context.Context, tenantID string) (*model.TenantQuota, error) {
	var quota model.TenantQuota
	query := `SELECT * FROM tenant_quotas WHERE tenant_id = $1`

	err := r.db.GetContext(ctx, &quota, query, tenantID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("quota not found for tenant: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	return &quota, nil
}

// UpdateQuota 更新配额
func (r *TenantRepository) UpdateQuota(ctx context.Context, quota *model.TenantQuota) error {
	quota.UpdatedAt = time.Now()

	query := `
		UPDATE tenant_quotas
		SET max_users = $2, max_agents = $3, max_tokens_per_month = $4,
		    max_storage_bytes = $5, max_concurrent_tasks = $6,
		    max_api_calls_per_minute = $7, updated_at = $8
		WHERE tenant_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		quota.TenantID,
		quota.MaxUsers,
		quota.MaxAgents,
		quota.MaxTokensPerMonth,
		quota.MaxStorageBytes,
		quota.MaxConcurrentTasks,
		quota.MaxAPICallsPerMinute,
		quota.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("quota not found for tenant: %s", quota.TenantID)
	}

	return nil
}

// CreateUsage 创建使用情况记录
func (r *TenantRepository) CreateUsage(ctx context.Context, usage *model.TenantUsage) error {
	usage.ID = uuid.New().String()
	usage.LastUpdated = time.Now()

	query := `
		INSERT INTO tenant_usage (
			id, tenant_id, current_users, current_agents, tokens_used_this_month,
			storage_used_bytes, active_tasks, api_calls_this_minute, last_updated
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		usage.ID,
		usage.TenantID,
		usage.CurrentUsers,
		usage.CurrentAgents,
		usage.TokensUsedThisMonth,
		usage.StorageUsedBytes,
		usage.ActiveTasks,
		usage.APICallsThisMinute,
		usage.LastUpdated,
	)

	return err
}

// GetUsage 获取使用情况
func (r *TenantRepository) GetUsage(ctx context.Context, tenantID string) (*model.TenantUsage, error) {
	var usage model.TenantUsage
	query := `SELECT * FROM tenant_usage WHERE tenant_id = $1`

	err := r.db.GetContext(ctx, &usage, query, tenantID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usage not found for tenant: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	return &usage, nil
}

// UpdateUsage 更新使用情况
func (r *TenantRepository) UpdateUsage(ctx context.Context, usage *model.TenantUsage) error {
	usage.LastUpdated = time.Now()

	query := `
		UPDATE tenant_usage
		SET current_users = $2, current_agents = $3, tokens_used_this_month = $4,
		    storage_used_bytes = $5, active_tasks = $6, api_calls_this_minute = $7,
		    last_updated = $8
		WHERE tenant_id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		usage.TenantID,
		usage.CurrentUsers,
		usage.CurrentAgents,
		usage.TokensUsedThisMonth,
		usage.StorageUsedBytes,
		usage.ActiveTasks,
		usage.APICallsThisMinute,
		usage.LastUpdated,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("usage not found for tenant: %s", usage.TenantID)
	}

	return nil
}

// IncrementUsage 增加使用量
func (r *TenantRepository) IncrementUsage(ctx context.Context, tenantID, usageType string, amount int64) error {
	var query string

	switch usageType {
	case "users":
		query = `UPDATE tenant_usage SET current_users = current_users + $2, last_updated = $3 WHERE tenant_id = $1`
	case "agents":
		query = `UPDATE tenant_usage SET current_agents = current_agents + $2, last_updated = $3 WHERE tenant_id = $1`
	case "tokens":
		query = `UPDATE tenant_usage SET tokens_used_this_month = tokens_used_this_month + $2, last_updated = $3 WHERE tenant_id = $1`
	case "storage":
		query = `UPDATE tenant_usage SET storage_used_bytes = storage_used_bytes + $2, last_updated = $3 WHERE tenant_id = $1`
	case "tasks":
		query = `UPDATE tenant_usage SET active_tasks = active_tasks + $2, last_updated = $3 WHERE tenant_id = $1`
	case "api_calls":
		query = `UPDATE tenant_usage SET api_calls_this_minute = api_calls_this_minute + $2, last_updated = $3 WHERE tenant_id = $1`
	default:
		return fmt.Errorf("unknown usage type: %s", usageType)
	}

	_, err := r.db.ExecContext(ctx, query, tenantID, amount, time.Now())
	return err
}

// ResetMonthlyUsage 重置月度使用量
func (r *TenantRepository) ResetMonthlyUsage(ctx context.Context, tenantID string) error {
	query := `UPDATE tenant_usage SET tokens_used_this_month = 0, last_updated = $2 WHERE tenant_id = $1`
	_, err := r.db.ExecContext(ctx, query, tenantID, time.Now())
	return err
}

// SetFeature 设置功能开关
func (r *TenantRepository) SetFeature(ctx context.Context, tenantID, feature string, enabled bool) error {
	// 先尝试更新
	query := `UPDATE tenant_features SET enabled = $3, updated_at = $4 WHERE tenant_id = $1 AND feature = $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, feature, enabled, time.Now())
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		// 如果没有更新，则插入
		insertQuery := `
			INSERT INTO tenant_features (id, tenant_id, feature, enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = r.db.ExecContext(ctx, insertQuery,
			uuid.New().String(), tenantID, feature, enabled, time.Now(), time.Now())
		return err
	}

	return nil
}

// GetFeatures 获取所有功能开关
func (r *TenantRepository) GetFeatures(ctx context.Context, tenantID string) (map[string]bool, error) {
	var features []model.TenantFeature
	query := `SELECT * FROM tenant_features WHERE tenant_id = $1`

	err := r.db.SelectContext(ctx, &features, query, tenantID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, f := range features {
		result[f.Feature] = f.Enabled
	}

	return result, nil
}
