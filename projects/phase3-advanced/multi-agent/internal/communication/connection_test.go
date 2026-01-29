package communication

import (
	"testing"
	"time"
)

func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager()

	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}

	if cm.connections == nil {
		t.Error("connections map not initialized")
	}

	if cm.agentConns == nil {
		t.Error("agentConns map not initialized")
	}
}

func TestConnectionManager_AddConnection(t *testing.T) {
	cm := NewConnectionManager()

	conn := &Connection{
		ID:      "conn-001",
		AgentID: "agent-001",
		Status:  ConnectionStatusConnected,
	}

	err := cm.AddConnection(conn)
	if err != nil {
		t.Fatalf("AddConnection failed: %v", err)
	}

	// Check connection was added
	retrieved, err := cm.GetConnection("conn-001")
	if err != nil {
		t.Fatalf("GetConnection failed: %v", err)
	}

	if retrieved.ID != conn.ID {
		t.Errorf("Expected ID %s, got %s", conn.ID, retrieved.ID)
	}

	// Check agentConns mapping
	byAgent, err := cm.GetConnectionByAgent("agent-001")
	if err != nil {
		t.Fatalf("GetConnectionByAgent failed: %v", err)
	}

	if byAgent.ID != conn.ID {
		t.Errorf("Expected connection ID %s, got %s", conn.ID, byAgent.ID)
	}
}

func TestConnectionManager_AddConnection_Duplicate(t *testing.T) {
	cm := NewConnectionManager()

	conn := &Connection{
		ID:      "conn-001",
		AgentID: "agent-001",
		Status:  ConnectionStatusConnected,
	}

	cm.AddConnection(conn)

	// Try to add again
	err := cm.AddConnection(conn)
	if err == nil {
		t.Error("Expected error when adding duplicate connection")
	}
}

func TestConnectionManager_RemoveConnection(t *testing.T) {
	cm := NewConnectionManager()

	conn := &Connection{
		ID:       "conn-001",
		AgentID:  "agent-001",
		Status:   ConnectionStatusConnected,
		SendChan: make(chan []byte, 1),
	}

	cm.AddConnection(conn)

	err := cm.RemoveConnection("conn-001")
	if err != nil {
		t.Fatalf("RemoveConnection failed: %v", err)
	}

	// Check connection was removed
	_, err = cm.GetConnection("conn-001")
	if err == nil {
		t.Error("Expected error after removing connection")
	}

	// Check agentConns mapping was removed
	_, err = cm.GetConnectionByAgent("agent-001")
	if err == nil {
		t.Error("Expected error after removing connection")
	}
}

func TestConnectionManager_ListConnections(t *testing.T) {
	cm := NewConnectionManager()

	// Add multiple connections
	for i := 0; i < 3; i++ {
		conn := &Connection{
			ID:      string(rune('a' + i)),
			AgentID: string(rune('1' + i)),
			Status:  ConnectionStatusConnected,
		}
		cm.AddConnection(conn)
	}

	list := cm.ListConnections()
	if len(list) != 3 {
		t.Errorf("Expected 3 connections, got %d", len(list))
	}
}

func TestConnectionManager_GetActiveConnections(t *testing.T) {
	cm := NewConnectionManager()

	// Add connections with different statuses
	conns := []*Connection{
		{ID: "conn-001", AgentID: "agent-001", Status: ConnectionStatusConnected},
		{ID: "conn-002", AgentID: "agent-002", Status: ConnectionStatusDisconnected},
		{ID: "conn-003", AgentID: "agent-003", Status: ConnectionStatusConnected},
	}

	for _, conn := range conns {
		cm.AddConnection(conn)
	}

	active := cm.GetActiveConnections()
	if len(active) != 2 {
		t.Errorf("Expected 2 active connections, got %d", len(active))
	}
}

func TestConnectionManager_BroadcastToAll(t *testing.T) {
	cm := NewConnectionManager()

	// Add connections
	for i := 0; i < 3; i++ {
		conn := &Connection{
			ID:       string(rune('a' + i)),
			AgentID:  string(rune('1' + i)),
			Status:   ConnectionStatusConnected,
			SendChan: make(chan []byte, 10),
		}
		cm.AddConnection(conn)
	}

	data := []byte("test message")
	err := cm.BroadcastToAll(data)
	if err != nil {
		t.Fatalf("BroadcastToAll failed: %v", err)
	}

	// Check all connections received message
	for _, conn := range cm.ListConnections() {
		select {
		case msg := <-conn.SendChan:
			if string(msg) != "test message" {
				t.Errorf("Expected 'test message', got '%s'", string(msg))
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for broadcast message")
		}
	}
}

func TestConnectionManager_BroadcastToAgents(t *testing.T) {
	cm := NewConnectionManager()

	// Add connections
	for i := 0; i < 5; i++ {
		conn := &Connection{
			ID:       string(rune('a' + i)),
			AgentID:  string(rune('1' + i)),
			Status:   ConnectionStatusConnected,
			SendChan: make(chan []byte, 10),
		}
		cm.AddConnection(conn)
	}

	// Broadcast to specific agents
	targetAgents := []string{"1", "3"}
	data := []byte("targeted message")

	err := cm.BroadcastToAgents(targetAgents, data)
	if err != nil {
		t.Fatalf("BroadcastToAgents failed: %v", err)
	}

	// Check only targeted agents received message
	for _, conn := range cm.ListConnections() {
		isTarget := false
		for _, agentID := range targetAgents {
			if conn.AgentID == agentID {
				isTarget = true
				break
			}
		}

		if isTarget {
			select {
			case msg := <-conn.SendChan:
				if string(msg) != "targeted message" {
					t.Errorf("Expected 'targeted message', got '%s'", string(msg))
				}
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Target agent %s did not receive message", conn.AgentID)
			}
		} else {
			select {
			case <-conn.SendChan:
				t.Errorf("Non-target agent %s received message", conn.AgentID)
			default:
				// Expected - no message
			}
		}
	}
}

func TestConnection_IsAlive(t *testing.T) {
	conn := &Connection{
		ID:            "conn-001",
		AgentID:       "agent-001",
		LastHeartbeat: time.Now(),
	}

	// Should be alive
	if !conn.IsAlive(1 * time.Second) {
		t.Error("Connection should be alive")
	}

	// Wait and check
	time.Sleep(1100 * time.Millisecond)

	// Should be dead
	if conn.IsAlive(1 * time.Second) {
		t.Error("Connection should be dead")
	}
}

func TestConnection_UpdateHeartbeat(t *testing.T) {
	conn := &Connection{
		ID:            "conn-001",
		AgentID:       "agent-001",
		LastHeartbeat: time.Now().Add(-2 * time.Second),
	}

	// Should be dead
	if conn.IsAlive(1 * time.Second) {
		t.Error("Connection should be dead before update")
	}

	// Update heartbeat
	conn.UpdateHeartbeat()

	// Should be alive
	if !conn.IsAlive(1 * time.Second) {
		t.Error("Connection should be alive after update")
	}
}

func TestConnectionManager_CheckHeartbeat(t *testing.T) {
	cm := NewConnectionManager()

	// Add connections with different heartbeat times
	conns := []*Connection{
		{
			ID:            "conn-001",
			AgentID:       "agent-001",
			LastHeartbeat: time.Now(),
			Status:        ConnectionStatusConnected,
		},
		{
			ID:            "conn-002",
			AgentID:       "agent-002",
			LastHeartbeat: time.Now().Add(-2 * time.Second),
			Status:        ConnectionStatusConnected,
		},
		{
			ID:            "conn-003",
			AgentID:       "agent-003",
			LastHeartbeat: time.Now().Add(-3 * time.Second),
			Status:        ConnectionStatusConnected,
		},
	}

	for _, conn := range conns {
		cm.AddConnection(conn)
	}

	// Check with 1 second timeout
	timeoutConns := cm.CheckHeartbeat(1 * time.Second)

	// Should have 2 timeout connections
	if len(timeoutConns) != 2 {
		t.Errorf("Expected 2 timeout connections, got %d", len(timeoutConns))
	}
}

func TestConnectionManager_GetConnectionCount(t *testing.T) {
	cm := NewConnectionManager()

	if cm.GetConnectionCount() != 0 {
		t.Error("Expected 0 connections initially")
	}

	// Add connections
	for i := 0; i < 5; i++ {
		conn := &Connection{
			ID:      string(rune('a' + i)),
			AgentID: string(rune('1' + i)),
		}
		cm.AddConnection(conn)
	}

	if cm.GetConnectionCount() != 5 {
		t.Errorf("Expected 5 connections, got %d", cm.GetConnectionCount())
	}
}

func TestConnectionManager_GetConnectionCountByStatus(t *testing.T) {
	cm := NewConnectionManager()

	// Add connections with different statuses
	statuses := []ConnectionStatus{
		ConnectionStatusConnected,
		ConnectionStatusConnected,
		ConnectionStatusDisconnected,
		ConnectionStatusReconnecting,
	}

	for i, status := range statuses {
		conn := &Connection{
			ID:      string(rune('a' + i)),
			AgentID: string(rune('1' + i)),
			Status:  status,
		}
		cm.AddConnection(conn)
	}

	counts := cm.GetConnectionCountByStatus()

	if counts[ConnectionStatusConnected] != 2 {
		t.Errorf("Expected 2 connected, got %d", counts[ConnectionStatusConnected])
	}

	if counts[ConnectionStatusDisconnected] != 1 {
		t.Errorf("Expected 1 disconnected, got %d", counts[ConnectionStatusDisconnected])
	}

	if counts[ConnectionStatusReconnecting] != 1 {
		t.Errorf("Expected 1 reconnecting, got %d", counts[ConnectionStatusReconnecting])
	}
}

func BenchmarkConnectionManager_AddConnection(b *testing.B) {
	cm := NewConnectionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := &Connection{
			ID:      string(rune(i)),
			AgentID: string(rune(i + 1000)),
			Status:  ConnectionStatusConnected,
		}
		cm.AddConnection(conn)
	}
}

func BenchmarkConnectionManager_GetConnectionByAgent(b *testing.B) {
	cm := NewConnectionManager()

	// Setup
	for i := 0; i < 100; i++ {
		conn := &Connection{
			ID:      string(rune('a' + i)),
			AgentID: string(rune('1' + i)),
			Status:  ConnectionStatusConnected,
		}
		cm.AddConnection(conn)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.GetConnectionByAgent(string(rune('1' + (i % 100))))
	}
}
