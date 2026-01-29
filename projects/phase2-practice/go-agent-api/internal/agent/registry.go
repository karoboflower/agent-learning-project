package agent

import (
	"fmt"
	"sync"
)

// AgentRegistry manages registered agents
type AgentRegistry struct {
	agents map[string]*Agent
	mu     sync.RWMutex
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{
		agents: make(map[string]*Agent),
	}
}

// Register registers a new agent
func (r *AgentRegistry) Register(agent *Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agent.ID]; exists {
		return fmt.Errorf("agent with ID %s already exists", agent.ID)
	}

	r.agents[agent.ID] = agent
	return nil
}

// Unregister removes an agent from the registry
func (r *AgentRegistry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[id]; !exists {
		return fmt.Errorf("agent with ID %s not found", id)
	}

	delete(r.agents, id)
	return nil
}

// Get retrieves an agent by ID
func (r *AgentRegistry) Get(id string) (*Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[id]
	if !exists {
		return nil, fmt.Errorf("agent with ID %s not found", id)
	}

	return agent, nil
}

// Update updates an agent in the registry
func (r *AgentRegistry) Update(agent *Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.agents[agent.ID]; !exists {
		return fmt.Errorf("agent with ID %s not found", agent.ID)
	}

	r.agents[agent.ID] = agent
	return nil
}

// List returns all registered agents
func (r *AgentRegistry) List() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0, len(r.agents))
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}

	return agents
}

// Count returns the number of registered agents
func (r *AgentRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.agents)
}

// GetByStatus returns agents with specific status
func (r *AgentRegistry) GetByStatus(status AgentStatus) []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]*Agent, 0)
	for _, agent := range r.agents {
		if agent.Status == status {
			agents = append(agents, agent)
		}
	}

	return agents
}
