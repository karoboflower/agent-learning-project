package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

// AgentService defines the interface for agent operations
type AgentService interface {
	CreateAgent(ctx context.Context, req *CreateAgentRequest) (*Agent, error)
	GetAgent(ctx context.Context, id string) (*Agent, error)
	ListAgents(ctx context.Context) ([]*Agent, error)
	DeleteAgent(ctx context.Context, id string) error
	ExecuteTask(ctx context.Context, agent *Agent, task *Task) (*TaskResult, error)
}

// agentService implements AgentService
type agentService struct {
	openaiClient *openai.Client
	registry     *AgentRegistry
}

// NewAgentService creates a new agent service
func NewAgentService(apiKey string) AgentService {
	return &agentService{
		openaiClient: openai.NewClient(apiKey),
		registry:     NewAgentRegistry(),
	}
}

// CreateAgent creates a new agent
func (s *agentService) CreateAgent(ctx context.Context, req *CreateAgentRequest) (*Agent, error) {
	agent := &Agent{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Type:      req.Type,
		Status:    AgentStatusIdle,
		Config:    req.Config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set default config if not provided
	if agent.Config.Model == "" {
		agent.Config.Model = "gpt-4"
	}
	if agent.Config.Temperature == 0 {
		agent.Config.Temperature = 0.7
	}
	if agent.Config.MaxTokens == 0 {
		agent.Config.MaxTokens = 2000
	}

	// Register agent
	if err := s.registry.Register(agent); err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (s *agentService) GetAgent(ctx context.Context, id string) (*Agent, error) {
	agent, err := s.registry.Get(id)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

// ListAgents lists all agents
func (s *agentService) ListAgents(ctx context.Context) ([]*Agent, error) {
	return s.registry.List(), nil
}

// DeleteAgent deletes an agent
func (s *agentService) DeleteAgent(ctx context.Context, id string) error {
	return s.registry.Unregister(id)
}

// ExecuteTask executes a task using the agent
func (s *agentService) ExecuteTask(ctx context.Context, agent *Agent, task *Task) (*TaskResult, error) {
	startTime := time.Now()

	// Update agent status
	agent.Status = AgentStatusBusy
	agent.UpdatedAt = time.Now()
	s.registry.Update(agent)

	defer func() {
		agent.Status = AgentStatusIdle
		agent.UpdatedAt = time.Now()
		s.registry.Update(agent)
	}()

	// Build system prompt based on agent type
	systemPrompt := s.buildSystemPrompt(agent)

	// Call OpenAI API
	resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: agent.Config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: task.Input,
			},
		},
		Temperature: agent.Config.Temperature,
		MaxTokens:   agent.Config.MaxTokens,
	})

	if err != nil {
		return &TaskResult{
			TaskID:    task.ID,
			Status:    TaskStatusFailed,
			Error:     err.Error(),
			CreatedAt: startTime,
			EndedAt:   time.Now(),
			Duration:  time.Since(startTime).Milliseconds(),
		}, err
	}

	output := ""
	if len(resp.Choices) > 0 {
		output = resp.Choices[0].Message.Content
	}

	result := &TaskResult{
		TaskID:    task.ID,
		Status:    TaskStatusCompleted,
		Output:    output,
		CreatedAt: startTime,
		EndedAt:   time.Now(),
		Duration:  time.Since(startTime).Milliseconds(),
		Metadata: map[string]interface{}{
			"model":        resp.Model,
			"tokens_used":  resp.Usage.TotalTokens,
			"finish_reason": resp.Choices[0].FinishReason,
		},
	}

	return result, nil
}

// buildSystemPrompt builds the system prompt based on agent type
func (s *agentService) buildSystemPrompt(agent *Agent) string {
	switch agent.Type {
	case AgentTypeCodeReview:
		return "You are an expert code reviewer. Analyze the provided code and give detailed feedback on code quality, potential bugs, security issues, and improvement suggestions."
	case AgentTypeDocQA:
		return "You are a helpful documentation assistant. Answer questions based on the provided documentation context accurately and concisely."
	case AgentTypeAPIHandler:
		return "You are an API request handler. Process API requests, validate inputs, and generate appropriate responses."
	default:
		return "You are a helpful AI assistant. Provide accurate, detailed, and helpful responses to user queries."
	}
}
