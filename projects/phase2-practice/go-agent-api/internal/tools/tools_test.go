package tools

import (
	"context"
	"testing"
)

func TestCodeTool(t *testing.T) {
	tool := NewCodeTool()
	ctx := context.Background()

	// Test code analysis
	code := `func main() {
		// This is a comment
		fmt.Println("Hello, World!")
	}`

	result, err := tool.Execute(ctx, "analyze:"+code)
	if err != nil {
		t.Fatalf("Failed to analyze code: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	t.Logf("Analysis result: %s", result)

	// Test code formatting
	result, err = tool.Execute(ctx, "format:"+code)
	if err != nil {
		t.Fatalf("Failed to format code: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Test syntax check
	result, err = tool.Execute(ctx, "check:"+code)
	if err != nil {
		t.Fatalf("Failed to check syntax: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	// Register tools
	codeTool := NewCodeTool()
	searchTool := NewMockSearchTool()

	err := registry.Register(codeTool)
	if err != nil {
		t.Fatalf("Failed to register code tool: %v", err)
	}

	err = registry.Register(searchTool)
	if err != nil {
		t.Fatalf("Failed to register search tool: %v", err)
	}

	// Test count
	if registry.Count() != 2 {
		t.Errorf("Expected 2 tools, got %d", registry.Count())
	}

	// Test retrieval
	tool, err := registry.Get("code")
	if err != nil {
		t.Fatalf("Failed to get code tool: %v", err)
	}

	if tool.Name() != "code" {
		t.Errorf("Expected tool name 'code', got '%s'", tool.Name())
	}

	// Test list names
	names := registry.ListNames()
	if len(names) != 2 {
		t.Errorf("Expected 2 tool names, got %d", len(names))
	}

	// Test has
	if !registry.Has("code") {
		t.Error("Expected registry to have 'code' tool")
	}

	if registry.Has("nonexistent") {
		t.Error("Expected registry to not have 'nonexistent' tool")
	}
}

func TestToolExecutor(t *testing.T) {
	registry := NewToolRegistry()
	registry.Register(NewMockSearchTool())

	executor := NewToolExecutor(registry)
	ctx := context.Background()

	// Test execution
	result, err := executor.Execute(ctx, "search", "test query")
	if err != nil {
		t.Fatalf("Failed to execute tool: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful execution, got error: %s", result.Error)
	}

	if result.Output == "" {
		t.Error("Expected non-empty output")
	}

	t.Logf("Execution result: %s", result.Output)

	// Test invalid tool
	_, err = executor.Execute(ctx, "nonexistent", "input")
	if err == nil {
		t.Error("Expected error when executing nonexistent tool")
	}
}
