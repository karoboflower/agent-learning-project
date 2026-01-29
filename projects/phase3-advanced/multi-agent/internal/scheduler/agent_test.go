package scheduler

import (
	"fmt"
	"testing"
	"time"
)

func TestNewAgentRegistry(t *testing.T) {
	registry := NewAgentRegistry()

	if registry == nil {
		t.Fatal("NewAgentRegistry returned nil")
	}

	if registry.agents == nil {
		t.Error("agents map not initialized")
	}
}

func TestAgentRegistry_Register(t *testing.T) {
	registry := NewAgentRegistry()

	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"code_review", "testing"},
		Status:       AgentStatusIdle,
		MaxTasks:     5,
	}

	err := registry.Register(agent)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Check defaults
	if agent.RegisteredAt.IsZero() {
		t.Error("RegisteredAt not set")
	}

	if agent.LastHeartbeat.IsZero() {
		t.Error("LastHeartbeat not set")
	}

	if agent.Metadata == nil {
		t.Error("Metadata not initialized")
	}

	// Check stored agent
	stored, err := registry.GetAgent("agent-001")
	if err != nil {
		t.Fatalf("GetAgent failed: %v", err)
	}

	if stored.ID != agent.ID {
		t.Errorf("Expected ID %s, got %s", agent.ID, stored.ID)
	}
}

func TestAgentRegistry_Register_Validation(t *testing.T) {
	registry := NewAgentRegistry()

	tests := []struct {
		name    string
		agent   *Agent
		wantErr bool
	}{
		{
			name: "empty ID",
			agent: &Agent{
				Name:         "Test",
				Capabilities: []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "empty name",
			agent: &Agent{
				ID:           "agent-001",
				Capabilities: []string{"test"},
			},
			wantErr: true,
		},
		{
			name: "no capabilities",
			agent: &Agent{
				ID:   "agent-001",
				Name: "Test",
			},
			wantErr: true,
		},
		{
			name: "valid agent",
			agent: &Agent{
				ID:           "agent-001",
				Name:         "Test",
				Capabilities: []string{"test"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.agent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentRegistry_Unregister(t *testing.T) {
	registry := NewAgentRegistry()

	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
	}

	registry.Register(agent)

	err := registry.Unregister("agent-001")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}

	_, err = registry.GetAgent("agent-001")
	if err == nil {
		t.Error("Expected error after unregister")
	}
}

func TestAgentRegistry_ListAgents(t *testing.T) {
	registry := NewAgentRegistry()

	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"test"}},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"test"}},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"test"}},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	list := registry.ListAgents()
	if len(list) != 3 {
		t.Errorf("Expected 3 agents, got %d", len(list))
	}
}

func TestAgentRegistry_FindAgentsByCapability(t *testing.T) {
	registry := NewAgentRegistry()

	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"code_review"}},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"testing"}},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"code_review", "testing"}},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	// Find code_review agents
	reviewAgents := registry.FindAgentsByCapability("code_review")
	if len(reviewAgents) != 2 {
		t.Errorf("Expected 2 code_review agents, got %d", len(reviewAgents))
	}

	// Find testing agents
	testAgents := registry.FindAgentsByCapability("testing")
	if len(testAgents) != 2 {
		t.Errorf("Expected 2 testing agents, got %d", len(testAgents))
	}
}

func TestAgentRegistry_FindAvailableAgents(t *testing.T) {
	registry := NewAgentRegistry()

	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 0},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"test"}, Status: AgentStatusBusy, MaxTasks: 5, CurrentTasks: 5},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"test"}, Status: AgentStatusIdle, MaxTasks: 5, CurrentTasks: 3},
		{ID: "agent-004", Name: "Agent 4", Capabilities: []string{"test"}, Status: AgentStatusOffline, MaxTasks: 5, CurrentTasks: 0},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	available := registry.FindAvailableAgents()
	if len(available) != 2 {
		t.Errorf("Expected 2 available agents, got %d", len(available))
	}
}

func TestAgentRegistry_UpdateAgentStatus(t *testing.T) {
	registry := NewAgentRegistry()

	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
		Status:       AgentStatusIdle,
	}

	registry.Register(agent)

	err := registry.UpdateAgentStatus("agent-001", AgentStatusBusy)
	if err != nil {
		t.Fatalf("UpdateAgentStatus failed: %v", err)
	}

	updated, _ := registry.GetAgent("agent-001")
	if updated.Status != AgentStatusBusy {
		t.Errorf("Expected status BUSY, got %s", updated.Status)
	}
}

func TestAgentRegistry_UpdateAgentLoad(t *testing.T) {
	registry := NewAgentRegistry()

	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
	}

	registry.Register(agent)

	err := registry.UpdateAgentLoad("agent-001", 0.5)
	if err != nil {
		t.Fatalf("UpdateAgentLoad failed: %v", err)
	}

	updated, _ := registry.GetAgent("agent-001")
	if updated.Load != 0.5 {
		t.Errorf("Expected load 0.5, got %f", updated.Load)
	}

	// Test invalid load
	err = registry.UpdateAgentLoad("agent-001", 1.5)
	if err == nil {
		t.Error("Expected error for invalid load")
	}
}

func TestAgentRegistry_TaskCount(t *testing.T) {
	registry := NewAgentRegistry()

	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
		MaxTasks:     3,
		CurrentTasks: 0,
		Status:       AgentStatusIdle,
	}

	registry.Register(agent)

	// Increment
	for i := 0; i < 3; i++ {
		err := registry.IncrementTaskCount("agent-001")
		if err != nil {
			t.Fatalf("IncrementTaskCount failed: %v", err)
		}
	}

	updated, _ := registry.GetAgent("agent-001")
	if updated.CurrentTasks != 3 {
		t.Errorf("Expected 3 tasks, got %d", updated.CurrentTasks)
	}

	if updated.Status != AgentStatusBusy {
		t.Errorf("Expected status BUSY, got %s", updated.Status)
	}

	// Try to exceed max
	err := registry.IncrementTaskCount("agent-001")
	if err == nil {
		t.Error("Expected error when exceeding max tasks")
	}

	// Decrement
	for i := 0; i < 3; i++ {
		err := registry.DecrementTaskCount("agent-001")
		if err != nil {
			t.Fatalf("DecrementTaskCount failed: %v", err)
		}
	}

	updated, _ = registry.GetAgent("agent-001")
	if updated.CurrentTasks != 0 {
		t.Errorf("Expected 0 tasks, got %d", updated.CurrentTasks)
	}

	if updated.Status != AgentStatusIdle {
		t.Errorf("Expected status IDLE, got %s", updated.Status)
	}
}

func TestAgentRegistry_Heartbeat(t *testing.T) {
	registry := NewAgentRegistry()

	agent := &Agent{
		ID:           "agent-001",
		Name:         "Test Agent",
		Capabilities: []string{"test"},
	}

	registry.Register(agent)

	// Update heartbeat
	time.Sleep(100 * time.Millisecond)
	err := registry.UpdateHeartbeat("agent-001")
	if err != nil {
		t.Fatalf("UpdateHeartbeat failed: %v", err)
	}

	// Check heartbeat timeout (should not timeout yet)
	timeoutAgents := registry.CheckHeartbeat(1 * time.Second)
	if len(timeoutAgents) != 0 {
		t.Error("Agent should not timeout yet")
	}

	// Wait and check timeout
	time.Sleep(1100 * time.Millisecond)
	timeoutAgents = registry.CheckHeartbeat(1 * time.Second)
	if len(timeoutAgents) != 1 {
		t.Errorf("Expected 1 timeout agent, got %d", len(timeoutAgents))
	}

	// Check status changed to offline
	updated, _ := registry.GetAgent("agent-001")
	if updated.Status != AgentStatusOffline {
		t.Errorf("Expected status OFFLINE, got %s", updated.Status)
	}
}

func TestAgentRegistry_GetAgentCount(t *testing.T) {
	registry := NewAgentRegistry()

	if registry.GetAgentCount() != 0 {
		t.Error("Expected 0 agents initially")
	}

	for i := 0; i < 5; i++ {
		agent := &Agent{
			ID:           fmt.Sprintf("agent-%03d", i),
			Name:         fmt.Sprintf("Agent %d", i),
			Capabilities: []string{"test"},
		}
		registry.Register(agent)
	}

	if registry.GetAgentCount() != 5 {
		t.Errorf("Expected 5 agents, got %d", registry.GetAgentCount())
	}
}

func TestAgentRegistry_GetAgentCountByStatus(t *testing.T) {
	registry := NewAgentRegistry()

	agents := []*Agent{
		{ID: "agent-001", Name: "Agent 1", Capabilities: []string{"test"}, Status: AgentStatusIdle},
		{ID: "agent-002", Name: "Agent 2", Capabilities: []string{"test"}, Status: AgentStatusIdle},
		{ID: "agent-003", Name: "Agent 3", Capabilities: []string{"test"}, Status: AgentStatusBusy},
		{ID: "agent-004", Name: "Agent 4", Capabilities: []string{"test"}, Status: AgentStatusOffline},
	}

	for _, agent := range agents {
		registry.Register(agent)
	}

	counts := registry.GetAgentCountByStatus()

	if counts[AgentStatusIdle] != 2 {
		t.Errorf("Expected 2 idle agents, got %d", counts[AgentStatusIdle])
	}

	if counts[AgentStatusBusy] != 1 {
		t.Errorf("Expected 1 busy agent, got %d", counts[AgentStatusBusy])
	}

	if counts[AgentStatusOffline] != 1 {
		t.Errorf("Expected 1 offline agent, got %d", counts[AgentStatusOffline])
	}
}

func BenchmarkAgentRegistry_Register(b *testing.B) {
	registry := NewAgentRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agent := &Agent{
			ID:           fmt.Sprintf("agent-%d", i),
			Name:         fmt.Sprintf("Agent %d", i),
			Capabilities: []string{"test"},
		}
		registry.Register(agent)
	}
}

func BenchmarkAgentRegistry_FindAgentsByCapability(b *testing.B) {
	registry := NewAgentRegistry()

	// Setup
	for i := 0; i < 100; i++ {
		agent := &Agent{
			ID:           fmt.Sprintf("agent-%d", i),
			Name:         fmt.Sprintf("Agent %d", i),
			Capabilities: []string{"test", "code_review"},
		}
		registry.Register(agent)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.FindAgentsByCapability("code_review")
	}
}
