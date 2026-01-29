package communication

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ConnectionStatus 连接状态
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "CONNECTED"
	ConnectionStatusDisconnected ConnectionStatus = "DISCONNECTED"
	ConnectionStatusReconnecting ConnectionStatus = "RECONNECTING"
)

// Connection Agent连接
type Connection struct {
	ID            string
	AgentID       string
	Conn          *websocket.Conn
	Status        ConnectionStatus
	ConnectedAt   time.Time
	LastHeartbeat time.Time
	SendChan      chan []byte
	mu            sync.RWMutex
}

// NewConnection 创建新连接
func NewConnection(id, agentID string, conn *websocket.Conn) *Connection {
	return &Connection{
		ID:            id,
		AgentID:       agentID,
		Conn:          conn,
		Status:        ConnectionStatusConnected,
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
		SendChan:      make(chan []byte, 256),
	}
}

// Send 发送消息
func (c *Connection) Send(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.Status != ConnectionStatusConnected {
		return fmt.Errorf("connection is not active")
	}

	select {
	case c.SendChan <- data:
		return nil
	default:
		return fmt.Errorf("send buffer is full")
	}
}

// UpdateHeartbeat 更新心跳时间
func (c *Connection) UpdateHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.LastHeartbeat = time.Now()
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Status = ConnectionStatusDisconnected
	close(c.SendChan)

	if c.Conn != nil {
		return c.Conn.Close()
	}

	return nil
}

// IsAlive 检查连接是否存活
func (c *Connection) IsAlive(timeout time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return time.Since(c.LastHeartbeat) < timeout
}

// ConnectionManager 连接管理器
type ConnectionManager struct {
	connections map[string]*Connection // connectionID -> Connection
	agentConns  map[string]*Connection // agentID -> Connection
	mu          sync.RWMutex
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
		agentConns:  make(map[string]*Connection),
	}
}

// AddConnection 添加连接
func (m *ConnectionManager) AddConnection(conn *Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[conn.ID]; exists {
		return fmt.Errorf("connection %s already exists", conn.ID)
	}

	m.connections[conn.ID] = conn
	m.agentConns[conn.AgentID] = conn

	return nil
}

// RemoveConnection 移除连接
func (m *ConnectionManager) RemoveConnection(connID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[connID]
	if !exists {
		return fmt.Errorf("connection %s not found", connID)
	}

	delete(m.connections, connID)
	delete(m.agentConns, conn.AgentID)

	return conn.Close()
}

// GetConnection 获取连接
func (m *ConnectionManager) GetConnection(connID string) (*Connection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[connID]
	if !exists {
		return nil, fmt.Errorf("connection %s not found", connID)
	}

	return conn, nil
}

// GetConnectionByAgent 根据AgentID获取连接
func (m *ConnectionManager) GetConnectionByAgent(agentID string) (*Connection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.agentConns[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s has no connection", agentID)
	}

	return conn, nil
}

// ListConnections 列出所有连接
func (m *ConnectionManager) ListConnections() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		conns = append(conns, conn)
	}

	return conns
}

// GetActiveConnections 获取活跃连接
func (m *ConnectionManager) GetActiveConnections() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns := make([]*Connection, 0)
	for _, conn := range m.connections {
		if conn.Status == ConnectionStatusConnected {
			conns = append(conns, conn)
		}
	}

	return conns
}

// BroadcastToAll 广播给所有连接
func (m *ConnectionManager) BroadcastToAll(data []byte) error {
	m.mu.RLock()
	conns := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		if conn.Status == ConnectionStatusConnected {
			conns = append(conns, conn)
		}
	}
	m.mu.RUnlock()

	errors := make([]error, 0)
	for _, conn := range conns {
		if err := conn.Send(data); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", conn.AgentID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d connections", len(errors))
	}

	return nil
}

// BroadcastToAgents 广播给指定Agent列表
func (m *ConnectionManager) BroadcastToAgents(agentIDs []string, data []byte) error {
	m.mu.RLock()
	conns := make([]*Connection, 0, len(agentIDs))
	for _, agentID := range agentIDs {
		if conn, exists := m.agentConns[agentID]; exists && conn.Status == ConnectionStatusConnected {
			conns = append(conns, conn)
		}
	}
	m.mu.RUnlock()

	errors := make([]error, 0)
	for _, conn := range conns {
		if err := conn.Send(data); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to %s: %w", conn.AgentID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast failed for %d agents", len(errors))
	}

	return nil
}

// CheckHeartbeat 检查心跳超时
func (m *ConnectionManager) CheckHeartbeat(timeout time.Duration) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	timeoutConns := make([]string, 0)
	for _, conn := range m.connections {
		if !conn.IsAlive(timeout) {
			timeoutConns = append(timeoutConns, conn.ID)
		}
	}

	return timeoutConns
}

// GetConnectionCount 获取连接数
func (m *ConnectionManager) GetConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.connections)
}

// GetConnectionCountByStatus 按状态统计连接数
func (m *ConnectionManager) GetConnectionCountByStatus() map[ConnectionStatus]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[ConnectionStatus]int)
	for _, conn := range m.connections {
		counts[conn.Status]++
	}

	return counts
}
