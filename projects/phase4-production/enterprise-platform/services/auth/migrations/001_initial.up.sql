-- Auth Service Database Migration
-- Version: 001
-- Description: Initial schema for authentication and authorization

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    parent_id VARCHAR(36),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_roles_is_system ON roles(is_system);
CREATE INDEX idx_roles_parent_id ON roles(parent_id);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id VARCHAR(36) PRIMARY KEY,
    role_id VARCHAR(36) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, permission)
);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id VARCHAR(36) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- 资源表
CREATE TABLE IF NOT EXISTS resources (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(500),
    owner VARCHAR(36) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_resources_tenant_id ON resources(tenant_id);
CREATE INDEX idx_resources_type ON resources(type);
CREATE INDEX idx_resources_owner ON resources(owner);

-- 资源权限表
CREATE TABLE IF NOT EXISTS resource_permissions (
    id VARCHAR(36) PRIMARY KEY,
    resource_id VARCHAR(36) NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    user_id VARCHAR(36),
    role_id VARCHAR(36),
    permission VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resource_id, user_id, permission),
    UNIQUE(resource_id, role_id, permission)
);

CREATE INDEX idx_resource_permissions_resource_id ON resource_permissions(resource_id);
CREATE INDEX idx_resource_permissions_user_id ON resource_permissions(user_id);
CREATE INDEX idx_resource_permissions_role_id ON resource_permissions(role_id);

-- 策略规则表（ABAC）
CREATE TABLE IF NOT EXISTS policy_rules (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    subject VARCHAR(255) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    effect VARCHAR(20) NOT NULL,
    conditions JSONB,
    priority INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policy_rules_tenant_id ON policy_rules(tenant_id);
CREATE INDEX idx_policy_rules_enabled ON policy_rules(enabled);
CREATE INDEX idx_policy_rules_priority ON policy_rules(priority DESC);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    username VARCHAR(100) NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(500),
    result VARCHAR(50) NOT NULL,
    details TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    duration BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_result ON audit_logs(result);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- 触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policy_rules_updated_at BEFORE UPDATE ON policy_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入系统角色
INSERT INTO roles (id, tenant_id, name, description, is_system)
VALUES
    ('system-admin', NULL, 'System Administrator', '完全的系统访问权限', true),
    ('tenant-admin', NULL, 'Tenant Administrator', '租户管理员', true),
    ('developer', NULL, 'Developer', '开发者', true),
    ('viewer', NULL, 'Viewer', '查看者', true),
    ('guest', NULL, 'Guest', '访客', true)
ON CONFLICT DO NOTHING;

-- 系统管理员权限
INSERT INTO role_permissions (id, role_id, permission)
SELECT
    gen_random_uuid()::text,
    'system-admin',
    unnest(ARRAY[
        'tool:execute', 'tool:register', 'tool:unregister', 'tool:list', 'tool:view',
        'resource:read', 'resource:write', 'resource:delete', 'resource:create',
        'user:manage', 'role:manage', 'permission:manage', 'audit:view',
        'agent:create', 'agent:execute', 'agent:view', 'agent:delete',
        'task:create', 'task:view', 'task:cancel', 'task:retry',
        'tenant:manage', 'tenant:view', 'quota:manage',
        'api:read', 'api:write', 'api:admin'
    ])
ON CONFLICT DO NOTHING;

-- 租户管理员权限
INSERT INTO role_permissions (id, role_id, permission)
SELECT
    gen_random_uuid()::text,
    'tenant-admin',
    unnest(ARRAY[
        'tool:execute', 'tool:list', 'tool:view',
        'resource:read', 'resource:write', 'resource:create',
        'user:manage', 'role:manage', 'audit:view',
        'agent:create', 'agent:execute', 'agent:view', 'agent:delete',
        'task:create', 'task:view', 'task:cancel', 'task:retry',
        'tenant:view', 'api:read', 'api:write'
    ])
ON CONFLICT DO NOTHING;

-- 开发者权限
INSERT INTO role_permissions (id, role_id, permission)
SELECT
    gen_random_uuid()::text,
    'developer',
    unnest(ARRAY[
        'tool:execute', 'tool:list', 'tool:view',
        'resource:read', 'resource:write', 'resource:create',
        'agent:create', 'agent:execute', 'agent:view',
        'task:create', 'task:view',
        'api:read', 'api:write'
    ])
ON CONFLICT DO NOTHING;

-- 查看者权限
INSERT INTO role_permissions (id, role_id, permission)
SELECT
    gen_random_uuid()::text,
    'viewer',
    unnest(ARRAY[
        'tool:list', 'tool:view',
        'resource:read',
        'agent:view',
        'task:view',
        'api:read'
    ])
ON CONFLICT DO NOTHING;

-- 访客权限
INSERT INTO role_permissions (id, role_id, permission)
SELECT
    gen_random_uuid()::text,
    'guest',
    unnest(ARRAY[
        'tool:list',
        'agent:view',
        'task:view'
    ])
ON CONFLICT DO NOTHING;

-- 注释
COMMENT ON TABLE users IS '用户表';
COMMENT ON TABLE roles IS '角色表';
COMMENT ON TABLE role_permissions IS '角色权限关联表';
COMMENT ON TABLE user_roles IS '用户角色关联表';
COMMENT ON TABLE resources IS '资源表';
COMMENT ON TABLE resource_permissions IS '资源权限表';
COMMENT ON TABLE policy_rules IS '策略规则表';
COMMENT ON TABLE audit_logs IS '审计日志表';
