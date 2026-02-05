package compression

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

// PromptCompressor Prompt压缩器
type PromptCompressor struct {
	stopwords map[string]bool
}

// NewPromptCompressor 创建Prompt压缩器
func NewPromptCompressor() *PromptCompressor {
	// 常见停用词
	stopwords := []string{
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "from", "as", "is", "was", "are", "were", "been",
		"be", "have", "has", "had", "do", "does", "did", "will", "would",
		"should", "could", "may", "might", "must", "can",
	}

	stopwordMap := make(map[string]bool)
	for _, word := range stopwords {
		stopwordMap[word] = true
	}

	return &PromptCompressor{
		stopwords: stopwordMap,
	}
}

// Compress 压缩Prompt
func (pc *PromptCompressor) Compress(prompt string, aggressiveness int) (string, model.PromptTemplate) {
	original := prompt
	originalLen := len(original)

	// 级别1：基础清理
	prompt = pc.basicCleanup(prompt)

	// 级别2：移除停用词
	if aggressiveness >= 2 {
		prompt = pc.removeStopwords(prompt)
	}

	// 级别3：缩写和简化
	if aggressiveness >= 3 {
		prompt = pc.abbreviate(prompt)
	}

	// 级别4：积极压缩
	if aggressiveness >= 4 {
		prompt = pc.aggressiveCompression(prompt)
	}

	compressedLen := len(prompt)
	ratio := float64(compressedLen) / float64(originalLen) * 100

	template := model.PromptTemplate{
		Template:         prompt,
		OriginalLength:   originalLen,
		CompressedLength: compressedLen,
		CompressionRatio: ratio,
	}

	return prompt, template
}

// basicCleanup 基础清理
func (pc *PromptCompressor) basicCleanup(text string) string {
	// 移除多余的空白字符
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// 移除行首行尾空白
	text = strings.TrimSpace(text)

	// 移除重复的标点符号
	text = regexp.MustCompile(`([.!?]){2,}`).ReplaceAllString(text, "$1")

	// 移除多余的换行符
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	return text
}

// removeStopwords 移除停用词
func (pc *PromptCompressor) removeStopwords(text string) string {
	words := strings.Fields(text)
	filtered := make([]string, 0, len(words))

	for _, word := range words {
		lower := strings.ToLower(word)
		// 保留长词和非停用词
		if len(word) > 4 || !pc.stopwords[lower] {
			filtered = append(filtered, word)
		}
	}

	return strings.Join(filtered, " ")
}

// abbreviate 缩写和简化
func (pc *PromptCompressor) abbreviate(text string) string {
	// 常见缩写
	abbreviations := map[string]string{
		"please":       "pls",
		"because":      "bc",
		"without":      "w/o",
		"with":         "w/",
		"information":  "info",
		"description":  "desc",
		"configuration": "config",
		"application":  "app",
		"documentation": "docs",
		"repository":   "repo",
		"database":     "db",
		"function":     "func",
		"parameter":    "param",
		"variable":     "var",
		"example":      "ex",
		"requirement":  "req",
		"response":     "resp",
		"request":      "req",
	}

	for full, abbr := range abbreviations {
		text = regexp.MustCompile(`(?i)\b`+full+`\b`).ReplaceAllString(text, abbr)
	}

	return text
}

// aggressiveCompression 积极压缩
func (pc *PromptCompressor) aggressiveCompression(text string) string {
	// 移除冗余短语
	redundantPhrases := []string{
		"in order to",
		"due to the fact that",
		"in the event that",
		"for the purpose of",
		"in spite of the fact that",
		"it is important to note that",
		"it should be noted that",
	}

	for _, phrase := range redundantPhrases {
		text = strings.ReplaceAll(text, phrase, "")
	}

	// 简化表达
	text = strings.ReplaceAll(text, "in my opinion", "I think")
	text = strings.ReplaceAll(text, "as a result of", "because")
	text = strings.ReplaceAll(text, "at this point in time", "now")

	return pc.basicCleanup(text)
}

// CompressMessages 压缩消息列表
func (pc *PromptCompressor) CompressMessages(messages []model.Message, targetTokens int) []model.Message {
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += msg.Tokens
	}

	if totalTokens <= targetTokens {
		return messages
	}

	// 压缩策略：保留重要消息，压缩其他消息
	compressed := make([]model.Message, 0, len(messages))

	for _, msg := range messages {
		if msg.Important {
			// 重要消息保留
			compressed = append(compressed, msg)
		} else {
			// 非重要消息压缩
			compressedContent, _ := pc.Compress(msg.Content, 3)
			msg.Content = compressedContent
			msg.Tokens = len(compressedContent) / 4 // 粗略估算Token数
			compressed = append(compressed, msg)
		}
	}

	return compressed
}

// SummarizeContext 总结上下文
func (pc *PromptCompressor) SummarizeContext(messages []model.Message) string {
	// 简单实现：提取关键信息
	var keyPoints []string

	for _, msg := range messages {
		if msg.Role == "user" {
			// 提取用户问题
			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if len(line) > 20 && (strings.Contains(line, "?") || strings.HasPrefix(line, "Please") || strings.HasPrefix(line, "How")) {
					keyPoints = append(keyPoints, line)
				}
			}
		}
	}

	if len(keyPoints) == 0 {
		return "Previous conversation context."
	}

	return fmt.Sprintf("Context: %s", strings.Join(keyPoints, "; "))
}

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*model.PromptTemplate
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*model.PromptTemplate),
	}
}

// SaveTemplate 保存模板
func (tm *TemplateManager) SaveTemplate(template *model.PromptTemplate) {
	tm.templates[template.ID] = template
}

// GetTemplate 获取模板
func (tm *TemplateManager) GetTemplate(id string) (*model.PromptTemplate, error) {
	template, ok := tm.templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	return template, nil
}

// RenderTemplate 渲染模板
func (tm *TemplateManager) RenderTemplate(id string, variables map[string]string) (string, error) {
	template, err := tm.GetTemplate(id)
	if err != nil {
		return "", err
	}

	result := template.Template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// ListTemplates 列出所有模板
func (tm *TemplateManager) ListTemplates() []*model.PromptTemplate {
	templates := make([]*model.PromptTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}
	return templates
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(id string) error {
	if _, ok := tm.templates[id]; !ok {
		return fmt.Errorf("template not found: %s", id)
	}
	delete(tm.templates, id)
	return nil
}
