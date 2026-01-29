package communication

import (
	"fmt"
	"sync"
)

// MessageHandler 消息处理器
type MessageHandler func(msg *Message) error

// MessageRouter 消息路由器
type MessageRouter struct {
	handlers map[string]MessageHandler // messageType -> handler
	mu       sync.RWMutex
}

// NewMessageRouter 创建消息路由器
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (r *MessageRouter) RegisterHandler(messageType string, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[messageType] = handler
}

// UnregisterHandler 注销消息处理器
func (r *MessageRouter) UnregisterHandler(messageType string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.handlers, messageType)
}

// Route 路由消息
func (r *MessageRouter) Route(msg *Message) error {
	r.mu.RLock()
	handler, exists := r.handlers[msg.Type]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler for message type: %s", msg.Type)
	}

	return handler(msg)
}

// HasHandler 检查是否有处理器
func (r *MessageRouter) HasHandler(messageType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.handlers[messageType]
	return exists
}

// GetHandlerCount 获取处理器数量
func (r *MessageRouter) GetHandlerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.handlers)
}

// Message 消息定义（简化版，与protocol包兼容）
type Message struct {
	MessageID string                 `json:"message_id"`
	Type      string                 `json:"type"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Timestamp string                 `json:"timestamp"`
	Priority  int                    `json:"priority,omitempty"`
	Payload   map[string]interface{} `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MessageQueue 消息队列（用于异步处理）
type MessageQueue struct {
	messages chan *Message
	size     int
	mu       sync.RWMutex
}

// NewMessageQueue 创建消息队列
func NewMessageQueue(size int) *MessageQueue {
	return &MessageQueue{
		messages: make(chan *Message, size),
		size:     size,
	}
}

// Enqueue 入队消息
func (q *MessageQueue) Enqueue(msg *Message) error {
	select {
	case q.messages <- msg:
		return nil
	default:
		return fmt.Errorf("message queue is full")
	}
}

// Dequeue 出队消息
func (q *MessageQueue) Dequeue() (*Message, error) {
	select {
	case msg := <-q.messages:
		return msg, nil
	default:
		return nil, fmt.Errorf("message queue is empty")
	}
}

// DequeueWait 阻塞等待出队
func (q *MessageQueue) DequeueWait() *Message {
	return <-q.messages
}

// Size 获取队列大小
func (q *MessageQueue) Size() int {
	return len(q.messages)
}

// IsEmpty 检查队列是否为空
func (q *MessageQueue) IsEmpty() bool {
	return len(q.messages) == 0
}

// IsFull 检查队列是否已满
func (q *MessageQueue) IsFull() bool {
	return len(q.messages) >= q.size
}

// MessageDispatcher 消息分发器
type MessageDispatcher struct {
	router     *MessageRouter
	connMgr    *ConnectionManager
	inQueue    *MessageQueue // 接收队列
	outQueue   *MessageQueue // 发送队列
	workerPool int
	mu         sync.RWMutex
}

// NewMessageDispatcher 创建消息分发器
func NewMessageDispatcher(router *MessageRouter, connMgr *ConnectionManager, queueSize, workerPool int) *MessageDispatcher {
	return &MessageDispatcher{
		router:     router,
		connMgr:    connMgr,
		inQueue:    NewMessageQueue(queueSize),
		outQueue:   NewMessageQueue(queueSize),
		workerPool: workerPool,
	}
}

// EnqueueIncoming 接收消息入队
func (d *MessageDispatcher) EnqueueIncoming(msg *Message) error {
	return d.inQueue.Enqueue(msg)
}

// EnqueueOutgoing 发送消息入队
func (d *MessageDispatcher) EnqueueOutgoing(msg *Message) error {
	return d.outQueue.Enqueue(msg)
}

// DispatchIncoming 分发接收的消息
func (d *MessageDispatcher) DispatchIncoming(msg *Message) error {
	// 路由到对应的处理器
	return d.router.Route(msg)
}

// DispatchOutgoing 分发发送的消息
func (d *MessageDispatcher) DispatchOutgoing(msg *Message) error {
	// 根据目标发送消息
	if msg.To == "broadcast" {
		// 广播消息
		return d.BroadcastMessage(msg)
	}

	// 单播消息
	return d.SendToAgent(msg.To, msg)
}

// SendToAgent 发送消息给指定Agent
func (d *MessageDispatcher) SendToAgent(agentID string, msg *Message) error {
	conn, err := d.connMgr.GetConnectionByAgent(agentID)
	if err != nil {
		return fmt.Errorf("failed to get connection for agent %s: %w", agentID, err)
	}

	// 序列化消息
	data, err := SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	return conn.Send(data)
}

// BroadcastMessage 广播消息
func (d *MessageDispatcher) BroadcastMessage(msg *Message) error {
	// 序列化消息
	data, err := SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	return d.connMgr.BroadcastToAll(data)
}

// SendToAgents 发送消息给多个Agent
func (d *MessageDispatcher) SendToAgents(agentIDs []string, msg *Message) error {
	// 序列化消息
	data, err := SerializeMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	return d.connMgr.BroadcastToAgents(agentIDs, data)
}

// GetInQueueSize 获取接收队列大小
func (d *MessageDispatcher) GetInQueueSize() int {
	return d.inQueue.Size()
}

// GetOutQueueSize 获取发送队列大小
func (d *MessageDispatcher) GetOutQueueSize() int {
	return d.outQueue.Size()
}
