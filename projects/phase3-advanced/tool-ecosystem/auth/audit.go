package auth

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AuditAction 审计动作
type AuditAction string

const (
	// 工具操作
	AuditActionToolExecute    AuditAction = "tool.execute"
	AuditActionToolRegister   AuditAction = "tool.register"
	AuditActionToolUnregister AuditAction = "tool.unregister"

	// 资源操作
	AuditActionResourceRead   AuditAction = "resource.read"
	AuditActionResourceWrite  AuditAction = "resource.write"
	AuditActionResourceCreate AuditAction = "resource.create"
	AuditActionResourceDelete AuditAction = "resource.delete"

	// 用户操作
	AuditActionUserCreate AuditAction = "user.create"
	AuditActionUserUpdate AuditAction = "user.update"
	AuditActionUserDelete AuditAction = "user.delete"
	AuditActionUserLogin  AuditAction = "user.login"
	AuditActionUserLogout AuditAction = "user.logout"

	// 角色操作
	AuditActionRoleCreate AuditAction = "role.create"
	AuditActionRoleUpdate AuditAction = "role.update"
	AuditActionRoleDelete AuditAction = "role.delete"
	AuditActionRoleAssign AuditAction = "role.assign"
	AuditActionRoleRevoke AuditAction = "role.revoke"

	// 权限操作
	AuditActionPermissionGrant  AuditAction = "permission.grant"
	AuditActionPermissionRevoke AuditAction = "permission.revoke"
)

// AuditResult 审计结果
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
	AuditResultDenied  AuditResult = "denied"
)

// AuditLog 审计日志
type AuditLog struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id"`
	Username   string                 `json:"username"`
	Action     AuditAction            `json:"action"`
	Resource   string                 `json:"resource"`   // 资源标识
	Result     AuditResult            `json:"result"`
	Details    string                 `json:"details"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Metadata   map[string]interface{} `json:"metadata"`
	Duration   time.Duration          `json:"duration"` // 操作耗时
}

// AuditLogger 审计日志记录器
type AuditLogger struct {
	logs     []*AuditLog
	maxLogs  int
	mu       sync.RWMutex
	handlers []AuditHandler
}

// AuditHandler 审计日志处理器接口
type AuditHandler interface {
	Handle(log *AuditLog) error
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(maxLogs int) *AuditLogger {
	if maxLogs <= 0 {
		maxLogs = 10000 // 默认最多保存10000条
	}

	return &AuditLogger{
		logs:     make([]*AuditLog, 0),
		maxLogs:  maxLogs,
		handlers: make([]AuditHandler, 0),
	}
}

// Log 记录审计日志
func (al *AuditLogger) Log(log *AuditLog) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	// 生成ID
	if log.ID == "" {
		log.ID = fmt.Sprintf("audit-%d-%d", time.Now().Unix(), len(al.logs)+1)
	}

	// 设置时间戳
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	// 初始化metadata
	if log.Metadata == nil {
		log.Metadata = make(map[string]interface{})
	}

	// 添加到日志列表
	al.logs = append(al.logs, log)

	// 如果超过最大数量，删除最旧的
	if len(al.logs) > al.maxLogs {
		al.logs = al.logs[len(al.logs)-al.maxLogs:]
	}

	// 调用所有处理器
	for _, handler := range al.handlers {
		go func(h AuditHandler) {
			if err := h.Handle(log); err != nil {
				// 记录错误但不影响主流程
				fmt.Printf("Audit handler error: %v\n", err)
			}
		}(handler)
	}

	return nil
}

// LogToolExecution 记录工具执行
func (al *AuditLogger) LogToolExecution(userID, username, toolID string, result AuditResult, details string, duration time.Duration) error {
	return al.Log(&AuditLog{
		UserID:   userID,
		Username: username,
		Action:   AuditActionToolExecute,
		Resource: toolID,
		Result:   result,
		Details:  details,
		Duration: duration,
	})
}

// LogResourceAccess 记录资源访问
func (al *AuditLogger) LogResourceAccess(userID, username, resourceID string, action AuditAction, result AuditResult, details string) error {
	return al.Log(&AuditLog{
		UserID:   userID,
		Username: username,
		Action:   action,
		Resource: resourceID,
		Result:   result,
		Details:  details,
	})
}

// LogUserAction 记录用户操作
func (al *AuditLogger) LogUserAction(userID, username string, action AuditAction, targetUser string, result AuditResult, details string) error {
	return al.Log(&AuditLog{
		UserID:   userID,
		Username: username,
		Action:   action,
		Resource: targetUser,
		Result:   result,
		Details:  details,
	})
}

// LogRoleAction 记录角色操作
func (al *AuditLogger) LogRoleAction(userID, username, roleID string, action AuditAction, result AuditResult, details string) error {
	return al.Log(&AuditLog{
		UserID:   userID,
		Username: username,
		Action:   action,
		Resource: roleID,
		Result:   result,
		Details:  details,
	})
}

// GetLogs 获取所有日志
func (al *AuditLogger) GetLogs() []*AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	// 返回副本
	logs := make([]*AuditLog, len(al.logs))
	copy(logs, al.logs)

	return logs
}

// GetLogsByUser 按用户获取日志
func (al *AuditLogger) GetLogsByUser(userID string) []*AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	logs := make([]*AuditLog, 0)
	for _, log := range al.logs {
		if log.UserID == userID {
			logs = append(logs, log)
		}
	}

	return logs
}

// GetLogsByAction 按动作获取日志
func (al *AuditLogger) GetLogsByAction(action AuditAction) []*AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	logs := make([]*AuditLog, 0)
	for _, log := range al.logs {
		if log.Action == action {
			logs = append(logs, log)
		}
	}

	return logs
}

// GetLogsByResult 按结果获取日志
func (al *AuditLogger) GetLogsByResult(result AuditResult) []*AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	logs := make([]*AuditLog, 0)
	for _, log := range al.logs {
		if log.Result == result {
			logs = append(logs, log)
		}
	}

	return logs
}

// GetLogsByTimeRange 按时间范围获取日志
func (al *AuditLogger) GetLogsByTimeRange(start, end time.Time) []*AuditLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	logs := make([]*AuditLog, 0)
	for _, log := range al.logs {
		if log.Timestamp.After(start) && log.Timestamp.Before(end) {
			logs = append(logs, log)
		}
	}

	return logs
}

// GetLogCount 获取日志总数
func (al *AuditLogger) GetLogCount() int {
	al.mu.RLock()
	defer al.mu.RUnlock()

	return len(al.logs)
}

// ClearLogs 清空日志（慎用）
func (al *AuditLogger) ClearLogs() {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.logs = make([]*AuditLog, 0)
}

// AddHandler 添加审计日志处理器
func (al *AuditLogger) AddHandler(handler AuditHandler) {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.handlers = append(al.handlers, handler)
}

// ConsoleAuditHandler 控制台审计日志处理器
type ConsoleAuditHandler struct{}

// Handle 处理审计日志
func (h *ConsoleAuditHandler) Handle(log *AuditLog) error {
	fmt.Printf("[AUDIT] %s | User: %s | Action: %s | Resource: %s | Result: %s | Details: %s\n",
		log.Timestamp.Format(time.RFC3339),
		log.Username,
		log.Action,
		log.Resource,
		log.Result,
		log.Details,
	)
	return nil
}

// FileAuditHandler 文件审计日志处理器
type FileAuditHandler struct {
	filepath string
	mu       sync.Mutex
}

// NewFileAuditHandler 创建文件审计日志处理器
func NewFileAuditHandler(filepath string) *FileAuditHandler {
	return &FileAuditHandler{
		filepath: filepath,
	}
}

// Handle 处理审计日志
func (h *FileAuditHandler) Handle(log *AuditLog) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 将日志序列化为JSON
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	// 这里简化处理，实际应该使用文件追加写入
	fmt.Printf("[FILE] Writing audit log to %s: %s\n", h.filepath, string(data))

	return nil
}

// AuditStatistics 审计统计
type AuditStatistics struct {
	TotalLogs      int                     `json:"total_logs"`
	SuccessCount   int                     `json:"success_count"`
	FailureCount   int                     `json:"failure_count"`
	DeniedCount    int                     `json:"denied_count"`
	ActionCounts   map[AuditAction]int     `json:"action_counts"`
	UserCounts     map[string]int          `json:"user_counts"`
	ResourceCounts map[string]int          `json:"resource_counts"`
	StartTime      time.Time               `json:"start_time"`
	EndTime        time.Time               `json:"end_time"`
}

// GetStatistics 获取审计统计信息
func (al *AuditLogger) GetStatistics() *AuditStatistics {
	al.mu.RLock()
	defer al.mu.RUnlock()

	stats := &AuditStatistics{
		TotalLogs:      len(al.logs),
		ActionCounts:   make(map[AuditAction]int),
		UserCounts:     make(map[string]int),
		ResourceCounts: make(map[string]int),
	}

	if len(al.logs) == 0 {
		return stats
	}

	stats.StartTime = al.logs[0].Timestamp
	stats.EndTime = al.logs[len(al.logs)-1].Timestamp

	for _, log := range al.logs {
		// 统计结果
		switch log.Result {
		case AuditResultSuccess:
			stats.SuccessCount++
		case AuditResultFailure:
			stats.FailureCount++
		case AuditResultDenied:
			stats.DeniedCount++
		}

		// 统计动作
		stats.ActionCounts[log.Action]++

		// 统计用户
		stats.UserCounts[log.UserID]++

		// 统计资源
		if log.Resource != "" {
			stats.ResourceCounts[log.Resource]++
		}
	}

	return stats
}
