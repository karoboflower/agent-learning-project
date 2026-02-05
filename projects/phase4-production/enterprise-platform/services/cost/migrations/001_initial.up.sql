-- Cost Service Database Migration
-- Version: 001
-- Description: Initial schema for cost control and monitoring

-- Token使用记录表（时序数据）
CREATE TABLE IF NOT EXISTS token_usage (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    agent_id VARCHAR(36),
    task_id VARCHAR(36),
    model VARCHAR(100) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    input_cost DECIMAL(10, 6) NOT NULL DEFAULT 0,
    output_cost DECIMAL(10, 6) NOT NULL DEFAULT 0,
    total_cost DECIMAL(10, 6) NOT NULL DEFAULT 0,
    duration BIGINT,
    cached BOOLEAN NOT NULL DEFAULT false,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_token_usage_tenant_id ON token_usage(tenant_id);
CREATE INDEX idx_token_usage_user_id ON token_usage(user_id);
CREATE INDEX idx_token_usage_agent_id ON token_usage(agent_id);
CREATE INDEX idx_token_usage_model ON token_usage(model);
CREATE INDEX idx_token_usage_timestamp ON token_usage(timestamp DESC);
CREATE INDEX idx_token_usage_tenant_timestamp ON token_usage(tenant_id, timestamp DESC);

-- 成本统计表
CREATE TABLE IF NOT EXISTS cost_statistics (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    period VARCHAR(20) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    request_count BIGINT NOT NULL DEFAULT 0,
    avg_tokens_per_request DECIMAL(10, 2) DEFAULT 0,
    avg_cost_per_request DECIMAL(10, 6) DEFAULT 0,
    cached_requests BIGINT NOT NULL DEFAULT 0,
    cache_hit_rate DECIMAL(5, 2) DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, period, period_start)
);

CREATE INDEX idx_cost_statistics_tenant_id ON cost_statistics(tenant_id);
CREATE INDEX idx_cost_statistics_period ON cost_statistics(period);
CREATE INDEX idx_cost_statistics_period_start ON cost_statistics(period_start DESC);

-- 成本预测表
CREATE TABLE IF NOT EXISTS cost_forecasts (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    forecast_date DATE NOT NULL,
    predicted_tokens BIGINT NOT NULL,
    predicted_cost DECIMAL(10, 2) NOT NULL,
    confidence DECIMAL(3, 2) NOT NULL,
    method VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, forecast_date)
);

CREATE INDEX idx_cost_forecasts_tenant_id ON cost_forecasts(tenant_id);
CREATE INDEX idx_cost_forecasts_forecast_date ON cost_forecasts(forecast_date);

-- 成本告警表
CREATE TABLE IF NOT EXISTS cost_alerts (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    current_value DECIMAL(10, 2),
    threshold_value DECIMAL(10, 2),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP,
    resolved_at TIMESTAMP
);

CREATE INDEX idx_cost_alerts_tenant_id ON cost_alerts(tenant_id);
CREATE INDEX idx_cost_alerts_status ON cost_alerts(status);
CREATE INDEX idx_cost_alerts_severity ON cost_alerts(severity);
CREATE INDEX idx_cost_alerts_created_at ON cost_alerts(created_at DESC);

-- 成本预算表
CREATE TABLE IF NOT EXISTS cost_budgets (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    period VARCHAR(20) NOT NULL,
    budget DECIMAL(10, 2) NOT NULL,
    spent DECIMAL(10, 2) NOT NULL DEFAULT 0,
    remaining DECIMAL(10, 2) NOT NULL,
    alert_at DECIMAL(5, 2) NOT NULL DEFAULT 80.0,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, period, start_date)
);

CREATE INDEX idx_cost_budgets_tenant_id ON cost_budgets(tenant_id);
CREATE INDEX idx_cost_budgets_period ON cost_budgets(period);
CREATE INDEX idx_cost_budgets_start_date ON cost_budgets(start_date);

-- 触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_cost_statistics_updated_at BEFORE UPDATE ON cost_statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cost_budgets_updated_at BEFORE UPDATE ON cost_budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 触发器：自动更新预算剩余
CREATE OR REPLACE FUNCTION update_budget_remaining()
RETURNS TRIGGER AS $$
BEGIN
    NEW.remaining = NEW.budget - NEW.spent;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_cost_budgets_remaining BEFORE INSERT OR UPDATE ON cost_budgets
    FOR EACH ROW EXECUTE FUNCTION update_budget_remaining();

-- 视图：每日成本统计
CREATE OR REPLACE VIEW daily_cost_summary AS
SELECT
    tenant_id,
    DATE(timestamp) as date,
    model,
    provider,
    SUM(total_tokens) as total_tokens,
    SUM(total_cost) as total_cost,
    COUNT(*) as request_count,
    AVG(total_tokens) as avg_tokens,
    AVG(total_cost) as avg_cost,
    SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_requests,
    (SUM(CASE WHEN cached THEN 1 ELSE 0 END)::float / COUNT(*)::float * 100) as cache_hit_rate
FROM token_usage
GROUP BY tenant_id, DATE(timestamp), model, provider
ORDER BY date DESC;

-- 视图：租户成本排名
CREATE OR REPLACE VIEW tenant_cost_ranking AS
SELECT
    tenant_id,
    SUM(total_cost) as total_cost,
    SUM(total_tokens) as total_tokens,
    COUNT(*) as request_count,
    RANK() OVER (ORDER BY SUM(total_cost) DESC) as cost_rank
FROM token_usage
WHERE timestamp >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY tenant_id
ORDER BY total_cost DESC;

-- 视图：模型使用统计
CREATE OR REPLACE VIEW model_usage_stats AS
SELECT
    model,
    provider,
    COUNT(*) as usage_count,
    SUM(total_tokens) as total_tokens,
    SUM(total_cost) as total_cost,
    AVG(total_tokens) as avg_tokens_per_request,
    AVG(total_cost) as avg_cost_per_request,
    AVG(duration) as avg_duration_ms
FROM token_usage
WHERE timestamp >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY model, provider
ORDER BY total_cost DESC;

-- 插入示例数据
INSERT INTO cost_budgets (id, tenant_id, period, budget, spent, start_date, end_date, alert_at)
VALUES
    (gen_random_uuid()::text, 'tenant-demo-001', 'monthly', 1000.00, 250.50, DATE_TRUNC('month', CURRENT_DATE), (DATE_TRUNC('month', CURRENT_DATE) + INTERVAL '1 month' - INTERVAL '1 day')::date, 80.0)
ON CONFLICT DO NOTHING;

-- 注释
COMMENT ON TABLE token_usage IS 'Token使用记录（时序数据）';
COMMENT ON TABLE cost_statistics IS '成本统计汇总';
COMMENT ON TABLE cost_forecasts IS '成本预测';
COMMENT ON TABLE cost_alerts IS '成本告警';
COMMENT ON TABLE cost_budgets IS '成本预算';
COMMENT ON VIEW daily_cost_summary IS '每日成本汇总';
COMMENT ON VIEW tenant_cost_ranking IS '租户成本排名';
COMMENT ON VIEW model_usage_stats IS '模型使用统计';
