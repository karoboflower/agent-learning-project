package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Host              string
	Port              int
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	PingInterval      time.Duration
	PongTimeout       time.Duration
	MessageQueueSize  int
	WorkerPoolSize    int
}

// DefaultWebSocketConfig 默认配置
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		Host:             "0.0.0.0",
		Port:             8080,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second,
		ReadTimeout:      60 * time.Second,
		WriteTimeout:     10 * time.Second,
		PingInterval:     30 * time.Second,
		PongTimeout:      60 * time.Second,
		MessageQueueSize: 1000,
		WorkerPoolSize:   10,
	}
}

// WebSocketServer WebSocket服务器
type WebSocketServer struct {
	config     *WebSocketConfig
	connMgr    *ConnectionManager
	router     *MessageRouter
	dispatcher *MessageDispatcher
	upgrader   websocket.Upgrader
	server     *http.Server

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.RWMutex
}

// NewWebSocketServer 创建WebSocket服务器
func NewWebSocketServer(config *WebSocketConfig) *WebSocketServer {
	if config == nil {
		config = DefaultWebSocketConfig()
	}

	connMgr := NewConnectionManager()
	router := NewMessageRouter()
	dispatcher := NewMessageDispatcher(router, connMgr, config.MessageQueueSize, config.WorkerPoolSize)

	ctx, cancel := context.WithCancel(context.Background())

	return &WebSocketServer{
		config:     config,
		connMgr:    connMgr,
		router:     router,
		dispatcher: dispatcher,
		upgrader: websocket.Upgrader{
			ReadBufferSize:   config.ReadBufferSize,
			WriteBufferSize:  config.WriteBufferSize,
			HandshakeTimeout: config.HandshakeTimeout,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动服务器
func (s *WebSocketServer) Start() error {
	// 启动消息处理worker
	for i := 0; i < s.config.WorkerPoolSize; i++ {
		s.wg.Add(1)
		go s.incomingMessageWorker(i)

		s.wg.Add(1)
		go s.outgoingMessageWorker(i)
	}

	// 启动心跳检查
	s.wg.Add(1)
	go s.heartbeatChecker()

	// 设置HTTP路由
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("WebSocket server starting on %s", addr)

	// 启动服务器（非阻塞）
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *WebSocketServer) Stop() error {
	s.cancel()

	// 关闭所有连接
	for _, conn := range s.connMgr.ListConnections() {
		conn.Close()
	}

	// 关闭HTTP服务器
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}

	s.wg.Wait()

	return nil
}

// handleWebSocket WebSocket连接处理
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级连接
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 获取AgentID（从查询参数）
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		log.Printf("Missing agent_id parameter")
		wsConn.Close()
		return
	}

	// 创建连接
	connID := uuid.New().String()
	conn := NewConnection(connID, agentID, wsConn)

	// 添加到连接管理器
	if err := s.connMgr.AddConnection(conn); err != nil {
		log.Printf("Failed to add connection: %v", err)
		wsConn.Close()
		return
	}

	log.Printf("Agent %s connected (connection: %s)", agentID, connID)

	// 启动读写协程
	s.wg.Add(2)
	go s.readPump(conn)
	go s.writePump(conn)
}

// handleHealth 健康检查
func (s *WebSocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"status":            "healthy",
		"connections":       s.connMgr.GetConnectionCount(),
		"active_connections": len(s.connMgr.GetActiveConnections()),
		"in_queue_size":     s.dispatcher.GetInQueueSize(),
		"out_queue_size":    s.dispatcher.GetOutQueueSize(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// readPump 读取消息
func (s *WebSocketServer) readPump(conn *Connection) {
	defer func() {
		s.wg.Done()
		s.connMgr.RemoveConnection(conn.ID)
		log.Printf("Agent %s disconnected (connection: %s)", conn.AgentID, conn.ID)
	}()

	// 设置读超时
	conn.Conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
		conn.UpdateHeartbeat()
		return nil
	})

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// 读取消息
		_, data, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for agent %s: %v", conn.AgentID, err)
			}
			return
		}

		// 反序列化消息
		msg, err := DeserializeMessage(data)
		if err != nil {
			log.Printf("Failed to deserialize message from agent %s: %v", conn.AgentID, err)
			continue
		}

		// 更新心跳
		conn.UpdateHeartbeat()

		// 入队处理
		if err := s.dispatcher.EnqueueIncoming(msg); err != nil {
			log.Printf("Failed to enqueue incoming message: %v", err)
		}
	}
}

// writePump 发送消息
func (s *WebSocketServer) writePump(conn *Connection) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return

		case data, ok := <-conn.SendChan:
			// 设置写超时
			conn.Conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))

			if !ok {
				// 通道已关闭
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 发送消息
			if err := conn.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to write message to agent %s: %v", conn.AgentID, err)
				return
			}

		case <-ticker.C:
			// 发送ping
			conn.Conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping to agent %s: %v", conn.AgentID, err)
				return
			}
		}
	}
}

// incomingMessageWorker 处理接收的消息
func (s *WebSocketServer) incomingMessageWorker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msg := s.dispatcher.inQueue.DequeueWait()
			if msg == nil {
				continue
			}

			// 分发消息
			if err := s.dispatcher.DispatchIncoming(msg); err != nil {
				log.Printf("Worker %d: failed to dispatch incoming message: %v", id, err)
			}
		}
	}
}

// outgoingMessageWorker 处理发送的消息
func (s *WebSocketServer) outgoingMessageWorker(id int) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msg := s.dispatcher.outQueue.DequeueWait()
			if msg == nil {
				continue
			}

			// 分发消息
			if err := s.dispatcher.DispatchOutgoing(msg); err != nil {
				log.Printf("Worker %d: failed to dispatch outgoing message: %v", id, err)
			}
		}
	}
}

// heartbeatChecker 心跳检查
func (s *WebSocketServer) heartbeatChecker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			timeoutConns := s.connMgr.CheckHeartbeat(s.config.PongTimeout)
			for _, connID := range timeoutConns {
				log.Printf("Connection %s heartbeat timeout", connID)
				s.connMgr.RemoveConnection(connID)
			}
		}
	}
}

// RegisterMessageHandler 注册消息处理器
func (s *WebSocketServer) RegisterMessageHandler(messageType string, handler MessageHandler) {
	s.router.RegisterHandler(messageType, handler)
}

// SendMessage 发送消息
func (s *WebSocketServer) SendMessage(msg *Message) error {
	return s.dispatcher.EnqueueOutgoing(msg)
}

// BroadcastMessage 广播消息
func (s *WebSocketServer) BroadcastMessage(msg *Message) error {
	msg.To = "broadcast"
	return s.dispatcher.EnqueueOutgoing(msg)
}

// GetConnectionManager 获取连接管理器
func (s *WebSocketServer) GetConnectionManager() *ConnectionManager {
	return s.connMgr
}

// GetRouter 获取路由器
func (s *WebSocketServer) GetRouter() *MessageRouter {
	return s.router
}

// GetDispatcher 获取分发器
func (s *WebSocketServer) GetDispatcher() *MessageDispatcher {
	return s.dispatcher
}
