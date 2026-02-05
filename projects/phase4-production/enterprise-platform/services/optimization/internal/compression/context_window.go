package compression

import (
	"fmt"
	"sort"

	"github.com/agent-learning/enterprise-platform/services/optimization/internal/model"
)

// ContextWindowManager 上下文窗口管理器
type ContextWindowManager struct {
	maxTokens  int
	compressor *PromptCompressor
}

// NewContextWindowManager 创建上下文窗口管理器
func NewContextWindowManager(maxTokens int) *ContextWindowManager {
	return &ContextWindowManager{
		maxTokens:  maxTokens,
		compressor: NewPromptCompressor(),
	}
}

// ManageWindow 管理上下文窗口
func (cwm *ContextWindowManager) ManageWindow(window *model.ContextWindow) error {
	// 计算当前Token数
	currentTokens := cwm.calculateTotalTokens(window.Messages)
	window.CurrentTokens = currentTokens
	window.UsagePercent = float64(currentTokens) / float64(window.MaxTokens) * 100

	// 如果未超过限制，直接返回
	if currentTokens <= window.MaxTokens {
		return nil
	}

	// 根据策略剪枝
	switch window.PruneStrategy {
	case "oldest":
		return cwm.pruneOldest(window)
	case "least_important":
		return cwm.pruneLeastImportant(window)
	case "summarize":
		return cwm.summarize(window)
	default:
		return cwm.pruneOldest(window)
	}
}

// calculateTotalTokens 计算总Token数
func (cwm *ContextWindowManager) calculateTotalTokens(messages []model.Message) int {
	total := 0
	for _, msg := range messages {
		if msg.Tokens == 0 {
			// 粗略估算：1 token ≈ 4 characters
			msg.Tokens = len(msg.Content) / 4
		}
		total += msg.Tokens
	}
	return total
}

// pruneOldest 剪枝最旧的消息
func (cwm *ContextWindowManager) pruneOldest(window *model.ContextWindow) error {
	// 保留系统消息和重要消息
	preserved := make([]model.Message, 0)
	removable := make([]model.Message, 0)

	for _, msg := range window.Messages {
		if msg.Role == "system" || msg.Important {
			preserved = append(preserved, msg)
		} else {
			removable = append(removable, msg)
		}
	}

	// 按时间排序（最旧的在前）
	sort.Slice(removable, func(i, j int) bool {
		return removable[i].Timestamp.Before(removable[j].Timestamp)
	})

	// 计算需要保留的Token数
	preservedTokens := cwm.calculateTotalTokens(preserved)
	targetTokens := window.MaxTokens - preservedTokens

	// 从最新的消息开始保留
	kept := make([]model.Message, 0)
	keptTokens := 0

	for i := len(removable) - 1; i >= 0; i-- {
		if keptTokens+removable[i].Tokens <= targetTokens {
			kept = append([]model.Message{removable[i]}, kept...)
			keptTokens += removable[i].Tokens
		} else {
			break
		}
	}

	// 合并保留的消息
	window.Messages = append(preserved, kept...)
	window.CurrentTokens = preservedTokens + keptTokens
	window.UsagePercent = float64(window.CurrentTokens) / float64(window.MaxTokens) * 100

	return nil
}

// pruneLeastImportant 剪枝最不重要的消息
func (cwm *ContextWindowManager) pruneLeastImportant(window *model.ContextWindow) error {
	// 保留系统消息和标记为重要的消息
	preserved := make([]model.Message, 0)
	removable := make([]model.Message, 0)

	for _, msg := range window.Messages {
		if msg.Role == "system" || msg.Important {
			preserved = append(preserved, msg)
		} else {
			removable = append(removable, msg)
		}
	}

	// 按重要性排序（这里使用时间作为代理，越新越重要）
	sort.Slice(removable, func(i, j int) bool {
		return removable[i].Timestamp.After(removable[j].Timestamp)
	})

	// 计算需要保留的Token数
	preservedTokens := cwm.calculateTotalTokens(preserved)
	targetTokens := window.MaxTokens - preservedTokens

	// 保留最重要的消息
	kept := make([]model.Message, 0)
	keptTokens := 0

	for _, msg := range removable {
		if keptTokens+msg.Tokens <= targetTokens {
			kept = append(kept, msg)
			keptTokens += msg.Tokens
		} else {
			break
		}
	}

	// 按时间排序恢复原始顺序
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].Timestamp.Before(kept[j].Timestamp)
	})

	window.Messages = append(preserved, kept...)
	window.CurrentTokens = preservedTokens + keptTokens
	window.UsagePercent = float64(window.CurrentTokens) / float64(window.MaxTokens) * 100

	return nil
}

// summarize 总结旧消息
func (cwm *ContextWindowManager) summarize(window *model.ContextWindow) error {
	if len(window.Messages) < 5 {
		// 消息太少，使用其他策略
		return cwm.pruneOldest(window)
	}

	// 分离系统消息、重要消息和可总结消息
	systemMsgs := make([]model.Message, 0)
	importantMsgs := make([]model.Message, 0)
	summarizableMsgs := make([]model.Message, 0)
	recentMsgs := make([]model.Message, 0)

	// 保留最近的20%消息不总结
	recentThreshold := len(window.Messages) * 4 / 5

	for i, msg := range window.Messages {
		if msg.Role == "system" {
			systemMsgs = append(systemMsgs, msg)
		} else if msg.Important {
			importantMsgs = append(importantMsgs, msg)
		} else if i < recentThreshold {
			summarizableMsgs = append(summarizableMsgs, msg)
		} else {
			recentMsgs = append(recentMsgs, msg)
		}
	}

	// 总结可总结的消息
	summary := cwm.compressor.SummarizeContext(summarizableMsgs)

	summaryMsg := model.Message{
		Role:      "system",
		Content:   summary,
		Tokens:    len(summary) / 4,
		Timestamp: summarizableMsgs[0].Timestamp,
		Important: false,
	}

	// 重建消息列表
	window.Messages = append(systemMsgs, summaryMsg)
	window.Messages = append(window.Messages, importantMsgs...)
	window.Messages = append(window.Messages, recentMsgs...)

	window.CurrentTokens = cwm.calculateTotalTokens(window.Messages)
	window.UsagePercent = float64(window.CurrentTokens) / float64(window.MaxTokens) * 100

	// 如果还是超了，使用剪枝策略
	if window.CurrentTokens > window.MaxTokens {
		return cwm.pruneOldest(window)
	}

	return nil
}

// AddMessage 添加消息到窗口
func (cwm *ContextWindowManager) AddMessage(window *model.ContextWindow, message model.Message) error {
	// 估算Token数
	if message.Tokens == 0 {
		message.Tokens = len(message.Content) / 4
	}

	// 添加消息
	window.Messages = append(window.Messages, message)

	// 管理窗口
	return cwm.ManageWindow(window)
}

// GetRemainingTokens 获取剩余Token数
func (cwm *ContextWindowManager) GetRemainingTokens(window *model.ContextWindow) int {
	return window.MaxTokens - window.CurrentTokens
}

// CanAddMessage 检查是否可以添加消息
func (cwm *ContextWindowManager) CanAddMessage(window *model.ContextWindow, messageTokens int) bool {
	return cwm.GetRemainingTokens(window) >= messageTokens
}

// OptimizeWindow 优化窗口
func (cwm *ContextWindowManager) OptimizeWindow(window *model.ContextWindow) error {
	// 压缩所有非重要消息
	for i := range window.Messages {
		if !window.Messages[i].Important && window.Messages[i].Role != "system" {
			compressed, _ := cwm.compressor.Compress(window.Messages[i].Content, 2)
			window.Messages[i].Content = compressed
			window.Messages[i].Tokens = len(compressed) / 4
		}
	}

	// 重新计算
	window.CurrentTokens = cwm.calculateTotalTokens(window.Messages)
	window.UsagePercent = float64(window.CurrentTokens) / float64(window.MaxTokens) * 100

	return nil
}

// GetWindowStats 获取窗口统计
func (cwm *ContextWindowManager) GetWindowStats(window *model.ContextWindow) map[string]interface{} {
	stats := map[string]interface{}{
		"max_tokens":     window.MaxTokens,
		"current_tokens": window.CurrentTokens,
		"usage_percent":  window.UsagePercent,
		"message_count":  len(window.Messages),
		"remaining_tokens": cwm.GetRemainingTokens(window),
	}

	// 按角色统计
	roleCount := make(map[string]int)
	roleTokens := make(map[string]int)

	for _, msg := range window.Messages {
		roleCount[msg.Role]++
		roleTokens[msg.Role] += msg.Tokens
	}

	stats["role_count"] = roleCount
	stats["role_tokens"] = roleTokens

	return stats
}

// Clone 克隆窗口
func (cwm *ContextWindowManager) Clone(window *model.ContextWindow) *model.ContextWindow {
	clone := &model.ContextWindow{
		MaxTokens:     window.MaxTokens,
		CurrentTokens: window.CurrentTokens,
		UsagePercent:  window.UsagePercent,
		PruneStrategy: window.PruneStrategy,
		Messages:      make([]model.Message, len(window.Messages)),
	}

	copy(clone.Messages, window.Messages)

	return clone
}

// Validate 验证窗口配置
func (cwm *ContextWindowManager) Validate(window *model.ContextWindow) error {
	if window.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be positive")
	}

	validStrategies := map[string]bool{
		"oldest":          true,
		"least_important": true,
		"summarize":       true,
	}

	if !validStrategies[window.PruneStrategy] {
		return fmt.Errorf("invalid prune strategy: %s", window.PruneStrategy)
	}

	return nil
}
