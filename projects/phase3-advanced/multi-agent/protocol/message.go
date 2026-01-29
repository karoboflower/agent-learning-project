package protocol

import (
	"time"
)

// MessageType 定义消息类型
type MessageType string

const (
	// 任务相关消息
	MessageTypeTaskRequest  MessageType = "TASK_REQUEST"
	MessageTypeTaskAccept   MessageType = "TASK_ACCEPT"
	MessageTypeTaskReject   MessageType = "TASK_REJECT"
	MessageTypeTaskComplete MessageType = "TASK_COMPLETE"
	MessageTypeTaskFailed   MessageType = "TASK_FAILED"

	// 状态相关消息
	MessageTypeHeartbeat      MessageType = "HEARTBEAT"
	MessageTypeStatusQuery    MessageType = "STATUS_QUERY"
	MessageTypeStatusResponse MessageType = "STATUS_RESPONSE"

	// 通用消息
	MessageTypeBroadcast MessageType = "BROADCAST"
	MessageTypeError     MessageType = "ERROR"
)

// AgentStatus 定义Agent状态
type AgentStatus string

const (
	AgentStatusActive      AgentStatus = "ACTIVE"
	AgentStatusIdle        AgentStatus = "IDLE"
	AgentStatusBusy        AgentStatus = "BUSY"
	AgentStatusMaintenance AgentStatus = "MAINTENANCE"
	AgentStatusError       AgentStatus = "ERROR"
)

// TaskStatus 定义任务状态
type TaskStatus string

const (
	TaskStatusSuccess TaskStatus = "SUCCESS"
	TaskStatusFailed  TaskStatus = "FAILED"
	TaskStatusPartial TaskStatus = "PARTIAL"
	TaskStatusTimeout TaskStatus = "TIMEOUT"
)

// RejectReason 定义拒绝原因
type RejectReason string

const (
	RejectReasonCapabilityMismatch   RejectReason = "CAPABILITY_MISMATCH"
	RejectReasonResourceUnavailable  RejectReason = "RESOURCE_UNAVAILABLE"
	RejectReasonOverloaded           RejectReason = "OVERLOADED"
	RejectReasonMaintenance          RejectReason = "MAINTENANCE"
	RejectReasonInvalidRequest       RejectReason = "INVALID_REQUEST"
)

// ErrorType 定义错误类型
type ErrorType string

const (
	ErrorTypeProtocol   ErrorType = "PROTOCOL_ERROR"
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"
	ErrorTypeExecution  ErrorType = "EXECUTION_ERROR"
	ErrorTypeTimeout    ErrorType = "TIMEOUT_ERROR"
	ErrorTypeResource   ErrorType = "RESOURCE_ERROR"
)

// ErrorSeverity 定义错误严重级别
type ErrorSeverity string

const (
	ErrorSeverityInfo     ErrorSeverity = "INFO"
	ErrorSeverityWarning  ErrorSeverity = "WARNING"
	ErrorSeverityError    ErrorSeverity = "ERROR"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// Message 定义基础消息结构
type Message struct {
	MessageID string                 `json:"message_id"`
	Type      MessageType            `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Timestamp string                 `json:"timestamp"`
	Priority  int                    `json:"priority,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// 安全相关字段
	Signature           string `json:"signature,omitempty"`
	Encrypted           bool   `json:"encrypted,omitempty"`
	EncryptionAlgorithm string `json:"encryption_algorithm,omitempty"`
}

// TaskRequestPayload 任务请求消息负载
type TaskRequestPayload struct {
	TaskID       string                 `json:"task_id"`
	TaskType     string                 `json:"task_type"`
	Input        interface{}            `json:"input"`
	Requirements map[string]interface{} `json:"requirements,omitempty"`
	Timeout      int                    `json:"timeout,omitempty"`
	CallbackURL  string                 `json:"callback_url,omitempty"`
}

// TaskAcceptPayload 任务接受消息负载
type TaskAcceptPayload struct {
	TaskID            string `json:"task_id"`
	EstimatedDuration int    `json:"estimated_duration"`
	AcceptedAt        string `json:"accepted_at"`
}

// TaskRejectPayload 任务拒绝消息负载
type TaskRejectPayload struct {
	TaskID          string       `json:"task_id"`
	Reason          RejectReason `json:"reason"`
	Message         string       `json:"message"`
	SuggestedAgents []string     `json:"suggested_agents,omitempty"`
}

// TaskCompletePayload 任务完成消息负载
type TaskCompletePayload struct {
	TaskID      string                 `json:"task_id"`
	Status      TaskStatus             `json:"status"`
	Output      map[string]interface{} `json:"output"`
	Duration    int64                  `json:"duration"`
	CompletedAt string                 `json:"completed_at"`
}

// TaskFailedPayload 任务失败消息负载
type TaskFailedPayload struct {
	TaskID        string                 `json:"task_id"`
	ErrorCode     string                 `json:"error_code"`
	ErrorMessage  string                 `json:"error_message"`
	ErrorDetails  map[string]interface{} `json:"error_details,omitempty"`
	RetryPossible bool                   `json:"retry_possible"`
}

// HeartbeatPayload 心跳消息负载
type HeartbeatPayload struct {
	Status       AgentStatus `json:"status"`
	Load         float64     `json:"load"`
	TasksRunning int         `json:"tasks_running"`
	TasksQueued  int         `json:"tasks_queued"`
	Capabilities []string    `json:"capabilities"`
}

// StatusQueryPayload 状态查询消息负载
type StatusQueryPayload struct {
	QueryType string `json:"query_type"`
	TaskID    string `json:"task_id,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
}

// StatusResponsePayload 状态响应消息负载
type StatusResponsePayload struct {
	QueryID            string `json:"query_id"`
	TaskID             string `json:"task_id,omitempty"`
	Status             string `json:"status"`
	Progress           int    `json:"progress,omitempty"`
	EstimatedRemaining int    `json:"estimated_remaining,omitempty"`
}

// BroadcastPayload 广播消息负载
type BroadcastPayload struct {
	Event     string `json:"event"`
	Message   string `json:"message"`
	Countdown int    `json:"countdown,omitempty"`
}

// ErrorPayload 错误消息负载
type ErrorPayload struct {
	ErrorType         ErrorType     `json:"error_type"`
	ErrorCode         string        `json:"error_code"`
	ErrorMessage      string        `json:"error_message"`
	OriginalMessageID string        `json:"original_message_id,omitempty"`
	Severity          ErrorSeverity `json:"severity"`
}

// NewMessage 创建新消息
func NewMessage(msgType MessageType, from, to string) *Message {
	return &Message{
		MessageID: generateMessageID(),
		Type:      msgType,
		From:      from,
		To:        to,
		Timestamp: time.Now().Format(time.RFC3339),
		Priority:  5, // 默认优先级
		Payload:   make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}
}

// SetPayload 设置消息负载
func (m *Message) SetPayload(payload interface{}) {
	// 将结构体转换为map
	// 实际实现中使用json序列化/反序列化
	m.Payload = interfaceToMap(payload)
}

// GetPayload 获取并解析消息负载
func (m *Message) GetPayload(target interface{}) error {
	// 将map转换为目标结构体
	// 实际实现中使用json序列化/反序列化
	return mapToInterface(m.Payload, target)
}

// SetMetadata 设置元数据
func (m *Message) SetMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata 获取元数据
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	value, exists := m.Metadata[key]
	return value, exists
}

// IsBroadcast 判断是否为广播消息
func (m *Message) IsBroadcast() bool {
	return m.To == "broadcast"
}

// IsHighPriority 判断是否为高优先级消息
func (m *Message) IsHighPriority() bool {
	return m.Priority >= 7
}

// 辅助函数
func generateMessageID() string {
	// 实际实现中使用uuid.New().String()
	return "msg-" + time.Now().Format("20060102150405")
}

func interfaceToMap(v interface{}) map[string]interface{} {
	// 简化实现，实际使用json.Marshal/Unmarshal
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}

func mapToInterface(m map[string]interface{}, target interface{}) error {
	// 简化实现，实际使用json.Marshal/Unmarshal
	return nil
}
