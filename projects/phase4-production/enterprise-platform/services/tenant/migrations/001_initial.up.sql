-- Tenant Service Database Migration
-- Version: 001
-- Description: Initial schema for tenant management

-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    company VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenants_email ON tenants(email);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);

-- 租户配额表
CREATE TABLE IF NOT EXISTS tenant_quotas (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    max_users INTEGER NOT NULL DEFAULT 5,
    max_agents INTEGER NOT NULL DEFAULT 3,
    max_tokens_per_month BIGINT NOT NULL DEFAULT 100000,
    max_storage_bytes BIGINT NOT NULL DEFAULT 1073741824,
    max_concurrent_tasks INTEGER NOT NULL DEFAULT 5,
    max_api_calls_per_minute INTEGER NOT NULL DEFAULT 60,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id)
);

CREATE INDEX idx_tenant_quotas_tenant_id ON tenant_quotas(tenant_id);

-- 租户使用情况表
CREATE TABLE IF NOT EXISTS tenant_usage (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    current_users INTEGER NOT NULL DEFAULT 0,
    current_agents INTEGER NOT NULL DEFAULT 0,
    tokens_used_this_month BIGINT NOT NULL DEFAULT 0,
    storage_used_bytes BIGINT NOT NULL DEFAULT 0,
    active_tasks INTEGER NOT NULL DEFAULT 0,
    api_calls_this_minute INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id)
);

CREATE INDEX idx_tenant_usage_tenant_id ON tenant_usage(tenant_id);
CREATE INDEX idx_tenant_usage_last_updated ON tenant_usage(last_updated);

-- 租户功能开关表
CREATE TABLE IF NOT EXISTS tenant_features (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    feature VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, feature)
);

CREATE INDEX idx_tenant_features_tenant_id ON tenant_features(tenant_id);
CREATE INDEX idx_tenant_features_feature ON tenant_features(feature);

-- 租户配置表
CREATE TABLE IF NOT EXISTS tenant_configs (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, key)
);

CREATE INDEX idx_tenant_configs_tenant_id ON tenant_configs(tenant_id);
CREATE INDEX idx_tenant_configs_key ON tenant_configs(key);

-- 租户审计日志表
CREATE TABLE IF NOT EXISTS tenant_audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id VARCHAR(36),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(255),
    details TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenant_audit_logs_tenant_id ON tenant_audit_logs(tenant_id);
CREATE INDEX idx_tenant_audit_logs_action ON tenant_audit_logs(action);
CREATE INDEX idx_tenant_audit_logs_created_at ON tenant_audit_logs(created_at DESC);

-- 触发器：自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_quotas_updated_at BEFORE UPDATE ON tenant_quotas
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_features_updated_at BEFORE UPDATE ON tenant_features
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tenant_configs_updated_at BEFORE UPDATE ON tenant_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入示例数据（可选）
INSERT INTO tenants (id, name, company, email, plan, status)
VALUES
    ('tenant-demo-001', 'Demo Company', 'Demo Corp', 'demo@example.com', 'pro', 'active')
ON CONFLICT DO NOTHING;

INSERT INTO tenant_quotas (id, tenant_id, max_users, max_agents, max_tokens_per_month, max_storage_bytes, max_concurrent_tasks, max_api_calls_per_minute)
VALUES
    ('quota-demo-001', 'tenant-demo-001', 100, 50, 10000000, 107374182400, 100, 1000)
ON CONFLICT DO NOTHING;

INSERT INTO tenant_usage (id, tenant_id, current_users, current_agents, tokens_used_this_month, storage_used_bytes, active_tasks, api_calls_this_minute)
VALUES
    ('usage-demo-001', 'tenant-demo-001', 5, 10, 1500000, 5368709120, 3, 15)
ON CONFLICT DO NOTHING;

-- 注释
COMMENT ON TABLE tenants IS '租户表';
COMMENT ON TABLE tenant_quotas IS '租户配额表';
COMMENT ON TABLE tenant_usage IS '租户使用情况表';
COMMENT ON TABLE tenant_features IS '租户功能开关表';
COMMENT ON TABLE tenant_configs IS '租户配置表';
COMMENT ON TABLE tenant_audit_logs IS '租户审计日志表';
