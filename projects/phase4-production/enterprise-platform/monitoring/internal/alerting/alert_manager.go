package alerting

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
)

// Alert 告警
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Level       AlertLevel             `json:"level"`
	Status      AlertStatus            `json:"status"`
	Message     string                 `json:"message"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	StartsAt    time.Time              `json:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string
	Query       string // Prometheus查询
	Duration    time.Duration
	Level       AlertLevel
	Threshold   float64
	Operator    string // >, <, >=, <=, ==, !=
	Annotations map[string]string
	Labels      map[string]string
}

// AlertManager 告警管理器
type AlertManager struct {
	alerts    map[string]*Alert
	rules     map[string]*AlertRule
	receivers []AlertReceiver
	mu        sync.RWMutex
}

// AlertReceiver 告警接收器接口
type AlertReceiver interface {
	Send(ctx context.Context, alert *Alert) error
	Name() string
}

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts:    make(map[string]*Alert),
		rules:     make(map[string]*AlertRule),
		receivers: make([]AlertReceiver, 0),
	}
}

// RegisterRule 注册告警规则
func (am *AlertManager) RegisterRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.rules[rule.Name] = rule
}

// UnregisterRule 注销告警规则
func (am *AlertManager) UnregisterRule(name string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.rules, name)
}

// AddReceiver 添加接收器
func (am *AlertManager) AddReceiver(receiver AlertReceiver) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.receivers = append(am.receivers, receiver)
}

// Fire 触发告警
func (am *AlertManager) Fire(ctx context.Context, alert *Alert) error {
	am.mu.Lock()

	// 检查是否已存在
	existingAlert, exists := am.alerts[alert.ID]
	if exists && existingAlert.Status == AlertStatusFiring {
		// 告警已存在且仍在触发中
		am.mu.Unlock()
		return nil
	}

	// 设置告警状态
	alert.Status = AlertStatusFiring
	if alert.StartsAt.IsZero() {
		alert.StartsAt = time.Now()
	}

	am.alerts[alert.ID] = alert

	receivers := make([]AlertReceiver, len(am.receivers))
	copy(receivers, am.receivers)

	am.mu.Unlock()

	// 发送告警到所有接收器
	var errs []error
	for _, receiver := range receivers {
		if err := receiver.Send(ctx, alert); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", receiver.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to send alert to some receivers: %v", errs)
	}

	return nil
}

// Resolve 解决告警
func (am *AlertManager) Resolve(ctx context.Context, alertID string) error {
	am.mu.Lock()

	alert, exists := am.alerts[alertID]
	if !exists {
		am.mu.Unlock()
		return fmt.Errorf("alert not found: %s", alertID)
	}

	if alert.Status == AlertStatusResolved {
		// 告警已解决
		am.mu.Unlock()
		return nil
	}

	// 更新告警状态
	alert.Status = AlertStatusResolved
	now := time.Now()
	alert.EndsAt = &now

	receivers := make([]AlertReceiver, len(am.receivers))
	copy(receivers, am.receivers)

	am.mu.Unlock()

	// 发送解决通知到所有接收器
	var errs []error
	for _, receiver := range receivers {
		if err := receiver.Send(ctx, alert); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", receiver.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to send resolved alert to some receivers: %v", errs)
	}

	return nil
}

// GetAlert 获取告警
func (am *AlertManager) GetAlert(alertID string) (*Alert, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alert, exists := am.alerts[alertID]
	if !exists {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	return alert, nil
}

// ListAlerts 列出告警
func (am *AlertManager) ListAlerts(status AlertStatus) []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range am.alerts {
		if status == "" || alert.Status == status {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// EmailReceiver 邮件接收器
type EmailReceiver struct {
	name      string
	sendEmail func(ctx context.Context, to, subject, body string) error
	recipients []string
}

// NewEmailReceiver 创建邮件接收器
func NewEmailReceiver(name string, recipients []string, sendEmail func(ctx context.Context, to, subject, body string) error) *EmailReceiver {
	return &EmailReceiver{
		name:       name,
		recipients: recipients,
		sendEmail:  sendEmail,
	}
}

// Name 返回接收器名称
func (er *EmailReceiver) Name() string {
	return er.name
}

// Send 发送告警
func (er *EmailReceiver) Send(ctx context.Context, alert *Alert) error {
	subject := fmt.Sprintf("[%s] %s - %s", alert.Level, alert.Status, alert.Name)
	body := fmt.Sprintf(`
Alert: %s
Level: %s
Status: %s
Message: %s
Value: %.2f
Threshold: %.2f
Starts At: %s
`, alert.Name, alert.Level, alert.Status, alert.Message, alert.Value, alert.Threshold, alert.StartsAt.Format(time.RFC3339))

	if alert.EndsAt != nil {
		body += fmt.Sprintf("Ends At: %s\n", alert.EndsAt.Format(time.RFC3339))
	}

	// 发送给所有收件人
	for _, recipient := range er.recipients {
		if err := er.sendEmail(ctx, recipient, subject, body); err != nil {
			return err
		}
	}

	return nil
}

// SlackReceiver Slack接收器
type SlackReceiver struct {
	name        string
	webhookURL  string
	sendWebhook func(ctx context.Context, url, payload string) error
}

// NewSlackReceiver 创建Slack接收器
func NewSlackReceiver(name, webhookURL string, sendWebhook func(ctx context.Context, url, payload string) error) *SlackReceiver {
	return &SlackReceiver{
		name:        name,
		webhookURL:  webhookURL,
		sendWebhook: sendWebhook,
	}
}

// Name 返回接收器名称
func (sr *SlackReceiver) Name() string {
	return sr.name
}

// Send 发送告警
func (sr *SlackReceiver) Send(ctx context.Context, alert *Alert) error {
	// 构建Slack消息
	color := "good"
	if alert.Level == AlertLevelWarning {
		color = "warning"
	} else if alert.Level == AlertLevelCritical {
		color = "danger"
	}

	payload := fmt.Sprintf(`{
		"attachments": [{
			"color": "%s",
			"title": "%s",
			"text": "%s",
			"fields": [
				{"title": "Level", "value": "%s", "short": true},
				{"title": "Status", "value": "%s", "short": true},
				{"title": "Value", "value": "%.2f", "short": true},
				{"title": "Threshold", "value": "%.2f", "short": true}
			],
			"footer": "Alert Manager",
			"ts": %d
		}]
	}`, color, alert.Name, alert.Message, alert.Level, alert.Status, alert.Value, alert.Threshold, alert.StartsAt.Unix())

	return sr.sendWebhook(ctx, sr.webhookURL, payload)
}

// WebhookReceiver Webhook接收器
type WebhookReceiver struct {
	name        string
	url         string
	sendWebhook func(ctx context.Context, url string, alert *Alert) error
}

// NewWebhookReceiver 创建Webhook接收器
func NewWebhookReceiver(name, url string, sendWebhook func(ctx context.Context, url string, alert *Alert) error) *WebhookReceiver {
	return &WebhookReceiver{
		name:        name,
		url:         url,
		sendWebhook: sendWebhook,
	}
}

// Name 返回接收器名称
func (wr *WebhookReceiver) Name() string {
	return wr.name
}

// Send 发送告警
func (wr *WebhookReceiver) Send(ctx context.Context, alert *Alert) error {
	return wr.sendWebhook(ctx, wr.url, alert)
}

// ConsoleReceiver 控制台接收器（用于开发/测试）
type ConsoleReceiver struct {
	name string
}

// NewConsoleReceiver 创建控制台接收器
func NewConsoleReceiver(name string) *ConsoleReceiver {
	return &ConsoleReceiver{
		name: name,
	}
}

// Name 返回接收器名称
func (cr *ConsoleReceiver) Name() string {
	return cr.name
}

// Send 发送告警
func (cr *ConsoleReceiver) Send(ctx context.Context, alert *Alert) error {
	fmt.Printf("[ALERT] %s - %s (%s): %s (Value: %.2f, Threshold: %.2f)\n",
		alert.Level, alert.Status, alert.Name, alert.Message, alert.Value, alert.Threshold)
	return nil
}

// CompositeReceiver 组合接收器
type CompositeReceiver struct {
	name      string
	receivers []AlertReceiver
}

// NewCompositeReceiver 创建组合接收器
func NewCompositeReceiver(name string, receivers ...AlertReceiver) *CompositeReceiver {
	return &CompositeReceiver{
		name:      name,
		receivers: receivers,
	}
}

// Name 返回接收器名称
func (cr *CompositeReceiver) Name() string {
	return cr.name
}

// Send 发送告警
func (cr *CompositeReceiver) Send(ctx context.Context, alert *Alert) error {
	var errs []error
	for _, receiver := range cr.receivers {
		if err := receiver.Send(ctx, alert); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", receiver.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some receivers failed: %v", errs)
	}

	return nil
}

// AlertEvaluator 告警评估器
type AlertEvaluator struct {
	manager *AlertManager
}

// NewAlertEvaluator 创建告警评估器
func NewAlertEvaluator(manager *AlertManager) *AlertEvaluator {
	return &AlertEvaluator{
		manager: manager,
	}
}

// EvaluateRule 评估规则
func (ae *AlertEvaluator) EvaluateRule(ctx context.Context, rule *AlertRule, value float64) error {
	// 评估条件
	shouldFire := false

	switch rule.Operator {
	case ">":
		shouldFire = value > rule.Threshold
	case "<":
		shouldFire = value < rule.Threshold
	case ">=":
		shouldFire = value >= rule.Threshold
	case "<=":
		shouldFire = value <= rule.Threshold
	case "==":
		shouldFire = value == rule.Threshold
	case "!=":
		shouldFire = value != rule.Threshold
	default:
		return fmt.Errorf("unsupported operator: %s", rule.Operator)
	}

	alertID := fmt.Sprintf("%s-%s", rule.Name, time.Now().Format("2006-01-02"))

	if shouldFire {
		// 触发告警
		alert := &Alert{
			ID:          alertID,
			Name:        rule.Name,
			Level:       rule.Level,
			Message:     fmt.Sprintf("Value %.2f %s threshold %.2f", value, rule.Operator, rule.Threshold),
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
			Value:       value,
			Threshold:   rule.Threshold,
		}

		return ae.manager.Fire(ctx, alert)
	} else {
		// 解决告警
		return ae.manager.Resolve(ctx, alertID)
	}
}
