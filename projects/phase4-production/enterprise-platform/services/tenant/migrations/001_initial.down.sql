-- Rollback Migration 001

DROP TRIGGER IF EXISTS update_tenant_configs_updated_at ON tenant_configs;
DROP TRIGGER IF EXISTS update_tenant_features_updated_at ON tenant_features;
DROP TRIGGER IF EXISTS update_tenant_quotas_updated_at ON tenant_quotas;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS tenant_audit_logs;
DROP TABLE IF EXISTS tenant_configs;
DROP TABLE IF EXISTS tenant_features;
DROP TABLE IF EXISTS tenant_usage;
DROP TABLE IF EXISTS tenant_quotas;
DROP TABLE IF EXISTS tenants;
