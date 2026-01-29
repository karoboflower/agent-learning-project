package tools

import (
	"context"
	"fmt"
	"net/http"
	"io"
	"net/url"
	"strings"
)

// SearchTool implements web search functionality
type SearchTool struct {
	BaseTool
	apiKey string
}

// NewSearchTool creates a new search tool
func NewSearchTool(apiKey string) *SearchTool {
	return &SearchTool{
		BaseTool: BaseTool{
			name:        "search",
			description: "Search the web for information. Input should be a search query.",
		},
		apiKey: apiKey,
	}
}

// Execute performs a web search
func (t *SearchTool) Execute(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("search query cannot be empty")
	}

	// Simulate search (in real implementation, use search API like Serper, Bing, etc.)
	query := url.QueryEscape(input)
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s", query)

	// For demonstration, return a mock result
	result := fmt.Sprintf("Search results for '%s':\n\n", input)
	result += "1. Result 1: Information related to " + input + "\n"
	result += "2. Result 2: More details about " + input + "\n"
	result += "3. Result 3: Additional resources for " + input + "\n\n"
	result += fmt.Sprintf("(Search URL: %s)", searchURL)

	return result, nil
}

// MockSearchTool is a simple mock implementation for testing
type MockSearchTool struct {
	BaseTool
}

// NewMockSearchTool creates a mock search tool
func NewMockSearchTool() *MockSearchTool {
	return &MockSearchTool{
		BaseTool: BaseTool{
			name:        "search",
			description: "Mock search tool for testing",
		},
	}
}

// Execute returns mock search results
func (t *MockSearchTool) Execute(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("Mock search results for: %s", input), nil
}

// WebFetchTool fetches content from a URL
type WebFetchTool struct {
	BaseTool
	client *http.Client
}

// NewWebFetchTool creates a new web fetch tool
func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		BaseTool: BaseTool{
			name:        "web_fetch",
			description: "Fetch content from a URL. Input should be a valid HTTP/HTTPS URL.",
		},
		client: &http.Client{},
	}
}

// Execute fetches content from a URL
func (t *WebFetchTool) Execute(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Validate URL
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return "", fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", input, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: status %d", resp.StatusCode)
	}

	// Read body (limit to 10KB for safety)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}
