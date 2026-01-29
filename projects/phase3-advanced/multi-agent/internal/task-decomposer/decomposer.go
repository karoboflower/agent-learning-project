package decomposer

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Decomposer 任务分解器
type Decomposer struct {
	config    *DecomposerConfig
	rules     []*DecompositionRule
	analyzer  *ComplexityAnalyzer
	generator *SubTaskGenerator
}

// DecompositionRule 分解规则
type DecompositionRule struct {
	Name        string
	TaskType    string
	Condition   func(*Task) bool
	Decompose   func(*Task) ([]*SubTask, error)
	Priority    int
}

// NewDecomposer 创建新的分解器
func NewDecomposer(config *DecomposerConfig) *Decomposer {
	if config == nil {
		config = DefaultConfig()
	}

	d := &Decomposer{
		config:    config,
		rules:     make([]*DecompositionRule, 0),
		analyzer:  NewComplexityAnalyzer(),
		generator: NewSubTaskGenerator(),
	}

	// 注册默认规则
	d.registerDefaultRules()

	return d
}

// DefaultConfig 默认配置
func DefaultConfig() *DecomposerConfig {
	return &DecomposerConfig{
		Strategy:           StrategyHybrid,
		MaxDepth:           3,
		MinSubTasks:        2,
		MaxSubTasks:        10,
		ParallelThreshold:  3,
		ComplexityAnalysis: true,
	}
}

// Decompose 分解任务
func (d *Decomposer) Decompose(task *Task) (*DecompositionResult, error) {
	// 验证任务
	if err := d.validateTask(task); err != nil {
		return nil, fmt.Errorf("invalid task: %w", err)
	}

	// 复杂度分析
	if d.config.ComplexityAnalysis {
		complexity := d.analyzer.Analyze(task)
		if complexity == ComplexitySimple {
			// 简单任务不需要分解
			return &DecompositionResult{
				OriginalTask: task,
				SubTasks:     []*SubTask{d.taskToSubTask(task, 0)},
				Strategy:     string(d.config.Strategy),
			}, nil
		}
	}

	// 选择分解策略
	var subTasks []*SubTask
	var err error

	switch d.config.Strategy {
	case StrategyDependency:
		subTasks, err = d.decomposeByDependency(task)
	case StrategyPriority:
		subTasks, err = d.decomposeByPriority(task)
	case StrategyCapability:
		subTasks, err = d.decomposeByCapability(task)
	case StrategyHybrid:
		subTasks, err = d.decomposeHybrid(task)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", d.config.Strategy)
	}

	if err != nil {
		return nil, fmt.Errorf("decomposition failed: %w", err)
	}

	// 构建依赖图
	graph := d.buildDependencyGraph(subTasks)

	// 计算层级
	if err := graph.CalculateLevels(); err != nil {
		return nil, fmt.Errorf("failed to calculate levels: %w", err)
	}

	// 更新子任务层级
	for _, subTask := range subTasks {
		subTask.Level = graph.GetLevel(subTask.ID)
	}

	return &DecompositionResult{
		OriginalTask: task,
		SubTasks:     subTasks,
		Graph:        graph,
		Strategy:     string(d.config.Strategy),
		Metadata: map[string]interface{}{
			"sub_task_count": len(subTasks),
			"max_level":      d.getMaxLevel(subTasks),
		},
	}, nil
}

// decomposeByDependency 基于依赖关系分解
func (d *Decomposer) decomposeByDependency(task *Task) ([]*SubTask, error) {
	subTasks := make([]*SubTask, 0)

	// 分析任务的依赖结构
	if len(task.Dependencies) == 0 {
		// 无依赖，直接转换为子任务
		return []*SubTask{d.taskToSubTask(task, 0)}, nil
	}

	// 有依赖，按依赖关系分解
	// 1. 为每个依赖创建子任务
	depTasks := make([]*SubTask, 0)
	for i, depID := range task.Dependencies {
		subTask := &SubTask{
			ID:           d.generateSubTaskID(task.ID, i),
			ParentID:     task.ID,
			Type:         fmt.Sprintf("dependency_%d", i),
			Description:  fmt.Sprintf("Handle dependency: %s", depID),
			Priority:     task.Priority,
			Dependencies: []string{depID},
			Level:        0,
		}
		depTasks = append(depTasks, subTask)
	}

	// 2. 创建主任务
	mainTask := &SubTask{
		ID:           d.generateSubTaskID(task.ID, len(task.Dependencies)),
		ParentID:     task.ID,
		Type:         task.Type,
		Description:  task.Description,
		Input:        task.Input,
		Priority:     task.Priority,
		Dependencies: d.extractSubTaskIDs(depTasks),
		Capabilities: task.Capabilities,
		Level:        1,
	}

	subTasks = append(subTasks, depTasks...)
	subTasks = append(subTasks, mainTask)

	return subTasks, nil
}

// decomposeByPriority 基于优先级分解
func (d *Decomposer) decomposeByPriority(task *Task) ([]*SubTask, error) {
	subTasks := make([]*SubTask, 0)

	// 按优先级分解为多个阶段
	phases := []struct {
		name     string
		priority int
	}{
		{"preparation", task.Priority - 1},
		{"execution", task.Priority},
		{"verification", task.Priority + 1},
	}

	for i, phase := range phases {
		subTask := &SubTask{
			ID:          d.generateSubTaskID(task.ID, i),
			ParentID:    task.ID,
			Type:        phase.name,
			Description: fmt.Sprintf("%s phase of %s", phase.name, task.Description),
			Priority:    phase.priority,
			Level:       i,
		}

		// 设置依赖（除了第一个阶段）
		if i > 0 {
			subTask.Dependencies = []string{subTasks[i-1].ID}
		}

		subTasks = append(subTasks, subTask)
	}

	return subTasks, nil
}

// decomposeByCapability 基于Agent能力分解
func (d *Decomposer) decomposeByCapability(task *Task) ([]*SubTask, error) {
	if len(task.Capabilities) == 0 {
		// 无特定能力要求，直接转换
		return []*SubTask{d.taskToSubTask(task, 0)}, nil
	}

	subTasks := make([]*SubTask, 0)

	// 为每个能力创建子任务
	for i, capability := range task.Capabilities {
		subTask := &SubTask{
			ID:           d.generateSubTaskID(task.ID, i),
			ParentID:     task.ID,
			Type:         fmt.Sprintf("capability_%s", capability),
			Description:  fmt.Sprintf("Execute %s capability for %s", capability, task.Description),
			Priority:     task.Priority,
			Capabilities: []string{capability},
			Level:        0,
		}

		subTasks = append(subTasks, subTask)
	}

	// 创建聚合任务
	aggregateTask := &SubTask{
		ID:           d.generateSubTaskID(task.ID, len(task.Capabilities)),
		ParentID:     task.ID,
		Type:         "aggregate",
		Description:  fmt.Sprintf("Aggregate results for %s", task.Description),
		Priority:     task.Priority + 1,
		Dependencies: d.extractSubTaskIDs(subTasks),
		Level:        1,
	}

	subTasks = append(subTasks, aggregateTask)

	return subTasks, nil
}

// decomposeHybrid 混合策略分解
func (d *Decomposer) decomposeHybrid(task *Task) ([]*SubTask, error) {
	// 1. 首先按依赖分解
	subTasks := make([]*SubTask, 0)

	// 2. 分析任务类型，应用特定规则
	for _, rule := range d.rules {
		if rule.Condition(task) {
			ruleTasks, err := rule.Decompose(task)
			if err != nil {
				continue
			}
			subTasks = append(subTasks, ruleTasks...)
			break
		}
	}

	// 3. 如果没有匹配的规则，使用默认策略
	if len(subTasks) == 0 {
		// 根据复杂度选择策略
		complexity := task.GetComplexity()
		switch complexity {
		case ComplexityVeryComplex:
			return d.decomposeByDependency(task)
		case ComplexityComplex:
			return d.decomposeByCapability(task)
		default:
			return d.decomposeByPriority(task)
		}
	}

	return subTasks, nil
}

// RegisterRule 注册分解规则
func (d *Decomposer) RegisterRule(rule *DecompositionRule) {
	d.rules = append(d.rules, rule)
}

// registerDefaultRules 注册默认规则
func (d *Decomposer) registerDefaultRules() {
	// 代码审查任务分解规则
	d.RegisterRule(&DecompositionRule{
		Name:     "code_review",
		TaskType: "code_review",
		Condition: func(t *Task) bool {
			return t.Type == "code_review"
		},
		Decompose: func(t *Task) ([]*SubTask, error) {
			return []*SubTask{
				{
					ID:           d.generateSubTaskID(t.ID, 0),
					ParentID:     t.ID,
					Type:         "syntax_check",
					Description:  "Check code syntax",
					Priority:     t.Priority,
					Capabilities: []string{"syntax_analysis"},
					Level:        0,
				},
				{
					ID:           d.generateSubTaskID(t.ID, 1),
					ParentID:     t.ID,
					Type:         "quality_check",
					Description:  "Check code quality",
					Priority:     t.Priority,
					Capabilities: []string{"quality_analysis"},
					Dependencies: []string{d.generateSubTaskID(t.ID, 0)},
					Level:        1,
				},
				{
					ID:           d.generateSubTaskID(t.ID, 2),
					ParentID:     t.ID,
					Type:         "security_check",
					Description:  "Check code security",
					Priority:     t.Priority,
					Capabilities: []string{"security_analysis"},
					Dependencies: []string{d.generateSubTaskID(t.ID, 0)},
					Level:        1,
				},
			}, nil
		},
		Priority: 10,
	})

	// 文档处理任务分解规则
	d.RegisterRule(&DecompositionRule{
		Name:     "document_processing",
		TaskType: "document_processing",
		Condition: func(t *Task) bool {
			return t.Type == "document_processing" || strings.Contains(t.Type, "doc")
		},
		Decompose: func(t *Task) ([]*SubTask, error) {
			return []*SubTask{
				{
					ID:           d.generateSubTaskID(t.ID, 0),
					ParentID:     t.ID,
					Type:         "parse",
					Description:  "Parse document",
					Priority:     t.Priority,
					Capabilities: []string{"document_parsing"},
					Level:        0,
				},
				{
					ID:           d.generateSubTaskID(t.ID, 1),
					ParentID:     t.ID,
					Type:         "analyze",
					Description:  "Analyze document content",
					Priority:     t.Priority,
					Capabilities: []string{"content_analysis"},
					Dependencies: []string{d.generateSubTaskID(t.ID, 0)},
					Level:        1,
				},
				{
					ID:           d.generateSubTaskID(t.ID, 2),
					ParentID:     t.ID,
					Type:         "summarize",
					Description:  "Generate document summary",
					Priority:     t.Priority,
					Capabilities: []string{"summarization"},
					Dependencies: []string{d.generateSubTaskID(t.ID, 1)},
					Level:        2,
				},
			}, nil
		},
		Priority: 10,
	})
}

// buildDependencyGraph 构建依赖图
func (d *Decomposer) buildDependencyGraph(subTasks []*SubTask) *DependencyGraph {
	graph := NewDependencyGraph()

	// 添加所有节点
	for _, subTask := range subTasks {
		graph.AddNode(subTask.ID)
	}

	// 添加边
	for _, subTask := range subTasks {
		for _, depID := range subTask.Dependencies {
			graph.AddEdge(depID, subTask.ID, 1)
		}
	}

	return graph
}

// 辅助方法

func (d *Decomposer) validateTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if task.Type == "" {
		return fmt.Errorf("task type cannot be empty")
	}
	return nil
}

func (d *Decomposer) taskToSubTask(task *Task, level int) *SubTask {
	return &SubTask{
		ID:           task.ID,
		ParentID:     "",
		Type:         task.Type,
		Description:  task.Description,
		Input:        task.Input,
		Priority:     task.Priority,
		Dependencies: task.Dependencies,
		Requirements: task.Requirements,
		Capabilities: task.Capabilities,
		Level:        level,
		Metadata:     task.Metadata,
	}
}

func (d *Decomposer) generateSubTaskID(parentID string, index int) string {
	return fmt.Sprintf("%s-sub-%d-%s", parentID, index, uuid.New().String()[:8])
}

func (d *Decomposer) extractSubTaskIDs(subTasks []*SubTask) []string {
	ids := make([]string, len(subTasks))
	for i, st := range subTasks {
		ids[i] = st.ID
	}
	return ids
}

func (d *Decomposer) getMaxLevel(subTasks []*SubTask) int {
	maxLevel := 0
	for _, st := range subTasks {
		if st.Level > maxLevel {
			maxLevel = st.Level
		}
	}
	return maxLevel
}
