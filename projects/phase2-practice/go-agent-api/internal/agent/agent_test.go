package agent

import (
	"context"
	"testing"
)

func TestAgentRegistry(t *testing.T) {
	registry := NewAgentRegistry()

	// Test registration
	agent := &Agent{
		ID:     "test-agent-1",
		Name:   "Test Agent",
		Type:   AgentTypeGeneral,
		Status: AgentStatusIdle,
		Config: AgentConfig{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   2000,
		},
	}

	err := registry.Register(agent)
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(agent)
	if err == nil {
		t.Fatal("Expected error when registering duplicate agent")
	}

	// Test retrieval
	retrieved, err := registry.Get("test-agent-1")
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Errorf("Expected agent ID %s, got %s", agent.ID, retrieved.ID)
	}

	// Test list
	agents := registry.List()
	if len(agents) != 1 {
		t.Errorf("Expected 1 agent, got %d", len(agents))
	}

	// Test count
	count := registry.Count()
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Test unregistration
	err = registry.Unregister("test-agent-1")
	if err != nil {
		t.Fatalf("Failed to unregister agent: %v", err)
	}

	// Verify agent is gone
	_, err = registry.Get("test-agent-1")
	if err == nil {
		t.Fatal("Expected error when getting unregistered agent")
	}
}

func TestAgentService(t *testing.T) {
	// Skip if no API key
	ctx := context.Background()
	service := NewAgentService("test-api-key")

	// Test agent creation
	req := &CreateAgentRequest{
		Name: "Test Agent",
		Type: AgentTypeGeneral,
		Config: AgentConfig{
			Model:       "gpt-4",
			Temperature: 0.7,
		},
	}

	agent, err := service.CreateAgent(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent.Name != req.Name {
		t.Errorf("Expected agent name %s, got %s", req.Name, agent.Name)
	}

	if agent.Status != AgentStatusIdle {
		t.Errorf("Expected agent status %s, got %s", AgentStatusIdle, agent.Status)
	}

	// Test agent retrieval
	retrieved, err := service.GetAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if retrieved.ID != agent.ID {
		t.Errorf("Expected agent ID %s, got %s", agent.ID, retrieved.ID)
	}

	// Test agent listing
	agents, err := service.ListAgents(ctx)
	if err != nil {
		t.Fatalf("Failed to list agents: %v", err)
	}

	if len(agents) == 0 {
		t.Error("Expected at least one agent")
	}

	// Test agent deletion
	err = service.DeleteAgent(ctx, agent.ID)
	if err != nil {
		t.Fatalf("Failed to delete agent: %v", err)
	}

	// Verify agent is deleted
	_, err = service.GetAgent(ctx, agent.ID)
	if err == nil {
		t.Error("Expected error when getting deleted agent")
	}
}
