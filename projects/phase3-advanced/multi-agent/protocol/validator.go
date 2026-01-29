package protocol

import (
	"errors"
	"fmt"
	"time"
)

// Validator 消息验证器
type Validator struct {
	maxMessageSize int
	strictMode     bool
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		maxMessageSize: 1024 * 1024, // 1MB
		strictMode:     true,
	}
}

// Validate 验证消息
func (v *Validator) Validate(msg *Message) error {
	if err := v.validateBasicFields(msg); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	if err := v.validatePayload(msg); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	return nil
}

// validateBasicFields 验证基本字段
func (v *Validator) validateBasicFields(msg *Message) error {
	// 验证必需字段
	if msg.MessageID == "" {
		return errors.New("message_id is required")
	}

	if msg.Type == "" {
		return errors.New("type is required")
	}

	if msg.From == "" {
		return errors.New("from is required")
	}

	if msg.To == "" {
		return errors.New("to is required")
	}

	if msg.Timestamp == "" {
		return errors.New("timestamp is required")
	}

	// 验证时间戳格式
	if _, err := time.Parse(time.RFC3339, msg.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	// 验证优先级范围
	if msg.Priority < 0 || msg.Priority > 10 {
		return fmt.Errorf("priority must be between 0 and 10, got %d", msg.Priority)
	}

	// 验证消息类型
	if !v.isValidMessageType(msg.Type) {
		return fmt.Errorf("invalid message type: %s", msg.Type)
	}

	return nil
}

// validatePayload 验证负载
func (v *Validator) validatePayload(msg *Message) error {
	if msg.Payload == nil {
		return errors.New("payload is required")
	}

	switch msg.Type {
	case MessageTypeTaskRequest:
		return v.validateTaskRequestPayload(msg.Payload)
	case MessageTypeTaskAccept:
		return v.validateTaskAcceptPayload(msg.Payload)
	case MessageTypeTaskReject:
		return v.validateTaskRejectPayload(msg.Payload)
	case MessageTypeTaskComplete:
		return v.validateTaskCompletePayload(msg.Payload)
	case MessageTypeTaskFailed:
		return v.validateTaskFailedPayload(msg.Payload)
	case MessageTypeHeartbeat:
		return v.validateHeartbeatPayload(msg.Payload)
	case MessageTypeStatusQuery:
		return v.validateStatusQueryPayload(msg.Payload)
	case MessageTypeStatusResponse:
		return v.validateStatusResponsePayload(msg.Payload)
	case MessageTypeBroadcast:
		return v.validateBroadcastPayload(msg.Payload)
	case MessageTypeError:
		return v.validateErrorPayload(msg.Payload)
	default:
		if v.strictMode {
			return fmt.Errorf("unknown message type: %s", msg.Type)
		}
		return nil
	}
}

// validateTaskRequestPayload 验证任务请求负载
func (v *Validator) validateTaskRequestPayload(payload map[string]interface{}) error {
	if _, ok := payload["task_id"]; !ok {
		return errors.New("task_id is required")
	}

	if _, ok := payload["task_type"]; !ok {
		return errors.New("task_type is required")
	}

	if _, ok := payload["input"]; !ok {
		return errors.New("input is required")
	}

	// 验证超时值
	if timeout, ok := payload["timeout"].(int); ok {
		if timeout <= 0 {
			return fmt.Errorf("timeout must be positive, got %d", timeout)
		}
	}

	return nil
}

// validateTaskAcceptPayload 验证任务接受负载
func (v *Validator) validateTaskAcceptPayload(payload map[string]interface{}) error {
	if _, ok := payload["task_id"]; !ok {
		return errors.New("task_id is required")
	}

	if _, ok := payload["accepted_at"]; !ok {
		return errors.New("accepted_at is required")
	}

	return nil
}

// validateTaskRejectPayload 验证任务拒绝负载
func (v *Validator) validateTaskRejectPayload(payload map[string]interface{}) error {
	if _, ok := payload["task_id"]; !ok {
		return errors.New("task_id is required")
	}

	if _, ok := payload["reason"]; !ok {
		return errors.New("reason is required")
	}

	if _, ok := payload["message"]; !ok {
		return errors.New("message is required")
	}

	return nil
}

// validateTaskCompletePayload 验证任务完成负载
func (v *Validator) validateTaskCompletePayload(payload map[string]interface{}) error {
	if _, ok := payload["task_id"]; !ok {
		return errors.New("task_id is required")
	}

	if _, ok := payload["status"]; !ok {
		return errors.New("status is required")
	}

	if _, ok := payload["completed_at"]; !ok {
		return errors.New("completed_at is required")
	}

	return nil
}

// validateTaskFailedPayload 验证任务失败负载
func (v *Validator) validateTaskFailedPayload(payload map[string]interface{}) error {
	if _, ok := payload["task_id"]; !ok {
		return errors.New("task_id is required")
	}

	if _, ok := payload["error_code"]; !ok {
		return errors.New("error_code is required")
	}

	if _, ok := payload["error_message"]; !ok {
		return errors.New("error_message is required")
	}

	return nil
}

// validateHeartbeatPayload 验证心跳负载
func (v *Validator) validateHeartbeatPayload(payload map[string]interface{}) error {
	if _, ok := payload["status"]; !ok {
		return errors.New("status is required")
	}

	return nil
}

// validateStatusQueryPayload 验证状态查询负载
func (v *Validator) validateStatusQueryPayload(payload map[string]interface{}) error {
	if _, ok := payload["query_type"]; !ok {
		return errors.New("query_type is required")
	}

	return nil
}

// validateStatusResponsePayload 验证状态响应负载
func (v *Validator) validateStatusResponsePayload(payload map[string]interface{}) error {
	if _, ok := payload["query_id"]; !ok {
		return errors.New("query_id is required")
	}

	if _, ok := payload["status"]; !ok {
		return errors.New("status is required")
	}

	return nil
}

// validateBroadcastPayload 验证广播负载
func (v *Validator) validateBroadcastPayload(payload map[string]interface{}) error {
	if _, ok := payload["event"]; !ok {
		return errors.New("event is required")
	}

	if _, ok := payload["message"]; !ok {
		return errors.New("message is required")
	}

	return nil
}

// validateErrorPayload 验证错误负载
func (v *Validator) validateErrorPayload(payload map[string]interface{}) error {
	if _, ok := payload["error_type"]; !ok {
		return errors.New("error_type is required")
	}

	if _, ok := payload["error_code"]; !ok {
		return errors.New("error_code is required")
	}

	if _, ok := payload["error_message"]; !ok {
		return errors.New("error_message is required")
	}

	return nil
}

// isValidMessageType 检查消息类型是否有效
func (v *Validator) isValidMessageType(msgType MessageType) bool {
	validTypes := []MessageType{
		MessageTypeTaskRequest,
		MessageTypeTaskAccept,
		MessageTypeTaskReject,
		MessageTypeTaskComplete,
		MessageTypeTaskFailed,
		MessageTypeHeartbeat,
		MessageTypeStatusQuery,
		MessageTypeStatusResponse,
		MessageTypeBroadcast,
		MessageTypeError,
	}

	for _, t := range validTypes {
		if msgType == t {
			return true
		}
	}

	return false
}

// SetStrictMode 设置严格模式
func (v *Validator) SetStrictMode(strict bool) {
	v.strictMode = strict
}

// SetMaxMessageSize 设置最大消息大小
func (v *Validator) SetMaxMessageSize(size int) {
	v.maxMessageSize = size
}
