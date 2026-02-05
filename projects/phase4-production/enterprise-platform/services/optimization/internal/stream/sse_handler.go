package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

// SSEHandler Server-Sent Events处理器
type SSEHandler struct {
	clients   map[string]chan model.StreamEvent
	register  chan *SSEClient
	unregister chan *SSEClient
}

// SSEClient SSE客户端
type SSEClient struct {
	ID       string
	TenantID string
	UserID   string
	Channel  chan model.StreamEvent
}

// NewSSEHandler 创建SSE处理器
func NewSSEHandler() *SSEHandler {
	handler := &SSEHandler{
		clients:    make(map[string]chan model.StreamEvent),
		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),
	}

	go handler.run()

	return handler
}

// run 运行SSE处理器
func (h *SSEHandler) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.ID] = client.Channel

		case client := <-h.unregister:
			if _, ok := h.clients[client.ID]; ok {
				close(client.Channel)
				delete(h.clients, client.ID)
			}
		}
	}
}

// ServeHTTP HTTP处理
func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取客户端信息
	clientID := r.URL.Query().Get("client_id")
	tenantID := r.URL.Query().Get("tenant_id")
	userID := r.URL.Query().Get("user_id")

	if clientID == "" {
		clientID = fmt.Sprintf("client-%d", time.Now().UnixNano())
	}

	// 创建客户端
	client := &SSEClient{
		ID:       clientID,
		TenantID: tenantID,
		UserID:   userID,
		Channel:  make(chan model.StreamEvent, 10),
	}

	// 注册客户端
	h.register <- client

	// 获取Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 发送连接成功事件
	connectEvent := model.StreamEvent{
		ID:        clientID,
		Type:      "connected",
		Data:      map[string]interface{}{"message": "Connection established"},
		Timestamp: time.Now(),
	}
	h.sendEvent(w, flusher, connectEvent)

	// 监听客户端断开和事件
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			// 客户端断开
			h.unregister <- client
			return

		case event, ok := <-client.Channel:
			if !ok {
				return
			}

			// 发送事件
			h.sendEvent(w, flusher, event)
		}
	}
}

// sendEvent 发送事件
func (h *SSEHandler) sendEvent(w http.ResponseWriter, flusher http.Flusher, event model.StreamEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "id: %s\n", event.ID)
	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", string(data))

	flusher.Flush()
}

// SendToClient 发送事件到指定客户端
func (h *SSEHandler) SendToClient(clientID string, event model.StreamEvent) error {
	channel, ok := h.clients[clientID]
	if !ok {
		return fmt.Errorf("client not found: %s", clientID)
	}

	select {
	case channel <- event:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending to client: %s", clientID)
	}
}

// Broadcast 广播事件到所有客户端
func (h *SSEHandler) Broadcast(event model.StreamEvent) {
	for _, channel := range h.clients {
		select {
		case channel <- event:
		default:
			// 如果通道满了，跳过
		}
	}
}

// BroadcastToTenant 广播事件到租户的所有客户端
func (h *SSEHandler) BroadcastToTenant(tenantID string, event model.StreamEvent) {
	// TODO: 需要在客户端注册时记录租户ID
	// 这里简化实现，实际应该维护租户ID到客户端的映射
	h.Broadcast(event)
}

// StreamExecutor 流式执行器
type StreamExecutor struct {
	sseHandler *SSEHandler
}

// NewStreamExecutor 创建流式执行器
func NewStreamExecutor(sseHandler *SSEHandler) *StreamExecutor {
	return &StreamExecutor{
		sseHandler: sseHandler,
	}
}

// ExecuteWithStream 流式执行任务
func (e *StreamExecutor) ExecuteWithStream(ctx context.Context, clientID string, taskFunc func(context.Context, chan<- model.StreamEvent) error) error {
	// 创建事件通道
	eventChan := make(chan model.StreamEvent, 10)

	// 启动转发goroutine
	go func() {
		for event := range eventChan {
			e.sseHandler.SendToClient(clientID, event)
		}
	}()

	// 发送开始事件
	eventChan <- model.StreamEvent{
		ID:        clientID,
		Type:      "start",
		Data:      map[string]interface{}{"status": "started"},
		Timestamp: time.Now(),
	}

	// 执行任务
	err := taskFunc(ctx, eventChan)

	// 发送结束或错误事件
	if err != nil {
		eventChan <- model.StreamEvent{
			ID:        clientID,
			Type:      "error",
			Data:      map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		}
	} else {
		eventChan <- model.StreamEvent{
			ID:        clientID,
			Type:      "end",
			Data:      map[string]interface{}{"status": "completed"},
			Timestamp: time.Now(),
		}
	}

	close(eventChan)

	return err
}

// SendProgress 发送进度更新
func (e *StreamExecutor) SendProgress(clientID string, progress int, message string) error {
	event := model.StreamEvent{
		ID:   clientID,
		Type: "progress",
		Data: map[string]interface{}{
			"progress": progress,
			"message":  message,
		},
		Timestamp: time.Now(),
	}

	return e.sseHandler.SendToClient(clientID, event)
}

// SendChunk 发送数据块
func (e *StreamExecutor) SendChunk(clientID string, chunk map[string]interface{}) error {
	event := model.StreamEvent{
		ID:        clientID,
		Type:      "chunk",
		Data:      chunk,
		Timestamp: time.Now(),
	}

	return e.sseHandler.SendToClient(clientID, event)
}
