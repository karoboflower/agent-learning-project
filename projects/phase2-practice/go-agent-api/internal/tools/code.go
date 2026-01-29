package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// CodeTool provides code analysis and manipulation functionality
type CodeTool struct {
	BaseTool
}

// NewCodeTool creates a new code tool
func NewCodeTool() *CodeTool {
	return &CodeTool{
		BaseTool: BaseTool{
			name:        "code",
			description: "Analyze and manipulate code. Supports syntax checking, formatting, and basic analysis.",
		},
	}
}

// Execute performs code operations
func (t *CodeTool) Execute(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("code input cannot be empty")
	}

	// Parse operation and code (format: "operation:code")
	parts := strings.SplitN(input, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid input format. Expected 'operation:code'")
	}

	operation := strings.TrimSpace(parts[0])
	code := strings.TrimSpace(parts[1])

	switch operation {
	case "analyze":
		return t.analyzeCode(code), nil
	case "format":
		return t.formatCode(code), nil
	case "check":
		return t.checkSyntax(code), nil
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

// analyzeCode analyzes code structure
func (t *CodeTool) analyzeCode(code string) string {
	lines := strings.Split(code, "\n")
	funcs := 0
	comments := 0

	// Simple pattern matching
	funcPattern := regexp.MustCompile(`func\s+\w+`)
	commentPattern := regexp.MustCompile(`^\s*(//|/\*|\*)`)

	for _, line := range lines {
		if funcPattern.MatchString(line) {
			funcs++
		}
		if commentPattern.MatchString(line) {
			comments++
		}
	}

	result := fmt.Sprintf("Code Analysis:\n")
	result += fmt.Sprintf("- Total lines: %d\n", len(lines))
	result += fmt.Sprintf("- Functions found: %d\n", funcs)
	result += fmt.Sprintf("- Comment lines: %d\n", comments)
	result += fmt.Sprintf("- Code quality: %s\n", t.assessQuality(len(lines), comments))

	return result
}

// formatCode formats code (basic)
func (t *CodeTool) formatCode(code string) string {
	lines := strings.Split(code, "\n")
	formatted := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			formatted = append(formatted, trimmed)
		}
	}

	return strings.Join(formatted, "\n")
}

// checkSyntax checks basic syntax
func (t *CodeTool) checkSyntax(code string) string {
	issues := []string{}

	// Check for balanced braces
	braces := 0
	for _, char := range code {
		if char == '{' {
			braces++
		} else if char == '}' {
			braces--
		}
	}

	if braces != 0 {
		issues = append(issues, "Unbalanced braces")
	}

	// Check for balanced parentheses
	parens := 0
	for _, char := range code {
		if char == '(' {
			parens++
		} else if char == ')' {
			parens--
		}
	}

	if parens != 0 {
		issues = append(issues, "Unbalanced parentheses")
	}

	if len(issues) == 0 {
		return "Syntax check passed: No issues found"
	}

	return "Syntax issues:\n- " + strings.Join(issues, "\n- ")
}

// assessQuality assesses code quality
func (t *CodeTool) assessQuality(totalLines, comments int) string {
	if totalLines == 0 {
		return "Unknown"
	}

	commentRatio := float64(comments) / float64(totalLines)

	if commentRatio > 0.3 {
		return "Excellent (well-documented)"
	} else if commentRatio > 0.15 {
		return "Good (adequately documented)"
	} else if commentRatio > 0.05 {
		return "Fair (needs more documentation)"
	}

	return "Poor (needs documentation)"
}
