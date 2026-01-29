package communication

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AckStatus 确认状态
type AckStatus string

const (
	AckStatusPending   AckStatus = "PENDING"
	AckStatusConfirmed AckStatus = "CONFIRMED"
	AckStatusTimeout   AckStatus = "TIMEOUT"
	AckStatusFailed    AckStatus = "FAILED"
)

// Acknowledgment 消息确认
type Acknowledgment struct {
	MessageID string
	Status    AckStatus
	Timestamp time.Time
	Error     string
}

// AckManager 确认管理器
type AckManager struct {
	acks    map[string]*Acknowledgment // messageID -> Ack
	waiters map[string]chan *Acknowledgment
	timeout time.Duration
	mu      sync.RWMutex
}

// NewAckManager 创建确认管理器
func NewAckManager(timeout time.Duration) *AckManager {
	return &AckManager{
		acks:    make(map[string]*Acknowledgment),
		waiters: make(map[string]chan *Acknowledgment),
		timeout: timeout,
	}
}

// RegisterMessage 注册等待确认的消息
func (m *AckManager) RegisterMessage(messageID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.acks[messageID] = &Acknowledgment{
		MessageID: messageID,
		Status:    AckStatusPending,
		Timestamp: time.Now(),
	}
	m.waiters[messageID] = make(chan *Acknowledgment, 1)
}

// Confirm 确认消息
func (m *AckManager) Confirm(messageID string, success bool, errorMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ack, exists := m.acks[messageID]
	if !exists {
		return fmt.Errorf("message %s not found", messageID)
	}

	if success {
		ack.Status = AckStatusConfirmed
	} else {
		ack.Status = AckStatusFailed
		ack.Error = errorMsg
	}
	ack.Timestamp = time.Now()

	// 通知等待者
	if waiter, ok := m.waiters[messageID]; ok {
		select {
		case waiter <- ack:
		default:
		}
		close(waiter)
		delete(m.waiters, messageID)
	}

	return nil
}

// WaitForAck 等待消息确认
func (m *AckManager) WaitForAck(messageID string) (*Acknowledgment, error) {
	m.mu.RLock()
	waiter, exists := m.waiters[messageID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("message %s not registered", messageID)
	}

	// 等待确认或超时
	timer := time.NewTimer(m.timeout)
	defer timer.Stop()

	select {
	case ack := <-waiter:
		return ack, nil
	case <-timer.C:
		m.mu.Lock()
		if ack, ok := m.acks[messageID]; ok {
			ack.Status = AckStatusTimeout
			ack.Timestamp = time.Now()
		}
		m.mu.Unlock()
		return nil, fmt.Errorf("ack timeout for message %s", messageID)
	}
}

// GetAck 获取确认信息
func (m *AckManager) GetAck(messageID string) (*Acknowledgment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ack, exists := m.acks[messageID]
	if !exists {
		return nil, fmt.Errorf("message %s not found", messageID)
	}

	return ack, nil
}

// RemoveAck 移除确认信息
func (m *AckManager) RemoveAck(messageID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.acks, messageID)
	if waiter, ok := m.waiters[messageID]; ok {
		close(waiter)
		delete(m.waiters, messageID)
	}
}

// CleanupExpired 清理过期的确认
func (m *AckManager) CleanupExpired(expireAfter time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	count := 0

	for messageID, ack := range m.acks {
		if now.Sub(ack.Timestamp) > expireAfter {
			delete(m.acks, messageID)
			if waiter, ok := m.waiters[messageID]; ok {
				close(waiter)
				delete(m.waiters, messageID)
			}
			count++
		}
	}

	return count
}

// GetPendingCount 获取待确认消息数
func (m *AckManager) GetPendingCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, ack := range m.acks {
		if ack.Status == AckStatusPending {
			count++
		}
	}

	return count
}

// GetAckStats 获取确认统计
func (m *AckManager) GetAckStats() map[AckStatus]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[AckStatus]int)
	for _, ack := range m.acks {
		stats[ack.Status]++
	}

	return stats
}

// SerializeMessage 序列化消息
func SerializeMessage(msg *Message) ([]byte, error) {
	return json.Marshal(msg)
}

// DeserializeMessage 反序列化消息
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &msg, nil
}

// MessageBuilder 消息构建器
type MessageBuilder struct {
	msg *Message
}

// NewMessageBuilder 创建消息构建器
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		msg: &Message{
			Metadata: make(map[string]interface{}),
			Payload:  make(map[string]interface{}),
		},
	}
}

// SetMessageID 设置消息ID
func (b *MessageBuilder) SetMessageID(id string) *MessageBuilder {
	b.msg.MessageID = id
	return b
}

// SetType 设置消息类型
func (b *MessageBuilder) SetType(msgType string) *MessageBuilder {
	b.msg.Type = msgType
	return b
}

// SetFrom 设置发送者
func (b *MessageBuilder) SetFrom(from string) *MessageBuilder {
	b.msg.From = from
	return b
}

// SetTo 设置接收者
func (b *MessageBuilder) SetTo(to string) *MessageBuilder {
	b.msg.To = to
	return b
}

// SetTimestamp 设置时间戳
func (b *MessageBuilder) SetTimestamp(timestamp string) *MessageBuilder {
	b.msg.Timestamp = timestamp
	return b
}

// SetPriority 设置优先级
func (b *MessageBuilder) SetPriority(priority int) *MessageBuilder {
	b.msg.Priority = priority
	return b
}

// SetPayload 设置负载
func (b *MessageBuilder) SetPayload(payload map[string]interface{}) *MessageBuilder {
	b.msg.Payload = payload
	return b
}

// AddPayloadField 添加负载字段
func (b *MessageBuilder) AddPayloadField(key string, value interface{}) *MessageBuilder {
	b.msg.Payload[key] = value
	return b
}

// SetMetadata 设置元数据
func (b *MessageBuilder) SetMetadata(metadata map[string]interface{}) *MessageBuilder {
	b.msg.Metadata = metadata
	return b
}

// AddMetadataField 添加元数据字段
func (b *MessageBuilder) AddMetadataField(key string, value interface{}) *MessageBuilder {
	b.msg.Metadata[key] = value
	return b
}

// Build 构建消息
func (b *MessageBuilder) Build() *Message {
	// 如果没有设置时间戳，自动设置
	if b.msg.Timestamp == "" {
		b.msg.Timestamp = time.Now().Format(time.RFC3339)
	}

	return b.msg
}

// MessageValidator 消息验证器
type MessageValidator struct{}

// NewMessageValidator 创建消息验证器
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{}
}

// Validate 验证消息
func (v *MessageValidator) Validate(msg *Message) error {
	if msg.MessageID == "" {
		return fmt.Errorf("message_id is required")
	}

	if msg.Type == "" {
		return fmt.Errorf("type is required")
	}

	if msg.From == "" {
		return fmt.Errorf("from is required")
	}

	if msg.To == "" {
		return fmt.Errorf("to is required")
	}

	if msg.Timestamp == "" {
		return fmt.Errorf("timestamp is required")
	}

	// 验证时间戳格式
	if _, err := time.Parse(time.RFC3339, msg.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	return nil
}

// ValidatePayload 验证负载
func (v *MessageValidator) ValidatePayload(msg *Message, requiredFields []string) error {
	for _, field := range requiredFields {
		if _, exists := msg.Payload[field]; !exists {
			return fmt.Errorf("required payload field missing: %s", field)
		}
	}

	return nil
}
