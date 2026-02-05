package alerter

import (
	"context"
	"fmt"
	"time"

	"github.com/agent-learning/enterprise-platform/services/cost/internal/model"
	"github.com/agent-learning/enterprise-platform/services/cost/internal/repository"
)

// Alerter 成本告警器
type Alerter struct {
	repo *repository.CostRepository
}

// NewAlerter 创建告警器
func NewAlerter(repo *repository.CostRepository) *Alerter {
	return &Alerter{
		repo: repo,
	}
}

// CheckBudgetAlerts 检查预算告警
func (a *Alerter) CheckBudgetAlerts(ctx context.Context, tenantID string) ([]*model.CostAlert, error) {
	// 获取预算
	budget, err := a.repo.GetCurrentBudget(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	alerts := make([]*model.CostAlert, 0)

	// 检查是否超预算
	if budget.IsOverBudget() {
		alert := &model.CostAlert{
			TenantID:       tenantID,
			AlertType:      "threshold",
			Severity:       "critical",
			Title:          "成本超出预算",
			Message:        fmt.Sprintf("当前成本 $%.2f 已超出预算 $%.2f", budget.Spent, budget.Budget),
			CurrentValue:   budget.Spent,
			ThresholdValue: budget.Budget,
			Status:         "active",
		}
		alerts = append(alerts, alert)
	} else if budget.ShouldAlert() {
		// 接近预算阈值
		usagePercent := budget.GetUsagePercent()
		severity := a.getSeverity(usagePercent)

		alert := &model.CostAlert{
			TenantID:       tenantID,
			AlertType:      "threshold",
			Severity:       severity,
			Title:          fmt.Sprintf("成本已使用 %.1f%%", usagePercent),
			Message:        fmt.Sprintf("当前成本 $%.2f，预算 $%.2f，剩余 $%.2f", budget.Spent, budget.Budget, budget.Remaining),
			CurrentValue:   budget.Spent,
			ThresholdValue: budget.Budget * (budget.AlertAt / 100),
			Status:         "active",
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CheckForecastAlerts 检查预测告警
func (a *Alerter) CheckForecastAlerts(ctx context.Context, tenantID string, forecast *model.CostForecast) ([]*model.CostAlert, error) {
	// 获取预算
	budget, err := a.repo.GetCurrentBudget(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	alerts := make([]*model.CostAlert, 0)

	// 检查预测成本是否会超预算
	if forecast.PredictedCost > budget.Budget {
		excessAmount := forecast.PredictedCost - budget.Budget
		excessPercent := (excessAmount / budget.Budget) * 100

		alert := &model.CostAlert{
			TenantID:       tenantID,
			AlertType:      "forecast",
			Severity:       "warning",
			Title:          "预测成本将超预算",
			Message:        fmt.Sprintf("根据当前趋势，预测下月成本 $%.2f 将超出预算 $%.2f (超出 %.1f%%)", forecast.PredictedCost, budget.Budget, excessPercent),
			CurrentValue:   forecast.PredictedCost,
			ThresholdValue: budget.Budget,
			Status:         "active",
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// CheckAnomalyAlerts 检查异常告警
func (a *Alerter) CheckAnomalyAlerts(ctx context.Context, tenantID string, isAnomaly bool, message string) (*model.CostAlert, error) {
	if !isAnomaly {
		return nil, nil
	}

	alert := &model.CostAlert{
		TenantID:  tenantID,
		AlertType: "anomaly",
		Severity:  "warning",
		Title:     "检测到成本异常",
		Message:   message,
		Status:    "active",
	}

	return alert, nil
}

// CheckDailySpike 检查日度激增
func (a *Alerter) CheckDailySpike(ctx context.Context, tenantID string) (*model.CostAlert, error) {
	// 获取今天和昨天的成本
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.AddDate(0, 0, -1)

	todayCost, err := a.repo.GetDailyCost(ctx, tenantID, today)
	if err != nil {
		return nil, err
	}

	yesterdayCost, err := a.repo.GetDailyCost(ctx, tenantID, yesterday)
	if err != nil {
		return nil, err
	}

	// 如果今天的成本比昨天增加50%以上
	if yesterdayCost > 0 {
		increasePercent := ((todayCost - yesterdayCost) / yesterdayCost) * 100

		if increasePercent > 50 {
			alert := &model.CostAlert{
				TenantID:       tenantID,
				AlertType:      "anomaly",
				Severity:       "warning",
				Title:          "成本激增",
				Message:        fmt.Sprintf("今日成本 $%.2f 比昨日 $%.2f 增加 %.1f%%", todayCost, yesterdayCost, increasePercent),
				CurrentValue:   todayCost,
				ThresholdValue: yesterdayCost * 1.5,
				Status:         "active",
			}
			return alert, nil
		}
	}

	return nil, nil
}

// CheckRateLimit 检查速率限制
func (a *Alerter) CheckRateLimit(ctx context.Context, tenantID string) (*model.CostAlert, error) {
	// 获取过去1小时的请求数
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	requestCount, err := a.repo.GetRequestCount(ctx, tenantID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// 假设限制是10000请求/小时
	rateLimit := int64(10000)

	if requestCount > rateLimit {
		alert := &model.CostAlert{
			TenantID:       tenantID,
			AlertType:      "threshold",
			Severity:       "warning",
			Title:          "API调用频率过高",
			Message:        fmt.Sprintf("过去1小时API调用 %d 次，超过限制 %d 次", requestCount, rateLimit),
			CurrentValue:   float64(requestCount),
			ThresholdValue: float64(rateLimit),
			Status:         "active",
		}
		return alert, nil
	}

	return nil, nil
}

// CreateAlert 创建告警
func (a *Alerter) CreateAlert(ctx context.Context, alert *model.CostAlert) error {
	// 检查是否已存在相同的未解决告警
	existingAlerts, err := a.repo.GetActiveAlerts(ctx, alert.TenantID, alert.AlertType)
	if err != nil {
		return err
	}

	// 如果已存在相同类型的活跃告警，不重复创建
	for _, existing := range existingAlerts {
		if existing.Title == alert.Title {
			return nil
		}
	}

	// 创建新告警
	return a.repo.CreateAlert(ctx, alert)
}

// AcknowledgeAlert 确认告警
func (a *Alerter) AcknowledgeAlert(ctx context.Context, alertID string) error {
	now := time.Now()
	return a.repo.UpdateAlertStatus(ctx, alertID, "acknowledged", &now, nil)
}

// ResolveAlert 解决告警
func (a *Alerter) ResolveAlert(ctx context.Context, alertID string) error {
	now := time.Now()
	return a.repo.UpdateAlertStatus(ctx, alertID, "resolved", nil, &now)
}

// GetActiveAlerts 获取活跃告警
func (a *Alerter) GetActiveAlerts(ctx context.Context, tenantID string) ([]*model.CostAlert, error) {
	return a.repo.GetActiveAlerts(ctx, tenantID, "")
}

// AutoResolveAlerts 自动解决已修复的告警
func (a *Alerter) AutoResolveAlerts(ctx context.Context, tenantID string) error {
	// 获取所有活跃的预算告警
	alerts, err := a.repo.GetActiveAlerts(ctx, tenantID, "threshold")
	if err != nil {
		return err
	}

	// 获取当前预算状态
	budget, err := a.repo.GetCurrentBudget(ctx, tenantID)
	if err != nil {
		return err
	}

	// 如果预算已恢复正常，自动解决相关告警
	if !budget.IsOverBudget() && !budget.ShouldAlert() {
		for _, alert := range alerts {
			if alert.Status == "active" {
				a.ResolveAlert(ctx, alert.ID)
			}
		}
	}

	return nil
}

// getSeverity 根据使用百分比确定严重程度
func (a *Alerter) getSeverity(usagePercent float64) string {
	if usagePercent >= 95 {
		return "critical"
	} else if usagePercent >= 85 {
		return "warning"
	}
	return "info"
}

// SendNotification 发送通知（待实现）
func (a *Alerter) SendNotification(ctx context.Context, alert *model.CostAlert, channels []string) error {
	// TODO: 实现通知发送
	// - Email
	// - Slack
	// - Webhook
	// - SMS

	fmt.Printf("[ALERT] %s - %s: %s\n", alert.Severity, alert.Title, alert.Message)

	return nil
}
