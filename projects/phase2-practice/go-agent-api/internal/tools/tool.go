package tools

import (
	"context"
	"fmt"
)

// Tool represents a tool that can be used by agents
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, input string) (string, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// BaseTool provides common functionality for tools
type BaseTool struct {
	name        string
	description string
}

// Name returns the tool name
func (t *BaseTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *BaseTool) Description() string {
	return t.description
}

// ExecuteWithResult wraps the execution and returns a structured result
func ExecuteWithResult(ctx context.Context, tool Tool, input string) *ToolResult {
	output, err := tool.Execute(ctx, input)
	if err != nil {
		return &ToolResult{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}
	}

	return &ToolResult{
		Success: true,
		Output:  output,
		Error:   "",
	}
}

// ToolExecutor executes tools safely
type ToolExecutor struct {
	registry *ToolRegistry
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(registry *ToolRegistry) *ToolExecutor {
	return &ToolExecutor{
		registry: registry,
	}
}

// Execute executes a tool by name
func (te *ToolExecutor) Execute(ctx context.Context, toolName, input string) (*ToolResult, error) {
	tool, err := te.registry.Get(toolName)
	if err != nil {
		return nil, fmt.Errorf("tool not found: %s", toolName)
	}

	return ExecuteWithResult(ctx, tool, input), nil
}

// ExecuteMultiple executes multiple tools sequentially
func (te *ToolExecutor) ExecuteMultiple(ctx context.Context, executions []ToolExecution) ([]*ToolResult, error) {
	results := make([]*ToolResult, 0, len(executions))

	for _, exec := range executions {
		result, err := te.Execute(ctx, exec.ToolName, exec.Input)
		if err != nil {
			return results, err
		}
		results = append(results, result)

		// Stop if a tool fails
		if !result.Success {
			return results, fmt.Errorf("tool %s failed: %s", exec.ToolName, result.Error)
		}
	}

	return results, nil
}

// ToolExecution represents a tool execution request
type ToolExecution struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
}
