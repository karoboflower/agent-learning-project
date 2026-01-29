package aggregator

import (
	"fmt"
	"sync"
	"time"
)

// MergeStrategy 合并策略
type MergeStrategy string

const (
	MergeStrategyVoting      MergeStrategy = "VOTING"       // 投票法
	MergeStrategyAveraging   MergeStrategy = "AVERAGING"    // 平均法
	MergeStrategyWeighted    MergeStrategy = "WEIGHTED"     // 加权法
	MergeStrategyConsensus   MergeStrategy = "CONSENSUS"    // 一致性法
	MergeStrategyPriority    MergeStrategy = "PRIORITY"     // 优先级法
	MergeStrategyHighestScore MergeStrategy = "HIGHEST_SCORE" // 最高分法
)

// ConflictResolutionStrategy 冲突解决策略
type ConflictResolutionStrategy string

const (
	ConflictResolutionVoting    ConflictResolutionStrategy = "VOTING"    // 投票
	ConflictResolutionMajority  ConflictResolutionStrategy = "MAJORITY"  // 多数
	ConflictResolutionHighScore ConflictResolutionStrategy = "HIGH_SCORE" // 高分
	ConflictResolutionManual    ConflictResolutionStrategy = "MANUAL"    // 手动
)

// ResultMerger 结果合并器
type ResultMerger struct {
	strategy          MergeStrategy
	conflictStrategy  ConflictResolutionStrategy
	minResults        int     // 最少结果数
	confidenceThreshold float64 // 置信度阈值
}

// NewResultMerger 创建合并器
func NewResultMerger(strategy MergeStrategy, conflictStrategy ConflictResolutionStrategy) *ResultMerger {
	return &ResultMerger{
		strategy:          strategy,
		conflictStrategy:  conflictStrategy,
		minResults:        1,
		confidenceThreshold: 0.5,
	}
}

// SetMinResults 设置最少结果数
func (m *ResultMerger) SetMinResults(min int) {
	m.minResults = min
}

// SetConfidenceThreshold 设置置信度阈值
func (m *ResultMerger) SetConfidenceThreshold(threshold float64) {
	m.confidenceThreshold = threshold
}

// Merge 合并结果
func (m *ResultMerger) Merge(taskID string, results []*TaskResult) (*AggregatedResult, error) {
	// 验证结果数量
	if len(results) < m.minResults {
		return nil, fmt.Errorf("insufficient results: got %d, need at least %d", len(results), m.minResults)
	}

	// 过滤已验证的结果
	validResults := make([]*TaskResult, 0)
	for _, r := range results {
		if r.Status == ResultStatusValidated {
			validResults = append(validResults, r)
		}
	}

	if len(validResults) == 0 {
		return nil, fmt.Errorf("no validated results available")
	}

	// 创建聚合结果
	aggregated := &AggregatedResult{
		TaskID:     taskID,
		Results:    validResults,
		MergedData: make(map[string]interface{}),
		Conflicts:  make([]*Conflict, 0),
		Strategy:   string(m.strategy),
		CreatedAt:  time.Now(),
	}

	// 根据策略合并
	switch m.strategy {
	case MergeStrategyVoting:
		m.mergeByVoting(aggregated)
	case MergeStrategyAveraging:
		m.mergeByAveraging(aggregated)
	case MergeStrategyWeighted:
		m.mergeByWeighted(aggregated)
	case MergeStrategyConsensus:
		m.mergeByConsensus(aggregated)
	case MergeStrategyPriority:
		m.mergeByPriority(aggregated)
	case MergeStrategyHighestScore:
		m.mergeByHighestScore(aggregated)
	default:
		return nil, fmt.Errorf("unknown merge strategy: %s", m.strategy)
	}

	// 检测冲突
	m.detectConflicts(aggregated)

	// 解决冲突
	if err := m.resolveConflicts(aggregated); err != nil {
		return nil, fmt.Errorf("failed to resolve conflicts: %w", err)
	}

	// 计算置信度
	aggregated.Confidence = m.calculateConfidence(aggregated)

	now := time.Now()
	aggregated.CompletedAt = &now

	return aggregated, nil
}

// mergeByVoting 投票法合并
func (m *ResultMerger) mergeByVoting(aggregated *AggregatedResult) {
	// 统计每个字段的值的出现次数
	fieldVotes := make(map[string]map[interface{}]int)

	for _, result := range aggregated.Results {
		for field, value := range result.Data {
			if fieldVotes[field] == nil {
				fieldVotes[field] = make(map[interface{}]int)
			}
			// 使用字符串表示作为key
			key := fmt.Sprintf("%v", value)
			fieldVotes[field][key]++
		}
	}

	// 选择票数最多的值
	for field, votes := range fieldVotes {
		maxVotes := 0
		var selectedValue interface{}

		for value, count := range votes {
			if count > maxVotes {
				maxVotes = count
				selectedValue = value
			}
		}

		aggregated.MergedData[field] = selectedValue
	}
}

// mergeByAveraging 平均法合并
func (m *ResultMerger) mergeByAveraging(aggregated *AggregatedResult) {
	fieldSums := make(map[string]float64)
	fieldCounts := make(map[string]int)

	for _, result := range aggregated.Results {
		for field, value := range result.Data {
			// 只对数值类型求平均
			switch v := value.(type) {
			case int:
				fieldSums[field] += float64(v)
				fieldCounts[field]++
			case int64:
				fieldSums[field] += float64(v)
				fieldCounts[field]++
			case float64:
				fieldSums[field] += v
				fieldCounts[field]++
			case float32:
				fieldSums[field] += float64(v)
				fieldCounts[field]++
			default:
				// 非数值类型使用投票法
				if aggregated.MergedData[field] == nil {
					aggregated.MergedData[field] = value
				}
			}
		}
	}

	// 计算平均值
	for field, sum := range fieldSums {
		aggregated.MergedData[field] = sum / float64(fieldCounts[field])
	}
}

// mergeByWeighted 加权法合并
func (m *ResultMerger) mergeByWeighted(aggregated *AggregatedResult) {
	// 使用结果分数作为权重
	fieldWeightedSums := make(map[string]float64)
	fieldTotalWeights := make(map[string]float64)

	for _, result := range aggregated.Results {
		weight := result.Score
		if weight <= 0 {
			weight = 1.0 // 默认权重
		}

		for field, value := range result.Data {
			switch v := value.(type) {
			case int:
				fieldWeightedSums[field] += float64(v) * weight
				fieldTotalWeights[field] += weight
			case int64:
				fieldWeightedSums[field] += float64(v) * weight
				fieldTotalWeights[field] += weight
			case float64:
				fieldWeightedSums[field] += v * weight
				fieldTotalWeights[field] += weight
			case float32:
				fieldWeightedSums[field] += float64(v) * weight
				fieldTotalWeights[field] += weight
			default:
				// 非数值类型使用最高权重的值
				if aggregated.MergedData[field] == nil {
					aggregated.MergedData[field] = value
				}
			}
		}
	}

	// 计算加权平均
	for field, sum := range fieldWeightedSums {
		if fieldTotalWeights[field] > 0 {
			aggregated.MergedData[field] = sum / fieldTotalWeights[field]
		}
	}
}

// mergeByConsensus 一致性法合并
func (m *ResultMerger) mergeByConsensus(aggregated *AggregatedResult) {
	// 只保留所有结果一致的字段
	if len(aggregated.Results) == 0 {
		return
	}

	// 以第一个结果为基准
	baseResult := aggregated.Results[0]
	for field, baseValue := range baseResult.Data {
		allMatch := true

		// 检查其他结果
		for i := 1; i < len(aggregated.Results); i++ {
			value, exists := aggregated.Results[i].Data[field]
			if !exists || fmt.Sprintf("%v", value) != fmt.Sprintf("%v", baseValue) {
				allMatch = false
				break
			}
		}

		if allMatch {
			aggregated.MergedData[field] = baseValue
		}
	}
}

// mergeByPriority 优先级法合并
func (m *ResultMerger) mergeByPriority(aggregated *AggregatedResult) {
	// 使用分数最高的结果
	if len(aggregated.Results) == 0 {
		return
	}

	highestScoreResult := aggregated.Results[0]
	for _, result := range aggregated.Results {
		if result.Score > highestScoreResult.Score {
			highestScoreResult = result
		}
	}

	// 复制数据
	for field, value := range highestScoreResult.Data {
		aggregated.MergedData[field] = value
	}
}

// mergeByHighestScore 最高分法合并
func (m *ResultMerger) mergeByHighestScore(aggregated *AggregatedResult) {
	m.mergeByPriority(aggregated) // 与优先级法相同
}

// detectConflicts 检测冲突
func (m *ResultMerger) detectConflicts(aggregated *AggregatedResult) {
	// 收集每个字段的所有值
	fieldValues := make(map[string]map[string][]string) // field -> value -> []agentID

	for _, result := range aggregated.Results {
		for field, value := range result.Data {
			if fieldValues[field] == nil {
				fieldValues[field] = make(map[string][]string)
			}
			valueStr := fmt.Sprintf("%v", value)
			fieldValues[field][valueStr] = append(fieldValues[field][valueStr], result.AgentID)
		}
	}

	// 查找有多个不同值的字段
	for field, values := range fieldValues {
		if len(values) > 1 {
			// 存在冲突
			conflict := &Conflict{
				Field:       field,
				Values:      make([]interface{}, 0),
				AgentIDs:    make([]string, 0),
				Description: fmt.Sprintf("Field '%s' has %d different values", field, len(values)),
			}

			for valueStr, agentIDs := range values {
				conflict.Values = append(conflict.Values, valueStr)
				conflict.AgentIDs = append(conflict.AgentIDs, agentIDs...)
			}

			aggregated.Conflicts = append(aggregated.Conflicts, conflict)
		}
	}
}

// resolveConflicts 解决冲突
func (m *ResultMerger) resolveConflicts(aggregated *AggregatedResult) error {
	for _, conflict := range aggregated.Conflicts {
		switch m.conflictStrategy {
		case ConflictResolutionVoting:
			m.resolveByVoting(conflict, aggregated.Results)
		case ConflictResolutionMajority:
			m.resolveByMajority(conflict, aggregated.Results)
		case ConflictResolutionHighScore:
			m.resolveByHighScore(conflict, aggregated.Results)
		case ConflictResolutionManual:
			conflict.Resolution = "Manual resolution required"
			continue
		default:
			return fmt.Errorf("unknown conflict resolution strategy: %s", m.conflictStrategy)
		}

		now := time.Now()
		conflict.ResolvedAt = &now
	}

	return nil
}

// resolveByVoting 投票解决冲突
func (m *ResultMerger) resolveByVoting(conflict *Conflict, results []*TaskResult) {
	votes := make(map[string]int)

	for _, result := range results {
		if value, exists := result.Data[conflict.Field]; exists {
			valueStr := fmt.Sprintf("%v", value)
			votes[valueStr]++
		}
	}

	maxVotes := 0
	var winner string
	for value, count := range votes {
		if count > maxVotes {
			maxVotes = count
			winner = value
		}
	}

	conflict.Resolution = fmt.Sprintf("Resolved by voting: %s (%d votes)", winner, maxVotes)
}

// resolveByMajority 多数解决冲突
func (m *ResultMerger) resolveByMajority(conflict *Conflict, results []*TaskResult) {
	m.resolveByVoting(conflict, results) // 与投票法相同
}

// resolveByHighScore 高分解决冲突
func (m *ResultMerger) resolveByHighScore(conflict *Conflict, results []*TaskResult) {
	var highestScoreResult *TaskResult
	maxScore := -1.0

	for _, result := range results {
		if _, exists := result.Data[conflict.Field]; exists {
			if result.Score > maxScore {
				maxScore = result.Score
				highestScoreResult = result
			}
		}
	}

	if highestScoreResult != nil {
		value := highestScoreResult.Data[conflict.Field]
		conflict.Resolution = fmt.Sprintf("Resolved by highest score: %v (score: %.2f, agent: %s)",
			value, highestScoreResult.Score, highestScoreResult.AgentID)
	}
}

// calculateConfidence 计算置信度
func (m *ResultMerger) calculateConfidence(aggregated *AggregatedResult) float64 {
	if len(aggregated.Results) == 0 {
		return 0.0
	}

	// 因素1: 结果数量 (越多越好)
	resultCountFactor := float64(len(aggregated.Results)) / float64(m.minResults)
	if resultCountFactor > 1.0 {
		resultCountFactor = 1.0
	}

	// 因素2: 平均分数
	totalScore := 0.0
	for _, result := range aggregated.Results {
		totalScore += result.Score
	}
	avgScore := totalScore / float64(len(aggregated.Results))
	scoreFactor := avgScore / 100.0 // 假设最高分为100
	if scoreFactor > 1.0 {
		scoreFactor = 1.0
	}

	// 因素3: 冲突数量 (越少越好)
	conflictFactor := 1.0
	if len(aggregated.MergedData) > 0 {
		conflictRatio := float64(len(aggregated.Conflicts)) / float64(len(aggregated.MergedData))
		conflictFactor = 1.0 - conflictRatio
		if conflictFactor < 0 {
			conflictFactor = 0
		}
	}

	// 综合置信度 (加权平均)
	confidence := (resultCountFactor*0.3 + scoreFactor*0.4 + conflictFactor*0.3)
	return confidence
}

// ResultAggregator 结果聚合器
type ResultAggregator struct {
	store     *ResultStore
	validator *ResultValidator
	merger    *ResultMerger
	mu        sync.RWMutex
}

// NewResultAggregator 创建聚合器
func NewResultAggregator(strategy MergeStrategy, conflictStrategy ConflictResolutionStrategy) *ResultAggregator {
	return &ResultAggregator{
		store:     NewResultStore(),
		validator: NewResultValidator(),
		merger:    NewResultMerger(strategy, conflictStrategy),
	}
}

// AddResult 添加结果
func (a *ResultAggregator) AddResult(result *TaskResult) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 验证结果
	if err := a.validator.ValidateAndMark(result); err != nil {
		// 即使验证失败也存储，以便查看
		a.store.AddResult(result)
		return err
	}

	// 存储结果
	return a.store.AddResult(result)
}

// AggregateTask 聚合任务结果
func (a *ResultAggregator) AggregateTask(taskID string) (*AggregatedResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	results := a.store.GetResultsByTask(taskID)
	return a.merger.Merge(taskID, results)
}

// GetResult 获取结果
func (a *ResultAggregator) GetResult(resultID string) (*TaskResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.GetResult(resultID)
}

// GetResultsByTask 获取任务的所有结果
func (a *ResultAggregator) GetResultsByTask(taskID string) []*TaskResult {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.GetResultsByTask(taskID)
}

// GetStore 获取存储
func (a *ResultAggregator) GetStore() *ResultStore {
	return a.store
}

// GetValidator 获取验证器
func (a *ResultAggregator) GetValidator() *ResultValidator {
	return a.validator
}

// GetMerger 获取合并器
func (a *ResultAggregator) GetMerger() *ResultMerger {
	return a.merger
}
