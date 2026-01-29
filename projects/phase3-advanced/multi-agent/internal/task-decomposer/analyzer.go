package decomposer

import (
	"fmt"
)

// ComplexityAnalyzer 复杂度分析器
type ComplexityAnalyzer struct {
	weights ComplexityWeights
}

// ComplexityWeights 复杂度权重配置
type ComplexityWeights struct {
	DependencyWeight  float64 // 依赖数量权重
	CapabilityWeight  float64 // 能力数量权重
	RequirementWeight float64 // 要求数量权重
	TypeWeight        float64 // 任务类型权重
}

// NewComplexityAnalyzer 创建复杂度分析器
func NewComplexityAnalyzer() *ComplexityAnalyzer {
	return &ComplexityAnalyzer{
		weights: ComplexityWeights{
			DependencyWeight:  0.3,
			CapabilityWeight:  0.3,
			RequirementWeight: 0.2,
			TypeWeight:        0.2,
		},
	}
}

// Analyze 分析任务复杂度
func (a *ComplexityAnalyzer) Analyze(task *Task) TaskComplexity {
	score := a.calculateScore(task)

	if score >= 8.0 {
		return ComplexityVeryComplex
	} else if score >= 5.0 {
		return ComplexityComplex
	} else if score >= 2.0 {
		return ComplexityModerate
	}
	return ComplexitySimple
}

// calculateScore 计算复杂度得分
func (a *ComplexityAnalyzer) calculateScore(task *Task) float64 {
	score := 0.0

	// 依赖数量得分
	depScore := float64(len(task.Dependencies))
	if depScore > 5 {
		depScore = 5 + (depScore-5)*0.5 // 超过5个后增长变慢
	}
	score += depScore * a.weights.DependencyWeight * 10

	// 能力数量得分
	capScore := float64(len(task.Capabilities))
	if capScore > 3 {
		capScore = 3 + (capScore-3)*0.5
	}
	score += capScore * a.weights.CapabilityWeight * 10

	// 要求数量得分
	reqScore := float64(len(task.Requirements))
	if reqScore > 5 {
		reqScore = 5 + (reqScore-5)*0.3
	}
	score += reqScore * a.weights.RequirementWeight * 10

	// 任务类型得分
	typeScore := a.getTypeComplexity(task.Type)
	score += typeScore * a.weights.TypeWeight * 10

	return score
}

// getTypeComplexity 获取任务类型的复杂度
func (a *ComplexityAnalyzer) getTypeComplexity(taskType string) float64 {
	complexTypes := map[string]float64{
		"code_review":          3.0,
		"refactoring":          4.0,
		"system_design":        5.0,
		"data_analysis":        3.5,
		"document_processing":  2.5,
		"simple_query":         1.0,
		"calculation":          1.5,
	}

	if complexity, ok := complexTypes[taskType]; ok {
		return complexity
	}

	// 默认中等复杂度
	return 2.0
}

// GetRecommendedStrategy 获取推荐的分解策略
func (a *ComplexityAnalyzer) GetRecommendedStrategy(task *Task) DecompositionStrategy {
	complexity := a.Analyze(task)

	switch complexity {
	case ComplexityVeryComplex:
		// 非常复杂的任务使用混合策略
		return StrategyHybrid
	case ComplexityComplex:
		// 复杂任务根据依赖情况选择
		if len(task.Dependencies) > 0 {
			return StrategyDependency
		}
		return StrategyCapability
	case ComplexityModerate:
		// 中等复杂度使用能力分解
		if len(task.Capabilities) > 0 {
			return StrategyCapability
		}
		return StrategyPriority
	default:
		// 简单任务不需要分解
		return StrategyPriority
	}
}

// EstimateSubTaskCount 估算子任务数量
func (a *ComplexityAnalyzer) EstimateSubTaskCount(task *Task) int {
	complexity := a.Analyze(task)

	baseCount := map[TaskComplexity]int{
		ComplexitySimple:      1,
		ComplexityModerate:    3,
		ComplexityComplex:     5,
		ComplexityVeryComplex: 8,
	}

	count := baseCount[complexity]

	// 根据依赖和能力调整
	if len(task.Dependencies) > 0 {
		count += len(task.Dependencies) / 2
	}

	if len(task.Capabilities) > 1 {
		count += len(task.Capabilities) - 1
	}

	return count
}

// AnalyzeReport 生成分析报告
type AnalysisReport struct {
	TaskID               string             `json:"task_id"`
	Complexity           TaskComplexity     `json:"complexity"`
	Score                float64            `json:"score"`
	RecommendedStrategy  DecompositionStrategy `json:"recommended_strategy"`
	EstimatedSubTasks    int                `json:"estimated_sub_tasks"`
	Factors              map[string]float64 `json:"factors"`
	Recommendations      []string           `json:"recommendations"`
}

// GenerateReport 生成分析报告
func (a *ComplexityAnalyzer) GenerateReport(task *Task) *AnalysisReport {
	complexity := a.Analyze(task)
	score := a.calculateScore(task)

	report := &AnalysisReport{
		TaskID:              task.ID,
		Complexity:          complexity,
		Score:               score,
		RecommendedStrategy: a.GetRecommendedStrategy(task),
		EstimatedSubTasks:   a.EstimateSubTaskCount(task),
		Factors: map[string]float64{
			"dependencies":  float64(len(task.Dependencies)),
			"capabilities":  float64(len(task.Capabilities)),
			"requirements":  float64(len(task.Requirements)),
			"type_complexity": a.getTypeComplexity(task.Type),
		},
		Recommendations: make([]string, 0),
	}

	// 生成建议
	if len(task.Dependencies) > 5 {
		report.Recommendations = append(report.Recommendations,
			"Consider reducing task dependencies or grouping related dependencies")
	}

	if len(task.Capabilities) > 3 {
		report.Recommendations = append(report.Recommendations,
			"Task requires multiple capabilities, consider capability-based decomposition")
	}

	if complexity == ComplexityVeryComplex {
		report.Recommendations = append(report.Recommendations,
			"Very complex task detected, recommend breaking down into multiple stages")
	}

	return report
}

// SubTaskGenerator 子任务生成器
type SubTaskGenerator struct {
	idCounter int
}

// NewSubTaskGenerator 创建子任务生成器
func NewSubTaskGenerator() *SubTaskGenerator {
	return &SubTaskGenerator{
		idCounter: 0,
	}
}

// Generate 生成子任务
func (g *SubTaskGenerator) Generate(parentTask *Task, count int) ([]*SubTask, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive")
	}

	subTasks := make([]*SubTask, count)

	for i := 0; i < count; i++ {
		subTasks[i] = &SubTask{
			ID:           g.generateID(parentTask.ID, i),
			ParentID:     parentTask.ID,
			Type:         fmt.Sprintf("%s_part_%d", parentTask.Type, i+1),
			Description:  fmt.Sprintf("Part %d of %s", i+1, parentTask.Description),
			Priority:     parentTask.Priority,
			Dependencies: make([]string, 0),
			Level:        0,
		}

		// 如果是序列任务，添加依赖
		if i > 0 {
			subTasks[i].Dependencies = []string{subTasks[i-1].ID}
			subTasks[i].Level = i
		}
	}

	return subTasks, nil
}

// GenerateWithPattern 按模式生成子任务
func (g *SubTaskGenerator) GenerateWithPattern(parentTask *Task, pattern string) ([]*SubTask, error) {
	switch pattern {
	case "sequential":
		return g.generateSequential(parentTask)
	case "parallel":
		return g.generateParallel(parentTask)
	case "pipeline":
		return g.generatePipeline(parentTask)
	default:
		return nil, fmt.Errorf("unknown pattern: %s", pattern)
	}
}

// generateSequential 生成顺序执行的子任务
func (g *SubTaskGenerator) generateSequential(parentTask *Task) ([]*SubTask, error) {
	phases := []string{"prepare", "execute", "verify"}
	subTasks := make([]*SubTask, len(phases))

	for i, phase := range phases {
		subTasks[i] = &SubTask{
			ID:          g.generateID(parentTask.ID, i),
			ParentID:    parentTask.ID,
			Type:        phase,
			Description: fmt.Sprintf("%s phase of %s", phase, parentTask.Description),
			Priority:    parentTask.Priority,
			Level:       i,
		}

		if i > 0 {
			subTasks[i].Dependencies = []string{subTasks[i-1].ID}
		}
	}

	return subTasks, nil
}

// generateParallel 生成并行执行的子任务
func (g *SubTaskGenerator) generateParallel(parentTask *Task) ([]*SubTask, error) {
	count := 3 // 默认3个并行任务
	if len(parentTask.Capabilities) > 0 {
		count = len(parentTask.Capabilities)
	}

	subTasks := make([]*SubTask, count)

	for i := 0; i < count; i++ {
		subTasks[i] = &SubTask{
			ID:          g.generateID(parentTask.ID, i),
			ParentID:    parentTask.ID,
			Type:        fmt.Sprintf("parallel_%d", i+1),
			Description: fmt.Sprintf("Parallel task %d of %s", i+1, parentTask.Description),
			Priority:    parentTask.Priority,
			Level:       0, // 所有任务同一层级
		}

		if i < len(parentTask.Capabilities) {
			subTasks[i].Capabilities = []string{parentTask.Capabilities[i]}
		}
	}

	return subTasks, nil
}

// generatePipeline 生成流水线模式的子任务
func (g *SubTaskGenerator) generatePipeline(parentTask *Task) ([]*SubTask, error) {
	stages := []string{"input", "process", "output"}
	subTasks := make([]*SubTask, len(stages))

	for i, stage := range stages {
		subTasks[i] = &SubTask{
			ID:          g.generateID(parentTask.ID, i),
			ParentID:    parentTask.ID,
			Type:        stage,
			Description: fmt.Sprintf("%s stage of %s", stage, parentTask.Description),
			Priority:    parentTask.Priority,
			Level:       i,
		}

		if i > 0 {
			subTasks[i].Dependencies = []string{subTasks[i-1].ID}
		}
	}

	return subTasks, nil
}

func (g *SubTaskGenerator) generateID(parentID string, index int) string {
	g.idCounter++
	return fmt.Sprintf("%s-sub-%d", parentID, g.idCounter)
}
