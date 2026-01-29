package tools

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileTool provides file operation functionality
type FileTool struct {
	BaseTool
	allowedPaths []string
}

// NewFileTool creates a new file tool
func NewFileTool(allowedPaths []string) *FileTool {
	return &FileTool{
		BaseTool: BaseTool{
			name:        "file",
			description: "Perform file operations (read, write, list). Input format: 'operation:path[:content]'",
		},
		allowedPaths: allowedPaths,
	}
}

// Execute performs file operations
func (t *FileTool) Execute(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("file operation input cannot be empty")
	}

	// Parse operation (format: "operation:path" or "operation:path:content")
	parts := strings.SplitN(input, ":", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid input format. Expected 'operation:path[:content]'")
	}

	operation := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])

	// Validate path
	if err := t.validatePath(path); err != nil {
		return "", err
	}

	switch operation {
	case "read":
		return t.readFile(path)
	case "write":
		if len(parts) < 3 {
			return "", fmt.Errorf("write operation requires content")
		}
		content := parts[2]
		return t.writeFile(path, content)
	case "list":
		return t.listFiles(path)
	case "exists":
		return t.checkExists(path), nil
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// validatePath checks if the path is allowed
func (t *FileTool) validatePath(path string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path is in allowed paths
	if len(t.allowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range t.allowedPaths {
			absAllowed, _ := filepath.Abs(allowedPath)
			if strings.HasPrefix(absPath, absAllowed) {
				allowed = true
				break
			}
		}

		if !allowed {
			return fmt.Errorf("path not allowed: %s", path)
		}
	}

	return nil
}

// readFile reads a file
func (t *FileTool) readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// writeFile writes to a file
func (t *FileTool) writeFile(path, content string) (string, error) {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

// listFiles lists files in a directory
func (t *FileTool) listFiles(path string) (string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to list directory: %w", err)
	}

	result := fmt.Sprintf("Files in %s:\n\n", path)
	for _, file := range files {
		fileType := "file"
		if file.IsDir() {
			fileType = "dir"
		}
		result += fmt.Sprintf("- %s (%s, %d bytes)\n", file.Name(), fileType, file.Size())
	}

	return result, nil
}

// checkExists checks if a file or directory exists
func (t *FileTool) checkExists(path string) string {
	if _, err := os.Stat(path); err == nil {
		return fmt.Sprintf("Path exists: %s", path)
	} else if os.IsNotExist(err) {
		return fmt.Sprintf("Path does not exist: %s", path)
	} else {
		return fmt.Sprintf("Error checking path: %s", err.Error())
	}
}
